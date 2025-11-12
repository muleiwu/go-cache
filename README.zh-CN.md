# go-cache

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/muleiwu/go-cache)](https://goreportcard.com/report/github.com/muleiwu/go-cache)

go-cache 是一个统一接口的 Go 缓存库，提供了多种缓存实现方式，包括内存缓存、Redis 缓存和空缓存。该库实现了 `gsr.Cacher` 接口，支持在不同缓存实现之间无缝切换，为应用程序提供灵活的缓存解决方案。

## 🚀 特性

- **统一接口**: 所有缓存实现都遵循 `gsr.Cacher` 接口，便于切换和测试
- **多种实现**: 支持内存缓存、Redis 缓存和空缓存实现
- **类型安全**: 使用反射确保类型安全的值赋值
- **TTL 支持**: 支持设置键的生存时间
- **缓存穿透保护**: 提供 `GetSet` 方法防止缓存穿透
- **可插拔序列化**: 支持 Gob（默认）和 JSON 序列化器，可自定义扩展
- **完整的 nil 值支持**: 完整支持 nil 指针、nil 切片和 nil map
- **过期管理**: 支持设置具体的过期时间或相对的 TTL
- **上下文支持**: 所有操作都支持 context.Context

## 📦 安装

使用 go get 安装 go-cache：

```bash
go get github.com/muleiwu/go-cache
```

## 🏗️ 架构概览

```
go-cache/
├── memory.go          # 内存缓存实现
├── redis.go           # Redis 缓存实现
├── none.go            # 空缓存实现
├── serializer/        # 序列化器包
│   ├── serializer.go  # 序列化器接口
│   ├── gob.go         # Gob 序列化器（默认）
│   └── json.go        # JSON 序列化器
└── cache_value/       # 缓存值处理
    └── cache_value.go # 序列化/反序列化逻辑
```

### 核心组件

1. **缓存接口** (`gsr.Cacher`): 定义了统一的缓存操作接口
2. **内存缓存** (`Memory`): 基于内存的缓存实现，适用于单机应用
3. **Redis缓存** (`Redis`): 基于 Redis 的分布式缓存实现
4. **空缓存** (`None`): 空操作实现，用于测试或禁用缓存场景
5. **序列化系统** (`serializer`): 可插拔的序列化器，支持 Gob 和 JSON
6. **值处理** (`cache_value`): 处理缓存值的序列化和反序列化

## 🚀 快速开始

### 内存缓存示例

```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/muleiwu/go-cache"
)

func main() {
	// 创建内存缓存，默认过期时间 5 分钟，清理间隔 10 分钟
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()
	
	// 设置缓存
	err := cache.Set(ctx, "user:123", &User{ID: 123, Name: "张三"}, 10*time.Minute)
	if err != nil {
		panic(err)
	}
	
	// 获取缓存
	var user User
	err = cache.Get(ctx, "user:123", &user)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("用户: %+v\n", user)
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

### Redis 缓存示例

```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/muleiwu/go-cache"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 无密码
		DB:       0,  // 默认 DB
	})
	
	// 创建 Redis 缓存，使用默认的 Gob 序列化器
	cache := go_cache.NewRedis(rdb)
	ctx := context.Background()
	
	// 设置缓存
	err := cache.Set(ctx, "product:456", &Product{ID: 456, Name: "商品A", Price: 99.99}, 30*time.Minute)
	if err != nil {
		panic(err)
	}
	
	// 获取缓存
	var product Product
	err = cache.Get(ctx, "product:456", &product)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("商品: %+v\n", product)
}

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
```

### 缓存穿透保护示例

```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/muleiwu/go-cache"
)

func main() {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()
	
	// 使用 GetSet 防止缓存穿透
	var user User
	err := cache.GetSet(ctx, "user:789", 10*time.Minute, &user, func(key string, obj any) error {
		// 缓存未命中时从数据库获取数据
		fmt.Println("从数据库获取用户数据...")
		user := obj.(*User)
		user.ID = 789
		user.Name = "李四"
		return nil
	})
	
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("用户: %+v\n", user)
	
	// 第二次调用会直接从缓存获取
	var user2 User
	err = cache.GetSet(ctx, "user:789", 10*time.Minute, &user2, func(key string, obj any) error {
		fmt.Println("这个回调不会被调用，因为缓存已存在")
		return nil
	})
	
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("用户2: %+v\n", user2)
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

### 可插拔序列化器示例

#### 使用 JSON 序列化器（跨语言、人类可读）

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
	
	// 使用 JSON 序列化器
	cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))
	ctx := context.Background()
	
	// JSON 序列化的数据在 Redis 中是人类可读的
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	user := User{ID: 1, Name: "张三"}
	cache.Set(ctx, "user:1", user, 10*time.Minute)
	
	// 在 Redis CLI 中可以直接查看：
	// redis-cli GET user:1
	// {"is_nil":false,"value":{"id":1,"name":"张三"}}
}
```

#### 使用 Gob 序列化器（默认，类型安全）

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
	
	// 显式指定 Gob 序列化器（默认已是 Gob）
	cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewGob()))
	ctx := context.Background()
	
	// Gob 完美处理复杂结构体和 nil 值
	type Config struct {
		Settings map[string]interface{}
		Version  *string
	}
	
	version := "v1.0.0"
	config := Config{
		Settings: map[string]interface{}{"timeout": 30},
		Version:  &version,
	}
	
	cache.Set(ctx, "config", config, 1*time.Hour)
	
	// 获取时完全还原类型，包括指针
	var cached Config
	cache.Get(ctx, "config", &cached)
	// cached.Version 指向正确的字符串
}
```

#### 序列化器对比

| 特性 | Gob | JSON |
|------|-----|------|
| **类型安全** | ✅ 完整 | ⚠️ 部分 |
| **性能（编码）** | 中等 (~1052 ns/op) | 快 (~161 ns/op) |
| **性能（解码）** | 慢 (~6199 ns/op) | 中等 (~1436 ns/op) |
| **跨语言支持** | ❌ 仅 Go | ✅ 全语言 |
| **人类可读** | ❌ 二进制 | ✅ 文本 |
| **复杂结构体** | ✅ 完美支持 | ✅ 良好支持 |
| **nil 值支持** | ✅ 完整 | ✅ 完整 |
| **调试友好** | ⚠️ 困难 | ✅ 容易 |

**使用建议**：
- **默认使用 Gob** - 适合纯 Go 应用，类型安全
- **跨语言用 JSON** - 适合微服务架构
- **调试时用 JSON** - 方便查看 Redis 中的数据

详细使用指南请参阅 [SERIALIZER_GUIDE.md](docs/SERIALIZER_GUIDE.md)

### Nil 值支持

go-cache 提供完整的 nil 值支持，允许您区分"键不存在"和"键存在但值为 nil"：

```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/muleiwu/go-cache"
)

func main() {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()
	
	// 存储 nil 指针
	var user *User = nil
	err := cache.Set(ctx, "user:123", user, 10*time.Minute)
	if err != nil {
		panic(err)
	}
	
	// 获取 nil 指针
	var result *User
	err = cache.Get(ctx, "user:123", &result)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("用户为 nil: %v\n", result == nil) // 输出: 用户为 nil: true
	
	// 检查键是否存在
	exists := cache.Exists(ctx, "user:123")
	fmt.Printf("键存在: %v\n", exists) // 输出: 键存在: true
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

详细的 nil 值使用方法请参阅 [NIL_VALUES.md](docs/NIL_VALUES.md)

## 📚 API 文档

### 接口定义

go-cache 实现了 `gsr.Cacher` 接口，定义了以下方法：

```go
type Cacher interface {
    // Exists 检查键是否存在
    Exists(ctx context.Context, key string) bool
    
    // Get 获取缓存值，将结果反序列化到 obj 中
    Get(ctx context.Context, key string, obj any) error
    
    // Set 设置缓存值，ttl 为生存时间
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    
    // GetSet 获取缓存值，如果不存在则通过回调函数获取并设置
    GetSet(ctx context.Context, key string, ttl time.Duration, obj any, funCallback CacheCallback) error
    
    // Del 删除缓存键
    Del(ctx context.Context, key string) error
    
    // ExpiresAt 设置键在特定时间过期
    ExpiresAt(ctx context.Context, key string, expiresAt time.Time) error
    
    // ExpiresIn 设置键在指定时间后过期
    ExpiresIn(ctx context.Context, key string, ttl time.Duration) error
}

type CacheCallback func(key string, obj any) error
```

### 内存缓存 (Memory)

#### 构造函数

```go
func NewMemory(defaultExpiration, cleanupInterval time.Duration) *Memory
```

- `defaultExpiration`: 默认过期时间
- `cleanupInterval`: 清理过期项的时间间隔

#### 特性

- 基于内存的缓存实现
- 使用 `github.com/patrickmn/go-cache` 作为底层存储
- 支持自动清理过期项
- 线程安全

#### 使用示例

```go
// 创建内存缓存
cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)

// 设置缓存
err := cache.Set(ctx, "key", "value", 10*time.Minute)

// 获取缓存
var result string
err = cache.Get(ctx, "key", &result)

// 检查键是否存在
exists := cache.Exists(ctx, "key")

// 删除键
err = cache.Del(ctx, "key")

// 设置过期时间
err = cache.ExpiresIn(ctx, "key", 5*time.Minute)
err = cache.ExpiresAt(ctx, "key", time.Now().Add(5*time.Minute))
```

### Redis 缓存 (Redis)

#### 构造函数

```go
func NewRedis(conn *redis.Client, opts ...RedisOption) *Redis
```

- `conn`: Redis 客户端连接
- `opts`: 可选配置（例如 WithRedisSerializer）

#### 选项

```go
// 使用自定义序列化器
cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))
```

#### 特性

- 基于 Redis 的分布式缓存
- **可插拔序列化**: 默认使用 Gob，可切换为 JSON 或自定义序列化器
- **完整类型安全**: Gob 序列化器保证类型安全
- **nil 值支持**: 完整支持 nil 指针、nil 切片和 nil map
- 支持所有 Redis 数据类型
- 适用于分布式系统

#### 使用示例

```go
// 创建 Redis 客户端
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// 创建使用默认 Gob 序列化器的 Redis 缓存
cache := go_cache.NewRedis(rdb)

// 或创建使用 JSON 序列化器的 Redis 缓存
cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))

// 使用方式与内存缓存相同
err := cache.Set(ctx, "key", "value", 10*time.Minute)
var result string
err = cache.Get(ctx, "key", &result)
```

### 空缓存 (None)

#### 构造函数

```go
func NewNone() *None
func NewCacheNone() *None  // 别名
```

#### 特性

- 所有操作都是空操作或返回错误
- 用于测试或禁用缓存的场景
- 不存储任何数据

#### 使用示例

```go
// 创建空缓存
cache := go_cache.NewNone()

// Set 操作总是成功但不存储数据
err := cache.Set(ctx, "key", "value", 10*time.Minute) // 返回 nil

// Get 操作总是返回错误
var result string
err = cache.Get(ctx, "key", &result) // 返回 "not implemented" 错误

// Exists 总是返回 false
exists := cache.Exists(ctx, "key") // 返回 false
```

## 🎯 使用场景和最佳实践

### 1. 缓存策略选择

#### 内存缓存适用场景
- 单机应用
- 对性能要求极高的场景
- 数据量不大的应用
- 开发和测试环境

#### Redis 缓存适用场景
- 分布式系统
- 需要持久化的缓存
- 大数据量应用
- 生产环境

#### 空缓存适用场景
- 单元测试
- 需要禁用缓存的环境
- 性能基准测试

### 2. 缓存模式

#### Cache-Aside 模式

```go
func GetUser(id int) (*User, error) {
    var user User
    
    // 先从缓存获取
    err := cache.Get(ctx, fmt.Sprintf("user:%d", id), &user)
    if err == nil {
        return &user, nil
    }
    
    // 缓存未命中，从数据库获取
    user, err = db.GetUser(id)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    cache.Set(ctx, fmt.Sprintf("user:%d", id), user, 10*time.Minute)
    
    return user, nil
}
```

#### Write-Through 模式

```go
func UpdateUser(user *User) error {
    // 先更新数据库
    err := db.UpdateUser(user)
    if err != nil {
        return err
    }
    
    // 同时更新缓存
    return cache.Set(ctx, fmt.Sprintf("user:%d", user.ID), user, 10*time.Minute)
}
```

#### Write-Behind 模式

```go
func UpdateUserAsync(user *User) error {
    // 立即更新缓存
    err := cache.Set(ctx, fmt.Sprintf("user:%d", user.ID), user, 10*time.Minute)
    if err != nil {
        return err
    }
    
    // 异步更新数据库
    go func() {
        db.UpdateUser(user)
    }()
    
    return nil
}
```

### 3. 缓存穿透保护

使用 `GetSet` 方法可以有效防止缓存穿透：

```go
func GetProduct(id int) (*Product, error) {
    var product Product
    
    // 使用 GetSet 防止缓存穿透
    err := cache.GetSet(ctx, fmt.Sprintf("product:%d", id), 30*time.Minute, &product, func(key string, obj any) error {
        // 缓存未命中时的回调函数
        p, err := db.GetProduct(id)
        if err != nil {
            return err
        }
        
        // 将结果赋值给 obj
        productPtr := obj.(*Product)
        *productPtr = *p
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    
    return &product, nil
}
```

## 🧪 测试

### 单元测试示例

```go
package main

import (
    "context"
    "testing"
    "time"
    
    "github.com/muleiwu/go-cache"
    "github.com/stretchr/testify/assert"
)

func TestMemoryCache(t *testing.T) {
    cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
    ctx := context.Background()
    
    // 测试设置和获取
    err := cache.Set(ctx, "test_key", "test_value", 10*time.Minute)
    assert.NoError(t, err)
    
    var result string
    err = cache.Get(ctx, "test_key", &result)
    assert.NoError(t, err)
    assert.Equal(t, "test_value", result)
    
    // 测试键存在性
    assert.True(t, cache.Exists(ctx, "test_key"))
    
    // 测试删除
    err = cache.Del(ctx, "test_key")
    assert.NoError(t, err)
    assert.False(t, cache.Exists(ctx, "test_key"))
}
```

### 基准测试

```go
func BenchmarkMemoryCacheSet(b *testing.B) {
    cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Set(ctx, fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), 10*time.Minute)
    }
}
```

## 📊 性能基准

基于在 Apple M4 Pro 上执行的测试：

### 内存缓存性能

```
BenchmarkMemorySet-14       26337690        45.49 ns/op       0 B/op    0 allocs/op
BenchmarkMemoryGet-14       22087718        54.17 ns/op      16 B/op    1 allocs/op
BenchmarkMemoryExists-14    34378371        35.17 ns/op       0 B/op    0 allocs/op
```

### 序列化器性能

```
Gob 序列化器:
- 编码: ~1052 ns/op
- 解码: ~6199 ns/op

JSON 序列化器:
- 编码: ~161 ns/op
- 解码: ~1436 ns/op
```

## 🚨 注意事项

### 1. 类型安全

- `Get` 和 `GetSet` 方法的 `obj` 参数必须是指针类型
- 确保传入的类型与存储的类型匹配，否则会返回类型不匹配错误

### 2. 序列化限制

- **Gob 序列化**（Redis 缓存默认使用）：
  - 不支持函数、通道等不可序列化的类型
  - 无法序列化未导出的字段（小写字段名）
  - 只能在 Go 应用之间使用
- **JSON 序列化**：
  - 不支持函数、通道和某些复杂类型
  - 某些数值类型可能会丢失精度
  - 可以跨不同语言使用
- 复杂结构体需要确保所有字段都可以被所选序列化器序列化

### 3. 内存管理

- 内存缓存会占用应用程序内存，注意监控内存使用情况
- 设置合适的清理间隔，避免内存泄漏

### 4. 并发安全

- 所有缓存实现都是并发安全的
- 但在回调函数中仍需要注意并发问题

### 5. 错误处理

- Redis 缓存可能会因为网络问题返回错误
- 建议实现重试机制或降级策略

## 🔗 相关链接

- [gsr 接口库](https://github.com/muleiwu/gsr)
- [patrickmn/go-cache](https://github.com/patrickmn/go-cache)
- [redis/go-redis](https://github.com/redis/go-redis)
- [Go encoding/gob](https://pkg.go.dev/encoding/gob)
- [Go encoding/json](https://pkg.go.dev/encoding/json)

## 📝 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 开发环境设置

```bash
# 克隆仓库
git clone https://github.com/muleiwu/go-cache.git
cd go-cache

# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 运行基准测试
go test -bench=. ./...
```

## 📄 附加文档

- [SERIALIZER_GUIDE.md](docs/SERIALIZER_GUIDE.md) - 详细的序列化器使用指南
- [NIL_VALUES.md](docs/NIL_VALUES.md) - Nil 值支持文档
- [GOB_MIGRATION.md](docs/GOB_MIGRATION.md) - 从 msgpack 迁移到 gob 的说明
- [IMPROVEMENTS.md](docs/IMPROVEMENTS.md) - 项目改进记录
- [test/README.md](test/README.md) - 测试文档

## 🆘 常见问题

### Q: 如何在内存缓存和 Redis 缓存之间切换？

A: 由于所有实现都遵循相同的接口，只需要更改初始化代码即可：

```go
// 内存缓存
cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)

// Redis 缓存
cache := go_cache.NewRedis(redisClient)

// 其余代码无需修改
```

### Q: 如何处理缓存中的 nil 值？

A: go-cache 完全支持 nil 值。您可以区分"键不存在"和"键存在但值为 nil"：

```go
// 检查键是否存在
if !cache.Exists(ctx, "key") {
    // 键不存在
} else {
    var value *SomeType
    if err := cache.Get(ctx, "key", &value); err == nil {
        if value == nil {
            // 键存在但值为 nil
        } else {
            // 键存在且有值
        }
    }
}
```

### Q: 如何监控缓存性能？

A: 可以通过包装器模式添加监控功能：

```go
type CacheWithMetrics struct {
    cache gsr.Cacher
}

func (c *CacheWithMetrics) Get(ctx context.Context, key string, obj any) error {
    start := time.Now()
    err := c.cache.Get(ctx, key, obj)
    duration := time.Since(start)
    
    // 记录指标
    metrics.RecordCacheGetDuration(duration)
    if err != nil {
        metrics.RecordCacheMiss()
    } else {
        metrics.RecordCacheHit()
    }
    
    return err
}
```

### Q: 如何选择 Gob 和 JSON 序列化器？

A: 
- **使用 Gob**（默认）适用于需要类型安全的纯 Go 应用
  - 完整的类型安全保证
  - 支持复杂的 Go 类型（接口、指针等）
  - 性能稍慢，但类型匹配更可靠
- **使用 JSON** 适用于跨语言场景或需要调试的情况
  - 编码和解码性能都更快（约 4-6 倍）
  - 人类可读，便于调试
  - 跨语言支持
  - 类型安全性较弱

---

如有其他问题，请提交 [Issue](https://github.com/muleiwu/go-cache/issues) 或联系维护者。