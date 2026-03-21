package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== Task 1: Select Multiplexer ===")
	task1SelectMultiplexer()

	fmt.Println("\n=== Task 2: Timeout ===")
	task2Timeout()

	fmt.Println("\n=== Task 3: Done Channel Cancellation ===")
	task3DoneChannel()

	fmt.Println("\n=== Task 4: Context with Timeout ===")
	task4ContextTimeout()

	fmt.Println("\n=== Task 5: Context Propagation ===")
	task5ContextPropagation()

	fmt.Println("\n=== Task 6: Rate Limiter ===")
	task6RateLimiter()
}

// -----------------------------------------------------------------------------
// Task 1 — Select Multiplexer
//
// Create two channels: fastCh (fires every 100ms) and slowCh (fires every 300ms).
// Launch goroutines that send incrementing integers on each channel.
// In a for+select loop, receive from whichever is ready and print:
//
//	"fast: N" or "slow: N"
//
// Stop after receiving 10 total values across both channels.
//
// Hint: use a counter variable and break out of the loop when it hits 10.
// -----------------------------------------------------------------------------
func task1SelectMultiplexer() {
	// TODO: Create fastCh and slowCh (both chan int)
	fastCh := make(chan int)
	slowCh := make(chan int)
	// TODO: Launch goroutine for fast sender (sends every 100ms)
	go func() {
		for i := range 10 {
			<-time.After(100 * time.Millisecond)
			fastCh <- i
		}
	}()
	// TODO: Launch goroutine for slow sender (sends every 300ms)
	go func() {
		for i := range 10 {
			<-time.After(300 * time.Millisecond)
			slowCh <- i
		}
	}()
	// TODO: for loop with select on fastCh and slowCh
	// TODO: Print source and value, count receives, break at 10
	for range 10 {
		select {
		case result := <-fastCh:
			fmt.Printf("fast: %d\n", result)
		case result := <-slowCh:
			fmt.Printf("slow: %d\n", result)
		}
	}
}

// -----------------------------------------------------------------------------
// Task 2 — Timeout
//
// fetchData simulates a slow async operation. It launches a goroutine that
// sleeps for `delay`, then sends a result string to the returned channel.
//
// In task2Timeout:
//   - Call fetchData(2*time.Second) with a 1-second timeout -> should time out
//   - Call fetchData(500*time.Millisecond) with a 1-second timeout -> should succeed
//
// -----------------------------------------------------------------------------
func fetchData(delay time.Duration) <-chan string {
	ch := make(chan string, 1)
	go func() {
		// TODO: Sleep for delay
		<-time.After(delay)
		// TODO: Send "data arrived" to ch
		ch <- "result"
	}()
	return ch
}

func task2Timeout() {
	timeout := 1 * time.Second

	fmt.Println("Attempt 1 (slow, should time out):")
	// TODO: select on fetchData(2*time.Second) and time.After(timeout)
	var res string
	ch := fetchData(2 * time.Second)
	select {
	case result := <-ch:
		res = result
	case <-time.After(timeout):
		res = "request timed out"
	}

	// TODO: Print result or "request timed out"
	fmt.Println(res)
	fmt.Println("Attempt 2 (fast, should succeed):")
	// TODO: select on fetchData(500*time.Millisecond) and time.After(timeout)
	ch = fetchData(500 * time.Millisecond)
	select {
	case result := <-ch:
		res = result
	case <-time.After(timeout):
		res = "request timed out"
	}
	// TODO: Print result or "request timed out"
	fmt.Println(res)
	fmt.Println("Attempt 2 (fast, should succeed):")
}

// -----------------------------------------------------------------------------
// Task 3 — Done Channel Cancellation
//
// Launch a worker goroutine that:
//   - Loops, selecting on done channel and a time.Tick(200ms)
//   - On tick: prints "working..."
//   - On done: prints "worker stopped" and returns
//
// In task3DoneChannel:
//   - Create done channel (chan struct{})
//   - Launch worker goroutine
//   - Sleep 1 second
//   - Close done channel
//   - Give worker goroutine a moment to print its exit message
//   - Print "all clean"
//
// -----------------------------------------------------------------------------
func workerWithDone(done <-chan struct{}) {
	// TODO: Create a time.Ticker (200ms) or use time.After in a loop
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	// TODO: Select on done and tick
	// TODO: On tick: print "working..."
	// TODO: On done: print "worker stopped", return
	for {
		select {
		case <-ticker.C:
			fmt.Println("Working")
		case <-done:
			fmt.Println("Worker Stopped")
			return
		}
	}
}

func task3DoneChannel() {
	done := make(chan struct{})
	// TODO: Launch workerWithDone goroutine
	go workerWithDone(done)
	// TODO: Sleep 1 second
	<-time.After(1 * time.Second)
	// TODO: Close done
	close(done)
	// TODO: Give goroutine time to print exit message (small sleep)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("all clean")
	_ = done
}

// -----------------------------------------------------------------------------
// Task 4 — Context with Timeout
//
// Same as Task 3 but use context.WithTimeout instead of a done channel.
//
// workerWithCtx receives a context.Context. It selects on ctx.Done() and a tick.
//   - On tick: prints "working..."
//   - On ctx.Done(): prints "worker cancelled:", ctx.Err() and returns
//
// In task4ContextTimeout:
//   - Create context with 1-second timeout
//   - Always defer cancel()
//   - Launch workerWithCtx
//   - Wait for worker to finish (small sleep or WaitGroup)
//
// -----------------------------------------------------------------------------
func workerWithCtx(ctx context.Context) {
	// TODO: Create ticker (200ms)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	// TODO: Select on ctx.Done() and ticker.C
	// TODO: On tick: print "working..."
	// TODO: On ctx.Done(): print cancellation reason and return
	for {
		select {
		case <-ticker.C:
			fmt.Println("working....")
		case <-ctx.Done():
			fmt.Println("Done! ", ctx.Err())
			return
		}
	}
}

func task4ContextTimeout() {
	// TODO: context.WithTimeout — 1 second
	// TODO: defer cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	// TODO: Launch workerWithCtx goroutine
	// TODO: Sleep long enough for context to expire + worker to print exit
	go workerWithCtx(ctx)
	<-ctx.Done()
	time.Sleep(50 * time.Millisecond)

}

// -----------------------------------------------------------------------------
// Task 5 — Context Propagation
//
// Write a chain: A(ctx) calls B(ctx) which calls C(ctx).
//
// C(ctx): loops every 100ms printing "C: working", selects on ctx.Done().
//
//	When done: prints "C: stopping", returns.
//
// B(ctx): calls C(ctx), after C returns prints "B: C finished, cleaning up".
// A(ctx): calls B(ctx), after B returns prints "A: done".
//
// In task5ContextPropagation:
//   - Create a context with 500ms timeout
//   - Call A(ctx)
//   - Print "propagation complete"
//
// -----------------------------------------------------------------------------
func C(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// TODO: Loop with select on ctx.Done() and ticker.C
	// TODO: On tick: print "C: working"
	// TODO: On done: print "C: stopping", return
	for {
		select {
		case <-ticker.C:
			fmt.Println("C: working")
		case <-ctx.Done():
			fmt.Println("C: stopping")
			return
		}
	}
}

func B(ctx context.Context) {
	// TODO: Call C(ctx)
	C(ctx)
	// TODO: Print "B: C finished, cleaning up"
	fmt.Println("B: C finished, cleaning up")
}

func A(ctx context.Context) {
	// TODO: Call B(ctx)
	B(ctx)
	// TODO: Print "A: done"
	fmt.Println("A: done")
}

func task5ContextPropagation() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// TODO: Call A(ctx)
	A(ctx)
	fmt.Println("propagation complete")
}

// -----------------------------------------------------------------------------
// Task 6 — Rate Limiter
//
// Process 10 requests with a rate limit of 3 per second (one every ~333ms).
// For each request, print:
//
//	"processed request N at <timestamp>"
//
// After all requests, print total elapsed time.
// Expected: ~3.3 seconds for 10 requests at 3/sec.
//
// Hint: time.NewTicker(333 * time.Millisecond)
// -----------------------------------------------------------------------------
func task6RateLimiter() {
	requests := make([]int, 10)
	for i := range requests {
		requests[i] = i + 1
	}

	start := time.Now()

	// TODO: Create ticker at 333ms interval
	// TODO: defer ticker.Stop()
	ticker := time.NewTicker(333 * time.Millisecond)
	defer ticker.Stop()
	// TODO: For each request, wait for tick, then print processed message with timestamp
	for _, val := range requests {
		tick := <-ticker.C
		fmt.Printf("processed request %d at %v\n", val, tick)

	}
	fmt.Printf("Total time: %v\n", time.Since(start))
}
