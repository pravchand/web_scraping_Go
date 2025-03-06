package main

import (
	"fmt"
	"os"
	"strconv"
	"proj3/parsingSupport"
	"proj3/tokenparse"
	)

const usage = `Usage: parse <number of links to parse> <number of threads> <work stealing (true/false)>
    <number of threads> = the number of goroutines to be part of the parallel version.`

func main() {
	if len(os.Args) < 2 || len(os.Args) > 4 {
		fmt.Println(usage)
		return
	}
	
	var c parsingSupport.Config
	numLinks, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Error: Invalid number of links")
		fmt.Println(usage)
		return
	}
	c.NumLinks = numLinks

	if len(os.Args) == 4 {
		c.Mode = "p"
		threads, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Error: Invalid number of threads")
			fmt.Println(usage)
			return
		}
		c.NumThreads = threads
		workStealing, err := strconv.ParseBool(os.Args[3])
		if err != nil {
			fmt.Println("Error: Invalid work stealing")
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
