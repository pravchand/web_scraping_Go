package tokenparse

// this is the workhorse package


// reference for pipeline implementation
// https://dev.to/johnscode/the-pipeline-pattern-in-go-2bho
// https://anupamgogoi.medium.com/go-pipeline-for-a-layman-4791fb4f1e2d
// https://www.youtube.com/watch?v=8Rn8yOQH62k&t=543s
// Channels in Go
// https://www.youtube.com/results?search_query=channels+in+Go
// https://www.youtube.com/watch?v=nNXhePi3xwE

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"proj3/parsingSupport"
	"proj3/queue"
	"strconv"
	"time"
)

// maximum number of links in my usecase
const Max int = 400

///////////////////////////////////////////////////////////////////////////  
// HELPER ACROSS PARALLEL AND SEQUENTIAL //
///////////////////////////////////////////////////////////////////////////  

// Random number generation is to help cause load imbalance among threads
// this is to generate uncertainity in the load of each thread

func getRandomNumbers(n, max int) []string {
	rand.Seed(time.Now().UnixNano())
	
	// was initially generating a lot of repeats-hence, the final file counts werent matching
	// so added a map to only have unique 
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
	// Gives back a slice of n random numbers from 0 to 400 
	return randomNumbers
}

// helper to send the GET REQUEST and send back the body of the response as a string
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
// MAIN FUNCTION RUN PARALLEL OR SEQUENTIAL //
///////////////////////////////////////////////////////////////////////////
func Run(c parsingSupport.Config) error {
	
	startTime := time.Now()
	var err error
	if c.Mode == "p" {
		err = runParallel(c)
	} else {
		err = runSequential(c)
	}
	
	elapsed := time.Since(startTime)
	fmt.Printf("BENCHMARK_TIME: %.5f\n", elapsed.Seconds())
	
	return err
}	

///////////////////////////////////////////////////////////////////////////
// SEQUENTIAL //
///////////////////////////////////////////////////////////////////////////
func runSequential(c parsingSupport.Config) error {
	
	randomNumbers := getRandomNumbers(c.NumLinks, Max)
	
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
// PARALLEL - Pipeline implementation
///////////////////////////////////////////////////////////////////////////
// currently,no error channel, but can be extended to include on

func runParallel(c parsingSupport.Config) error {
	links := generatePipelineLinks(c.NumLinks)
    // process stage fetches the content and tokenizes it
	processed := processStage(links, c)
    // write stage writes the results to the file
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
	// Currenly, all the queues are filled and then sent to the workers
    // if i do not have to use queues, i could probably just send the tasks to the worker channels
	queueChannels := make([]chan *queue.BoundedDequeue, c.NumThreads)
	
	for i := range deques {
		deques[i] = queue.NewBoundedDequeue(int64(math.Ceil(float64(c.NumLinks)/float64(c.NumThreads))))
		queueChannels[i] = make(chan *queue.BoundedDequeue, 1) 
	}

    // to track the completion of each worker processing the html
	completed := make(chan int, c.NumThreads)
	
	for i := 0; i < c.NumThreads; i++ {
		go func(id int) {
            // receive the queue from the channel
			myQueue := <-queueChannels[id]
			
			processWorkerAndSendResults(id, myQueue, deques, out, c.WorkStealing)
			completed <- id
		}(i)
	}
	
    // A single go routine to fill the queues and send them to the workers
	go func() {
		workerIdx := 0
		taskCount := 0
		for task := range in {
            // Round robin to fill the queues
			if err := deques[workerIdx].PushBottom(parsingSupport.Task{Url: task.Url}); err == nil {
				taskCount++
			} else {
				fmt.Printf("Failed to queue task: %v", err)
			}
            // This helps increment and wrap back to the first queue
			workerIdx = (workerIdx + 1) % c.NumThreads
		}
        // Send the queues to the workers
		for i := 0; i < c.NumThreads; i++ {
			queueChannels[i] <- deques[i]
		}
	}()

	// to close the out channel once all the workers are done
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
	// done used as a flag to check if the worker has finished its tasks + stealing if needed
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

		// I have not implemented failed stealingretry- only one attempt at stealing once my own tasks are done
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
                    // leaving the variable stolenCount as is.Currenly useless. In an extended version, 
                    // i can use it to track the number of steals for retry logic
					stolenCount++
					stolen = true
					break
				}
			}
			
			if !stolen {
				done = true
			}
        // no work stealing
		} else {
			done = true
		}
	}
}

///////////////////////////////////////////////////////////////////////////
// WRITE //
///////////////////////////////////////////////////////////////////////////
func writeStage(in <-chan parsingSupport.Result, numWriters int) <-chan struct{} {
    // The final completion channel to end the program
	completed := make(chan struct{})

    // individual completion channels for each writer
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