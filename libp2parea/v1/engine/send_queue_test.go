package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestSendQueue(t *testing.T) {
	example1()
}

const loopCount = 10

func example1() {
	c, cance := context.WithCancel(context.Background())
	sq := NewSendQueue(SendQueueCacheNum, c, cance, "")
	go send(sq)
	write(sq)
}

func send(sq *SendQueue) {
	for i := 0; i < loopCount; i++ {
		temp := i
		go func(j int) {
			fmt.Println("开始发送:", j)
			err := sq.AddAndWaitTimeout(nil, time.Second*time.Duration(j+1))
			if err != nil {
				fmt.Println("发送失败:", j, err.Error())
			} else {
				// fmt.Println("发送成功", j)
			}
		}(temp)
	}
}

func write(sq *SendQueue) {
	c := sq.GetQueueChan()
	var sp *SendPacket
	var isClose bool
	var err error
	for i := 0; i < loopCount; i++ {
		err = nil
		sp, isClose = <-c
		if !isClose {
			fmt.Println("关闭通道")
			return
		}
		log.Println("接收顺序:", sp.ID)
		// time.Sleep(time.Second / 2)
		if sp.ID == 5 {
			err = errors.New("模拟关闭连接EOF")
		} else if sp.ID == 9 {
			sq.Destroy()
			continue
		}
		sq.SetResult(sp.ID, err)
	}
}
