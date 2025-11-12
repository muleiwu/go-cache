# 可插拔序列化系统使用指南

go-cache 现在支持可插拔的序列化系统，允许你根据需求选择不同的序列化器。

## 概述

go-cache提供以下序列化器：

1. **Gob序列化器** (默认) - Go原生，类型安全
2. **JSON序列化器** - 跨语言，人类可读

## 序列化器对比

| 特性 | Gob | JSON |
|------|-----|------|
| **类型安全** | ✅ 完整 | ⚠️ 部分 |
| **性能（编码）** | 中等 (~1052 ns/op) | 快 (~161 ns/op) |
| **性能（解码）** | 慢 (~6199 ns/op) | 中等 (~1436 ns/op) |
| **跨语言支持** | ❌ 仅Go | ✅ 全语言 |
| **人类可读** | ❌ 二进制 | ✅ 文本 |
| **复杂结构体** | ✅ 完美支持 | ✅ 良好支持 |
| **nil值支持** | ✅ 完整 | ✅ 完整 |
| **调试友好** | ⚠️ 困难 | ✅ 容易 |

## 使用方法

### 1. 使用默认序列化器（Gob）

```go
package main

import (
    "context"
    "time"
    
    "github.com/muleiwu/go-cache"
    "github.com/redis/go-redis/v9"
)

func main() {
    // 创建Redis客户端
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 使用默认的Gob序列化器
    cache := go_cache.NewRedis(rdb)
    
    // 正常使用
    ctx := context.Background()
    cache.Set(ctx, "key", "value", 10*time.Minute)
}
```

### 2. 使用JSON序列化器

```go
package main

import (
    "context"
    "time"
    
    "github.com/muleiwu/go-cache"
    "github.com/muleiwu/go-cache/serializer"
    "github.com/redis/go-redis/v9"
)

func main() {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 使用JSON序列化器
    cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))
    
    ctx := context.Background()
    
    // JSON序列化的数据在Redis中是人类可读的
    type User struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }
    
    user := User{ID: 1, Name: "张三"}
    cache.Set(ctx, "user:1", user, 10*time.Minute)
    
    // 可以在Redis中直接查看：
    // redis-cli GET user:1
    // {"is_nil":false,"value":{"id":1,"name":"张三"}}
}
```

### 3. 在同一个应用中使用不同的序列化器

```go
package main

import (
    "github.com/muleiwu/go-cache"
    "github.com/muleiwu/go-cache/serializer"
    "github.com/redis/go-redis/v9"
)

func main() {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 用于内部数据，使用Gob（更快，类型安全）
    internalCache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewGob()))
    
    // 用于API数据，使用JSON（跨语言，可读）
    apiCache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))
    
    // 分别使用
    // internalCache.Set(...)
    // apiCache.Set(...)
}
```

## 使用场景建议

### 使用Gob的场景

✅ **推荐场景：**
- 纯Go微服务架构
- 需要存储复杂结构体（嵌套结构、指针等）
- 性能敏感的应用（解码较慢但编码快）
- 不需要跨语言访问缓存

❌ **不推荐场景：**
- 需要与其他语言共享缓存
- 需要在Redis中人工检查数据
- 调试阶段

### 使用JSON的场景

✅ **推荐场景：**
- 跨语言微服务架构
- 需要人工检查Redis数据
- 调试阶段
- API响应缓存
- 与前端/移动端共享数据

❌ **不推荐场景：**
- 极致性能要求（解码比Gob快，但编码慢）
- 需要存储不可JSON化的类型（如channel）

## 完整示例

### 示例1：使用JSON序列化器的API缓存

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/muleiwu/go-cache"
    "github.com/muleiwu/go-cache/serializer"
    "github.com/redis/go-redis/v9"
)

type Product struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
}

func main() {
    // 创建使用JSON序列化器的Redis缓存
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    defer rdb.Close()
    
    cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))
    ctx := context.Background()
    
    // 缓存产品数据
    product := Product{
        ID:          100,
        Name:        "iPhone 15",
        Price:       999.99,
        Description: "Latest iPhone",
    }
    
    err := cache.Set(ctx, "product:100", product, 30*time.Minute)
    if err != nil {
        panic(err)
    }
    
    // 获取产品数据
    var cached Product
    err = cache.Get(ctx, "product:100", &cached)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Cached Product: %+v\n", cached)
    
    // 在Redis CLI中可以直接查看：
    // redis-cli GET product:100
    // 输出是人类可读的JSON
}
```

### 示例2：使用Gob的内部缓存

```go
package main

import (
    "context"
    "time"
    
    "github.com/muleiwu/go-cache"
    "github.com/muleiwu/go-cache/serializer"
    "github.com/redis/go-redis/v9"
)

type InternalConfig struct {
    Settings map[string]interface{}
    Flags    []string
    Version  *string
}

func main() {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    defer rdb.Close()
    
    // 使用Gob序列化器（默认）
    cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewGob()))
    ctx := context.Background()
    
    // Gob完美处理复杂结构
    version := "v1.0.0"
    config := InternalConfig{
        Settings: map[string]interface{}{
            "timeout": 30,
            "retries": 3,
        },
        Flags:   []string{"debug", "verbose"},
        Version: &version,
    }
    
    cache.Set(ctx, "internal:config", config, 1*time.Hour)
    
    var cached InternalConfig
    cache.Get(ctx, "internal:config", &cached)
    
    // cached 完全还原，包括指针
    // cached.Version 指向正确的字符串
}
```

### 示例3：自定义序列化器

你也可以实现自己的序列化器：

```go
package main

import (
    "github.com/muleiwu/go-cache/serializer"
)

// 实现serializer.Serializer接口
type CustomSerializer struct{}

func (c *CustomSerializer) Name() string {
    return "custom"
}

func (c *CustomSerializer) Encode(value interface{}) ([]byte, error) {
    // 你的编码逻辑
    return nil, nil
}

func (c *CustomSerializer) Decode(data []byte, obj any) error {
    // 你的解码逻辑
    return nil
}

func main() {
    // 使用自定义序列化器
    // cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(&CustomSerializer{}))
}
```

## 性能基准测试

基于 `TestSerializerUser{ID: 1, Name: "Benchmark User", Age: 30}`

### Gob序列化器

```
BenchmarkGobSerializer/Encode-14    1132600    1052 ns/op
BenchmarkGobSerializer/Decode-14     196642    6199 ns/op
```

### JSON序列化器

```
BenchmarkJsonSerializer/Encode-14   7507710     161.6 ns/op
BenchmarkJsonSerializer/Decode-14    811050    1436 ns/op
```

### 性能总结

- **JSON编码** 比 Gob **快6.5倍**
- **JSON解码** 比 Gob **快4.3倍**
- **JSON总体性能** 优于 Gob

但是：
- Gob类型安全性更高
- Gob支持更复杂的类型

## 注意事项

### 1. 序列化器不可混用

不同序列化器编码的数据不能互相解码：

```go
// ❌ 错误示例
cacheGob := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewGob()))
cacheJson := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))

cacheGob.Set(ctx, "key", "value", ttl)
cacheJson.Get(ctx, "key", &value) // ❌ 解码失败！
```

### 2. 切换序列化器需要清空缓存

如果你要切换序列化器，需要先清空Redis：

```bash
redis-cli FLUSHALL
```

或者使用不同的key前缀：

```go
gobCache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewGob()))
jsonCache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))

gobCache.Set(ctx, "gob:key", value, ttl)
jsonCache.Set(ctx, "json:key", value, ttl)
```

### 3. JSON序列化限制

JSON不支持某些Go特性：

```go
// ❌ JSON不支持
type Unsupported struct {
    Ch chan int  // 通道
    Fn func()    // 函数
}
```

### 4. Gob只能在Go中使用

如果其他语言（Python、Java等）需要访问缓存，必须使用JSON。

## 最佳实践

### 1. 根据数据类型选择序列化器

```go
// 内部配置 → Gob
configCache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewGob()))

// API响应 → JSON  
apiCache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))

// 用户session → JSON（可能需要跨服务访问）
sessionCache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))
```

### 2. 使用key前缀区分序列化器

```go
const (
    GobPrefix  = "gob:"
    JsonPrefix = "json:"
)

gobCache.Set(ctx, GobPrefix+"config", config, ttl)
jsonCache.Set(ctx, JsonPrefix+"api_response", response, ttl)
```

### 3. 在配置中指定序列化器

```go
type CacheConfig struct {
    SerializerType string // "gob" or "json"
}

func NewCacheFromConfig(config CacheConfig, rdb *redis.Client) *go_cache.Redis {
    var ser serializer.Serializer
    
    switch config.SerializerType {
    case "json":
        ser = serializer.NewJson()
    case "gob":
        fallthrough
    default:
        ser = serializer.NewGob()
    }
    
    return go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(ser))
}
```

## 故障排查

### 问题1：解码失败

**症状：** `Decode() error = ...`

**原因：** 使用了错误的序列化器读取数据

**解决：** 确保读写使用相同的序列化器

### 问题2：类型不匹配

**症状：** `type mismatch: expected X, got Y`

**原因：** 目标类型与存储类型不匹配

**解决：** 确保Get时的类型与Set时的类型一致

### 问题3：JSON解析失败

**症状：** `json decode error: ...`

**原因：** 结构体字段缺少json标签

**解决：** 添加json标签：

```go
type User struct {
    ID   int    `json:"id"`    // ✅ 有标签
    Name string `json:"name"`  // ✅ 有标签
}
```

## 总结

- **默认使用Gob** - 适合纯Go应用
- **跨语言用JSON** - 适合微服务架构
- **调试时用JSON** - 方便查看数据
- **生产环境评估性能** - 根据实际测试选择

选择合适的序列化器可以提升应用性能和开发体验！
