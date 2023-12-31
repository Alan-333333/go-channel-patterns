// package producerconsumer implements a producer and consumer model
// to handle data generation and processing in goroutines.
package producerconsumer

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gomonkey "github.com/agiledragon/gomonkey/v2"
)

func TestNewProducer(t *testing.T) {

	// Test various bufferSize
	p1 := NewProducer(10, 1)
	if p1 == nil || cap(p1.Buffer) != 10 {
		t.Error("Expected buffer size 10")
	}

	p2 := NewProducer(100, 2)
	if p2 == nil || cap(p2.Buffer) != 100 {
		t.Error("Expected buffer size 100")
	}

	// Test various numProcs
	if p1.NumProcs != 1 {
		t.Error("Expected 1 proc")
	}

	if p2.NumProcs != 2 {
		t.Error("Expected 2 procs")
	}
}

func TestProducer_Run(t *testing.T) {

	// 设置一个1分钟的超时时间
	timeout := time.After(time.Minute)

	// 用一个channel来标识done
	done := make(chan bool)

	p := NewProducer(1, 3)

	patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(p), "runProc",
		func(_ context.Context, _ *sync.WaitGroup) {
			// 自定义逻辑
		})

	defer patch.Reset()

	var wg sync.WaitGroup
	var counter int32

	const EXPECTED_PROCS = 3

	patchWg := gomonkey.ApplyMethod(reflect.TypeOf(&wg), "Wait", func() {})
	defer patchWg.Reset()

	for i := 0; i < EXPECTED_PROCS; i++ {
		wg.Add(1)
	}

	// 使用 atomic 原子操作来更新计数
	atomic.AddInt32(&counter, int32(EXPECTED_PROCS))

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		p.Run(ctx)
		wg.Done()
		done <- true
	}()

	select {

	case <-timeout:
		t.Fatal("test timeout")

	case <-done:
		// pass
	}
	time.Sleep(time.Millisecond * 100)

	// 对比 atomic 的计数和期望值
	if atomic.LoadInt32(&counter) != int32(EXPECTED_PROCS) {
		t.Errorf("Expected %d procs, got %d", EXPECTED_PROCS, counter)
	}

	cancel()

	wg.Wait()

}

func TestProducer_runProc(t *testing.T) {

	// 创建一个 Producer
	p := &Producer{
		Buffer: make(chan interface{}, 10),
		ProduceFunc: func() (interface{}, error) {
			return "data", nil
		},
	}
	patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(p), "applyBackpressure",
		func() {
			// 自定义逻辑
		})

	defer patch.Reset()
	// 用 WaitGroup 记录执行次数
	var wg sync.WaitGroup
	wg.Add(1)

	// 用 Context 可以取消 runProc
	ctx, cancel := context.WithCancel(context.Background())

	// 执行 runProc
	go p.runProc(ctx, &wg)

	// 让它执行一段时间
	time.Sleep(time.Millisecond * 100)

	// 校验data是否写入 Buffer
	data := <-p.Buffer
	if data != "data" {
		t.Error("Expected 'data' in buffer")
	}

	// 取消 Context
	cancel()

	// 等待结束
	wg.Wait()

	// Buffer 不应有其他数据
	if len(p.Buffer) != 0 {
		t.Error("Buffer should be empty after cancel")
	}
}

func TestProducer_Inject(t *testing.T) {

	p := &Producer{
		Buffer: make(chan interface{}, 10),
	}

	// 校验通知被调用
	notified := false
	p.Notifier = func(string) {
		notified = true
	}

	// 用来接收注入的数据
	out := make(chan interface{}, 10)

	// 用 Context 取消注入
	ctx, cancel := context.WithCancel(context.Background())

	// 开始注入
	go p.Inject(ctx, out)

	// 向 Buffer 写入数据
	p.Buffer <- "foo"
	p.Buffer <- "bar"

	// 校验 out 通道接收到的数据
	if <-out != "foo" {
		t.Error("Expected 'foo' in out")
	}

	if <-out != "bar" {
		t.Error("Expected 'bar' in out")
	}

	cancel()

	// 等待结束
	<-time.After(time.Millisecond * 100)

	// 校验通知被调用
	if !notified {
		t.Error("InjectFinished notification not received")
	}

	// Buffer 和 out 通道应该为空
	if len(p.Buffer) != 0 || len(out) != 0 {
		t.Error("Channels should be empty after cancel")
	}
}

func TestProducer_Close(t *testing.T) {

	p := &Producer{
		Buffer: make(chan interface{}, 1),
	}

	closed := make(chan bool)

	// Buffer 通道应该开始是开启的
	if cap(p.Buffer) != 1 {
		t.Error("Buffer channel should be open")
	}

	// 调用 Close 方法
	p.Close()

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
		p.Buffer <- "data"
		closed <- true
	}()
}

func TestProducer_HandleError(t *testing.T) {

	p := NewProducer(1, 1)

	err := fmt.Errorf("mock error")
	// 模拟数据生成函数返回错误
	p.ProduceFunc = func() (interface{}, error) {
		return nil, err
	}

	handledCount := 0
	// Mock Notifier
	p.Notifier = func(s string) {}

	p.HandleError(func(error) {
		handledCount++
	})

	p.handleError(err)

	// 检查错误是否被处理函数接收到
	if handledCount == 0 {
		t.Error("Expected error to be handled")
	}
}

func TestProducer_tryReadBuffer(t *testing.T) {

	buffer := make(chan interface{}, 1)

	p := &Producer{
		Buffer: buffer,
	}

	// 当 buffer 为空时,tryReadBuffer 应该返回 false
	_, ok := p.tryReadBuffer()
	if ok {
		t.Error("Should return false when buffer is empty")
	}

	// 当 buffer 写入数据后,tryReadBuffer 应该返回数据
	buffer <- "data"
	d, ok := p.tryReadBuffer()
	if !ok || d != "data" {
		t.Error("Should return data when buffer has data")
	}
}

func TestProducer_isCancelled(t *testing.T) {

	p := &Producer{}
	// 未取消的 Context
	ctx := context.Background()
	if p.isCancelled(ctx) {
		t.Error("Unpancelled context should return false")
	}

	// 已取消的 Context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if !p.isCancelled(ctx) {
		t.Error("Cancelled context should return true")
	}
}

func TestProducer_applyBackpressure(t *testing.T) {

	p := &Producer{}
	// 模拟数据填满 buffer
	p.Buffer = make(chan interface{}, 1)
	p.Buffer <- "data"

	// 定义通知函数,检测是否被调用
	var notified bool
	p.Notifier = func(string) {
		notified = true
	}

	// 调用 applyBackpressure
	p.applyBackpressure()

	// 检查通知函数是否被调用
	if !notified {
		t.Error("Notifier should be called")
	}

	// 检查是否有睡眠
	start := time.Now()
	p.applyBackpressure()
	if time.Since(start) < time.Millisecond*50 {
		t.Error("Should sleep on backpressure")
	}
}

func TestProducer_handleError(t *testing.T) {

	p := &Producer{}

	// 定义错误处理函数
	handleCount := 0
	p.ErrHandler = func(err error) {
		handleCount++
	}

	// 定义通知函数
	var notified bool
	p.Notifier = func(string) {
		notified = true
	}

	// 传入错误,调用 handleError
	err := errors.New("test error")
	p.handleError(err)

	// 检查通知和错误处理是否被调用
	if !notified {
		t.Error("Notifier should be called")
	}

	if handleCount != 1 {
		t.Error("ErrHandler should be called")
	}
}
