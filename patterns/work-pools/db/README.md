# Database Connection Pool

这个包实现了数据库连接池。

## 特性

- 设置最小和最大连接数
- 连接重用,避免重复创建连接  
- 连接超时设置
- 健康检查机制,关闭失效连接
- 定期清理过期连接

## 用法

```go
pool := dbpool.New(max, min, timeout)

pool.Open() 

conn, err := pool.Acquire()

// 使用连接

pool.Release(conn)

pool.Close()
```

## 接口

- `New` 创建连接池,传入最大连接数、最小连接数和获取连接超时时间
- `Open` 打开连接池,初始化连接
- `Acquire` 获取一个连接
- `Release` 释放使用完的连接  
- `Close` 关闭连接池
- `Cleaner` 定期清理过期连接
- `Check` 健康检查连接

## 实现

- 使用channel管理连接池
- 打开连接后放入连接池供重用
- 获取连接时优先返回已有连接
- 定期清理过期和失效连接
- 小于最小连接数时打开新连接  

## TODO

- 添加连接池统计和指标
- 从配置文件初始化连接池
- 连接泄漏检测
