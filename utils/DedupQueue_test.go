package utils

import (
	"fmt"
	"testing"
	"time"
)

var q = NewDedupQueue[int64](
	10000,                // 队列容量
	4,                    // worker 数量
	500*time.Millisecond, // worker处理间隙
	// 业务处理
	func(userID int64) {

		fmt.Println("process:", userID)

		// 模拟 panic
		if userID == 0 {
			panic("test panic")
		}
	},
)

func TestDedupQueue_Enqueue(t *testing.T) {

	q.Enqueue(1213)
	q.Enqueue(1213) // 不会重复
	q.Enqueue(1214)
	q.Enqueue(0)

	time.Sleep(5 * time.Second)

	fmt.Println("queue len:", q.Len())

	q.Stop()
}

func TestDedupQueue_Dequeue(t *testing.T) {
	for i := 0; i < 10000; i++ {
		q.Enqueue(1)
		q.Enqueue(2)
		q.Enqueue(3)
		q.Enqueue(4)
		q.Enqueue(4)
		q.Enqueue(4)
		q.Enqueue(4)
		time.Sleep(time.Millisecond * 100)
	}

	time.Sleep(time.Minute)
}
