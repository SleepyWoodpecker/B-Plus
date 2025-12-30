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
