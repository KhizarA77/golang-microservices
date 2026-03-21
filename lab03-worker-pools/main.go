package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== Task 1: Worker Pool ===")
	task1WorkerPool()

	fmt.Println("\n=== Task 2: Pipeline ===")
	task2Pipeline()

	fmt.Println("\n=== Task 3: Fan-Out / Fan-In ===")
	task3FanOutFanIn()

	fmt.Println("\n=== Task 4: Semaphore ===")
	task4Semaphore()

	fmt.Println("\n=== Task 5: Pipeline with Cancellation ===")
	task5PipelineCancel()
}

// =============================================================================
// Task 1 — Basic Worker Pool
// =============================================================================

// Job holds the ID of a unit of work
type Job struct {
	ID int
}

// Result holds the output of processing a Job
type Result struct {
	WorkerID int
	JobID    int
	Value    int
}

// poolWorker processes jobs from the jobs channel and sends results to results channel.
// It prints "worker W processing job J" for each job.
func poolWorker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		// TODO: Simulate work with random sleep (50–200ms)
		time.Sleep(time.Duration(rand.Intn(151)+50) * time.Millisecond)
		// TODO: Compute value = job.ID * job.ID
		val := job.ID * job.ID
		// TODO: Print "worker <id> processing job <job.ID>"
		fmt.Printf("worker %d processing job %d\n", id, job.ID)
		// TODO: Send Result{WorkerID: id, JobID: job.ID, Value: value} to results
		results <- Result{id, job.ID, val}
	}
}

func task1WorkerPool() {
	const numJobs = 20
	const numWorkers = 5

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)
	var wg sync.WaitGroup

	// TODO: Launch numWorkers goroutines (poolWorker), add to WaitGroup
	for i := range numWorkers {
		wg.Add(1)
		go poolWorker(i+1, jobs, results, &wg)
	}
	// TODO: Send jobs 1–20 to jobs channel, then close it
	for i := range numJobs {
		jobs <- Job{i + 1}
	}
	close(jobs)
	// TODO: Launch a goroutine that waits for all workers then closes results
	go func() {
		wg.Wait()
		close(results)
	}()
	// TODO: Range over results and print "Job J -> Value V (by worker W)"
	for result := range results {
		fmt.Printf("Job %d -> Value %d (by worker %d)\n", result.JobID, result.Value, result.WorkerID)
	}
}

// =============================================================================
// Task 2 — Pipeline
// =============================================================================

// generate sends each number in nums to a new channel, then closes it.
func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		// TODO: Send each num to out
		for _, num := range nums {
			out <- num
		}
	}()
	return out
}

// square reads from in, squares each value, sends to output channel.
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		// TODO: Range over in, send n*n to out
		for val := range in {
			out <- val * val
		}
	}()
	return out
}

// filter reads from in, only passes values > threshold to output channel.
func filter(in <-chan int, threshold int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		// TODO: Range over in, send value to out only if value > threshold
		for val := range in {
			if val > threshold {
				out <- val
			}
		}
	}()
	return out
}

/*
	Generate -> Square -> Filter
*/

func task2Pipeline() {
	// Chain: generate(1..10) -> square -> filter(>25)
	// TODO: Build the pipeline
	out := filter(square(generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)), 25)
	// TODO: Range over final stage, print each value
	for res := range out {
		fmt.Printf("%d, ", res)
	}
	// Expected: 36, 49, 64, 81, 100
}

// =============================================================================
// Task 3 — Fan-Out / Fan-In
// =============================================================================

// fetchURL simulates fetching a URL. It sleeps randomly then sends result string.
func fetchURL(workerID int, urls <-chan string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range urls {
		// TODO: Sleep random 50–150ms
		time.Sleep(time.Duration(rand.Intn(101)+50) * time.Millisecond)
		// TODO: Send fmt.Sprintf("worker %d fetched %s", workerID, url) to results
		results <- fmt.Sprintf("worker %d fetched %s", workerID, url)
	}
}

// fanIn merges multiple result channels into one.
func fanIn(channels ...<-chan string) <-chan string {
	merged := make(chan string, 20)
	var wg sync.WaitGroup

	forwardFrom := func(ch <-chan string) {
		defer wg.Done()
		// TODO: Range over ch, send each value to merged
		for s := range ch {
			merged <- s
		}
	}
	wg.Add(len(channels))
	for _, ch := range channels {
		// TODO: Launch forwardFrom goroutine for each channel
		go forwardFrom(ch)
	}

	// TODO: Launch goroutine: wait for all forwarders, then close merged
	go func() {
		wg.Wait()
		close(merged)
	}()
	return merged
}

func task3FanOutFanIn() {
	const numURLs = 12
	const numWorkers = 3

	urlCh := make(chan string, numURLs)
	for i := 1; i <= numURLs; i++ {
		urlCh <- fmt.Sprintf("url-%d", i)
	}
	close(urlCh)
	// TODO: Launch numWorkers goroutines, each calling fetchURL
	//       Collect result channels from each worker into a slice
	// Hint: each worker needs its own results channel, OR use a single shared one
	// var _wg sync.WaitGroup
	// resultChannels := make([]<-chan string, 0, 3)
	// for i := range numWorkers {
	// 	ch := make(chan string)
	// 	resultChannels = append(resultChannels, ch)
	// 	_wg.Add(1)
	// 	go func() {
	// 		fetchURL(i+1, urlCh, ch, &_wg)
	// 		close(ch)
	// 	}()
	// }
	// res := fanIn(resultChannels...)
	// for str := range res {
	// 	fmt.Println(str)
	// }
	// _wg.Wait()
	// Simpler approach: use a single shared results channel (fan-out to same channel)
	results := make(chan string, numURLs)
	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		// TODO: Launch fetchURL goroutine for worker w
		go fetchURL(w, urlCh, results, &wg)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	// // TODO: Launch goroutine: wg.Wait() then close(results)
	// // TODO: Range over results and print each
	for r := range results {
		fmt.Println(r)
	}
}

// =============================================================================
// Task 4 — Semaphore
// =============================================================================

func task4Semaphore() {
	const numTasks = 50
	const maxConcurrent = 5

	// A buffered channel used as a semaphore
	sem := make(chan struct{}, maxConcurrent)

	var running int64 // atomic counter of currently running tasks
	var wg sync.WaitGroup
	start := time.Now()

	for i := 1; i <= numTasks; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()

			// TODO: Acquire semaphore (send to sem)
			sem <- struct{}{}
			// TODO: defer release (receive from sem)
			defer func() { <-sem }()
			current := atomic.AddInt64(&running, 1)
			fmt.Printf("task %2d started (running: %d)\n", taskID, current)

			// Simulate work
			time.Sleep(100 * time.Millisecond)

			atomic.AddInt64(&running, -1)
			// TODO: Release is handled by defer above

		}(i)
	}

	wg.Wait()
	fmt.Printf("All tasks done in %v\n", time.Since(start))
	fmt.Println("Expected ~1 second (10 batches of 5 × 100ms)")
}

// =============================================================================
// Task 5 — Pipeline with Cancellation
// =============================================================================

// generateCtx is like generate but respects context cancellation.
func generateCtx(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				// TODO: Print "generate: cancelled", return
				fmt.Println("generate: cancelled")
				return
			}
		}
	}()
	return out
}

// squareCtx is like square but respects context cancellation.
func squareCtx(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			select {
			// TODO: Send v*v to out
			case out <- (v * v):
			case <-ctx.Done():
				// TODO: Print "square: cancelled", return
				fmt.Println("square: cancelled")
				return
			}
		}
	}()
	return out
}

func task5PipelineCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := make([]int, 10)
	for i := range nums {
		nums[i] = i + 1
	}

	// TODO: Build pipeline using generateCtx and squareCtx
	out := squareCtx(ctx, generateCtx(ctx, nums...))
	// TODO: Collect first 3 results from the pipeline, print them
	// TODO: After 3 results, call cancel()
	// TODO: Print "pipeline cancelled"
	for range 3 {
		res := <-out
		fmt.Println(res)
	}
	cancel()
	fmt.Println("Pipeline cancelled")
}

// =============================================================================
// Helpers
// =============================================================================

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}
