package leakybucket

import (
	"fmt"
	"time"
)

var bucket = New(100, 1000) // 全局漏桶

func handleRequest(i int) {
	if !bucket.Allow() {
		fmt.Println("Request", i, "limited")
		return
	}

	fmt.Println("Handling request", i)
	// 处理请求...
}

func main() {
	for i := 1; i <= 20; i++ {
		handleRequest(i)
		time.Sleep(50 * time.Millisecond) // 增加间隔
	}

	// 等待漏桶重新填充
	time.Sleep(1 * time.Second)

	for i := 21; i <= 30; i++ {
		handleRequest(i)
		time.Sleep(10 * time.Millisecond) // 增加间隔
	}
}
