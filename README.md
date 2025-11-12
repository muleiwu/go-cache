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
- **Nil Value Support**: Full support for nil pointers, nil slices, and nil maps
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
│   ├── gob.go         # Gob serializer
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

For detailed usage guide, see [SERIALIZER_GUIDE.md](SERIALIZER_GUIDE.md)

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
