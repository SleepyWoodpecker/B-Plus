package main

import (
	"cmp"
	"fmt"
)

var (
	ORDER                = 4 // each node has at most ORDER - 1 keys
	MAX_KEYS_PER_NODE    = ORDER - 1
	MAX_LEAF_POINTERS    = MAX_KEYS_PER_NODE
	LEAF_SPLIT_INDEX     = MAX_LEAF_POINTERS / 2
	MAX_NONLEAF_POINTERS = ORDER
)

// exceptions

type Tree[T cmp.Ordered] struct {
	Root *Node[T]
}

type Node[T cmp.Ordered] struct {
	IsLeaf bool

	// On a nonleaf node, this would be pointers to child nodes
	//
	// On a leaf node, this would be pointers to record values
	// Additionally, the final pointer is used to point to the next child in the line, from left to right
	//
	// The interface{} type helps this function as a void*
	Pointers []interface{}

	// use a generic []int type since you dont know how many keys there will be
	Keys    []T
	NumKeys int

	Parent *Node[T]
}

func NewTree[T cmp.Ordered]() *Tree[T] {
	return &Tree[T]{
		Root: nil,
	}
}

func NewNode[T cmp.Ordered]() *Node[T] {
	node := Node[T]{
		Pointers: make([]interface{}, ORDER),
		Keys:     make([]T, MAX_KEYS_PER_NODE),
		NumKeys:  0,
	}
	for i := range node.Pointers {
		node.Pointers[i] = nil
	}

	return &node
}

// insertion functions
func (t *Tree[T]) Insert(record Record[T]) {
	// set up an empty tree
	if t.Root == nil {
		t.Root = NewNode[T]()
		t.Root.IsLeaf = true
		t.Root.Parent = nil
	}

	nodeToInsertValue := t.FindNodeToInsertValue(record)
	indexToInsertVal := findInsertionIndex(nodeToInsertValue, record)

	if nodeToInsertValue.NumKeys < MAX_KEYS_PER_NODE {
		for i := nodeToInsertValue.NumKeys - 1; i >= indexToInsertVal; i-- {
			nodeToInsertValue.Keys[i+1] = nodeToInsertValue.Keys[i]
			nodeToInsertValue.Pointers[i+1] = nodeToInsertValue.Pointers[i]
		}

		nodeToInsertValue.Keys[indexToInsertVal] = record.GetHashableVal()
		nodeToInsertValue.Pointers[indexToInsertVal] = record
		nodeToInsertValue.NumKeys++
		return
	}

	// split the node
	tempKeys := make([]T, MAX_NONLEAF_POINTERS)
	tempPointers := make([]interface{}, MAX_NONLEAF_POINTERS)

	for i, j := 0, 0; i < MAX_NONLEAF_POINTERS; i++ {
		if i == indexToInsertVal {
			tempKeys[i] = record.GetHashableVal()
			tempPointers[i] = record
			continue
		}

		tempKeys[i] = nodeToInsertValue.Keys[j]
		tempPointers[i] = nodeToInsertValue.Pointers[j]
		j++
	}

	// put the keys and pointers into the original node
	for i := range MAX_LEAF_POINTERS {
		if i <= LEAF_SPLIT_INDEX {
			nodeToInsertValue.Keys[i] = tempKeys[i]
			nodeToInsertValue.Pointers[i] = tempPointers[i]
		} else {
			nodeToInsertValue.Pointers[i] = nil
			nodeToInsertValue.NumKeys--
		}
	}

	newNode := NewNode[T]()
	newNode.IsLeaf = true

	for i, j := 0, LEAF_SPLIT_INDEX+1; j < MAX_NONLEAF_POINTERS; i, j = i+1, j+1 {
		newNode.Keys[i] = tempKeys[j]
		newNode.Pointers[i] = tempPointers[j]
		newNode.NumKeys++
	}

	newRoot := NewNode[T]()
	newRoot.Keys[0] = newNode.Keys[0]
	newRoot.NumKeys++
	newRoot.IsLeaf = false

	nodeToInsertValue.Parent = newRoot
	newNode.Parent = newRoot
	newRoot.Pointers[0] = nodeToInsertValue
	newRoot.Pointers[1] = newNode

	t.Root = newRoot
}

func (t *Tree[T]) FindNodeToInsertValue(record Record[T]) *Node[T] {
	if t.Root == nil {
		panic("Tree is empty")
	}

	var currentNode *Node[T] = t.Root
	val := record.GetHashableVal()

	for !currentNode.IsLeaf {
		ptrIdx := currentNode.NumKeys
		// find the right key
		for i, key := range currentNode.Keys {
			if val < key {
				ptrIdx = i
				break
			}
		}
		if node, ok := currentNode.Pointers[ptrIdx].(*Node[T]); ok {
			currentNode = node
		} else {
			panic("Found a non node in a nonleaf pointer")
		}
	}

	return currentNode
}

// given a record and value, find the right place to insert the new value
// TODO: are we supposed to be able to do binary search here? I cant think of a way to do that
func findInsertionIndex[T cmp.Ordered](currentSearchNode *Node[T], record Record[T]) int {
	if !currentSearchNode.IsLeaf {
		panic("Cannot find insertion index for something that is not a child node")
	}

	if currentSearchNode.NumKeys == 0 {
		return 0
	}

	target := record.GetHashableVal()

	for i, key := range currentSearchNode.Keys {
		if target < key {
			return i
		}
	}

	return len(currentSearchNode.Keys)
}

// struct for printing
type nodeWithDepth[T cmp.Ordered] struct {
	*Node[T]
	Depth int
}

func (t *Tree[T]) String() string {
	treeString := ""

	if t.Root == nil {
		return ""
	}

	queue := make([]*nodeWithDepth[T], 1)
	queue[0] = &nodeWithDepth[T]{
		t.Root,
		0,
	}

	currentDepth := 0
	for len(queue) > 0 {
		top := queue[0]
		queue = queue[1:]

		if top.Depth > currentDepth {
			treeString += "\n"
			currentDepth = top.Depth
		} else if top.Depth < currentDepth {
			panic("Did not BFS")
		}

		// if not a leaf, print the keys
		if !top.IsLeaf {
			for i := range top.NumKeys {
				treeString += fmt.Sprintf("%v ", top.Keys[i])
			}

			// add children to the queue
			for _, child := range top.Pointers {
				if node, ok := child.(*Node[T]); ok {
					queue = append(
						queue,
						&nodeWithDepth[T]{
							node,
							top.Depth + 1,
						},
					)
				}
			}
		} else { // if is a leaf, print the values
			for i := range top.NumKeys {
				if formattedRecord, ok := top.Pointers[i].(Record[T]); ok {
					treeString += formattedRecord.String() + " "
				} else {
					fmt.Printf("Could not format into a record\n")
				}
			}
		}
		treeString += "|"
	}

	return treeString
}
