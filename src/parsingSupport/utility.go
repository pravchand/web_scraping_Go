package parsingSupport

import (
	"fmt"
	"os"
	"strings"
)

type Config struct{
	Mode string
	NumLinks int
	NumThreads int
	WorkStealing bool
}

// this struct can be expanded in the future to hold other task specific things
type Task struct{
	Url string
}

type Result struct{
	Url string
	Content []string
}	

// https://pkg.go.dev/os#WriteFile
func WriteResult(result Result, count int) {
    urlParts := ExtractFilenameFromURL(result.Url)
    dataPath := "/home/praveenc/project-3-pravchand/proj3/outputdata/" + urlParts + ".txt"
    
    os.MkdirAll("/home/praveenc/project-3-pravchand/proj3/outputdata", 0755)
    
    contentBytes := []byte(strings.Join(result.Content, "\n"))
    err := os.WriteFile(dataPath, contentBytes, 0644)
    if err != nil {
        fmt.Printf("ERROR: Could not write to file %v", err)
		return
	}
	
}