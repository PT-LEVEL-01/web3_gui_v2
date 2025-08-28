package utils

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewChanDynamic(*testing.T) {
	// chandynamicExample()
}

func chandynamicExample() {

	total := 1000000
	chanDy := NewChanDynamic(10000)

	group := new(sync.WaitGroup)
	group.Add(2)
	go add(chanDy, total, group)
	go get(chanDy, total, group)
	group.Wait()
}

var c = make(chan int, 100000)

func add(cd *ChanDynamic, total int, group *sync.WaitGroup) {
	start := time.Now()
	for i := 0; i < total; i++ {
		// fmt.Println("add", i)
		cd.Add(i)
		// c <- i
	}
	fmt.Println("ChanDynamic add spend:", time.Now().Sub(start))
	group.Done()
}

func get(cd *ChanDynamic, total int, group *sync.WaitGroup) {
	ctx, _ := context.WithCancel(context.Background())
	// time.Sleep(time.Second * 2)
	start := time.Now()
	for i := 0; i < total; i++ {
		// fmt.Println("get", i)
		cd.Get(ctx)
		// <-c
		// fmt.Println("get number:", cd.Get())
	}
	fmt.Println("ChanDynamic get spend:", time.Now().Sub(start))
	group.Done()
}
