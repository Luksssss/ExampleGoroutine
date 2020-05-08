package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Многопоточный
// https://www.coursera.org/learn/golang-webservices-1/programming/uEXU0/konvieiier-funktsii

func ExecutePipeline(jobs ...job) {

	var in, out chan interface{}
	wg := &sync.WaitGroup{}

	for _, j := range jobs {
		out = make(chan interface{})
		wg.Add(1)
		go worker(wg, j, in, out)
		in = out
	}
	wg.Wait()

}

func worker(wg *sync.WaitGroup, j job, in, out chan interface{}) {
	defer wg.Done()
	j(in, out)

	close(out)
}

func SingleHash(in, out chan interface{}) {

	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	// читаем из канала пока он не закрыт
	for el := range in {
		wg.Add(1)
		go func(el int) {
			defer wg.Done()
			elStr := strconv.Itoa(el)

			mu.Lock()
			// должен быть доступ только у одного потока
			md5 := DataSignerMd5(elStr)
			mu.Unlock()

			chCrc := make(chan string)
			go func() {
				chCrc <- DataSignerCrc32(elStr)
			}()

			chCrcMd5 := make(chan string)
			go func() {
				chCrcMd5 <- DataSignerCrc32(md5)
			}()

			tmp := <-chCrc + "~" + <-chCrcMd5
			fmt.Println(el, "SH", tmp)
			out <- tmp
		}(el.(int))
	}

	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	th := [6]string{"0", "1", "2", "3", "4", "5"}

	// WaitGroup ожидает пока выполнится вся группа
	wgMh := &sync.WaitGroup{}

	for sh := range in {
		wgMh.Add(1)
		go func(sh interface{}) {
			defer wgMh.Done()
			hashMap := map[string]string{}
			listRes := make([]string, 0, len(th))

			mu := &sync.Mutex{}
			wg := &sync.WaitGroup{}

			for i := range th {
				fmt.Println(i, "MH", sh)
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					el := DataSignerCrc32(th[i] + sh.(string))

					mu.Lock()
					// в map можно писать только в синхронном режиме (!)
					hashMap[th[i]] = el
					mu.Unlock()

				}(i)
			}
			wg.Wait()

			// из map нужно вытащить в порядке сортировки
			for _, el := range th {
				listRes = append(listRes, hashMap[el])
			}

			resStr := strings.Join(listRes, "")
			//fmt.Println(*data, resStr)
			out <- resStr
		}(sh)

	}
	wgMh.Wait()
}

func CombineResults(in, out chan interface{}) {

	res := make([]string, 0, len(in))

	for x := range in {
		res = append(res, x.(string))
		fmt.Println("CR", res)
	}

	sort.Strings(res)

	//fmt.Println(res)

	out <- strings.Join(res, "_")

}

// func main() {
// 	inputData := []int{0, 1, 1, 2, 3, 5, 8}
// 	// inputData := []int{0, 1}
// 	hashSignJobs := []job{
// 		job(func(in, out chan interface{}) {
// 			for _, fibNum := range inputData {
// 				out <- fibNum
// 			}
// 		}),
// 		job(SingleHash),
// 		job(MultiHash),
// 		job(CombineResults),
// 		job(func(in, out chan interface{}) {
// 			dataRaw := <-in
// 			data, ok := dataRaw.(string)
// 			if !ok {
// 				fmt.Println("cant convert result data to string")
// 			}
// 			fmt.Println("result:", data)
// 		}),
// 	}

// 	start := time.Now()
// 	ExecutePipeline(hashSignJobs...)
// 	end := time.Since(start)
// 	expectedTime := time.Second * 3

// 	if end > expectedTime {
// 		fmt.Printf("execition too long\nGot: %s\nExpected: <%s\n", end, expectedTime)
// 	}
// }
