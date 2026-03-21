# Lab 01 — Goroutines & Channels

**Level:** Beginner
**Topic:** Go Concurrency Foundations

---

## Background

### What is a Goroutine?

A goroutine is a lightweight thread managed by the Go runtime — not by the OS. You can run thousands of goroutines concurrently without the overhead of OS threads. Launching one is as simple as adding `go` before a function call.

```go
go someFunction()    // runs concurrently
go func() { ... }()  // anonymous goroutine
```

Goroutines are **not** OS threads. The Go scheduler multiplexes goroutines onto a small pool of OS threads using an M:N threading model.

### What is a Channel?

A channel is a typed conduit for sending and receiving values between goroutines. Channels make communication safe — they are the Go way to share memory by communicating (rather than communicating by sharing memory).

```go
ch := make(chan int)       // unbuffered: send blocks until receive
ch := make(chan int, 5)    // buffered: send blocks only when full

ch <- 42      // send
val := <-ch   // receive
```

**Unbuffered channels** synchronize sender and receiver — both must be ready.
**Buffered channels** allow sending without an immediate receiver (up to the buffer size).

### sync.WaitGroup

Used to wait for a collection of goroutines to finish. Think of it like a counter:

```go
var wg sync.WaitGroup
wg.Add(1)       // increment before launching goroutine
go func() {
    defer wg.Done()  // decrement when goroutine finishes
    // work here
}()
wg.Wait()        // block until counter reaches 0
```

### sync.Mutex

A mutual exclusion lock. Used to protect shared data from concurrent access (data races).

```go
var mu sync.Mutex
mu.Lock()
// critical section — only one goroutine at a time
mu.Unlock()
```

`sync.RWMutex` allows multiple concurrent readers but exclusive writers — use it when reads greatly outnumber writes.

### The Go Memory Model

Key rule: **"Do not communicate by sharing memory; instead, share memory by communicating."**
Channels handle synchronization automatically. If two goroutines share data without synchronization, you have a **data race** — undefined behavior in Go.

Run `go run -race .` to detect data races with the built-in race detector.

---

## Learning Objectives

By the end of this lab you will be able to:

- Launch goroutines and understand their lifecycle
- Create and use unbuffered and buffered channels
- Close channels and range over them
- Use `sync.WaitGroup` to coordinate goroutines
- Protect shared state with `sync.Mutex`
- Use the race detector to find data races

---

## Tasks

Open `main.go` and implement each task. Each task has a `TODO` block. Do not modify the function signatures.

### Task 1 — Hello Goroutines
Launch 5 goroutines, each printing `"Worker N started"` and `"Worker N done"`. Use a `WaitGroup` to ensure `main` does not exit before all goroutines finish.

**Why:** Without `Wait()`, main may exit before goroutines run — they are silently killed.

### Task 2 — Unbuffered Channel (Producer / Consumer)
Implement a producer that sends integers 1–10 into a channel. Implement a consumer that reads from the channel and prints `"Received: N"`. The producer must close the channel when done. Use `range` to receive.

**Why:** `range ch` loops until the channel is closed — it's the idiomatic way to consume all values.

### Task 3 — Buffered Channel (Batch Jobs)
Create a buffered channel with capacity 3. Send 6 job IDs into it from one goroutine. Have 2 worker goroutines each drain the channel. Print which worker processed which job.

**Why:** Buffered channels decouple the producer from the consumer — the producer can continue without blocking as long as buffer space exists.

### Task 4 — Thread-Safe Counter
Spawn 1000 goroutines, each incrementing a shared `counter` variable by 1. First, do it **without** a mutex and observe the data race (`go run -race .`). Then fix it with a `sync.Mutex`.

**Expected final value:** 1000 (without mutex, you'll often get less due to race conditions).

### Task 5 — Ping Pong
Create two goroutines: `ping` and `pong`. Pass an integer counter through two channels. Ping increments and sends to pong, pong increments and sends back to ping. Repeat 5 times, then print the final counter value.

**Why:** Demonstrates bidirectional channel-based communication between goroutines.

---

## Tips

- A goroutine that sends to an unbuffered channel will block until a receiver is ready — and vice versa.
- Closing a channel signals to receivers there are no more values; receiving from a closed channel returns the zero value and `false`: `val, ok := <-ch`.
- Never close a channel from the receiver side.
- `defer wg.Done()` is safer than placing `wg.Done()` manually — it runs even if the goroutine panics.

---

## Running Your Solution

```bash
cd lab01-goroutines-channels
go run .

# Run with race detector
go run -race .
```

---

## Expected Behaviour

- All tasks print output to stdout
- No race conditions (race detector should report none after Task 4 fix)
- Program exits cleanly — no hanging goroutines
