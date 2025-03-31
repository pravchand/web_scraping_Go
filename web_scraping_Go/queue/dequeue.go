package queue

import (
    "errors"
    "sync/atomic"
    "proj3/parsingSupport"
)

// Reference, Art of Multiprocessor Programming, section 16.5.1
// Converted to Go from Java

type BoundedDequeue struct {
    tasks    []parsingSupport.Task
    // using int64 which can be split into top and stamp
    head     atomic.Int64 
    tail     int64 
    capacity int64 
}

func NewBoundedDequeue(capacity int64) *BoundedDequeue {  
    return &BoundedDequeue{
        tasks:    make([]parsingSupport.Task, capacity),
        tail:     0,
        capacity: capacity,
    }
}

func (bd *BoundedDequeue) PushBottom(task parsingSupport.Task) error {
    if bd.tail >= bd.capacity {
        return errors.New("dequeue is full")
    }
    
    bd.tasks[bd.tail] = task
    bd.tail++
    return nil
}

func (bd *BoundedDequeue) PopBottom() (parsingSupport.Task, error) {
    if bd.tail == 0 {
        return parsingSupport.Task{}, errors.New("dequeue is empty")
    }
    
    bd.tail--
    t := bd.tasks[bd.tail]
    
    oldValue := bd.head.Load()

    // For example, if bd.head.Load()
    //  bd.head.Load() = 00010010 00110100 01010110 01111000 10101011 11001101 11101111 00010010
    // 0xFFFFFFFF     = 00000000 00000000 00000000 00000000 11111111 11111111 11111111 11111111
    // -------------------------------------------------------------------------------------------------
    // Result         = 00000000 00000000 00000000 00000000 10101011 11001101 11101111 00010010
    // we get the lower 32 bits

    oldTop := int64(oldValue & 0xFFFFFFFF)
    oldStamp := oldValue >> 32
    newTop := int64(0)
    newStamp := oldStamp + 1

    if bd.tail > oldTop {
        return t, nil
    }
    
    if bd.tail == oldTop {
        bd.tail = 0
        
        if bd.head.CompareAndSwap(oldValue, newStamp<<32|newTop) {
            return t, nil
        }
        return parsingSupport.Task{}, errors.New("failed to pop bottom")
    }
    bd.tail = 0
    bd.head.Store(newStamp<<32|newTop)
    return parsingSupport.Task{}, errors.New("failed to pop bottom")
}


func (bd *BoundedDequeue) PopTop() (parsingSupport.Task, error) {
    for {
        oldValue := bd.head.Load()
        oldTop := int64(oldValue & 0xFFFFFFFF) 
        oldStamp := oldValue >> 32
        
        if bd.tail <= oldTop {
            return parsingSupport.Task{}, errors.New("dequeue is empty")
        }
        
        t := bd.tasks[oldTop]
        
        newStamp := oldStamp + 1
        newTop := oldTop + 1
        
        // If someone has already taken this task, it will fail
        if bd.head.CompareAndSwap(oldValue, newStamp<<32|newTop) {
            return t, nil
        }
        
        // If swap fails, either:
        // 1. Another thread stole the task
        // 2. Queue state changed
        // So we retry or return empty- i have not implemented retry
        return parsingSupport.Task{}, errors.New("failed to pop top")
    }
}

func (bd *BoundedDequeue) Len() int64 {
    head := bd.head.Load() & 0xFFFFFFFF
    return bd.tail - head
}

func (bd *BoundedDequeue) IsEmpty() bool {
    head := bd.head.Load() & 0xFFFFFFFF
    return head >= bd.tail
}