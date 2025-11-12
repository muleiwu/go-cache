# go-cache

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
- **Serialization Support**: Uses Go's native gob encoding for efficient and type-safe serialization
- **Nil Value Support**: Supports storing and retrieving nil pointers, slices, and maps
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
└── cache_value/       # Cache value processing
    └── cache_value.go # Serialization/deserialization logic
```

### Core Components

1. **Cache Interface** (`gsr.Cacher`): Defines unified cache operation interface
2. **Memory Cache** (`Memory`): Memory-based cache implementation, suitable for single-machine applications
3. **Redis Cache** (`Redis`): Redis-based distributed cache implementation
4. **Null Cache** (`None`): No-op implementation for testing or disabling cache scenarios
5. **Value Processing** (`cache_value`): Handles serialization and deserialization of cache values

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
	
	// Create Redis cache
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

#### Usage Example

```go
// Create memory cache
cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)

// Set cache
err := cache.Set(ctx, "key", "value", 10*time.Minute)

// Get cache
var result string
err = cache.Get(ctx, "key", &result)

// Check if key exists
exists := cache.Exists(ctx, "key")

// Delete key
err = cache.Del(ctx, "key")

// Set expiration time
err = cache.ExpiresIn(ctx, "key", 5*time.Minute)
err = cache.ExpiresAt(ctx, "key", time.Now().Add(5*time.Minute))
```

### Redis Cache (Redis)

#### Constructor

```go
func NewRedis(conn *redis.Client) *Redis
```

- `conn`: Redis client connection

#### Features

- Redis-based distributed cache
- Uses Go's native gob encoding for serialization
- Full type safety with automatic type registration
- Supports complex structs, pointers, slices, maps, and nil values
- Suitable for distributed systems

#### Usage Example

```go
// Create Redis client
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// Create Redis cache
cache := go_cache.NewRedis(rdb)

// Usage is the same as memory cache
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

- All operations are no-ops or return errors
- Used for testing or disabling cache scenarios
- Does not store any data

#### Usage Example

```go
// Create null cache
cache := go_cache.NewNone()

// Set operation always succeeds but doesn't store data
err := cache.Set(ctx, "key", "value", 10*time.Minute) // returns nil

// Get operation always returns error
var result string
err = cache.Get(ctx, "key", &result) // returns "not implemented" error

// Exists always returns false
exists := cache.Exists(ctx, "key") // returns false
```

## 🎯 Use Cases and Best Practices

### 1. Cache Strategy Selection

#### Memory Cache Use Cases
- Single-machine applications
- Scenarios with extremely high performance requirements
- Applications with small data volume
- Development and testing environments

#### Redis Cache Use Cases
- Distributed systems
- Scenarios requiring persistent cache
- Applications with large data volume
- Production environments

#### Null Cache Use Cases
- Unit testing
- Environments where cache needs to be disabled
- Performance benchmarking

### 2. Cache Patterns

#### Cache-Aside Pattern

```go
func GetUser(id int) (*User, error) {
    var user User
    
    // First try to get from cache
    err := cache.Get(ctx, fmt.Sprintf("user:%d", id), &user)
    if err == nil {
        return &user, nil
    }
    
    // Cache miss, get from database
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
    
    // Update cache at the same time
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

Using the `GetSet` method can effectively prevent cache penetration:

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

### 4. Cache Avalanche Prevention

```go
// Set different expiration times for different keys
func SetUserWithRandomTTL(user *User) error {
    // Base TTL is 10 minutes
    baseTTL := 10 * time.Minute
    
    // Add random offset to prevent simultaneous expiration
    randomOffset := time.Duration(rand.Intn(300)) * time.Second // 0-5 minutes random offset
    
    ttl := baseTTL + randomOffset
    
    return cache.Set(ctx, fmt.Sprintf("user:%d", user.ID), user, ttl)
}
```

### 5. Cache Warmup

```go
func WarmupCache() error {
    // Preload hot data
    hotUsers := []int{1, 2, 3, 4, 5}
    
    for _, id := range hotUsers {
        var user User
        err := db.GetUser(id, &user)
        if err != nil {
            continue
        }
        
        cache.Set(ctx, fmt.Sprintf("user:%d", id), user, 30*time.Minute)
    }
    
    return nil
}
```

## 🔧 Advanced Configuration

### Memory Cache Tuning

```go
// High-frequency access scenarios: short expiration time, frequent cleanup
cache := go_cache.NewMemory(1*time.Minute, 2*time.Minute)

// Low-frequency access scenarios: long expiration time, infrequent cleanup
cache := go_cache.NewMemory(30*time.Minute, 1*time.Hour)

// Large data volume scenarios: increase cleanup frequency
cache := go_cache.NewMemory(10*time.Minute, 5*time.Minute)
```

### Redis Configuration Optimization

```go
// Use connection pool
rdb := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,  // Connection pool size
    MinIdleConns: 5,   // Minimum idle connections
    MaxRetries:   3,   // Maximum retry count
})

cache := go_cache.NewRedis(rdb)
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
    
    // Test deletion
    err = cache.Del(ctx, "test_key")
    assert.NoError(t, err)
    assert.False(t, cache.Exists(ctx, "test_key"))
}

func TestCacheGetSet(t *testing.T) {
    cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
    ctx := context.Background()
    
    var result string
    callCount := 0
    
    // First call, cache miss
    err := cache.GetSet(ctx, "test_key", 10*time.Minute, &result, func(key string, obj any) error {
        callCount++
        str := obj.(*string)
        *str = "callback_value"
        return nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "callback_value", result)
    assert.Equal(t, 1, callCount)
    
    // Second call, cache hit
    err = cache.GetSet(ctx, "test_key", 10*time.Minute, &result, func(key string, obj any) error {
        callCount++
        return nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "callback_value", result)
    assert.Equal(t, 1, callCount) // Callback function not called
}
```

### Benchmark Testing

```go
func BenchmarkMemoryCacheSet(b *testing.B) {
    cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Set(ctx, fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), 10*time.Minute)
    }
}

func BenchmarkMemoryCacheGet(b *testing.B) {
    cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
    ctx := context.Background()
    
    // Preset data
    for i := 0; i < 1000; i++ {
        cache.Set(ctx, fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), 10*time.Minute)
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var result string
        cache.Get(ctx, fmt.Sprintf("key_%d", i%1000), &result)
    }
}
```

## 🚨 Important Notes

### 1. Type Safety

- The `obj` parameter for `Get` and `GetSet` methods must be a pointer type
- Ensure the passed type matches the stored type, otherwise a type mismatch error will be returned

### 2. Serialization Support

- Redis cache uses Go's native gob encoding
- ✅ Fully supports: all Go types including complex structs, nested types, pointers, slices, maps
- ✅ Supports nil values: nil pointers, nil slices, nil maps
- ❌ Does not support: functions, channels (gob limitation)
- Type information is automatically preserved during serialization

### 3. Memory Management

- Memory cache occupies application memory, monitor memory usage
- Set appropriate cleanup intervals to avoid memory leaks

### 4. Concurrency Safety

- All cache implementations are concurrency-safe
- But still need to pay attention to concurrency issues in callback functions

### 5. Error Handling

- Redis cache may return errors due to network issues
- It is recommended to implement retry mechanisms or fallback strategies

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork this repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Development Environment Setup

```bash
# Clone repository
git clone https://github.com/muleiwu/go-cache.git
cd go-cache

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run benchmark tests
go test -bench=. ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Links

- [gsr Interface Library](https://github.com/muleiwu/gsr)
- [patrickmn/go-cache](https://github.com/patrickmn/go-cache)
- [redis/go-redis](https://github.com/redis/go-redis)
- [Go encoding/gob](https://pkg.go.dev/encoding/gob)

## 📊 Performance Comparison

| Operation | Memory Cache | Redis Cache | Null Cache |
|-----------|--------------|------------|------------|
| Set       | ~100ns       | ~1ms       | ~10ns      |
| Get       | ~100ns       | ~1ms       | ~10ns      |
| Del       | ~100ns       | ~1ms       | ~10ns      |

*Note: The above data are reference values, actual performance depends on hardware configuration and network environment*

## 🆘 Frequently Asked Questions

### Q: How to switch between memory cache and Redis cache?

A: Since all implementations follow the same interface, you only need to change the initialization code:

```go
// Memory cache
cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)

// Redis cache
cache := go_cache.NewRedis(redisClient)

// The rest of the code does not need to be modified
```

### Q: How to handle nil values in cache?

A: go-cache fully supports nil values (after switching to gob encoding):

```go
// Store and retrieve nil pointer
var user *User  // nil pointer
cache.Set(ctx, "user:123", user, 10*time.Minute)  // ✅ Works

var retrieved *User
cache.Get(ctx, "user:123", &retrieved)  // retrieved will be nil

// Store and retrieve nil slice
var tags []string  // nil slice
cache.Set(ctx, "tags:456", tags, 10*time.Minute)  // ✅ Works

// Store and retrieve nil map
var metadata map[string]int  // nil map
cache.Set(ctx, "metadata:789", metadata, 10*time.Minute)  // ✅ Works

// Can distinguish between nil value and key not existing
cache.Set(ctx, "key1", (*User)(nil), 10*time.Minute)
cache.Exists(ctx, "key1")  // true - key exists with nil value
cache.Exists(ctx, "key2")  // false - key doesn't exist
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

---

For other questions, please submit an [Issue](https://github.com/muleiwu/go-cache/issues) or contact the maintainer.
