package main

import (
	"testing"
)

func TestTree(t *testing.T) {

	var tests = []struct {
		input  []int
		output string
	}{
		{[]int{3, 2}, "2 3 |"},
		{[]int{3, 2, 1}, "1 2 3 |"},
		{[]int{3, 2, 1, 4}, "3 |\n1 2 |3 4 |"},
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
