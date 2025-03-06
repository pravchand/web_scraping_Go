package tokenparse

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"proj3/parsingSupport"
	"proj3/queue"
	"strconv"
	"time"
)

const Max int = 400


///////////////////////////////////////////////////////////////////////////  
// HELPER ACROSS PARALLEL AND SEQUENTIAL //
///////////////////////////////////////////////////////////////////////////  
func getRandomNumbers(n, max int) []string {
	rand.Seed(time.Now().UnixNano())
	
	uniqueNumbers := make(map[int]bool)
	for len(uniqueNumbers) < n {
		if len(uniqueNumbers) >= max {
			break
		}
		uniqueNumbers[rand.Intn(max)] = true
	}
	
	randomNumbers := make([]string, 0, len(uniqueNumbers))
	for num := range uniqueNumbers {
		var numberString string
		if num < 10 {
			numberString = "0" + strconv.Itoa(num)
		} else {
			numberString = strconv.Itoa(num)
		}
		randomNumbers = append(randomNumbers, numberString)
	}
	
	return randomNumbers
}

func fetch(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	
    
	return string(body)
}

///////////////////////////////////////////////////////////////////////////
// RUN PARALLEL OR SEQUENTIAL //
///////////////////////////////////////////////////////////////////////////
func Run(c parsingSupport.Config) error {
	// Initialize logger first
	
	// Start overall timing
	startTime := time.Now()
	

	
	var err error
	if c.Mode == "p" {
		err = runParallel(c)
	} else {
		err = runSequential(c)
	}
	
	// End timing and print result
	elapsed := time.Since(startTime)
	fmt.Printf("BENCHMARK_TIME: %.5f\n", elapsed.Seconds())
	
	return err
}	

///////////////////////////////////////////////////////////////////////////
// SEQUENTIAL //
///////////////////////////////////////////////////////////////////////////
func runSequential(c parsingSupport.Config) error {
	
	randomNumbers := getRandomNumbers(c.NumLinks, Max)
	
	os.MkdirAll("data", 0755)
	
	for i, num := range randomNumbers {
		url := "https://saket-choudhary.me/seenunseencap/" + num + ".html"
		content := fetch(url)
		tokenizedTexts := parsingSupport.ParseAndTokenize(content)
		result := parsingSupport.Result{
			Url:     url,
			Content: tokenizedTexts,
		}
		parsingSupport.WriteResult(result, i)
	}
	
	return nil
}

///////////////////////////////////////////////////////////////////////////
// PARALLEL //
///////////////////////////////////////////////////////////////////////////
func runParallel(c parsingSupport.Config) error {
	links := generatePipelineLinks(c.NumLinks)
	processed := processStage(links, c)
	written := writeStage(processed, c.NumThreads)
	<-written
	return nil
}

///////////////////////////////////////////////////////////////////////////
// LINK GENERATION //
///////////////////////////////////////////////////////////////////////////
func generatePipelineLinks(numLinks int) <-chan parsingSupport.Task {
	out := make(chan parsingSupport.Task)
	
	go func() {
		defer close(out)
		randomNumbers := getRandomNumbers(numLinks, Max)
		// randomNumbers = append(randomNumbers,randomNumbers...)
		for _, num := range randomNumbers {
			out <- parsingSupport.Task{Url: "https://saket-choudhary.me/seenunseencap/" + num + ".html"}
		}
	}()
	
	return out
}

///////////////////////////////////////////////////////////////////////////
// PROCESS //
///////////////////////////////////////////////////////////////////////////
func processStage(in <-chan parsingSupport.Task, c parsingSupport.Config) <-chan parsingSupport.Result {
	out := make(chan parsingSupport.Result)
	deques := make([]*queue.BoundedDequeue, c.NumThreads)
	
	// i create a channel for each worker to send its queue to
	// this is used to avoid blocking when the worker is waiting for a queue
	queueChannels := make([]chan *queue.BoundedDequeue, c.NumThreads)
	
	for i := range deques {
		deques[i] = queue.NewBoundedDequeue(int64(math.Ceil(float64(c.NumLinks)/float64(c.NumThreads))))
		queueChannels[i] = make(chan *queue.BoundedDequeue, 1) 
	}

	completed := make(chan int, c.NumThreads)
	
	for i := 0; i < c.NumThreads; i++ {
		go func(id int) {
			myQueue := <-queueChannels[id]
			
			processWorkerAndSendResults(id, myQueue, deques, out, c.WorkStealing)
			completed <- id
		}(i)
	}
	
	go func() {
		workerIdx := 0
		taskCount := 0
		for task := range in {
			if err := deques[workerIdx].PushBottom(parsingSupport.Task{Url: task.Url}); err == nil {
				taskCount++
			} else {
				fmt.Printf("Failed to queue task: %v", err)
			}
			workerIdx = (workerIdx + 1) % c.NumThreads
		}
		
		for i := 0; i < c.NumThreads; i++ {
			queueChannels[i] <- deques[i]
		}
	}()
	
	go func() {
		for i := 0; i < c.NumThreads; i++ {
			<-completed
		}
		close(out)
	}()

	return out
}

///////////////////////////////////////////////////////////////////////////
// EACH WORKER PROCESSING THEIR OWN TASKS //
///////////////////////////////////////////////////////////////////////////

func processWorkerAndSendResults(id int, myDeque *queue.BoundedDequeue, 
	allDeques []*queue.BoundedDequeue, out chan<- parsingSupport.Result, enableWorkStealing bool) {
	var taskCount, stolenCount int
	


	done := false
	
	for !done {
		
		if task, err := myDeque.PopBottom(); err == nil {
			content := fetch(task.Url)
			result := parsingSupport.Result{
				Url:     task.Url,
				Content: parsingSupport.ParseAndTokenize(content),
			}
			
			out <- result
			taskCount++
			continue    
		}

		
		if enableWorkStealing {
			stolen := false
			numberQueues := len(allDeques)
			startIdx := rand.Intn(numberQueues) 
			
			for attempt := 0; attempt < numberQueues; attempt++ {
				victimIdx := (startIdx + attempt) % numberQueues
				
				if victimIdx == id {
					continue
				}
				
				if task, err := allDeques[victimIdx].PopTop(); err == nil {
					content := fetch(task.Url)
					result := parsingSupport.Result{
						Url:     task.Url,
						Content: parsingSupport.ParseAndTokenize(content),
					}
					
					out <- result
					stolenCount++
					stolen = true
					break
				}
			}
			
			if !stolen {
				done = true
			}
		} else {
			done = true
		}
	}
}

///////////////////////////////////////////////////////////////////////////
// WRITE //
///////////////////////////////////////////////////////////////////////////
func writeStage(in <-chan parsingSupport.Result, numWriters int) <-chan struct{} {
	completed := make(chan struct{})
	writerComplete := make(chan int, numWriters)
	
	for i := 0; i < numWriters; i++ {
		go func(id int) {
			fileWriter(id, in, writerComplete)
		}(i)
	}
	
	go func() {
		for i := 0; i < numWriters; i++ {
		<-writerComplete
		}
		close(completed)
	}()
	
	return completed
}

func fileWriter(id int, in <-chan parsingSupport.Result, done chan<- int) {
	count := 0
	
	for result := range in {
		parsingSupport.WriteResult(result, count)
		count++
	}
	
	done <- id
}
///////////////////////////////////////////////////////////////////////////