# Counter Rate Limiter

这个包实现了基于计数器的限流算法。

## 特性

- 使用计数器记录时间窗口内的请求数

- 滑动时间窗口,超时后重置计数器

- 支持直接设置允许的每秒请求数(RPS)

- 并发安全的计数器

- 首次请求不限流,避免冷启动问题

- 简单易用

## 示例

```go
limiter := counter.New(100) // 100 RPS

for {
  if limiter.Allow() {
    // 处理请求
  } else {
    // 限流逻辑
  } 
}
```

## 接口

- `New` 创建限流器,传入允许的RPS

- `Allow` 处理请求,检查是否超过限流

- `Counter` 限流器结构体

## 实现逻辑

- 使用计数器记录当前时间窗口的请求数

- 计算时间窗口内的允许请求数 = RPS * 时间间隔

- 比较当前请求数和允许请求数判断是否限流

- 滑动时间窗口,周期性重置计数器

## 优点

- 实现简单,资源消耗低
- 支持直接精确限流
- 自动滑动时间窗口
- 无状态,可横向扩展
