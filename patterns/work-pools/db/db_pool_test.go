package dbpool

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey"
	_ "github.com/go-sql-driver/mysql"
)

func TestNew(t *testing.T) {

	maxConn := 10
	minConn := 5
	timeout := time.Second * 30

	pool := New(maxConn, minConn, timeout)

	if pool.maxConnections != maxConn {
		t.Errorf("maxConnections not set correctly")
	}

	if pool.minConnections != minConn {
		t.Errorf("minConnections not set correctly")
	}

	if pool.waitTimeout != timeout {
		t.Errorf("waitTimeout not set correctly")
	}

	if cap(pool.conns) != maxConn {
		t.Errorf("conns channel capacity not set correctly")
	}

}

func TestOpen(t *testing.T) {

	// mock OpenConnection
	openConnInvoked := 0
	mockOpenConn := func() (*DBConn, error) {
		openConnInvoked++
		return &DBConn{}, nil
	}

	// create pool
	pool := New(10, 5, 30*time.Second)
	pool.OpenConnection = mockOpenConn

	// call Open
	err := pool.Open()
	if err != nil {
		t.Fatal(err)
	}

	// check OpenConnection invoked
	if openConnInvoked != pool.maxConnections {
		t.Errorf("OpenConnection not invoked expected times")
	}

	// check connections
	if len(pool.conns) != pool.maxConnections {
		t.Errorf("number of connections not equals maxConnections")
	}

	// check cleanup goroutine
	if pool.cleanupTicker == nil {
		t.Error("cleanup goroutine not started")
	}

}

func TestAcquire(t *testing.T) {

	// mock connection
	mockConn := &DBConn{}

	// mock close with gomonkey
	var closed bool
	patches := gomonkey.ApplyMethod(reflect.TypeOf(mockConn.DB), "Close", func(*sql.DB) error {
		closed = true
		return nil
	})
	defer patches.Reset()

	// create pool
	pool := New(10, 5, 30*time.Second)

	// mark mockConn expired
	mockConn.HeartBeat = time.Now().Add(-1 * time.Hour)

	// add connection to pool
	pool.conns <- mockConn

	// acquire connection
	conn, _ := pool.Acquire()

	// assert closed
	if !closed {
		t.Error("did not close expired connection")
	}

	// assert returned connection
	if conn != nil {
		t.Error("should return nil on expired connection")
	}

}

func TestIsConnectionExpired(t *testing.T) {
	// create pool
	conn := &DBConn{
		HeartBeat: time.Now(),
		TimeOut:   10 * time.Second,
	}
	p := New(10, 5, 30*time.Second)

	// not expired
	if p.isConnectionExpired(conn) {
		t.Error("new connection should not be expired")
	}

	// update HeartBeat to 11 second ago
	conn.HeartBeat = time.Now().Add(-11 * time.Second)

	// expired
	if !p.isConnectionExpired(conn) {
		t.Error("expired connection should be expired")
	}

}

func TestRelease(t *testing.T) {
	// mock connection
	conn := &DBConn{HeartBeat: time.Now()}

	// create pool
	pool := New(10, 5, 30*time.Second)

	// call Release
	pool.Release(conn)

	// check connection heartbeat reset
	if conn.HeartBeat.IsZero() {
		t.Error("connection heartbeat not reset")
	}
}

func TestClose(t *testing.T) {
	// mock connection
	mockConn := &DBConn{}

	// mock close with gomonkey
	var closed bool
	patchesClose := gomonkey.ApplyMethod(reflect.TypeOf(mockConn.DB), "Close", func(*sql.DB) error {
		closed = true
		return nil
	})
	defer patchesClose.Reset()

	// create pool with connection
	pool := New(10, 5, 30*time.Second)
	pool.conns <- mockConn

	pool.cleanupTicker = time.NewTicker(time.Minute)

	// call Close
	pool.Close()

	// check connection closed
	if !closed {
		t.Error("connection not closed")
	}

	if len(pool.cleanupTicker.C) != 0 {
		t.Error("ticker not stop")
	}
}

func TestCloseExpiredConnections(t *testing.T) {

	// Creat ConnectionPool
	p := &ConnectionPool{
		conns: make(chan *DBConn, 3),
	}

	expiredConn := &DBConn{HeartBeat: time.Now().Add(-time.Hour)}
	normalConn := &DBConn{HeartBeat: time.Now()}

	p.conns <- expiredConn
	p.conns <- normalConn

	patchesClose := gomonkey.ApplyMethod(reflect.TypeOf((*sql.DB)(nil)), "Close", func(*sql.DB) error {
		return nil
	})
	defer patchesClose.Reset()

	connsBefore := len(p.conns)
	// connsBefore
	t.Logf("conns before close: %d", connsBefore)

	// CloseExpiredConnections
	p.CloseExpiredConnections()

	connsAfter := len(p.conns)
	// connsAfter
	t.Logf("conns after close: %d", connsAfter)

	if connsAfter >= connsBefore {
		t.Error("expired conn not closed")
	}
}

func TestMaintainMinConnections(t *testing.T) {
	// mock open connection
	openConn := 0
	// create pool
	pool := New(10, 5, 30*time.Second)
	pool.OpenConnection = func() (*DBConn, error) {
		openConn++
		return &DBConn{}, nil
	}

	// maintain min
	pool.MaintainMinConnections()

	// check opened connections
	if openConn != pool.minConnections {
		t.Error("did not open enough connections")
	}
}
