package redispool

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

// RedisConn encapsulates the Redis connection.
type RedisConn struct {
	Conn      *redis.Client
	HeartBeat time.Time
	TimeOut   time.Duration
}

// RedisConnectionPool manages a set of Redis connections.
type RedisConnectionPool struct {
	conns chan *RedisConn

	maxConnections int
	minConnections int

	waitTimeout time.Duration

	OpenConnection func() (*RedisConn, error)

	cleanupTicker *time.Ticker
}

// New Creates a new Redis connection pool
func New(maxConn, minConn int, waitTimeout time.Duration) *RedisConnectionPool {
	return &RedisConnectionPool{
		conns:          make(chan *RedisConn, maxConn),
		maxConnections: maxConn,
		minConnections: minConn,
		waitTimeout:    waitTimeout,
	}
}

// Open Initialize the connection pool
func (pool *RedisConnectionPool) Open() error {
	// Open the maximum number of connections
	for i := 0; i < pool.maxConnections; i++ {
		conn, err := pool.OpenConnection()
		if err != nil {
			return err
		}
		if conn.HeartBeat.IsZero() {
			conn.HeartBeat = time.Now()
		}
		pool.conns <- conn
	}

	// Start a goroutine to periodically clean up expired connections.
	pool.cleanupTicker = time.NewTicker(time.Minute)
	go func() {
		for range pool.cleanupTicker.C {
			pool.Cleaner()
		}
	}()

	return nil
}

// Acquire Acquire a connection
func (pool *RedisConnectionPool) Acquire() (*RedisConn, error) {
	select {
	case conn := <-pool.conns:
		// Check if the connection has expired
		if pool.isConnectionExpired(conn) {
			conn.Conn.Close()
			return nil, errors.New("connection expired")
		}
		return conn, nil
	case <-time.After(pool.waitTimeout):
		return nil, fmt.Errorf("timeout waiting for connection")
	}
}

// Release releases connections to the pool
func (pool *RedisConnectionPool) Release(conn *RedisConn) {
	conn.HeartBeat = time.Now()
	pool.conns <- conn
}

// Close closes the connection pool
func (pool *RedisConnectionPool) Close() {
	close(pool.conns)
	for conn := range pool.conns {
		conn.Conn.Close()
	}
}

// Checker checks if the connection is available
func (pool *RedisConnectionPool) Check(conn *RedisConn) bool {
	_, err := conn.Conn.Ping().Result()
	return err == nil
}

// Cleaner Clean up expired connections while maintaining minimum number of connections
func (pool *RedisConnectionPool) Cleaner() {
	pool.CloseExpiredConnections()
	pool.MaintainMinConnections()
}

// CloseExpiredConnections Close expired connections
func (pool *RedisConnectionPool) CloseExpiredConnections() {
	const timePerConn = 10 * time.Millisecond

	var timeout = time.Duration(len(pool.conns)) * timePerConn

	newConns := make(chan *RedisConn)
	// Loop to close expired connections
	for {
		select {
		case conn := <-pool.conns:
			if !conn.HeartBeat.Add(conn.TimeOut).Before(time.Now()) {
				newConns <- conn
			} else {
				conn.Conn.Close()
			}

		case <-time.After(timeout):
			// over time return
			pool.conns = newConns
			return
		}
	}
}

// MaintainMinConnections maintaining minimum number of connections
func (pool *RedisConnectionPool) MaintainMinConnections() {
	// Loop to open connections
	for i := len(pool.conns); i < pool.minConnections; i++ {
		conn, err := pool.OpenConnection()
		if err != nil {
			continue
		}
		pool.conns <- conn
	}
}

// isConnectionExpired Check if the connection has expired
func (pool *RedisConnectionPool) isConnectionExpired(conn *RedisConn) bool {
	return conn.HeartBeat.Add(conn.TimeOut).Before(time.Now())
}
