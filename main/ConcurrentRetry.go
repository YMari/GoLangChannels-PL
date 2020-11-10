package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Result struct {
	index int
	result string
	retry int
}
type Job = func() (string, error)

func main() {
	var tasks []func() (string, error)
	for i:=0; i < 20; i++ {
		tasks = append(tasks, genericTask)
	}

	result1 := ConcurrentRetry(tasks, 5, 3)
	result2 := ConcurrentRetry(tasks, 10, 3)

	i := 0
	j := 0

	for r := range result1 {
		fmt.Println("Thread 1 Result:", r)
		if r.result == "Passed" {
			i++
		}
	}

	for r := range result2 {
		fmt.Println("Thread 2 Result:", r)
		if r.result == "Passed" {
			j++
		}
	}

	fmt.Println("\nThread Processes Finished.")
	fmt.Println("\nFinal Results:")
	fmt.Println("Thread 1:", i, "Passed.")
	fmt.Println("Thread 2:", j, "Passed.")
}

func genericTask() (string, error) {
	x := rand.Intn(100)
	runDuration := time.Duration(rand.Intn(1e3)) * time.Millisecond
	time.Sleep(runDuration)
	if x <= 60 { // 60% failure chance
		return "Failed", errors.New("error")
	}

	return "Passed", nil
}

func worker(ID int, threads <-chan Job, results chan <-Result, Retry int, wg *sync.WaitGroup) {

	// traverse threat channel
	fmt.Println("Task started... (Task ID:", ID,")")
	for job := range threads {
		var r Result
		for i:=0; i < Retry; i++ {
			str, e := job()

			r = Result { // save result
				index: ID,
				result: str,
				retry: i+1,
			}

			if e == nil { // break if no errors
				break
			}
		}

		results <- r // send result to channel

	}

	wg.Done()

}

func ConcurrentRetry(tasks []func() (string, error), concurrent int, retry int) <-chan Result {
	threads := make(chan Job, len(tasks))
	results := make(chan Result, len(tasks))

	var wg sync.WaitGroup

	fmt.Println("Starting", concurrent, "worker tasks.")

	go func() {
		for _, job := range tasks { // fill thread channel with jobs
			threads <- job
		}

		for i:=1; i <= concurrent; i++ { // initialize workers
			wg.Add(1)
			go worker(i, threads, results, retry, &wg)
		}

		for { // close threads channel once finished
			if len(threads) == 0 {
				close(threads)
				break
			}
		}

		wg.Wait()
		close(results) // close results channel once finished
	}()

	return results
}