# go-cache

[中文文档](README.zh-CN.md) | English

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/muleiwu/go-cache)](https://goreportcard.com/report/github.com/muleiwu/go-cache)

go-cache is a unified interface Go cache library that provides multiple cache implementations, including memory cache, Redis cache, and null cache. The library implements the `gsr.Cacher` interface, supporting seamless switching between different cache implementations to provide flexible caching solutions for applications.

## 🚀 Features

- **Unified Interface**: All cache implementations follow the `gsr.Cacher` interface, making it easy to switch and test
- **Multiple Implementations**: Support for memory cache, Redis cache, and null cache implementations
- **Type Safety**: Uses reflection to ensure type-safe value assignment
- **TTL Support**: Supports setting time-to-live for keys
- **Cache Penetration Protection**: Provides `GetSet` method to prevent cache penetration
- **Pluggable Serialization**: Supports Gob (default) and JSON serializers, extensible with custom serializers
- **Complete Nil Value Support**: Full support for nil pointers, nil slices, and nil maps
- **Expiration Management**: Supports setting specific expiration times or relative TTL
- **Context Support**: All operations support context.Context

## 📦 Installation

Install go-cache using go get:

```bash
go get github.com/muleiwu/go-cache
```

## 🏗️ Architecture Overview

```
go-cache/
├── memory.go          # Memory cache implementation
├── redis.go           # Redis cache implementation
├── none.go            # Null cache implementation
├── serializer/        # Serializer package
│   ├── serializer.go  # Serializer interface
│   ├── gob.go         # Gob serializer (default)
│   └── json.go        # JSON serializer
└── cache_value/       # Cache value processing
    └── cache_value.go # Serialization/deserialization logic
```

### Core Components

1. **Cache Interface** (`gsr.Cacher`): Defines unified cache operation interface
2. **Memory Cache** (`Memory`): Memory-based cache implementation, suitable for single-machine applications
3. **Redis Cache** (`Redis`): Redis-based distributed cache implementation
4. **Null Cache** (`None`): No-op implementation for testing or disabling cache scenarios
5. **Serialization System** (`serializer`): Pluggable serializers supporting Gob and JSON
6. **Value Processing** (`cache_value`): Handles serialization and deserialization of cache values

## 🚀 Quick Start

### Memory Cache Example

```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/muleiwu/go-cache"
)

func main() {
	// Create memory cache with default expiration 5 minutes, cleanup interval 10 minutes
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()
	
	// Set cache
	err := cache.Set(ctx, "user:123", &User{ID: 123, Name: "John Doe"}, 10*time.Minute)
	if err != nil {
		panic(err)
	}
	
	// Get cache
	var user User
	err = cache.Get(ctx, "user:123", &user)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("User: %+v\n", user)
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

### Redis Cache Example

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
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password
		DB:       0,  // Default DB
	})
	
	// Create Redis cache with default Gob serializer
	cache := go_cache.NewRedis(rdb)
	ctx := context.Background()
	
	// Set cache
	err := cache.Set(ctx, "product:456", &Product{ID: 456, Name: "Product A", Price: 99.99}, 30*time.Minute)
	if err != nil {
		panic(err)
	}
	
	// Get cache
	var product Product
	err = cache.Get(ctx, "product:456", &product)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("Product: %+v\n", product)
}

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
```

### Cache Penetration Protection Example

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
	
	// Use GetSet to prevent cache penetration
	var user User
	err := cache.GetSet(ctx, "user:789", 10*time.Minute, &user, func(key string, obj any) error {
		// Fetch data from database when cache miss occurs
		fmt.Println("Fetching user data from database...")
		user := obj.(*User)
		user.ID = 789
		user.Name = "Jane Smith"
		return nil
	})
	
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("User: %+v\n", user)
	
	// Second call will get directly from cache
	var user2 User
	err = cache.GetSet(ctx, "user:789", 10*time.Minute, &user2, func(key string, obj any) error {
		fmt.Println("This callback won't be called because cache already exists")
		return nil
	})
	
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("User2: %+v\n", user2)
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

### Pluggable Serializer Examples

#### Using JSON Serializer (Cross-language, Human-readable)

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
	
	// Use JSON serializer
	cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))
	ctx := context.Background()
	
	// JSON-serialized data is human-readable in Redis
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	user := User{ID: 1, Name: "Alice"}
	cache.Set(ctx, "user:1", user, 10*time.Minute)
	
	// You can view directly in Redis CLI:
	// redis-cli GET user:1
	// {"is_nil":false,"value":{"id":1,"name":"Alice"}}
}
```

#### Using Gob Serializer (Default, Type-safe)

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
	
	// Explicitly specify Gob serializer (Gob is already default)
	cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewGob()))
	ctx := context.Background()
	
	// Gob perfectly handles complex structs and nil values
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
	
	// Type is fully restored when retrieved, including pointers
	var cached Config
	cache.Get(ctx, "config", &cached)
	// cached.Version points to the correct string
}
```

#### Serializer Comparison

| Feature | Gob | JSON |
|---------|-----|------|
| **Type Safety** | ✅ Complete | ⚠️ Partial |
| **Performance (Encode)** | Medium (~1052 ns/op) | Fast (~161 ns/op) |
| **Performance (Decode)** | Slow (~6199 ns/op) | Medium (~1436 ns/op) |
| **Cross-language** | ❌ Go only | ✅ All languages |
| **Human-readable** | ❌ Binary | ✅ Text |
| **Complex Structs** | ✅ Perfect | ✅ Good |
| **Nil Value Support** | ✅ Complete | ✅ Complete |
| **Debug-friendly** | ⚠️ Difficult | ✅ Easy |

**Recommendations**:
- **Use Gob by default** - Suitable for pure Go applications, type-safe
- **Use JSON for cross-language** - Suitable for microservices architecture
- **Use JSON for debugging** - Easy to view data in Redis

For detailed usage guide, see [SERIALIZER_GUIDE.md](docs/SERIALIZER_GUIDE.md)

### Nil Value Support

go-cache provides complete support for nil values, allowing you to distinguish between "key doesn't exist" and "key exists but value is nil":

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
	
	// Store nil pointer
	var user *User = nil
	err := cache.Set(ctx, "user:123", user, 10*time.Minute)
	if err != nil {
		panic(err)
	}
	
	// Retrieve nil pointer
	var result *User
	err = cache.Get(ctx, "user:123", &result)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("User is nil: %v\n", result == nil) // Output: User is nil: true
	
	// Check if key exists
	exists := cache.Exists(ctx, "user:123")
	fmt.Printf("Key exists: %v\n", exists) // Output: Key exists: true
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

For detailed nil value usage, see [NIL_VALUES.md](docs/NIL_VALUES.md)

## 📚 API Documentation

### Interface Definition

go-cache implements the `gsr.Cacher` interface, which defines the following methods:

```go
type Cacher interface {
    // Exists checks if a key exists
    Exists(ctx context.Context, key string) bool
    
    // Get gets cache value and deserializes the result into obj
    Get(ctx context.Context, key string, obj any) error
    
    // Set sets cache value, ttl is the time-to-live
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    
    // GetSet gets cache value, if not exists, gets and sets through callback function
    GetSet(ctx context.Context, key string, ttl time.Duration, obj any, funCallback CacheCallback) error
    
    // Del deletes cache key
    Del(ctx context.Context, key string) error
    
    // ExpiresAt sets key to expire at specific time
    ExpiresAt(ctx context.Context, key string, expiresAt time.Time) error
    
    // ExpiresIn sets key to expire after specified time
    ExpiresIn(ctx context.Context, key string, ttl time.Duration) error
}

type CacheCallback func(key string, obj any) error
```

### Memory Cache (Memory)

#### Constructor

```go
func NewMemory(defaultExpiration, cleanupInterval time.Duration) *Memory
```

- `defaultExpiration`: Default expiration time
- `cleanupInterval`: Time interval for cleaning up expired items

#### Features

- Memory-based cache implementation
- Uses `github.com/patrickmn/go-cache` as underlying storage
- Supports automatic cleanup of expired items
- Thread-safe

### Redis Cache (Redis)

#### Constructor

```go
func NewRedis(conn *redis.Client, opts ...RedisOption) *Redis
```

- `conn`: Redis client connection
- `opts`: Optional configuration (e.g., WithRedisSerializer)

#### Options

```go
// Use custom serializer
cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))
```

#### Features

- Redis-based distributed cache
- **Pluggable Serialization**: Uses Gob by default, switchable to JSON or custom serializers
- **Complete Type Safety**: Gob serializer guarantees type safety
- **Nil Value Support**: Full support for nil pointers, nil slices, and nil maps
- Supports all Redis data types
- Suitable for distributed systems

#### Usage Example

```go
// Create Redis client
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// Create Redis cache with default Gob serializer
cache := go_cache.NewRedis(rdb)

// Or create with JSON serializer
cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))

// Usage is the same
err := cache.Set(ctx, "key", "value", 10*time.Minute)
var result string
err = cache.Get(ctx, "key", &result)
```

### Null Cache (None)

#### Constructor

```go
func NewNone() *None
func NewCacheNone() *None  // Alias
```

#### Features

- All operations are no-op or return errors
- Used for testing or disabling cache scenarios
- Doesn't store any data

## 🎯 Use Cases and Best Practices

### 1. Cache Strategy Selection

#### Memory Cache Use Cases
- Single-machine applications
- Performance-critical scenarios
- Small to medium data volumes
- Development and testing environments

#### Redis Cache Use Cases
- Distributed systems
- Persistent cache requirements
- Large data volumes
- Production environments

#### Null Cache Use Cases
- Unit testing
- Cache-disabled environments
- Performance benchmarking

### 2. Cache Patterns

#### Cache-Aside Pattern

```go
func GetUser(id int) (*User, error) {
    var user User
    
    // Try to get from cache first
    err := cache.Get(ctx, fmt.Sprintf("user:%d", id), &user)
    if err == nil {
        return &user, nil
    }
    
    // Cache miss, fetch from database
    user, err = db.GetUser(id)
    if err != nil {
        return nil, err
    }
    
    // Write to cache
    cache.Set(ctx, fmt.Sprintf("user:%d", id), user, 10*time.Minute)
    
    return user, nil
}
```

#### Write-Through Pattern

```go
func UpdateUser(user *User) error {
    // Update database first
    err := db.UpdateUser(user)
    if err != nil {
        return err
    }
    
    // Update cache simultaneously
    return cache.Set(ctx, fmt.Sprintf("user:%d", user.ID), user, 10*time.Minute)
}
```

#### Write-Behind Pattern

```go
func UpdateUserAsync(user *User) error {
    // Update cache immediately
    err := cache.Set(ctx, fmt.Sprintf("user:%d", user.ID), user, 10*time.Minute)
    if err != nil {
        return err
    }
    
    // Update database asynchronously
    go func() {
        db.UpdateUser(user)
    }()
    
    return nil
}
```

### 3. Cache Penetration Protection

Use the `GetSet` method to effectively prevent cache penetration:

```go
func GetProduct(id int) (*Product, error) {
    var product Product
    
    // Use GetSet to prevent cache penetration
    err := cache.GetSet(ctx, fmt.Sprintf("product:%d", id), 30*time.Minute, &product, func(key string, obj any) error {
        // Callback function when cache miss occurs
        p, err := db.GetProduct(id)
        if err != nil {
            return err
        }
        
        // Assign result to obj
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

## 🧪 Testing

### Unit Test Example

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
    
    // Test set and get
    err := cache.Set(ctx, "test_key", "test_value", 10*time.Minute)
    assert.NoError(t, err)
    
    var result string
    err = cache.Get(ctx, "test_key", &result)
    assert.NoError(t, err)
    assert.Equal(t, "test_value", result)
    
    // Test key existence
    assert.True(t, cache.Exists(ctx, "test_key"))
    
    // Test delete
    err = cache.Del(ctx, "test_key")
    assert.NoError(t, err)
    assert.False(t, cache.Exists(ctx, "test_key"))
}
```

### Benchmark Tests

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

## 📊 Performance Benchmarks

Based on tests performed on Apple M4 Pro:

### Memory Cache Performance

```
BenchmarkMemorySet-14       26337690        45.49 ns/op       0 B/op    0 allocs/op
BenchmarkMemoryGet-14       22087718        54.17 ns/op      16 B/op    1 allocs/op
BenchmarkMemoryExists-14    34378371        35.17 ns/op       0 B/op    0 allocs/op
```

### Serializer Performance

```
Gob Serializer:
- Encode: ~1052 ns/op
- Decode: ~6199 ns/op

JSON Serializer:
- Encode: ~161 ns/op
- Decode: ~1436 ns/op
```

## 🚨 Important Notes

### 1. Type Safety

- The `obj` parameter in `Get` and `GetSet` methods must be a pointer type
- Ensure the passed type matches the stored type, otherwise a type mismatch error will be returned

### 2. Serialization Limitations

- **Gob serialization** (default for Redis cache):
  - Doesn't support non-serializable types like functions and channels
  - Cannot serialize unexported fields (lowercase field names)
  - Only works between Go applications
- **JSON serialization**:
  - Doesn't support functions, channels, and complex types
  - May lose precision with some numeric types
  - Works across different languages
- Complex structs must ensure all fields are serializable by the chosen serializer

### 3. Memory Management

- Memory cache consumes application memory, monitor memory usage
- Set appropriate cleanup intervals to avoid memory leaks

### 4. Concurrency Safety

- All cache implementations are thread-safe
- However, still need to pay attention to concurrency issues in callback functions

### 5. Error Handling

- Redis cache may return errors due to network issues
- It's recommended to implement retry mechanisms or fallback strategies

## 🔗 Related Links

- [gsr Interface Library](https://github.com/muleiwu/gsr)
- [patrickmn/go-cache](https://github.com/patrickmn/go-cache)
- [redis/go-redis](https://github.com/redis/go-redis)
- [Go encoding/gob](https://pkg.go.dev/encoding/gob)
- [Go encoding/json](https://pkg.go.dev/encoding/json)

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Environment Setup

```bash
# Clone repository
git clone https://github.com/muleiwu/go-cache.git
cd go-cache

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...
```

## 📄 Additional Documentation

- [SERIALIZER_GUIDE.md](docs/SERIALIZER_GUIDE.md) - Detailed serializer usage guide
- [NIL_VALUES.md](docs/NIL_VALUES.md) - Nil value support documentation
- [GOB_MIGRATION.md](docs/GOB_MIGRATION.md) - Migration from msgpack to gob
- [IMPROVEMENTS.md](docs/IMPROVEMENTS.md) - Project improvement records
- [test/README.md](test/README.md) - Test documentation

## 🆘 Frequently Asked Questions

### Q: How to switch between memory cache and Redis cache?

A: Since all implementations follow the same interface, you only need to change the initialization code:

```go
// Memory cache
cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)

// Redis cache
cache := go_cache.NewRedis(redisClient)

// The rest of the code remains unchanged
```

### Q: How to handle nil values in cache?

A: go-cache fully supports nil values. You can distinguish between "key doesn't exist" and "key exists but value is nil":

```go
// Check if key exists
if !cache.Exists(ctx, "key") {
    // Key doesn't exist
} else {
    var value *SomeType
    if err := cache.Get(ctx, "key", &value); err == nil {
        if value == nil {
            // Key exists but value is nil
        } else {
            // Key exists and has value
        }
    }
}
```

### Q: How to monitor cache performance?

A: You can add monitoring functionality through the wrapper pattern:

```go
type CacheWithMetrics struct {
    cache gsr.Cacher
}

func (c *CacheWithMetrics) Get(ctx context.Context, key string, obj any) error {
    start := time.Now()
    err := c.cache.Get(ctx, key, obj)
    duration := time.Since(start)
    
    // Record metrics
    metrics.RecordCacheGetDuration(duration)
    if err != nil {
        metrics.RecordCacheMiss()
    } else {
        metrics.RecordCacheHit()
    }
    
    return err
}
```

### Q: How to choose between Gob and JSON serializers?

A: 
- **Use Gob** (default) for pure Go applications requiring type safety
  - Complete type safety guarantee
  - Supports complex Go types (interfaces, pointers, etc.)
  - Slightly slower, but more reliable type matching
- **Use JSON** for cross-language scenarios or when debugging is needed
  - Faster encoding and decoding performance (4-6x faster)
  - Human-readable, easier to debug
  - Cross-language support
  - Weaker type safety

---

For other questions, please submit an [Issue](https://github.com/muleiwu/go-cache/issues) or contact the maintainer.
