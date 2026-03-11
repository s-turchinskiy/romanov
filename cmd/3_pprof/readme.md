GOGC=off go test -bench . -benchmem -cpuprofile=cpu.pprof -memprofile=mem.pprof && go tool pprof -http=":9090" mem.pprof

GOGC=off go test -bench . -benchmem -cpuprofile=cpu.pprof -memprofile=mem.pprof
Для чистоты добавляем GOGC=off, чтобы в графике не было видно освобождение памяти и он выглядел более чисто

Оптимизация оперативной памяти
go tool pprof -http=":9090" mem.pprof
откроется http://localhost:9090/ui/
<img width="1034" height="491" alt="Снимок экрана от 2026-03-11 01-31-41" src="https://github.com/user-attachments/assets/d4e60194-d53a-49d5-9fb3-3122bdb72e42" />
выделяем синим функцию для оптимизации и нажимаем Refine -> Show from. Наша функция потребила 1300+ мб оперативки

оптимизация 1. regexp.MatchString\
<img width="1227" height="881" alt="Снимок экрана от 2026-03-11 01-43-04" src="https://github.com/user-attachments/assets/00a52998-6525-46bc-b896-fa7f0af19e9b" />
скриншот 2. больше всего потребляет regexp.MatchString, которая вызывает regexp.Compile. https://habr.com/ru/companies/badoo/articles/301990/ из этой статьи известно, что надо заранее компилировать, а не каждый раз
заменим в коде 2 вызова regexp.MatchString("Android", browser) и regexp.MatchString("MSIE", browser) на
var patternAndroid = regexp.MustCompile("Android")
var patternMSIE = regexp.MustCompile("MSIE")
patternAndroid.MatchString(browser)
patternMSIE.MatchString(browser)
Количество выделенной памяти почему-то даже стало больше после 1 оптимизации, 1548мб, но regexp.MatchString ушел из анализа

оптимизация 2. io.ReadAll\
<img width="1227" height="881" alt="Снимок экрана от 2026-03-11 02-05-14" src="https://github.com/user-attachments/assets/ed23ec60-1537-4ccc-8fee-9b613795827b" />
скриншот 4. теперь самое жирное io.ReadAll, 793 мб. чтение всего файла происходит полностью одномоментно, надо читать по блокам
заменил
fileContents, err := ioutil.ReadAll(file)
lines := strings.Split(string(fileContents), "\n")
for _, line := range lines {
на
scanner := bufio.NewScanner(file)
for scanner.Scan() {
line := scanner.Text()
Количество выделенной памяти стало 726мб

оптимизация 3. easyjson\
<img width="1634" height="834" alt="Снимок экрана от 2026-03-11 22-10-15" src="https://github.com/user-attachments/assets/fb21b5f1-e914-4c25-bfd4-2f57e5abd000" />
скриншот 5. теперь самое жирное json.Unmarshal заменил на easyjson
<img width="1634" height="834" alt="Снимок экрана от 2026-03-11 22-29-56" src="https://github.com/user-attachments/assets/19d5b382-33f1-4fc7-adb0-5f521a5ca92f" />
скриншот 6. память увеличилась до 1280 мб после добавления сериализации для type User map[string]any
370	   3844916 ns/op	 2798023 B/op	   37738 allocs/op

в коде из джейсона берется 3 поля browsers,email,name
type User заменил на
type User struct {
Browsers []string `json:"browsers"`
Email    string   `json:"email"`
Name     string   `json:"name"`
}
Количество выделенной памяти стало 1050мб, но все остальные цифры уменьшились
439	   2449544 ns/op	 2058975 B/op	    9821 allocs/op

<img width="1387" height="826" alt="Снимок экрана от 2026-03-11 22-59-44" src="https://github.com/user-attachments/assets/e4ddd80c-2bc3-4493-a116-d39f54976448" />
Скриншот 7. Открыл View - Source
больше всего flat выделяется на строчках
line := scanner.Text()
err = easyjson.Unmarshal([]byte(line), &user)
не заметил этого раньше, по сути сначала из байт преобразуется в строку, а потом обратно
заменил на err = easyjson.Unmarshal(scanner.Bytes(), &user)
<img width="1605" height="622" alt="Снимок экрана от 2026-03-11 23-04-40" src="https://github.com/user-attachments/assets/44d99ebf-da64-42d1-8587-2cb42cfde25a" />
скриншот 8. количество выделяемой памяти значительно снизилось до 595 мб
613	   2110532 ns/op	  871508 B/op	    7821 allocs/op

оптимизация 4\
<img width="743" height="437" alt="Снимок экрана от 2026-03-11 23-09-06" src="https://github.com/user-attachments/assets/1da76df3-7073-4fa9-8f1c-1b898c28c3da" />
скриншот 9. видно что много памяти выделяется на users = append(users, user)
при этом слайс изначально определяется нулевого размера users := make([]data.User, 0)
надо определить его размер равным количеству строк
добавил функцию lineCounter
count, err := lineCounter(file)
_, err = file.Seek(0, io.SeekStart)
users := make([]data.User, 0, count)
637	   2092235 ns/op	  920841 B/op	    7813 allocs/op
улучшилось только количество операций, все остальное или осталось неизменным или чуть ухудшилось

оптимизация 5\
View - Source больше всего flat выделяется теперь на строчке
foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
foundUsers - это строка. правильно суммировать строки не через +=, а через strings.Builder
заменил на него
655	   1946396 ns/op	  750327 B/op	    7741 allocs/op
<img width="1826" height="581" alt="Снимок экрана от 2026-03-11 23-40-20" src="https://github.com/user-attachments/assets/34121519-0cdb-4758-bb6c-3267cc7840fc" />\
скриншот 10. память стала равна 529мб и значительно упало выделение памяти на операцию с 920841 до 750327

оптимизация 6\
тек значение памяти 750327 B/op, нужно 559910 B/op
без профайлера видно, что создается слайс users и на него выделяется память. правильно сделать через горутины, получили из файла юзера и передали его в горутину-обработчик.
но в задании явно сказано, что нельзя использовать горутины, так что просто объединил циклы, чтобы сразу после получения юзера он обрабатывался, а не помещался в слайс
ну и даже в source сейчас строчки с самыми большими значениеми
users := make([]data.User, 0, count)
users = append(users, user)
663	   1902897 ns/op	  578345 B/op	    7738 allocs/op
выделение памяти упало до нужного значения

Результат\
<img width="1826" height="581" alt="Снимок экрана от 2026-03-11 23-55-26" src="https://github.com/user-attachments/assets/758951e7-a2da-49cd-bff9-4760ea4d20b0" />\
Должно быть: BenchmarkSolution-8 500 2782432 ns/op 559910 B/op 10422 allocs/op\
Результат: BenchmarkFast-16    649	1883548 ns/op 573611 B/op 7656 allocs/op\
500 -> 649\
2782432 ns/op -> 1883548 ns/op\
559910 B/op -> 573611 B/op\
10422 allocs/op -> 7656 allocs/op


//////////////////////////////////////////////////////////////////////////////////////////////////////////

Есть функиця, которая что-то там ищет по файлу. Но делает она это не очень быстро. Надо её оптимизировать.

Задание на работу с профайлером pprof.

Цель задания - научиться работать с pprof, находить горячие места в коде, уметь строить профиль потребления cpu и памяти, оптимизировать код с учетом этой информации. Написание самого быстрого решения не является целью задания.

Для генерации графа вам понадобится graphviz. Для пользователей windows не забудьте добавить его в PATH чтобы была доступна команда dot.

Рекомендую внимательно прочитать доп. материалы на русском - там ещё много примеров оптимизации и объяснений как работать с профайлером. Фактически там есть вся информация для выполнения этого задания.

* https://habrahabr.ru/company/badoo/blog/301990/
* https://habrahabr.ru/company/badoo/blog/324682/
* https://habrahabr.ru/company/badoo/blog/332636/
* https://habrahabr.ru/company/mailru/blog/331784/

Есть с десяток мест где можно оптимизировать.
Вам надо писать отчет, где вы заоптимайзили и что. Со скриншотами и объяснением что делали. Чтобы именно научиться в pprof находить проблемы, а не прикинуть мозгами и решить что вот тут медленно.

Для выполнения задания необходимо чтобы один из параметров ( ns/op, B/op, allocs/op ) был быстрее чем в *BenchmarkSolution* ( fast < solution ) и ещё один лучше *BenchmarkSolution* + 20% ( fast < solution * 1.2), например ( fast allocs/op < 10422*1.2=12506 ).

По памяти ( B/op ) и количеству аллокаций ( allocs/op ) можно ориентироваться ровно на результаты *BenchmarkSolution* ниже, по времени ( ns/op ) - нет, зависит от системы.

Параллелить (использовать горутины) или sync.Pool в это задании не нужно.

Результат в fast.go в функцию FastSearch (изначально там то же самое что в SlowSearch).

Пример результатов с которыми будет сравниваться:
```
$ go test -bench . -benchmem

goos: windows

goarch: amd64

BenchmarkSlow-8 10 142703250 ns/op 336887900 B/op 284175 allocs/op

BenchmarkSolution-8 500 2782432 ns/op 559910 B/op 10422 allocs/op

PASS

ok coursera/hw3 3.897s
```

Запуск:
* `go test -v` - чтобы проверить что ничего не сломалось
* `go test -bench . -benchmem` - для просмотра производительности
* `go tool pprof -http=:8083 /path/ho/bin /path/to/out` - веб-интерфейс для pprof, пользуйтесь им для поиска горячих мест. Не забывайте, что у вас 2 режиме - cpu и mem, там разные out-файлы.

Советы:
* Смотрите где мы аллоцируем память
* Смотрите где мы накапливаем весь результат, хотя нам все значения одновременно не нужны
* Смотрите где происходят преобразования типов, которые можно избежать
* Смотрите не только на графе, но и в pprof в текстовом виде (list FastSearch) - там прямо по исходнику можно увидеть где что
* Задание предполагает использование easyjson. На сервере эта библиотека есть, подключать можно. Но сгенерированный через easyjson код вам надо поместить в файл с вашей функцией
* Можно сделать без easyjson

Примечание:
* easyjson основан на рефлекции и не может работать с пакетом main. Для генерации кода вам необходимо вынести вашу структуру в отдельный пакет, сгенерить там код, потом забрать его в main
