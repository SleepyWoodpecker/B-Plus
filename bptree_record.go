package main

import (
	"cmp"
	"strconv"
)

type Record[T cmp.Ordered] interface {
	// get value to sort the record by
	GetHashableVal() T
	String() string
}

type NumRecord struct {
	Value int
}

func NewIntRecord(val int) *NumRecord {
	return &NumRecord{
		Value: val,
	}
}

func (n *NumRecord) GetHashableVal() int {
	return n.Value
}

func (n *NumRecord) String() string {
	return strconv.FormatInt(int64(n.Value), 10)
}

type Iterator[T cmp.Ordered] interface {
	// bool field tells you if there is a next value
	Next() Record[T]
}

type NumIntRecordIterator[T cmp.Ordered] struct {
	// give a *Node[T] and int so that you dont run into the edge case where the end record is nil
	IteratorEnd    *Node[T]
	IteratorEndIdx int

	CurrentIdx   int
	CurrentNode  *Node[T]
	isFirstEntry bool
	takeLast     bool
}

func (n *NumIntRecordIterator[T]) Next() Record[T] {
	if !n.isFirstEntry {
		n.CurrentIdx++
	} else {
		n.isFirstEntry = false
	}
	// if you have reached the next pointer
	if n.CurrentIdx == n.CurrentNode.NumKeys && n.CurrentNode != n.IteratorEnd {
		if node, ok := n.CurrentNode.Pointers[MAX_LEAF_POINTERS].(*Node[T]); ok {
			n.CurrentNode = node
			n.CurrentIdx = 0
		} else if n.CurrentNode.Pointers[MAX_LEAF_POINTERS] == nil { // this is in the case that the end of the tree's leaves was reached
			return nil
		}
	}

	if n.CurrentNode != n.IteratorEnd || n.IteratorEndIdx != n.CurrentIdx {
		if r, ok := n.CurrentNode.Pointers[n.CurrentIdx].(Record[T]); ok {
			return r
		} else {
			panic("Could not cast into a record")
		}
	} else {
		return nil
	}
}
