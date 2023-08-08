package dbpool

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var once sync.Once

// DBConn 封装数据库连接
type DBConn struct {
	DB        *sql.DB
	HeartBeat time.Time
	TimeOut   time.Duration
}

// ConnectionPool manages a pool of connections.
type ConnectionPool struct {

	// conns is the pool of connections.
	conns chan *DBConn

	// maxConnections is the maximum number of connections in the pool.
	maxConnections int

	// minConnections is the minimum number of connections in the pool.
	minConnections int

	// waitTimeout is the timeout for getting a connection.
	waitTimeout time.Duration

	// OpenConnection opens a new connection.
	OpenConnection func() (*DBConn, error)

	// cleanupTicker ticks periodically for cleaning up expired connections.
	cleanupTicker *time.Ticker
}

// New creates a new ConnectionPool.
func New(maxConnections, minConnections int, waitTimeout time.Duration) *ConnectionPool {

	return &ConnectionPool{

		// Initialize the pool.
		conns: make(chan *DBConn, maxConnections),

		// Set config fields.
		maxConnections: maxConnections,
		minConnections: minConnections,
		waitTimeout:    waitTimeout,
	}
}

// Open initializes the connection pool.
func (p *ConnectionPool) Open() error {

	// Open maximum connections.
	for i := 0; i < p.maxConnections; i++ {

		conn, err := p.OpenConnection()
		if err != nil {
			return err
		}

		p.conns <- conn
	}

	// Start a goroutine to clean up expired connections periodically.
	p.cleanupTicker = time.NewTicker(time.Minute)
	go func() {
		for range p.cleanupTicker.C {
			p.Cleaner()
		}
	}()

	return nil
}

// Acquire retrieves a connection from the pool.
func (p *ConnectionPool) Acquire() (*DBConn, error) {

	// Try to get a connection before timeout.
	select {

	case conn := <-p.conns:
		// Check connection health before reusing it.
		if p.isConnectionExpired(conn) {
			conn.DB.Close()
			return nil, errors.New("connection expired")
		}
		return conn, nil

	case <-time.After(p.waitTimeout):
		return nil, fmt.Errorf("timeout waiting for connection")
	}
}

// Check if connection has expired.
func (p *ConnectionPool) isConnectionExpired(conn *DBConn) bool {
	return conn.HeartBeat.Add(conn.TimeOut).Before(time.Now())
}

// Release puts a connection back into the pool.
func (p *ConnectionPool) Release(conn *DBConn) {

	// Mark connection as active again before releasing.
	conn.HeartBeat = time.Now()

	p.conns <- conn
}

// Close closes the connection pool.
func (p *ConnectionPool) Close() {

	// Stop the cleaner.
	p.cleanupTicker.Stop()

	// Close all connections.
	once.Do(func() {
		close(p.conns)
	})

	for conn := range p.conns {
		conn.DB.Close()
	}
}

// CleanUpClosedConnections closes expired connections and
// opens new connections to maintain min connections.
func (p *ConnectionPool) Cleaner() {
	// closes expired connections.
	p.CloseExpiredConnections()

	//MaintainMinConnections opens connections if below min.
	p.MaintainMinConnections()
}

// CloseExpiredConnections closes expired connections.
func (p *ConnectionPool) CloseExpiredConnections() {

	const timePerConn = 10 * time.Millisecond

	var timeout = time.Duration(len(p.conns)) * timePerConn

	newConns := make(chan *DBConn)
	// Loop to close expired connections
	for {
		select {
		case conn := <-p.conns:
			if !conn.HeartBeat.Add(conn.TimeOut).Before(time.Now()) {
				newConns <- conn
			} else {
				conn.DB.Close()
			}

		case <-time.After(timeout):
			// over time return
			p.conns = newConns
			return
		}
	}

}

// MaintainMinConnections opens connections if below min.
func (p *ConnectionPool) MaintainMinConnections() {

	// Loop to open connections
	for i := len(p.conns); i < p.minConnections; i++ {
		conn, err := p.OpenConnection()
		if err != nil {
			continue
		}
		p.conns <- conn
	}
}

// Check returns true if connection is healthy.
func (p *ConnectionPool) Check(conn *DBConn) bool {

	// Check connection health with a ping.
	return conn.DB.Ping() == nil
}
