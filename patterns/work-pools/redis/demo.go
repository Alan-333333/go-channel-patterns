package redispool

import (
	"time"

	"github.com/go-redis/redis"
)

func main() {

	// Creating a Connection Pool
	pool := New(10, 5, time.Minute)

	// Setting the function that opens the connection
	pool.OpenConnection = func() (*RedisConn, error) {
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
		return &RedisConn{Conn: client, TimeOut: time.Minute}, nil
	}

	// Open Connection Pool
	err := pool.Open()
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	// Get a connection from the pool
	conn, err := pool.Acquire()
	if err != nil {
		panic(err)
	}

	// Using Connections
	conn.Conn.Set("key", "value", 0)

	// Release Connections
	pool.Release(conn)

}
