package queue

import (
    "errors"
    "sync/atomic"
    "proj3/parsingSupport"
)



type BoundedDequeue struct {
    tasks    []parsingSupport.Task
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
    oldTop := int64(oldValue & 0xFFFFFFFF)
    oldStamp := oldValue >> 32
    newTop := int64(0)
    newStamp := oldStamp + 1
    if bd.tail > oldTop {
        return t, nil
    }
    
    if bd.tail == oldTop {
        // newTail := bd.tail 
        bd.tail = 0
        
        if bd.head.CompareAndSwap(oldValue, newStamp<<32|newTop) {
            return t, nil
        }
        // bd.tail = newTail + 1
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
        
        // If someone has already stolen this task, it will fail
        if bd.head.CompareAndSwap(oldValue, newStamp<<32|newTop) {
            return t, nil
        }
        
        // If swap fails, either:
        // 1. Another thread stole the task
        // 2. Queue state changed
        // So we retry or return empty
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