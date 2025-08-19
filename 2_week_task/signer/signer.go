package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"sync"
)

var debugMode = flag.Bool("debug", false, "включить режим отладки")

func ExecutePipeline(jobs ...job) {
	fmt.Println("start")
	//отладочная информация отображается с флагом -debug
	flag.Parse()
	if !*debugMode {
		log.SetOutput(io.Discard)
	}

	if len(jobs) == 0 {
		return
	}

	wg := &sync.WaitGroup{}
	var inb chan interface{}
	//запускаем джобы в горутинах, прокидывая каналы
	for _, j := range jobs {
		out := make(chan interface{})
		in := inb
		currentJob := j
		wg.Add(1)
		go func() {
			defer func() {
				close(out)
				wg.Done()
			}()
			currentJob(in, out)
		}()
		inb = out
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	m := &sync.Mutex{}
	defer log.Println("SingleHash завершился")
	log.Println("SingleHash запустился")
	wg := &sync.WaitGroup{}

	for v := range in {
		wg.Add(1)
		go func(v interface{}) {
			defer wg.Done()

			outCrc32 := make(chan string)
			outCrc32FromMd5 := make(chan string)

			go func() {
				outCrc32 <- DataSignerCrc32(strconv.Itoa(v.(int)))
			}()

			go func() {
				//блокируем для избежания перегрева (подробнее в readme.md)
				m.Lock()
				md5 := DataSignerMd5(strconv.Itoa(v.(int)))
				m.Unlock()
				outCrc32FromMd5 <- DataSignerCrc32(md5)
			}()

			crc32 := <-outCrc32
			crc32FromMd5 := <-outCrc32FromMd5
			result := crc32 + "~" + crc32FromMd5
			out <- result

		}(v)
	}

	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	defer log.Println("---MultiHash завершился")
	log.Println("---MultiHash запущен")

	wg := &sync.WaitGroup{}
	m := &sync.Mutex{}

	for v := range in {
		wg.Add(1)
		go func(v interface{}) {
			defer wg.Done()
			wg1 := &sync.WaitGroup{}
			log.Println("---MultiHash получил:", v)
			arr := make([]string, 6)

			wg1.Add(6)
			for th := 0; th <= 5; th++ {
				go func(th int, v interface{}) {
					defer wg1.Done()

					prS := strconv.Itoa(th) + v.(string)
					s := DataSignerCrc32(prS)

					//записываем в слайс под соответствующим индексом
					//для сохранения порядка
					m.Lock()
					arr[th] = s
					m.Unlock()

				}(th, v)
			}

			wg1.Wait()
			var result string
			//соединяем результаты в единую строку
			for _, v := range arr {
				result += v
			}
			log.Println("---MultiHash отправил:", result)
			out <- result

		}(v)
	}

	wg.Wait()
}

// сортируем и объединяем результаты
func CombineResults(in, out chan interface{}) {
	defer log.Println("---------CombineResults завершился")

	var strings []string
	for v := range in {
		log.Println("---------CombineResults получил:", v)
		strings = append(strings, v.(string))
	}

	log.Println("---------CombineResults слайс:", strings)

	sort.Strings(strings)
	var result string
	for i, v := range strings {
		if i == len(strings)-1 {
			result += v
		} else {
			result = result + v + "_"
		}
	}

	log.Println("---------CombineResults отправил:", result)
	out <- result
}
