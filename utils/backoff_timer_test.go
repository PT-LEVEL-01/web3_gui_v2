package utils

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewBackoffTimerChan(*testing.T) {
	// backoffTimerExample()
}

func backoffTimerExample() {
	fmt.Println("start test TestNewBackoffTimerChan")
	btc := NewBackoffTimerChan(time.Second, time.Second*2, time.Second*4)
	c, cancel := context.WithCancel(context.Background())
	go TimeOut(cancel)
	go LoopWait(btc, c)
	fmt.Println("end")
}

func LoopWait(btc *BackoffTimerChan, c context.Context) {
	for {
		n := btc.Wait(c)
		fmt.Println("wait", n)
		if n == 0 {
			break
		}
	}
}

func TimeOut(cancel context.CancelFunc) {
	time.Sleep(time.Second * 5)
	cancel()
}
