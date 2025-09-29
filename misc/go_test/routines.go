package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func worker(id int, ch chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Duration(rand.Intn(6)) * time.Second)
	fmt.Printf("Hi from %v \n", id)
	time.Sleep(1 * time.Second)
	ch <- id
}

func main() {
	var wg sync.WaitGroup
	ch := make(chan int)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		// time.Sleep(1 * time.Second)
		go worker(i, ch, &wg)
	}

	go func() {
		// IDIOM: a go routine (like python lambda) specifically to wait till all workers exit
		// and close the channel, which will terminate the loop in main function
		wg.Wait()
		close(ch)
	}()

	for msg := range ch {
		fmt.Printf("Closed %v\n", msg)
	}

}
