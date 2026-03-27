package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ExecutePipeline(jobs ...job) {
	wg := sync.WaitGroup{}
	wg.Add(len(jobs))

	inCh := make(chan any, MaxInputDataLen)
	for _, j := range jobs {
		outCh := make(chan any, MaxInputDataLen)

		go func(inCh chan any) {
			defer wg.Done()
			j(inCh, outCh)
			// log.Println("job end", time.Now()) // как получить имена джоб?
			close(outCh)
		}(inCh)

		inCh = outCh
	}

	wg.Wait()
}

func SingleHash(in, out chan any) {
	wg := sync.WaitGroup{}

	for untypedValue := range in {
		data, err := stringFromUntypedValue(untypedValue)
		if err != nil {
			log.Fatal(err)
		}

		crc32First := make(chan string)
		crc32Second := make(chan string)

		go func() {
			crc32First <- DataSignerCrc32(data)
			close(crc32First)
		}()

		md5 := DataSignerMd5(data)
		go func() {
			crc32Second <- DataSignerCrc32(md5)
			close(crc32Second)
		}()

		wg.Add(1)
		go func(crc32First, crc32Second chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			result := <-crc32First + "~" + <-crc32Second
			out <- result
		}(crc32First, crc32Second, &wg)
	}

	wg.Wait()
}

func MultiHash(in, out chan any) {
	hashNumbers := 6
	wgForMultiHash := sync.WaitGroup{}

	for untypedValue := range in {
		data, err := stringFromUntypedValue(untypedValue)
		if err != nil {
			log.Fatal(err)
		}

		results := make([]string, hashNumbers)

		wg := sync.WaitGroup{}
		wg.Add(hashNumbers)

		for i := range hashNumbers {
			go func(i int) {
				defer wg.Done()
				results[i] = DataSignerCrc32(strconv.Itoa(i) + data)
			}(i)
		}

		wgForMultiHash.Add(1)
		go func(out chan any, wgForMultiHash *sync.WaitGroup) {
			defer wgForMultiHash.Done()
			wg.Wait()

			var sb strings.Builder
			for _, value := range results {
				sb.WriteString(value)
			}
			out <- sb.String()
		}(out, &wgForMultiHash)
	}
	wgForMultiHash.Wait()
}

func CombineResults(in, out chan any) {
	var results []string
	for untypedValue := range in {
		data, err := stringFromUntypedValue(untypedValue)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, data)
	}

	sort.Strings(results)

	res := strings.Join(results, "_")
	out <- res
}

func stringFromUntypedValue(untypedValue any) (string, error) {
	var data string

	switch valType := untypedValue.(type) {
	case string:
		data = untypedValue.(string)
	case int:
		data = strconv.Itoa(untypedValue.(int))
	default:
		err := fmt.Errorf("unklown type input channel %v", valType)
		return "", err
	}

	return data, nil
}

func main() {
	inputData := []int{0, 1, 1, 2, 3, 5, 8}

	hashSignJobs := []job{
		job(func(_, out chan any) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, _ chan any) {
			dataRaw := <-in
			_, ok := dataRaw.(string)
			if !ok {
				log.Fatal("cant convert result data to string")
			}
			// testResult = data
		}),
	}

	start := time.Now()

	ExecutePipeline(hashSignJobs...)

	end := time.Since(start)
	fmt.Println(end)
}
