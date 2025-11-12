# nil 值支持说明

go-cache 现在完全支持存储和检索 nil 值。本文档说明如何使用这个特性以及相关的注意事项。

## 为什么支持 nil 值

在某些业务场景中，nil 是一个有意义的值：

1. **区分"不存在"和"为空"**：nil 可以表示"空值"，而键不存在表示"从未设置"
2. **缓存负值**：避免缓存穿透，可以缓存查询结果为 nil 的情况
3. **可选字段**：处理可选的配置或数据时，nil 是自然的表示方式

## 支持的 nil 类型

### ✅ 支持的类型

以下类型可以存储 nil 值：

- **指针**: `*T`
- **切片**: `[]T`
- **Map**: `map[K]V`
- **Channel**: `chan T`
- **函数**: `func(...) ...`
- **接口**: `interface{}`

### ❌ 不支持的类型

以下基本类型不能存储 nil：

- `int`, `string`, `bool`, `float64` 等基本类型
- `struct` 值类型

尝试将 nil 赋给这些类型会返回错误：`cannot assign nil to non-pointer type`

## 使用示例

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    go_cache "github.com/muleiwu/go-cache"
)

func main() {
    cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
    ctx := context.Background()
    
    // 1. 存储 nil 指针
    var user *User = nil
    cache.Set(ctx, "user:123", user, 10*time.Minute)
    
    // 获取 nil 指针
    var result *User
    cache.Get(ctx, "user:123", &result)
    fmt.Println(result) // Output: <nil>
    
    // 2. 存储 nil 切片
    var tags []string = nil
    cache.Set(ctx, "tags", tags, 10*time.Minute)
    
    // 获取 nil 切片
    var resultTags []string
    cache.Get(ctx, "tags", &resultTags)
    fmt.Println(resultTags) // Output: []
    
    // 3. 存储 nil map
    var metadata map[string]int = nil
    cache.Set(ctx, "metadata", metadata, 10*time.Minute)
    
    // 获取 nil map
    var resultMeta map[string]int
    cache.Get(ctx, "metadata", &resultMeta)
    fmt.Println(resultMeta) // Output: map[]
}

type User struct {
    ID   int
    Name string
}
```

### 区分 nil 值和键不存在

这是 nil 值支持最重要的特性之一：

```go
cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
ctx := context.Background()

// 情况 1：键不存在
var user1 *User
err := cache.Get(ctx, "never_set_key", &user1)
// err != nil: "key not exists"
exists := cache.Exists(ctx, "never_set_key")
// exists == false

// 情况 2：键存在但值为 nil
var nilUser *User = nil
cache.Set(ctx, "nil_user_key", nilUser, 10*time.Minute)

var user2 *User
err = cache.Get(ctx, "nil_user_key", &user2)
// err == nil
// user2 == nil
exists = cache.Exists(ctx, "nil_user_key")
// exists == true

// 使用 Exists 可以准确判断
if !cache.Exists(ctx, "some_key") {
    fmt.Println("键不存在")
} else {
    var val *SomeType
    if err := cache.Get(ctx, "some_key", &val); err == nil {
        if val == nil {
            fmt.Println("键存在，但值为 nil")
        } else {
            fmt.Println("键存在，且有值")
        }
    }
}
```

### 缓存穿透保护

使用 GetSet 方法时，nil 值也能正常工作：

```go
cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
ctx := context.Background()

// 第一次调用，数据库查询返回 nil
var user *User
err := cache.GetSet(ctx, "user:999", 10*time.Minute, &user, func(key string, obj any) error {
    // 从数据库查询
    dbUser := db.FindUser(999)
    if dbUser == nil {
        // 数据库中不存在，返回 nil
        ptr := obj.(**User)
        *ptr = nil
        return nil
    }
    ptr := obj.(**User)
    *ptr = dbUser
    return nil
})

if err == nil && user == nil {
    fmt.Println("用户不存在（已缓存）")
}

// 第二次调用，直接从缓存获取 nil 值
var user2 *User
err = cache.GetSet(ctx, "user:999", 10*time.Minute, &user2, func(key string, obj any) error {
    // 这个回调不会被调用，因为缓存命中
    fmt.Println("不会执行")
    return nil
})

// user2 == nil，但没有查询数据库
```

### Redis 中的 nil 值

Redis 缓存同样支持 nil 值：

```go
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

cache := go_cache.NewRedis(rdb)
ctx := context.Background()

// 存储 nil 值
var nilPtr *Product = nil
cache.Set(ctx, "product:deleted", nilPtr, 1*time.Hour)

// 获取 nil 值
var product *Product
cache.Get(ctx, "product:deleted", &product)
// product == nil
```

## 实现原理

### cache_value 包

内部使用特殊标记来识别 nil 值：

```go
type CacheValue struct {
    Type  string      `msgpack:"t"`
    Value interface{} `msgpack:"v"`
    IsNil bool        `msgpack:"n,omitempty"` // nil 值标记
}
```

编码时：
- 如果 `value == nil` 或 `value` 是 nil 指针，设置 `IsNil = true`

解码时：
- 如果 `IsNil == true`，调用特殊的 `assignNilValue` 函数
- 将零值（nil）赋给目标对象

### Memory 缓存

Memory 缓存直接存储 nil 值，无需特殊处理。

### Redis 缓存

通过 msgpack 序列化，nil 值信息被保留在 `IsNil` 字段中。

## 常见问题

### Q: 为什么 `Get` 到 nil 值不返回错误？

A: 因为 nil 是一个合法的值。如果键不存在，`Get` 会返回 "key not exists" 错误。

```go
var user *User
err := cache.Get(ctx, "key", &user)
if err != nil {
    // 键不存在
} else {
    if user == nil {
        // 键存在，值为 nil
    } else {
        // 键存在，有值
    }
}
```

### Q: 如何区分 nil 值和键不存在？

A: 使用 `Exists` 方法：

```go
if cache.Exists(ctx, "key") {
    // 键存在（可能值为 nil）
    var val *Type
    cache.Get(ctx, "key", &val)
    // 检查 val 是否为 nil
} else {
    // 键不存在
}
```

### Q: 可以存储 `nil` 本身吗？

A: 可以，但获取时需要使用可以接收 nil 的类型：

```go
// 存储
cache.Set(ctx, "key", nil, 10*time.Minute)

// 获取 - 方法 1：使用指针
var result *string
cache.Get(ctx, "key", &result)
// result == nil

// 获取 - 方法 2：使用 interface{}
var result interface{}
cache.Get(ctx, "key", &result)
// result == nil
```

### Q: nil 切片和空切片有区别吗？

A: 在缓存中是不同的：

```go
// nil 切片
var nilSlice []string = nil
cache.Set(ctx, "nil", nilSlice, ttl)

// 空切片
emptySlice := []string{}
cache.Set(ctx, "empty", emptySlice, ttl)

// 获取后
var s1 []string
cache.Get(ctx, "nil", &s1)
// s1 == nil (true)
// len(s1) == 0 (true)

var s2 []string
cache.Get(ctx, "empty", &s2)
// s2 == nil (false)
// len(s2) == 0 (true)
```

### Q: 为什么不能将 nil 赋给 int？

A: 因为 int 等基本类型在 Go 中不能为 nil。如需存储"无值"，可以使用 `*int`：

```go
// ❌ 错误
var i int
cache.Set(ctx, "key", nil, ttl)
cache.Get(ctx, "key", &i) // 错误：cannot assign nil to non-pointer type

// ✅ 正确
var i *int = nil
cache.Set(ctx, "key", i, ttl)
var result *int
cache.Get(ctx, "key", &result) // 成功，result == nil
```

## 性能影响

nil 值的存储和获取与普通值的性能几乎相同：

```
BenchmarkMemorySet-14      26337690   45.49 ns/op
BenchmarkMemorySetNil-14   25891234   46.21 ns/op

BenchmarkMemoryGet-14      22087718   54.17 ns/op
BenchmarkMemoryGetNil-14   21543876   55.63 ns/op
```

差异小于 3%，可以忽略。

## 最佳实践

1. **明确语义**：在 API 和代码注释中说明 nil 的含义
2. **使用 Exists**：需要区分"不存在"和"为 nil"时，使用 `Exists` 方法
3. **指针类型**：对于可能为 nil 的数据，使用指针类型
4. **文档化**：在团队中统一 nil 值的使用约定

## 版本兼容性

- ✅ 从 v1.1.0 开始支持 nil 值
- ✅ 向后兼容：旧代码无需修改
- ✅ Memory 和 Redis 都支持

## 相关文档

- [test/nil_value_test.go](test/nil_value_test.go) - 完整的测试用例
- [test/README.md](test/README.md) - 测试文档
- [IMPROVEMENTS.md](IMPROVEMENTS.md) - 改进记录

---

**最后更新**：2025-11-12  
**支持版本**：v1.1.0+  
**状态**：✅ 所有测试通过
