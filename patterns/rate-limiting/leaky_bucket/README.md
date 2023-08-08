# Leaky Bucket Rate Limiter

该包实现了基于漏桶算法的限流器。

## 特性

- 支持设置桶的容量和流出速率

- 滑动时间窗口,周期性重置

- 平滑限流,防止突发流量击穿

- 简单易用

## 示例 

```go
bucket := leakybucket.New(1000, 100) // 1000 capacity, 100 requests/sec

for {
  if bucket.Allow() {
    // handle request
  } else {
    // limit request
  }
}
```

## 接口

- `New` 创建漏桶,传入容量和流出速率

- `Allow` 处理请求,检查是否限流

- `LeakyBucket` 漏桶结构体

## 实现原理

- 桶中存储的“水”表示可以处理的请求数

- 漏洞以固定速率流出水,控制处理速率

- 当桶中水量不足时,限制新的请求

- 滑动时间窗口周期性重置水量

## 优点

- 限流平滑,避免突增流量击穿
- 容易配置限流速率
- 无状态,可以扩展
