# 测试用例说明

本目录包含 go-cache 项目的所有测试用例。

## 测试文件结构

```
test/
├── cache_value_test.go        # cache_value 包的测试用例
├── memory_test.go             # Memory 缓存实现的测试用例
├── none_test.go               # None 缓存实现的测试用例
├── redis_integration_test.go  # Redis 缓存集成测试用例
├── nil_value_test.go          # nil值支持的测试用例
└── README.md                  # 本文件
```

## 运行测试

### 运行所有测试

```bash
go test ./test/... -v
```

### 运行特定文件的测试

```bash
# 只测试 cache_value
go test ./test/ -run TestEncode -v

# 只测试 Memory 缓存
go test ./test/ -run TestMemory -v

# 只测试 None 缓存
go test ./test/ -run TestNone -v

# 只测试 Redis 缓存（需要Redis运行）
go test ./test/ -run TestRedis -v
```

### 运行基准测试

```bash
go test ./test/... -bench=. -v
```

### 运行单个测试

```bash
go test ./test/ -run '^TestMemorySetAndGet$' -v
```

## 测试覆盖范围

### cache_value_test.go

测试 `cache_value` 包的序列化和反序列化功能：

- **TestEncode**: 测试各种类型的编码（字符串、整数、结构体、指针、切片、map）
- **TestDecode**: 测试字符串的解码
- **TestDecodeTypeMismatch**: 测试类型不匹配时的错误处理
- **TestDecodeWithNilObj**: 测试传入nil对象时的错误处理
- **TestDecodeWithNonPointer**: 测试传入非指针对象时的错误处理

**注意事项**：
- cache_value 包使用可插拔的序列化系统，默认使用 Gob 序列化器
- Gob 序列化器完整支持所有 Go 类型（包括复杂结构体、切片、map 等）
- Redis 缓存也支持 JSON 序列化器，可通过 `WithRedisSerializer` 选项配置

### memory_test.go

测试 Memory 缓存的所有功能：

**功能测试**：
- **TestMemorySetAndGet**: 测试设置和获取各种类型的缓存数据
- **TestMemoryExists**: 测试检查键是否存在
- **TestMemoryDel**: 测试删除缓存键
- **TestMemoryGetSet**: 测试 GetSet 方法（缓存不存在时执行回调）
- **TestMemoryExpiresIn**: 测试设置相对过期时间（已跳过，见下方说明）
- **TestMemoryExpiresAt**: 测试设置绝对过期时间（已跳过，见下方说明）
- **TestMemoryGetNonExistentKey**: 测试获取不存在的键
- **TestMemoryTypeMismatch**: 测试类型不匹配时的错误处理
- **TestMemoryWithZeroTTL**: 测试 TTL 为 0 或负数的情况
- **TestMemoryConcurrentAccess**: 测试并发访问的安全性

**基准测试**：
- **BenchmarkMemorySet**: 设置操作的性能测试
- **BenchmarkMemoryGet**: 获取操作的性能测试
- **BenchmarkMemoryExists**: 检查存在操作的性能测试

### none_test.go

测试 None 缓存（空操作缓存）的所有功能：

**功能测试**：
- **TestNoneExists**: 验证 Exists 总是返回 false
- **TestNoneGet**: 验证 Get 总是返回 "not implemented" 错误
- **TestNoneSet**: 验证 Set 操作成功但不存储数据
- **TestNoneGetSet**: 验证 GetSet 总是返回 "not implemented" 错误
- **TestNoneDel**: 验证 Del 操作总是成功
- **TestNoneExpiresAt**: 验证 ExpiresAt 操作总是成功
- **TestNoneExpiresIn**: 验证 ExpiresIn 操作总是成功
- **TestNewCacheNone**: 测试别名构造函数
- **TestNoneMultipleOperations**: 测试多个操作的组合
- **TestNoneWithDifferentContexts**: 测试使用不同 context 的行为

**基准测试**：
- **BenchmarkNoneSet**: Set 操作的性能测试
- **BenchmarkNoneGet**: Get 操作的性能测试
- **BenchmarkNoneExists**: Exists 操作的性能测试
- **BenchmarkNoneDel**: Del 操作的性能测试

### redis_integration_test.go

测试 Redis 缓存的所有功能（集成测试）：

**功能测试**：
- **TestRedisSetAndGet**: 测试设置和获取各种类型的缓存数据
- **TestRedisExists**: 测试检查键是否存在
- **TestRedisDel**: 测试删除缓存键
- **TestRedisGetSet**: 测试 GetSet 方法（缓存不存在时执行回调）
- **TestRedisExpiresIn**: 测试设置相对过期时间
- **TestRedisExpiresAt**: 测试设置绝对过期时间
- **TestRedisGetNonExistentKey**: 测试获取不存在的键
- **TestRedisWithZeroTTL**: 测试 TTL 为 0 的情况（永不过期）
- **TestRedisWithNegativeTTL**: 测试负数 TTL 的情况
- **TestRedisConcurrentAccess**: 测试并发访问的安全性
- **TestRedisComplexStruct**: 测试复杂结构体的存储（注意限制）
- **TestRedisConnectionFailure**: 测试连接失败的情况

**基准测试**：
- **BenchmarkRedisSet**: 设置操作的性能测试
- **BenchmarkRedisGet**: 获取操作的性能测试
- **BenchmarkRedisExists**: 检查存在操作的性能测试

**环境要求**：
- 需要本地或远程Redis服务器运行
- 默认连接 `localhost:6379`
- 使用DB 15进行测试，避免影响其他数据
- 可通过环境变量 `REDIS_ADDR` 指定地址
- 可通过环境变量 `SKIP_REDIS_TESTS=1` 跳过测试

**已知限制**：
- Redis的过期时间最小单位是1秒，小于1秒的值会被截断为1秒
- Gob 序列化器（默认）只能在 Go 应用之间使用，如需跨语言支持请使用 JSON 序列化器
- JSON 序列化器不支持某些复杂类型（如 channel、func 等）

### nil_value_test.go

测试对nil值的支持：

**功能测试**：
- **TestCacheValueEncodeDecodeNil**: 测试cache_value对nil值的编解码
- **TestMemorySetGetNil**: 测试Memory缓存对nil值的支持
  - nil指针
  - nil切片
  - nil map
  - nil interface
  - 非指针类型的错误处理
- **TestRedisSetGetNil**: 测试Redis缓存对nil值的支持
- **TestMemoryGetSetWithNil**: 测试GetSet方法对nil值的支持
- **TestNilVsKeyNotExists**: 测试区分nil值和键不存在

**基准测试**：
- **BenchmarkMemorySetNil**: 存储nil值的性能测试
- **BenchmarkMemoryGetNil**: 获取nil值的性能测试

**重要特性**：
- 支持存储和获取nil指针、nil切片、nil map等
- 可以区分nil值和键不存在（通过`Exists`方法）
- nil值无法赋给非指针类型，会返回错误

## 测试数据结构

测试中使用的公共数据结构：

```go
type TestUser struct {
    ID   int
    Name string
    Age  int
}
```

## 测试覆盖率

查看测试覆盖率：

```bash
go test ./test/... -cover
```

生成详细的覆盖率报告：

```bash
go test ./test/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 注意事项

1. **并发测试**：`TestMemoryConcurrentAccess` 和 `TestRedisConcurrentAccess` 测试了并发安全性，确保缓存在多个 goroutine 同时访问时不会出现数据竞争
2. **基准测试**：运行基准测试时，建议多次运行以获得更准确的结果
3. **过期时间测试**：
   - Memory测试使用100-200毫秒的过期时间
   - Redis测试使用1-1.5秒的过期时间（Redis最小单位是1秒）
4. **Redis集成测试**：
   - 需要Redis服务器运行才能执行
   - 如果Redis不可用，测试会自动跳过
   - 可以通过 `SKIP_REDIS_TESTS=1` 强制跳过

## 贡献测试

添加新的测试用例时，请遵循以下规范：

1. 测试函数名称使用 `Test<功能名称>` 格式
2. 使用表驱动测试（table-driven tests）处理多个测试用例
3. 每个测试应该独立，不依赖其他测试的执行顺序
4. 测试失败时提供清晰的错误信息
5. 使用中文注释说明测试目的和预期行为
