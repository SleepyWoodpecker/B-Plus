package main

import (
	"reflect"
	"testing"
)

func TestTreeInsert(t *testing.T) {

	var tests = []struct {
		input  []int
		output string
	}{
		{[]int{3, 2, 1}, "1 2 3 |"},
		{[]int{3, 2, 1, 4}, "3 |\n1 2 |3 4 |"},
		{[]int{1, 2, 3, 4, 5, 6}, "3 5 |\n1 2 |3 4 |5 6 |"},
		{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, "5 |\n3 |7 9 |\n1 2 |3 4 |5 6 |7 8 |9 10 |"},
		{[]int{10, 4, 5, 7, 8, 1, 2, 6, 3, 9, 11, 12}, "7 |\n4 |9 11 |\n1 2 3 |4 5 6 |7 8 |9 10 |11 12 |"},
	}

	for _, test := range tests {
		tree := NewTree[int]()

		for _, input := range test.input {
			tree.Insert(
				NewIntRecord(input),
			)
		}

		stringRep := tree.String()
		if stringRep != test.output {
			t.Errorf("Format incorrect:\ngot:\n%s\nexpected:\n%s\n", stringRep, test.output)
		}
	}
}

// test point and range lookups in the tree
func TestTreeLookup(t *testing.T) {
	tree := NewTree[int]()
	vals := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, val := range vals {
		tree.Insert(NewIntRecord(val))
	}

	var tests = []struct {
		val      Record[int]
		isNotNil bool
	}{
		{NewIntRecord(10), true},
		{NewIntRecord(11), false},
	}

	for _, test := range tests {
		record := tree.FindPoint(test.val.GetHashableVal())

		if test.isNotNil {
			if record == nil {
				t.Errorf("Expected query to be non-nil, but got nil\n")
			} else if !reflect.DeepEqual(test.val, record) {
				t.Errorf("Expected: %+v, Got: %+v\n", test.val, record)
			}
		} else if !test.isNotNil && record != nil {
			t.Errorf("Expected query to be nil, but returned nil")
		}
	}
}

func TestTreeDeletion(t *testing.T) {
	tree := NewTree[int]()
	vals := []int{10, 4, 5, 7, 8, 1, 2, 6, 3, 9, 11, 12}
	for _, val := range vals {
		tree.Insert(NewIntRecord(val))
	}

	tests := []struct {
		toDelete     int
		expectedTree string
		exists       bool
	}{
		{2, "7 |\n4 |9 11 |\n1 3 |4 5 6 |7 8 |9 10 |11 12 |", true},
		{67, "7 |\n4 |9 11 |\n1 3 |4 5 6 |7 8 |9 10 |11 12 |", false},
		{11, "7 |\n4 |9 |\n1 3 |4 5 6 |7 8 |9 10 12 |", true},
		{1, "7 |\n5 | 9 |\n3 4 | 5 6 | 7 8 | 9 10 12 |", true},
		{3, "7 9 |\n4 5 6 | 7 8 | 9 10 12 |", true},
	}

	for _, test := range tests {
		recordToDelete := NewIntRecord(test.toDelete)
		result, exists := tree.Delete(recordToDelete.GetHashableVal())

		if exists != test.exists {
			t.Errorf("Expected target's existence to be: %v, but got: %v, when node to delete was: %v\n", test.exists, exists, recordToDelete)
		} else if test.exists && !reflect.DeepEqual(result, recordToDelete) {
			t.Errorf("Expected result to be: %+v, but actual result was %+v\n", recordToDelete, result)
		} else if tree.String() != test.expectedTree {
			t.Errorf("Format incorrect:\ngot:\n%s\nexpected:\n%s\n", tree.String(), test.expectedTree)
		}
	}
}

// for now, the tests only run on the leaf node
func TestFindInsertionIndex(t *testing.T) {
	customTree := NewTree[int]()

	// insert dummy value that will not be used
	customTree.Insert(NewIntRecord(1))

	tests := []struct {
		keys           []int
		recordToInsert Record[int]
		expectedIndex  int
	}{
		{[]int{3, 7, 16}, NewIntRecord(1), 0},
		{[]int{3, 7, 16}, NewIntRecord(3), 1},
		{[]int{3, 7, 16}, NewIntRecord(8), 2},
		{[]int{3, 7, 16}, NewIntRecord(17), 3},
	}

	for _, test := range tests {
		customTree.Root.Keys = test.keys
		customTree.Root.NumKeys = len(customTree.Root.Keys)

		insertionIndex := findInsertionIndex(customTree.Root, test.recordToInsert)
		if insertionIndex != test.expectedIndex {
			t.Errorf("Did not find the correct insertion node in pointers")
		}
	}

}
