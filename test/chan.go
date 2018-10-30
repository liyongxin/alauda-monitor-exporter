package test

import (
	"fmt"
	"time"
)

var ch chan int = make(chan int, 1)

func main() {
	go aaa()

	select {
	case <-ch: //拿到锁
		fmt.Println("call")
	case <-time.After(5 * time.Second): //超时5s
		fmt.Println("5 sec call")
	}
}

func aaa() {
	time.Sleep(time.Second * 6)
	ch <- 1
}