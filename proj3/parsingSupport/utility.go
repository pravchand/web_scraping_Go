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

type Task struct{
	Url string
}

type Result struct{
	Url string
	Content []string
}	

func WriteResult(result Result, count int) {
    urlParts := ExtractFilenameFromURL(result.Url)
    dataPath := "/home/praveenc/project-3-pravchand/proj3/outputdata/" + urlParts + ".txt"
    
    os.MkdirAll("/home/praveenc/project-3-pravchand/proj3/outputdata", 0755)
    
    contentBytes := []byte(strings.Join(result.Content, "\n"))
    err := os.WriteFile(dataPath, contentBytes, 0644)
    if err != nil {
        fmt.Printf("ERROR: Could not write to file %s: %v", dataPath, err)
		return
	}
	
}