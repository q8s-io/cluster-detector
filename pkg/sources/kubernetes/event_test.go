package kubernetes

import (
	"log"
	"testing"
	"time"
)

func TestInspection(t *testing.T) {
	node := node{
		localBuffer: make(chan int, 100),
		stopChannel: make(chan struct{}, 0),
	}
	go node.watch()
	cnt := 0
Loop:
	for {
		if cnt > node.count {
			break Loop
		}
		select {
		case n := <-node.localBuffer:
			log.Println(n)
			cnt++
		case <-time.After(time.Second):
			log.Println("time out !!!")
			return
		}
	}
	return
}

func (this *node) watch() {
	this.count = 10
	for i := 0; i < 10; i++ {
		this.localBuffer <- i
		time.Sleep(time.Second * 5)
	}
}

type node struct {
	localBuffer chan int
	stopChannel chan struct{}
	count       int
}
