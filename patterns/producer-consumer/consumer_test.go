// package Consumerconsumer implements a Consumer and consumer model
// to handle data generation and processing in goroutines.
package producerconsumer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewConsumer(t *testing.T) {
	c := NewConsumer(10, 5)

	if c.Buffer == nil {
		t.Error("Expected Buffer channel to be created")
	}

	if c.NumProcs != 5 {
		t.Error("Expected NumProcs to be set correctly")
	}
}

func TestCallbacks(t *testing.T) {

	var errHandler ErrHandler
	var notifier Notifier

	c := &Consumer{}
	c.HandleError(func(err error) {
		errHandler(err)
	})

	c.Notify(func(msg string) {
		notifier(msg)
	})

	// 验证回调函数被设置
	if c.ErrHandler == nil {
		t.Error("Expected ErrHandler to be set")
	}

	if c.Notifier == nil {
		t.Error("Expected Notifier to be set")
	}

}

func TestRunProc(t *testing.T) {

	handleCount := 0
	// 设置测试依赖
	consumeFunc := func(data interface{}) error {
		handleCount++
		return nil
	}

	c := &Consumer{
		Buffer:      make(chan interface{}, 10),
		ConsumeFunc: consumeFunc,
	}
	c.Buffer <- "data"
	ctx := context.Background()
	// 初始化 WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	// 运行并验证
	c.runProc(ctx, &wg)

	// 等待 goroutine 完成
	wg.Wait()

	// 验证数据被消费
	if len(c.Buffer) != 0 {
		t.Error("Expected Buff be consumed")
	}

	if handleCount == 0 {
		t.Error("Expected ConsumeFunc to be called")
	}

}

func TestClose(t *testing.T) {

	c := &Consumer{
		Buffer: make(chan interface{}, 1),
	}

	closed := make(chan bool)

	// Buffer 通道应该开始是开启的
	if cap(c.Buffer) != 1 {
		t.Error("Buffer channel should be open")
	}

	// 调用 Close 方法
	c.Close()

	// Buffer 通道应该被关闭
	// 尝试向已关闭通道写入应该会被阻塞
	select {
	case <-closed:
		t.Error("Write to closed channel should block")
	case <-time.After(100 * time.Millisecond):
	}

	// 向关闭的 Buffer 通道写入数据应该 panic
	func() {
		defer func() {
			if recover() == nil {
				t.Error("Write to closed channel should panic")
			}
		}()
		c.Buffer <- "data"
		closed <- true
	}()

}
func TestIsCancelled(t *testing.T) {

	// 测试 ctx 未取消的情况
	ctx := context.Background()
	c := &Consumer{}

	if c.isCancelled(ctx) {
		t.Error("isCancelled should return false for non-cancelled context")
	}

	// 测试 ctx 被取消的情况
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if !c.isCancelled(ctx) {
		t.Error("isCancelled should return true for cancelled context")
	}

}
func TestHandleError(t *testing.T) {

	// 模拟依赖
	var calledNotifier bool
	notifier := func(msg string) {
		calledNotifier = true
	}

	var calledErrHandler bool
	errHandler := func(err error) {
		calledErrHandler = true
	}

	c := &Consumer{
		Notifier:   notifier,
		ErrHandler: errHandler,
	}

	// 测试调用
	c.handleError(errors.New("test error"))

	// 验证
	if !calledNotifier {
		t.Error("Expected Notifier to be called")
	}

	if !calledErrHandler {
		t.Error("Expected ErrHandler to be called")
	}

}

func TestTryReadBuffer(t *testing.T) {

	// 测试通道为空的情况
	c := &Consumer{
		Buffer: make(chan interface{}, 1),
	}

	data, ok := c.tryReadBuffer()

	if ok {
		t.Error("Should return false when buffer is empty")
	}

	// 测试通道有数据的情况
	c.Buffer <- "test"

	data, ok = c.tryReadBuffer()

	if !ok {
		t.Error("Should return true when buffer has data")
	}

	if data != "test" {
		t.Error("Should return channel data correctly")
	}

}
func TestIsTimedOut(t *testing.T) {

	c := &Consumer{}

	// 短暂 sleep 期间不应超时
	timeout := 10 * time.Millisecond
	time.Sleep(5 * time.Millisecond)
	require.False(t, c.isTimedOut(timeout))

	// 长时间 sleep 期间应超时
	timeout = 10 * time.Millisecond
	timer := time.After(20 * time.Millisecond)

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	select {
	case <-timer:
	case <-timeoutCtx.Done():
		// 处理接收被阻塞问题
		require.True(t, c.isTimedOut(timeout))
	}
}
