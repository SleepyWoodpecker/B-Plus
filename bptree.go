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
	MIN_NONLEAF_KEYS     = MAX_KEYS_PER_NODE / 2
	MIN_LEAF_KEYS        = ORDER / 2
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
		IsLeaf:   false,
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

	nodeToInsertValue := t.findNode(record.GetHashableVal())
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

	// make the old node point to the new node with the last pointer
	// this will help support range queries
	nodeToInsertValue.Pointers[MAX_LEAF_POINTERS] = newNode

	for i, j := 0, LEAF_SPLIT_INDEX+1; j < MAX_NONLEAF_POINTERS; i, j = i+1, j+1 {
		newNode.Keys[i] = tempKeys[j]
		newNode.Pointers[i] = tempPointers[j]
		newNode.NumKeys++
	}

	if record, ok := newNode.Pointers[0].(Record[T]); ok {
		t.insertIntoParentNode(nodeToInsertValue, newNode, nodeToInsertValue.Parent, record.GetHashableVal())
	} else {
		panic("No hashable value for the new node")
	}
}

// assume that left is the original node that was not split before this
func (t *Tree[T]) insertIntoParentNode(left *Node[T], right *Node[T], parent *Node[T], separator T) {
	// since left was the original node, if it does not have a parent node, it must be the original root
	if parent == nil {
		newRoot := NewNode[T]()
		newRoot.Keys[0] = separator
		newRoot.NumKeys++
		newRoot.IsLeaf = false

		left.Parent = newRoot
		right.Parent = newRoot

		newRoot.Pointers[0] = left
		newRoot.Pointers[1] = right

		t.Root = newRoot
		return
	}

	// find index of left node
	// if the number keys has not yet been maxed, the index of the key to modify is idx
	foundIdx := t.getNodeIndexInParent(left, parent)
	if foundIdx == -1 {
		panic("Could not find node idx")
	}

	indexToInsertNewNode := foundIdx + 1
	if parent.NumKeys < MAX_KEYS_PER_NODE {
		// copy all the keys over
		for i := parent.NumKeys; i >= indexToInsertNewNode; i-- {
			parent.Keys[i] = parent.Keys[i-1]
			parent.Pointers[i+1] = parent.Pointers[i]
		}

		parent.Keys[indexToInsertNewNode-1] = separator
		parent.Pointers[indexToInsertNewNode] = right
		parent.NumKeys++

		right.Parent = parent

		return
	}

	// if not, split the parent node
	// when trying to split a nonleaf node, there will be one more pointer than key
	tempKeys := make([]T, MAX_NONLEAF_POINTERS)
	tempPointers := make([]interface{}, MAX_NONLEAF_POINTERS+1)

	for i, j := 0, 0; i < MAX_NONLEAF_POINTERS+1; i++ {
		if i == indexToInsertNewNode {
			tempPointers[i] = right
			continue
		}
		tempPointers[i] = parent.Pointers[j]
		j++
	}

	for i, j := 0, 0; i < MAX_NONLEAF_POINTERS; i++ {
		if i == indexToInsertNewNode-1 {
			if record, ok := right.Pointers[0].(Record[T]); ok {
				tempKeys[i] = record.GetHashableVal()
				continue
			}
		}

		tempKeys[i] = parent.Keys[j]
		j++
	}

	for i := range MAX_LEAF_POINTERS {
		if i < LEAF_SPLIT_INDEX {
			parent.Keys[i] = tempKeys[i]
			parent.Pointers[i] = tempPointers[i]
		} else {
			parent.Pointers[i] = nil
			parent.NumKeys--
		}
	}
	parent.Pointers[LEAF_SPLIT_INDEX] = tempPointers[LEAF_SPLIT_INDEX]
	nodeSeparator := tempKeys[LEAF_SPLIT_INDEX]

	newNode := NewNode[T]()
	for i, j := 0, LEAF_SPLIT_INDEX+1; j < MAX_NONLEAF_POINTERS+1; i, j = i+1, j+1 {
		if j < MAX_NONLEAF_POINTERS {
			newNode.Keys[i] = tempKeys[j]
			newNode.NumKeys++
		}

		// some issue abuout not setting parent
		newNode.Pointers[i] = tempPointers[j]
		if nn, ok := newNode.Pointers[i].(*Node[T]); ok {
			nn.Parent = newNode
		}
	}

	t.insertIntoParentNode(parent, newNode, parent.Parent, nodeSeparator)
}

func (t *Tree[T]) getNodeIndexInParent(node *Node[T], parent *Node[T]) int {
	for i, ptr := range parent.Pointers {
		if ptr == node {
			return i
		}
	}

	return -1
}

// Find node that would contain the desired value
// This does not guarantee that the value is found, only that the desired node is found
func (t *Tree[T]) findNode(val T) *Node[T] {
	if t.Root == nil {
		panic("Tree is empty")
	}

	var currentNode *Node[T] = t.Root

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
			panic(fmt.Sprintf("Found a non node in a nonleaf pointer: %+v, \n\n%s\n", currentNode, t))
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

	return currentSearchNode.NumKeys
}

// function to search for an item using equality
func (t *Tree[T]) FindPoint(val T) Record[T] {
	targetNode := t.findNode(val)
	record, _ := findItemIndex(targetNode, val)
	return record
}

// if there is a match with the item, return the associated record
// if there is no match, return nil
func findItemIndex[T cmp.Ordered](currentNode *Node[T], val T) (Record[T], int) {
	if !currentNode.IsLeaf {
		panic("Cannot find insertion index for something that is not a child node")
	}

	for i, ptr := range currentNode.Pointers {
		if record, ok := ptr.(Record[T]); ok && record.GetHashableVal() == val {
			return record, i
		}
	}

	return nil, -1
}

func (t *Tree[T]) Delete(val T) bool {
	// first confirm that the desired value exists
	targetNode := t.findNode(val)

	// if the value exists, locate its current node,
	// find the index of the record in the node and remove the value from the node

	recordToDelete, recordToDeleteIdxInNode := findItemIndex(targetNode, val)
	if recordToDelete == nil {
		return false
	}
	removeKeyAndPointerFromLeaf(targetNode, recordToDeleteIdxInNode)

	if t.Root == targetNode {
		return true
	}

	if targetNode.NumKeys >= MIN_LEAF_KEYS {
		return true
	}

	targetNodeIdxInParent := -1
	// in a non-leaf node, the number of pointers is equal to numKeys + 1
	for i := 0; i < targetNode.Parent.NumKeys+1; i++ {
		if targetNode.Parent.Pointers[i] == targetNode {
			targetNodeIdxInParent = i
			break
		}
	}

	if targetNodeIdxInParent == -1 {
		panic("Could not find index of node in parent")
	}

	deleteCleanup(targetNode, targetNodeIdxInParent)
	return true
}

func deleteFromNonLeaf[T cmp.Ordered](targetNode *Node[T], targetNodeIdxInParent int) {
	removeKeyAndPointerFromNonLeaf(targetNode, targetNodeIdxInParent)

	if targetNode.NumKeys >= MIN_NONLEAF_KEYS {
		return
	}

	deleteCleanup(targetNode, targetNodeIdxInParent)
}

func deleteCleanup[T cmp.Ordered](targetNode *Node[T], targetNodeIdxInParent int) {
	var neighborNodeIdx, separatorKeyIdx int
	var separator T
	var neighborNode *Node[T]

	if targetNodeIdxInParent != 0 {
		neighborNodeIdx = targetNodeIdxInParent - 1
		separatorKeyIdx = neighborNodeIdx
	} else {
		neighborNodeIdx = 1
		separatorKeyIdx = 0
	}

	if nbn, ok := targetNode.Parent.Pointers[neighborNodeIdx].(*Node[T]); ok {
		neighborNode = nbn
		separator = targetNode.Parent.Keys[separatorKeyIdx]
	} else {
		panic(fmt.Sprintf("Neighbor node was invalid: %T", targetNode.Parent.Pointers[neighborNodeIdx]))
	}

	if targetNode.NumKeys+neighborNode.NumKeys <= MAX_KEYS_PER_NODE {
		if targetNodeIdxInParent != 0 {
			coalesce(neighborNode, targetNode, targetNodeIdxInParent, targetNode.Parent, separator)
		} else {
			coalesce(targetNode, neighborNode, neighborNodeIdx, targetNode.Parent, separator)
		}

		return
	}

	redistributeNodes[T]()
}

func removeKeyAndPointerFromLeaf[T cmp.Ordered](node *Node[T], recordToDeleteIdx int) {
	for i := recordToDeleteIdx; i < node.NumKeys-1; i++ {
		node.Keys[i] = node.Keys[i+1]
		node.Pointers[i] = node.Pointers[i+1]
	}

	node.NumKeys--
}

func removeKeyAndPointerFromNonLeaf[T cmp.Ordered](node *Node[T], targetNodeIdxInParent int) {
	// stop at NumKeys here since targetNodeIdx is the pointer index and the total number of pointers == node.NumKeys
	for i := targetNodeIdxInParent; i < node.NumKeys; i++ {
		node.Keys[i-1] = node.Keys[i]
		node.Pointers[i] = node.Pointers[i+1]
	}

	node.NumKeys--
}

func coalesce[T cmp.Ordered](left *Node[T], right *Node[T], rightIdx int, parent *Node[T], separator T) {
	// move all records that were in the right node into the left node

	// if it was a leaf, just copy directly
	if left.IsLeaf {
		for i, j := left.NumKeys, 0; j < right.NumKeys; i, j = i+1, j+1 {
			left.Keys[i] = right.Keys[j]
			left.Pointers[i] = right.Pointers[j]
			left.NumKeys++
		}
	} else {
		left.Keys[left.NumKeys] = separator
		left.NumKeys++
		for i, j := left.NumKeys, 0; j < right.NumKeys; i, j = i+1, j+1 {
			left.Keys[i] = right.Keys[j]
			left.Pointers[i] = right.Pointers[j]

			// adjust them all to point to the same parent
			if l, ok := left.Pointers[i].(*Node[T]); ok {
				l.Parent = left.Parent
			} else {
				panic("Did not insert a node")
			}

			left.NumKeys++
		}

		// copy over the final pointer
		left.Pointers[left.NumKeys] = right.Pointers[right.NumKeys]

	}

	// now remove the key from the top
	deleteFromNonLeaf(parent, rightIdx)
}

func redistributeNodes[T cmp.Ordered]() {}

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
			for i := range top.NumKeys + 1 {
				if i < top.NumKeys {
					treeString += fmt.Sprintf("%v ", top.Keys[i])

				}
				if childNode, ok := top.Pointers[i].(*Node[T]); ok {
					queue = append(
						queue,
						&nodeWithDepth[T]{
							childNode,
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
