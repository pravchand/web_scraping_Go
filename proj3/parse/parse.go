package main

// the main interface for the workflow
// use it like "go run parse.go 100 2 true" - 100 links, 2 threads and workstealing
// "go run parse.go 100" - 100 links sequential

import (
	"fmt"
	"os"
	"strconv"
	"proj3/parsingSupport"
	"proj3/tokenparse"
	)

const usage = `Usage: parse <number of links to parse> <number of threads> <work stealing (true/false)>
    <number of threads> = the number of goroutines that will parallely scrape and store.`

func main() {
	if len(os.Args) < 2 || len(os.Args) > 4 {
		fmt.Println(usage)
		return
	}
	
	var c parsingSupport.Config
	// The podcast website has only 400 links. Hence preventing excess input
	numLinks, err := strconv.Atoi(os.Args[1])
	if err != nil || numLinks > 400{
		fmt.Println("Error: Invalid number of links or above 400 links entered")
		fmt.Println(usage)
		return
	}
	c.NumLinks = numLinks

	if len(os.Args) == 4 {
		c.Mode = "p"
		threads, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid number of threads")
			fmt.Println(usage)
			return
		}
		c.NumThreads = threads
		workStealing, err := strconv.ParseBool(os.Args[3])
		if err != nil {
			fmt.Println("Invalid work stealing")
			fmt.Println(usage)
			return
		}
		c.WorkStealing = workStealing
	} else {
		c.Mode = "s"
		c.NumThreads = 1
	}

	tokenparse.Run(c)
}
