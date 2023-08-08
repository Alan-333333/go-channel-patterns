package dbpool

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

func openNewConnection() (*DBConn, error) {
	db, err := sql.Open("mysql", "reader:123456@tcp(127.0.0.1:3306)/mysql")
	if err != nil {
		return nil, err
	}

	c := &DBConn{
		DB:        db,
		HeartBeat: time.Now(),
		TimeOut:   50 * time.Minute,
	}

	return c, nil
}
func main() {
	pool := New(10, 5, 5*time.Second)
	pool.OpenConnection = openNewConnection

	err := pool.Open()
	if err != nil {
		fmt.Println("Error opening pool:", err)
		return
	}

	conn, err := pool.Acquire()
	if err != nil {
		fmt.Println("Error acquiring connection:", err)
		return
	}

	hc := pool.Check(conn)
	fmt.Println(hc)
	// 使用连接

	pool.Release(conn)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, _ := pool.Acquire()
			hc := pool.Check(conn)
			fmt.Println(hc)
			// 使用连接
			pool.Release(conn)
		}()
	}
	wg.Wait()
	// 应用退出前关闭所有连接
	defer pool.Close()
}
