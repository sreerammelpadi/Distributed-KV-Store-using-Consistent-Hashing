package main

import (
	"fmt"
	"sync"
)

func factorial(n int) int {
	if n <= 2 {
		return n
	}
	return n * factorial(n-1)
}

func FactorialWorker(n int, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	f := factorial(n)
	t := "%v! = %v"
	ch <- fmt.Sprintf(t, n, f)
}

func main() {
	var wg sync.WaitGroup
	nums := []int{3, 5, 7}
	ch := make(chan string)
	for _, n := range nums {
		wg.Add(1)
		go FactorialWorker(n, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for msg := range ch {
		fmt.Println(msg)
	}

}
