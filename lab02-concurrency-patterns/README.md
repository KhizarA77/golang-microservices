# Lab 02 — Concurrency Patterns

**Level:** Intermediate
**Topic:** Select, Timeouts, Context, Done Channels, Rate Limiting

---

## Background

### The `select` Statement

`select` lets a goroutine wait on multiple channel operations simultaneously — like a switch statement for channels. The first case that's ready executes.

```go
select {
case msg := <-ch1:
    fmt.Println("from ch1:", msg)
case msg := <-ch2:
    fmt.Println("from ch2:", msg)
case <-time.After(2 * time.Second):
    fmt.Println("timeout")
default:
    fmt.Println("no channel ready") // non-blocking
}
```

If multiple cases are ready simultaneously, Go picks one at random — ensuring fairness.

### Timeouts with `time.After`

`time.After(d)` returns a channel that receives a value after duration `d`. Combined with `select`, it implements timeouts:

```go
select {
case result := <-work():
    fmt.Println(result)
case <-time.After(1 * time.Second):
    fmt.Println("timed out")
}
```

### The Done Channel Pattern

A conventional way to signal cancellation: pass a `done <-chan struct{}` to goroutines. When the caller closes `done`, all listeners unblock and can clean up.

```go
func worker(done <-chan struct{}) {
    for {
        select {
        case <-done:
            return   // cancelled
        default:
            // do work
        }
    }
}

done := make(chan struct{})
go worker(done)
close(done)  // signal all workers to stop
```

`struct{}` is used because it has zero size — it's purely a signal, no data needed.

### `context.Context`

`context.Context` is the idiomatic Go way to carry cancellation, deadlines, and request-scoped values through a call chain. It supersedes the manual done channel pattern for most use cases.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  // always call cancel to free resources

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// In a goroutine or function:
select {
case <-ctx.Done():
    fmt.Println("cancelled:", ctx.Err())
case result := <-doWork():
    fmt.Println(result)
}
```

Key rules:
- Always `defer cancel()` immediately after creating a context
- Pass `ctx` as the **first argument** to functions (by convention)
- Never store context in a struct — pass it explicitly

### Rate Limiting with `time.Ticker`

A `time.Ticker` fires on a regular interval. Use it to throttle the rate of operations:

```go
limiter := time.NewTicker(200 * time.Millisecond)
defer limiter.Stop()

for req := range requests {
    <-limiter.C   // block until next tick
    fmt.Println("processing", req)
}
```

### Graceful Shutdown Pattern

Production services listen for OS signals to shut down cleanly:

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

// Start serving...

<-quit  // block until signal received
// do cleanup
```

---

## Learning Objectives

By the end of this lab you will be able to:

- Use `select` to multiplex over multiple channels
- Implement timeouts without blocking indefinitely
- Cancel goroutines with a done channel
- Use `context.WithCancel` and `context.WithTimeout`
- Build a basic rate limiter
- Implement a graceful shutdown sequence

---

## Tasks

### Task 1 — Select Multiplexer
Create two channels: `fastCh` and `slowCh`. Launch goroutines that send values on each at different intervals (e.g., every 100ms and 300ms respectively). Use a `for` + `select` loop to receive from whichever is ready first, and print the source and value. Stop after 10 total receives.

### Task 2 — Timeout
Write a function `fetchData(delay time.Duration) <-chan string` that simulates a slow operation by sleeping for `delay` then sending a result string. In `task2Timeout`, call it with a 2-second delay and a 1-second timeout. If data arrives in time, print it; otherwise print `"request timed out"`.

Then call it again with a 500ms delay and the same 1-second timeout — this time the data should arrive.

### Task 3 — Done Channel Cancellation
Launch a worker goroutine that loops, printing `"working..."` every 200ms. After 1 second, signal it to stop using a done channel. The worker must print `"worker stopped"` before exiting. The main goroutine prints `"all clean"` after the worker exits.

### Task 4 — Context with Timeout
Rewrite Task 3 using `context.WithTimeout` instead of a done channel. The worker receives a `ctx context.Context` and selects on `ctx.Done()`. After the context expires, print `ctx.Err()` inside the worker.

### Task 5 — Context Propagation
Write a chain of three functions: `A(ctx) -> B(ctx) -> C(ctx)`. C does the actual "work" (loops every 100ms). When the top-level context is cancelled (after 500ms), C should stop, B should clean up, and A should report the cancellation. Demonstrate how cancellation propagates through the call chain.

### Task 6 — Rate Limiter
You have a slice of 10 requests. Process them with a rate limit of 3 requests per second (one every ~333ms). Print the timestamp and request number for each. At the end, show that the total elapsed time is approximately what you'd expect at that rate.

---

## Tips

- `select` with a `default` case is non-blocking — it executes default immediately if no other case is ready.
- `context.WithDeadline` takes an absolute `time.Time`; `context.WithTimeout` takes a duration — prefer `WithTimeout` for simplicity.
- Closing a channel broadcasts to **all** receivers simultaneously — useful for fan-out cancellation.
- `time.Ticker` must be stopped with `.Stop()` to free its goroutine: `defer ticker.Stop()`.
- When a context is cancelled, `ctx.Done()` channel is closed — any goroutine selecting on it will unblock.

---

## Running Your Solution

```bash
cd lab02-concurrency-patterns
go run .
```

---

## Expected Behaviour

- Task 1: alternating messages from fast/slow channels
- Task 2: first call times out, second call succeeds
- Task 3 & 4: worker runs for ~1 second then exits cleanly
- Task 5: all three functions report cancellation propagation
- Task 6: requests processed at ~3/sec with correct timestamps
