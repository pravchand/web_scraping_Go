package parsingSupport

import (
	"strings"
)

func ParseAndTokenize(html string) []string {
	startTag := "<div class='text'>"
	endTag := "</div>"
	currentIndex := 0
	tokenizedTexts := []string{}
	for {
		startIndex := strings.Index(html[currentIndex:], startTag)
		if startIndex == -1 {
			break
		}
		startIndex += currentIndex 
		
		endIndex := strings.Index(html[startIndex:], endTag)
		// my HTML has multiple text classes, hence i will not ever traverse till the last line
		if endIndex == -1 {
			break 
		}
		endIndex += startIndex 

		textStart := startIndex + len(startTag)
		text := html[textStart:endIndex]
		currentIndex = endIndex + len(endTag) 
		tokenizedTexts = append(tokenizedTexts, strings.Split(text, " ")...)
	}

	return tokenizedTexts
}

// https://pkg.go.dev/strings#LastIndex
func ExtractFilenameFromURL(url string) string {
    // Extract just the filename part (the number) from the URL
    // For example, from "https://saket-choudhary.me/seenunseencap/42.html",extract "42"
    
    slash := strings.LastIndex(url, "/")
    dot := strings.LastIndex(url, ".")
    
    return url[slash+1:dot]
}

