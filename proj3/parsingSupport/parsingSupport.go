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

func ExtractFilenameFromURL(url string) string {
    // Extract just the filename part (the number) from the URL
    // For example, from "https://saket-choudhary.me/seenunseencap/42.html"
    // extract "42"
    
    // Find the last slash and the dot before extension
    lastSlash := strings.LastIndex(url, "/")
    lastDot := strings.LastIndex(url, ".")
    
    return url[lastSlash+1:lastDot]
}

