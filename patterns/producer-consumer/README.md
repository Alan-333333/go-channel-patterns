# Producer-Consumer

这个包实现了生产者-消费者模式,通过Goroutine和channel进行数据传输和处理。

## 特性

- Producer通过Goroutine并发生成数据,写入Buffer Channel

- Consumer通过Goroutine从Buffer Channel并发读取数据处理

- 使用缓冲Channel进行解耦

- 支持配置生产者和消费者的并发数

- 非阻塞写保证生产者效率

- 超时机制防止永久阻塞

- 反压机制防止生产速度过快

- 自定义错误处理

## 示例

```go
p := newProducer() 

c := newConsumer()

// 实现生产和消费函数

p.Run(ctx) 

c.Run(ctx)

p.Inject(ctx, c.Buffer)
```

## 接口

- `NewProducer` 和 `NewConsumer` 创建实例

- `Producer.Run` 启动生产goroutine 

- `Consumer.Run` 启动消费goroutine

- `Producer.Inject` 将数据从生产者输入消费者channel

- `Producer.Close` 和 `Consumer.Close` 关闭

## 生产者接口

- `SetProduceFunc` 自定义生产函数

- `SetErrorHandler` 错误处理

- `SetNotifier` 生命周期通知

## 消费者接口

- `SetConsumeFunc` 自定义消费函数 

- `SetErrorHandler` 错误处理

- `SetNotifier` 生命周期通知
