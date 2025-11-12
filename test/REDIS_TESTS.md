# Redis 集成测试说明

本文档说明如何运行和理解Redis集成测试。

## 环境要求

### Redis服务器

集成测试需要一个运行中的Redis服务器：

```bash
# 检查Redis是否运行
redis-cli ping
# 应该返回: PONG

# 如果未运行，启动Redis
redis-server
```

### 环境变量

可以通过以下环境变量配置测试：

```bash
# 指定Redis地址（默认：localhost:6379）
export REDIS_ADDR="localhost:6379"

# 跳过Redis集成测试
export SKIP_REDIS_TESTS=1
```

## 运行测试

### 运行所有Redis测试

```bash
go test ./test/ -run TestRedis -v
```

### 运行特定测试

```bash
# 只测试GetSet功能
go test ./test/ -run TestRedisGetSet -v

# 测试过期时间
go test ./test/ -run TestRedisExpires -v

# 测试并发访问
go test ./test/ -run TestRedisConcurrent -v
```

### 运行基准测试

```bash
go test ./test/ -bench=BenchmarkRedis -benchmem
```

## 测试覆盖

### 功能测试（12个）

1. **TestRedisSetAndGet** - 基本的设置和获取操作
   - 字符串
   - 浮点数
   - 布尔值

2. **TestRedisExists** - 检查键是否存在

3. **TestRedisDel** - 删除键

4. **TestRedisGetSet** - 缓存穿透保护
   - 缓存未命中时调用回调
   - 缓存命中时直接返回

5. **TestRedisExpiresIn** - 设置相对过期时间

6. **TestRedisExpiresAt** - 设置绝对过期时间

7. **TestRedisGetNonExistentKey** - 获取不存在的键

8. **TestRedisWithZeroTTL** - TTL为0（永不过期）

9. **TestRedisWithNegativeTTL** - 负数TTL

10. **TestRedisConcurrentAccess** - 并发安全性测试
    - 10个goroutine
    - 每个执行50次读写操作

11. **TestRedisComplexStruct** - 复杂结构体存储
    - 使用 Gob 序列化器，完整支持复杂结构体

12. **TestRedisConnectionFailure** - 连接失败处理

### 基准测试（3个）

```
BenchmarkRedisSet-14       13887    84919 ns/op    528 B/op    19 allocs/op
BenchmarkRedisGet-14       13731    83440 ns/op    472 B/op    20 allocs/op
BenchmarkRedisExists-14    14876    82544 ns/op    264 B/op    12 allocs/op
```

**性能对比（Apple M4 Pro）**：

| 操作 | Memory | Redis | 差距 |
|------|--------|-------|------|
| Set | 45 ns/op | 84,919 ns/op | ~1,887x |
| Get | 54 ns/op | 83,440 ns/op | ~1,545x |
| Exists | 35 ns/op | 82,544 ns/op | ~2,358x |

Redis操作需要网络往返，因此比内存缓存慢1000-2000倍是正常的。

## 测试数据库

测试使用Redis的DB 15，避免影响其他数据：

```go
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    DB:   15,  // 测试专用数据库
})
```

每个测试结束后会自动清理数据：
- 测试前：`FLUSHDB` 清空DB 15
- 测试后：`FLUSHDB` 再次清空

## 已知限制和注意事项

### 1. 过期时间精度

Redis的过期时间最小单位是1秒：

```go
// ❌ 小于1秒会被截断为1秒
cache.ExpiresIn(ctx, key, 100*time.Millisecond)

// ✅ 使用1秒或更长
cache.ExpiresIn(ctx, key, 1*time.Second)
```

### 2. 序列化器选择

Redis 缓存支持可插拔的序列化系统：

**Gob 序列化器**（默认）：
```go
// 默认使用 Gob
cache := go_cache.NewRedis(rdb)

// 完整支持复杂结构体
type Person struct {
    Name string
    Age  int
}

cache.Set(ctx, "key", Person{Name: "张三", Age: 30}, ttl)
var p Person
cache.Get(ctx, "key", &p)  // ✅ 完美工作
```

**JSON 序列化器**（可选）：
```go
// 使用 JSON 序列化器
cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))

// 优点：跨语言支持、人类可读
// 缺点：性能略低、类型安全性较弱
```

### 3. Redis警告信息

测试时可能看到以下警告（可以忽略）：

```
redis: auto mode fallback: maintnotifications disabled due to handshake error
```

这是Redis客户端的兼容性问题，不影响测试功能。

## 跳过测试

如果Redis不可用或想跳过集成测试：

```bash
# 方法1：设置环境变量
export SKIP_REDIS_TESTS=1
go test ./test/... -v

# 方法2：只运行其他测试
go test ./test/ -run "Test(Memory|None|Encode)" -v
```

如果Redis不可用，测试会自动跳过并显示：

```
--- SKIP: TestRedisSetAndGet (0.00s)
    redis_integration_test.go:35: Redis不可用，跳过集成测试: <error>
```

## 故障排查

### Redis连接失败

**问题**：测试输出 "Redis不可用，跳过集成测试"

**解决方案**：
1. 确认Redis正在运行：`redis-cli ping`
2. 检查端口是否正确：`lsof -i :6379`
3. 检查防火墙设置

### 测试超时

**问题**：测试运行时间过长

**原因**：
- Redis服务器响应慢
- 网络延迟高

**解决方案**：
- 使用本地Redis实例
- 增加测试超时时间

### 数据残留

**问题**：测试后Redis中有残留数据

**说明**：测试使用DB 15，不影响其他数据库

**清理**：
```bash
redis-cli -n 15 FLUSHDB
```

## 持续集成（CI）

在CI环境中运行测试：

### GitHub Actions

```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      - run: go test ./test/... -v
```

### GitLab CI

```yaml
test:
  image: golang:1.25
  services:
    - redis:7-alpine
  variables:
    REDIS_ADDR: "redis:6379"
  script:
    - go test ./test/... -v
```

## 最佳实践

1. **本地开发**：始终运行完整测试套件
2. **CI/CD**：包含Redis集成测试
3. **生产部署前**：运行基准测试验证性能
4. **监控**：关注Redis操作的延迟和错误率

## 相关文档

- [test/README.md](README.md) - 完整测试文档
- [../IMPROVEMENTS.md](../IMPROVEMENTS.md) - 改进记录
- [Redis官方文档](https://redis.io/documentation)

---

**最后更新**：2025-11-12  
**Redis版本**：7.x  
**测试状态**：✅ 所有测试通过（12个功能测试 + 3个基准测试）
