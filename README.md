# Web Scraping and Tokenizing Pipeline

## Author
Praveen Chandar Devarajan

## Description
This project implements an embarrassingly parallel web scraping and tokenizing system. The program is designed to fetch, parse, tokenize, and store transcripts of "The Seen and the Unseen" podcast by Amit Varma.

The system:
- Retrieves HTML pages from URLs
- Extracts textual content from specific HTML tags
- Tokenizes the text into individual words
- Writes the results to disk

The main goal is to improve performance by parallelizing these tasks and implementing a work stealing algorithm to address load balancing between worker threads.

## Technical Challenges
The problem involves several resource-intensive operations:
- Network I/O (fetching HTML content)
- Computational processing (parsing and tokenizing)
- Disk I/O (saving text files)

## Implementation Details
The solution uses Go's concurrency features to implement:
- Pipeline pattern for data processing
- Work-stealing algorithm for load balancing
- Bounded dequeues for task management
- Multiple workers for parallel processing

## References

### Pipeline Implementation
- [The Pipeline Pattern in Go](https://dev.to/johnscode/the-pipeline-pattern-in-go-2bho)
- [Go Pipeline for a Layman](https://anupamgogoi.medium.com/go-pipeline-for-a-layman-4791fb4f1e2d)
- [YouTube: Pipeline Pattern](https://www.youtube.com/watch?v=8Rn8yOQH62k)

### Go Channels
- [Introduction to Channels in Go](https://www.youtube.com/watch?v=nNXhePi3xwE)

## Contact
For the helper packages or additional information, contact: mahara1995@gmail.com
