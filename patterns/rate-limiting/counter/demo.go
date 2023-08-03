package counter

import (
	"fmt"
)

func main() {
	limit := New(2) // 2请求每秒

	for i := 0; i < 10; i++ {
		if limit.Allow() {
			fmt.Println(i)
		} else {
			fmt.Println("Limit exceeded")
		}
	}
}
