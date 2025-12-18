# GoDiscat > README.md << 'EOF'
<div align="center">

# 🚀 GoDis

### **Go**lang + Re**dis** = Production-Ready In-Memory Key-Value Store

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

*A lightweight, high-performance in-memory key-value store built from scratch in Go, implementing core Redis functionality with thread-safe operations, LRU eviction, and AOF persistence.*

[Features](#-features) • [Quick Start](#-quick-start) • [Architecture](#-architecture) • [Commands](#-supported-commands) • [Performance](#-performance)

</div>

---

## What is this?

I'm rebuilding a simplified version of Redis to learn:
- How databases handle concurrent connections
- How in-memory storage actually works
- What makes Redis so fast (spoiler: it's single-threaded!)
- How to implement persistence without slowing down reads/writes

Redis is basically a super-fast key-value store that lives in RAM. It's used everywhere - caching, session storage, pub/sub messaging, etc. But how does it actually work under the hood?

## Why Redis and not something else?

- **Redis is simpler than PostgreSQL** - no SQL parser, no complex query optimizer, just key-value ops
- **Redis is more interesting than Memcached** - has persistence (AOF), more data structures, better protocol
- **Small enough to build in a few weeks** - can actually finish this unlike building a full DBMS
- **Used in production everywhere** - understanding Redis internals is genuinely useful

## What's working so far

- [x] TCP server that accepts multiple connections
- [x] Basic RESP protocol parsing (can talk to redis-cli!)
- [x] Thread-safe storage with RWMutex
- [x] Core commands: SET, GET, DEL, EXISTS, PING
- [ ] TTL/EXPIRE - currently implementing
- [ ] LRU eviction when memory limit hit
- [ ] AOF persistence (write-ahead log)
- [ ] Background cleanup of expired keys

---
### The interesting challenges

**1. Why RWMutex instead of regular Mutex?**

Redis is read-heavy (tons of GETs, fewer SETs). RWMutex lets multiple goroutines read simultaneously, but only one can write. This makes GET operations way faster since they don't block each other.
```go
// Multiple readers can acquire this at once
mu.RLock()
value := store[key]
mu.RUnlock()

// Only one writer allowed
mu.Lock()
store[key] = value
mu.Unlock()
```

**2. RESP protocol is actually simple**

Redis uses RESP (REdis Serialization Protocol). It's just text over TCP:

Client sends: *2\r\n$3\r\nGET\r\n$3\r\nkey\r\n
Server replies: $5\r\nvalue\r\n

Simple but efficient. That's why `redis-cli` can connect to GoDis!

**3. LRU eviction is tricky**

Need O(1) for both access and eviction. Solution: HashMap + Doubly Linked List
- HashMap gives O(1) lookups
- Doubly linked list tracks order (most recent at head, least recent at tail)
- When memory full, pop from tail

**4. AOF persistence without killing performance**

Can't do synchronous writes - would slow down every SET command. Solution:
- Writes go to a buffered channel
- Separate goroutine drains the channel and writes to disk
- Configure fsync policy (always/everysec/never)

---

## Supported Commands

| Command | Status | Example |
|---------|--------|---------|
| `PING` | ✅ | `PING` → `PONG` |
| `SET` | ✅ | `SET key value` |
| `GET` | ✅ | `GET key` |
| `DEL` | ✅ | `DEL key` |
| `EXISTS` | ✅ | `EXISTS key` |
| `EXPIRE` | 🚧 | `EXPIRE key 60` |
| `TTL` | 🚧 | `TTL key` |
| `FLUSHALL` | 📝 | Clear everything |

More commands coming as I learn more about Redis internals.

---

Pretty standard Go project layout. Keeping it simple.

---

## What I learned building this

**Goroutines aren't magic**

Initially spawned a goroutine for every single operation. Bad idea - too much overhead. Now: one goroutine per client connection, separate goroutine for AOF writes, one for expiration cleanup. That's it.

**Mutex placement matters**

First attempt: locked the entire command handler. This made everything sequential even though I had goroutines. Now: lock only around the actual map access. Way faster.

**RESP protocol is clever**

Why use a text protocol instead of binary? Because it's:
1. Easy to debug (can read it with telnet)
2. Fast enough for most use cases
3. Simple to implement correctly

Binary protocols are faster but Redis proves text is fast enough.

**Testing concurrent code is hard**

The `-race` flag in Go is a lifesaver. Found 3 race conditions I would've never spotted otherwise. Always run `go test -race`.

**Implementing LRU taught me more than I expected**

Thought it would be simple - just a linked list, right? Wrong. Need to:
- Update on every access (GET counts!)
- Handle edge cases (empty cache, single element)
- Make it thread-safe without killing performance
- Integrate with the main storage map

Took way longer than expected but learned a lot about data structures.

---

## Testing
```bash
# Run tests
make test

# Run with race detector (finds concurrency bugs)
make test-race

# Benchmarks
make bench
```

Current benchmark on my laptop (M1 MacBook):
- ~45K GET ops/sec (single connection)
- ~28K SET ops/sec (single connection)
- Latency: <1ms p50, ~3ms p99

Not as fast as real Redis but way faster than I thought I could make it!

---

## 🗺️ Roadmap

### Phase 1: Foundation ✅
- [x] TCP server with goroutines
- [x] RESP protocol parser
- [x] Basic commands (SET, GET, DEL)
- [x] Thread-safe storage

### Phase 2: Advanced Features (In Progress)
- [x] TTL/Expiration
- [x] Background cleanup
- [ ] LRU eviction
- [ ] Memory limits

### Phase 3: Persistence (Planned)
- [ ] AOF writer
- [ ] AOF replay
- [ ] Configurable fsync
- [ ] AOF compaction

### Phase 4: Production Polish (Planned)
- [ ] Structured logging
- [ ] Metrics endpoint
- [ ] Docker support
- [ ] Comprehensive docs

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📚 Learning Resources

This project was built to understand:
- Go concurrency patterns (goroutines, channels, mutexes)
- Network programming (TCP, sockets)
- Data structures (hash maps, doubly linked lists)
- Persistence mechanisms (AOF, fsync)
- Protocol design (RESP)

### Recommended Reading
- [Redis Protocol Specification](https://redis.io/docs/reference/protocol-spec/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)

---

## 👨‍💻 Author

**Kartheek Budime** 

[![GitHub](https://img.shields.io/badge/GitHub-kartheek0107-181717?style=flat&logo=github)](https://github.com/kartheek0107)
[![LinkedIn](https://img.shields.io/badge/LinkedIn-kartheek--budime-0077B5?style=flat&logo=linkedin)](https://linkedin.com/in/kartheek-budime)
[![Email](https://img.shields.io/badge/Email-kartheekbudime%40gmail.com-D14836?style=flat&logo=gmail)](mailto:kartheekbudime@gmail.com)

---

<div align="center">

**⭐ Star this repo if you find it helpful!**

Made with ❤️ and lots of ☕ by Kartheek

</div>
EOF