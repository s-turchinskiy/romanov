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
	numWorkers := len(jobs)

	wg := sync.WaitGroup{}
	wg.Add(numWorkers)

	dataChs := make([]chan any, numWorkers+1)

	for i := 0; i < numWorkers+1; i++ {
		dataChs[i] = make(chan any, MaxInputDataLen)
	}

	for i, job := range jobs {
		go runJob(job, dataChs[i], dataChs[i+1], &wg, i == 0)
	}

	wg.Wait()
}

func runJob(job job, in, out chan any, wg *sync.WaitGroup, firstJob bool) {
	defer wg.Done()
	job(in, out)
	// а как получить имена джоб?
	// log.Println("job end", time.Now())
	close(out)
	if firstJob {
		close(in)
	}
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
		}()

		md5 := DataSignerMd5(data)
		go func(md5 string) {
			crc32Second <- DataSignerCrc32(md5)
		}(md5)

		wg.Add(1)
		go func(crc32First, crc32Second chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			result := <-crc32First + "~" + <-crc32Second
			out <- result
			close(crc32First)
			close(crc32Second)
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

		for i := 0; i < hashNumbers; i++ {
			go func(i int) {
				defer wg.Done()
				results[i] = DataSignerCrc32(strconv.Itoa(i) + data)
			}(i)
		}

		wgForMultiHash.Add(1)
		go func(out chan any, wgForMultiHash *sync.WaitGroup) {
			defer wgForMultiHash.Done()
			wg.Wait()
			var res string
			for _, value := range results {
				res += value
			}
			out <- res
		}(out, &wgForMultiHash)
	}
	wgForMultiHash.Wait()
}

func CombineResults(in, out chan any) {
	var results []string
	for value := range in {
		results = append(results, value.(string))
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
		job(func(in, out chan any) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan any) {
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
