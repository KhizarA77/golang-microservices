package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println("=== Task 1: Hello Goroutines ===")
	task1HelloGoroutines()

	fmt.Println("\n=== Task 2: Unbuffered Channel ===")
	task2UnbufferedChannel()

	fmt.Println("\n=== Task 3: Buffered Channel ===")
	task3BufferedChannel()

	fmt.Println("\n=== Task 4: Thread-Safe Counter ===")
	task4ThreadSafeCounter()

	fmt.Println("\n=== Task 5: Ping Pong ===")
	task5PingPong()
}

// -----------------------------------------------------------------------------
// Task 1 — Hello Goroutines
//
// Launch 5 goroutines. Each goroutine should:
//   - Print "Worker N started"
//   - Do some work (hint: you can just print for now)
//   - Print "Worker N done"
//
// Use a sync.WaitGroup so main() waits for all goroutines to finish before
// printing "All workers done".
// -----------------------------------------------------------------------------
func task1HelloGoroutines() {
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		// TODO: Add to WaitGroup, launch goroutine for worker i
		wg.Add(1)
		// Inside the goroutine:
		//   - Print "Worker N started"
		//   - Print "Worker N done"
		//   - Signal Done to WaitGroup
		//
		// IMPORTANT: capture i correctly (pass as argument or use := inside loop)
		go func() {
			defer wg.Done()
			fmt.Printf("Worker %d started\n", i)
			fmt.Printf("Worker %d done\n", i)
		}()
		_ = i
	}

	// TODO: Wait for all goroutines to finish
	// TODO: Print "All workers done"
	wg.Wait()
	fmt.Println("All workers done")
}

// -----------------------------------------------------------------------------
// Task 2 — Unbuffered Channel (Producer / Consumer)
//
// producer() sends integers 1 through 10 on the channel, then closes it.
// consumer() uses range to receive all values and prints "Received: N" for each.
//
// In main (task2UnbufferedChannel):
//   - Create an unbuffered int channel
//   - Launch producer as a goroutine
//   - Call consumer (blocking) or launch as goroutine + WaitGroup
//
// -----------------------------------------------------------------------------
func producer(ch chan<- int) {
	// TODO: Send integers 1–10 to ch
	for i := range 10 {
		ch <- i
	}
	// TODO: Close the channel when done
	close(ch)
}

func consumer(ch <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	// TODO: Use range to receive all values from ch
	// TODO: Print "Received: N" for each value
	for i := range ch {
		fmt.Printf("Received: %d\n", i)
	}
}

func task2UnbufferedChannel() {
	var wg sync.WaitGroup
	// TODO: Create unbuffered channel
	ch := make(chan int)
	// TODO: Launch producer as goroutine
	go producer(ch)
	// TODO: Add 1 to WaitGroup, launch consumer as goroutine
	wg.Add(1)
	go consumer(ch, &wg)
	// TODO: Wait
	wg.Wait()
}

// -----------------------------------------------------------------------------
// Task 3 — Buffered Channel (Batch Jobs)
//
// Create a buffered channel with capacity 3.
// Send job IDs 1–6 into the channel (all from one goroutine).
// Launch 2 worker goroutines. Each worker:
//   - Reads from the channel until it's closed
//   - Prints "Worker W processed job J"
//
// Use a WaitGroup to wait for both workers to finish.
// -----------------------------------------------------------------------------
func batchWorker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	// TODO: Range over jobs channel
	// TODO: Print "Worker <id> processed job <job>"
	for i := range jobs {
		fmt.Printf("Worker %d processed job %d\n", id, i)
	}
}

func task3BufferedChannel() {
	var wg sync.WaitGroup
	// TODO: Create buffered channel with capacity 3
	ch := make(chan int, 3)
	// TODO: Launch 2 worker goroutines (add to WaitGroup before launching)
	wg.Add(2)
	go batchWorker(1, ch, &wg)
	go batchWorker(2, ch, &wg)
	// TODO: Send jobs 1–6 into the channel
	for i := range 6 {
		ch <- i
	}
	// TODO: Close the channel so workers stop ranging
	close(ch)
	// TODO: Wait for workers
	wg.Wait()
}

// -----------------------------------------------------------------------------
// Task 4 — Thread-Safe Counter
//
// Part A (racy): Spawn 1000 goroutines each incrementing `counter` by 1.
// Run with `go run -race .` and observe the data race warning.
//
// Part B (safe): Fix it by protecting counter with a sync.Mutex.
//
// Print the final counter value after all goroutines finish.
// Expected: 1000
// -----------------------------------------------------------------------------

// Unsafe version — has a data race
func task4Unsafe() int {
	var counter int
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// TODO: increment counter (no lock — intentionally racy)
			counter++
		}()
	}
	wg.Wait()
	return counter
}

// Safe version — protected with Mutex
func task4Safe() int {
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer mu.Unlock()
			// TODO: Lock, increment counter, Unlock
			mu.Lock()
			counter++
		}()
	}
	wg.Wait()
	return counter
}

func task4ThreadSafeCounter() {
	// Uncomment the unsafe call to see data race (run with -race flag)
	// unsafeResult := task4Unsafe()
	// fmt.Printf("Unsafe counter (racy): %d\n", unsafeResult)

	safeResult := task4Safe()
	fmt.Printf("Safe counter: %d (expected 1000)\n", safeResult)
}

// -----------------------------------------------------------------------------
// Task 5 — Ping Pong
//
// Create two channels: pingCh and pongCh (both chan int).
// Launch a "ping" goroutine:
//   - Receives a value from pingCh
//   - Increments it
//   - Sends to pongCh
//   - Repeats 5 times, then returns
//
// Launch a "pong" goroutine:
//   - Receives a value from pongCh
//   - Increments it
//   - Sends to pingCh
//   - Repeats 5 times, then returns
//
// Start the game by sending 0 to pingCh.
// After both goroutines finish, print the final counter value.
// Expected: 10 (incremented once per send, 5 sends each side)
// -----------------------------------------------------------------------------
func ping(pingCh <-chan int, pongCh chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	// TODO: Loop 5 times
	// TODO: Receive from pingCh, increment, send to pongCh
	// TODO: Print "Ping: N"
	for range 5 {
		val := <-pingCh
		fmt.Printf("Ping: %d\n", (val + 1))
		pongCh <- val + 1

	}
	finalVal := <-pingCh
	fmt.Printf("Final result is %d\n", finalVal)
}

func pong(pongCh <-chan int, pingCh chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	// TODO: Loop 5 times
	// TODO: Receive from pongCh, increment, send to pingCh
	// TODO: Print "Pong: N"
	for range 5 {
		val := <-pongCh
		fmt.Printf("Pong: %d\n", (val + 1))
		pingCh <- val + 1
	}
}

func task5PingPong() {
	var wg sync.WaitGroup
	// TODO: Create pingCh and pongCh channels
	// Buffer size 1: absorbs pong's final send (value 10) after ping is done,
	// so pong doesn't block and can call wg.Done().
	pingCh := make(chan int)
	pongCh := make(chan int)
	// TODO: Add 2 to WaitGroup, launch ping and pong goroutines
	wg.Add(2)
	go ping(pingCh, pongCh, &wg)
	go pong(pongCh, pingCh, &wg)
	// TODO: Kick off the game by sending 0 to pingCh
	pingCh <- 0

	// Wait for BOTH goroutines to finish first.
	// Only then read the final value that pong left in the buffer.
	wg.Wait()
}
