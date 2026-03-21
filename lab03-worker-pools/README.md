# Lab 03 — Worker Pools & Pipelines

**Level:** Advanced
**Topic:** Worker Pool, Pipeline, Fan-Out, Fan-In, Semaphore

---

## Background

### Worker Pool Pattern

A worker pool limits the number of goroutines doing concurrent work. Instead of launching one goroutine per job (which could be thousands), you launch a fixed pool of N workers that all read from a shared jobs channel.

```
jobs channel  ──►  [worker 1]  ──►  results channel
               ──►  [worker 2]  ──►
               ──►  [worker 3]  ──►
```

```go
jobs    := make(chan Job, 100)
results := make(chan Result, 100)

// Start fixed pool
for w := 1; w <= numWorkers; w++ {
    go worker(w, jobs, results)
}

// Feed jobs
for _, j := range allJobs {
    jobs <- j
}
close(jobs)

// Collect results
for range allJobs {
    r := <-results
    fmt.Println(r)
}
```

### Pipeline Pattern

A pipeline is a chain of processing stages connected by channels. Each stage:
- Receives values from an upstream channel
- Transforms or processes them
- Sends results to a downstream channel

```
generate ──► stage1 ──► stage2 ──► stage3 ──► consumer
```

Each stage is a goroutine (or pool of goroutines). The pipeline processes data concurrently — stage1 can be working on item N+1 while stage2 is still processing item N.

```go
func generate(nums ...int) <-chan int { ... }
func square(in <-chan int) <-chan int { ... }
func filter(in <-chan int, pred func(int)bool) <-chan int { ... }

// Chain:
out := filter(square(generate(1, 2, 3, 4, 5)), func(n int) bool { return n > 5 })
for v := range out { fmt.Println(v) }
```

### Fan-Out

Fan-out distributes work from a single channel to multiple worker goroutines reading from the **same** channel. Each item goes to exactly one worker.

```
         ──► [worker A]
in chan  ──► [worker B]
         ──► [worker C]
```

All workers compete to receive from the same input channel — Go's scheduler distributes work naturally.

### Fan-In (Merge)

Fan-in multiplexes multiple input channels into one output channel. Useful when you have multiple independent result streams and want to process them in one place.

```
[result ch A] ──►
[result ch B] ──► merged channel ──► consumer
[result ch C] ──►
```

```go
func merge(cs ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)

    output := func(c <-chan int) {
        defer wg.Done()
        for v := range c { out <- v }
    }
    wg.Add(len(cs))
    for _, c := range cs { go output(c) }

    go func() { wg.Wait(); close(out) }()
    return out
}
```

### Semaphore Pattern

A semaphore limits the number of concurrent goroutines accessing a resource. In Go, a buffered channel works as a semaphore:

```go
sem := make(chan struct{}, maxConcurrency)

// Acquire (blocks if at max)
sem <- struct{}{}
defer func() { <-sem }()  // Release
```

This is useful when you want to launch goroutines freely but limit how many run simultaneously (e.g., limit concurrent HTTP requests to an external API).

### Cancellation in Pipelines

When using pipelines, a consumer that stops early (e.g., only needs the first N results) must signal upstream stages to stop. Pattern:

```go
func stage(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            select {
            case out <- process(v):
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

---

## Learning Objectives

By the end of this lab you will be able to:

- Build a fixed-size worker pool with jobs and results channels
- Construct a multi-stage processing pipeline
- Distribute work via fan-out and collect results via fan-in
- Use a buffered channel as a semaphore to limit concurrency
- Propagate context cancellation through a pipeline

---

## Tasks

### Task 1 — Basic Worker Pool
You have 20 "jobs" (integers 1–20). Each job's "work" is sleeping for a random duration (50–200ms) then returning `job * job`.

Build a worker pool with 5 workers:
- `jobs` channel (buffered, size 20)
- `results` channel (buffered, size 20)
- Each worker prints `"worker W processing job J"` and sends result
- Main collects all 20 results and prints them

### Task 2 — Pipeline: Number Processing
Build a 3-stage pipeline:
1. `generate(nums ...int) <-chan int` — emits numbers into a channel
2. `square(in <-chan int) <-chan int` — receives numbers, sends their squares
3. `filter(in <-chan int, threshold int) <-chan int` — only passes values > threshold

Chain them: `generate(1..10) -> square -> filter(>25)`. Collect and print all results.

Expected output: 36, 49, 64, 81, 100 (squares of 6–10)

### Task 3 — Fan-Out / Fan-In
Given a stream of 12 URLs (simulate as strings "url-1" through "url-12"), "fetch" each one concurrently using fan-out to 3 workers. Each fetch takes a random 50–150ms. Fan-in the results back to a single channel and print each result.

Structure:
- `fanOut(in <-chan string, workers int) []<-chan string` — creates N workers, returns N result channels
- `fanIn(channels ...<-chan string) <-chan string` — merges N channels into 1
- Each worker prints which URL it fetched and on which worker

### Task 4 — Semaphore
You have 50 tasks. Each task takes 100ms. Without limiting, all 50 would run at once. Implement a semaphore that allows at most 5 tasks to run concurrently. Print how many are running at each point in time (use an atomic counter). Measure total elapsed time — should be ~1 second (10 batches of 5 × 100ms).

### Task 5 — Pipeline with Cancellation
Extend the pipeline from Task 2 with context cancellation. The consumer only wants the first 3 results. After collecting 3, cancel the context. Upstream stages must stop gracefully. Print `"pipeline cancelled"` when done.

---

## Tips

- Close the jobs channel (not the results channel) after sending all jobs — this signals workers to stop ranging.
- Collect results before waiting for workers — otherwise you deadlock if results channel is full.
- In fan-in, a goroutine per input channel forwards values to the merged channel; a separate goroutine calls `wg.Wait()` then closes the output.
- Use `sync/atomic` for atomic counters instead of mutex when you only need simple increment/decrement.
- Benchmark with `time.Since(start)` to verify your concurrency actually speeds things up.

---

## Running Your Solution

```bash
cd lab03-worker-pools
go run .
```
