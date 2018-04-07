package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// TYPES
////////

type Node interface {
	Checksum() []byte
	AsString(func([]byte) string, int) string
}

type Branch struct {
	checksum []byte
	left     Node
	right    Node
}

type Leaf struct {
	checksum []byte
	block    []byte
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// CONSTRUCTORS
///////////////

func NewLeaf(sum func([]byte) []byte, block []byte) *Leaf {
	return &Leaf{
		checksum: sum(block),
		block:    block,
	}
}

func NewBranch(sum func([]byte) []byte, left Node, right Node) *Branch {
	return &Branch{
		checksum: sum(append(left.Checksum(), right.Checksum()...)),
		left:     left,
		right:    right,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// METHODS
//////////

func (m *Branch) Checksum() []byte {
	return m.checksum
}

func (m *Branch) AsString(checksumToString func([]byte) string, n int) string {
	c := checksumToString(m.checksum)
	l := m.left.AsString(checksumToString, n+2)
	r := m.right.AsString(checksumToString, n+2)

	return fmt.Sprintf("\n"+indent(n, "(B root: %s %s %s)"), c, l, r)
}

func (m *Leaf) Checksum() []byte {
	return m.checksum
}

func (m *Leaf) AsString(f func([]byte) string, n int) string {
	return fmt.Sprintf("\n"+indent(n, "(L root: %s)"), f(m.checksum))
}

func CreateTree(sum func([]byte) []byte, blocks [][]byte) Node {
	levels := int(math.Ceil(math.Log2(float64(len(blocks)+len(blocks)%2))) + 1)

	// represents each row in the tree, where rows[0] is the base and rows[len(rows)-1] is the root
	rows := make([][]Node, levels)

	// build our base of leaves
	for i := 0; i < len(blocks); i++ {
		rows[0] = append(rows[0], NewLeaf(sum, blocks[i]))
	}

	// build upwards until we hit the root
	for i := 1; i < levels; i++ {
		prev := rows[i-1]

		// each iteration creates a branch from a pair of values originating from the previous level
		for j := 0; j < len(prev); j = j + 2 {
			var l, r Node

			// if we don't have enough to make a pair, duplicate the left
			if j+1 >= len(prev) {
				l = prev[j]
				r = l
			} else {
				l = prev[j]
				r = prev[j+1]
			}

			rows[i] = append(rows[i], NewBranch(sum, l, r))
		}
	}

	root := rows[len(rows)-1][0]

	return root
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// UTILITIES
////////////

func indent(spaces int, orig string) string {
	str := ""
	for i := 0; i < spaces; i++ {
		str += " "
	}

	return str + orig
}

func main() {
	stuff := [][]byte{[]byte("alpha"), []byte("beta"), []byte("kappa"), []byte("gamma"), []byte("epsilon"), []byte("omega")}

	doubleSha256 := func(data []byte) []byte {
		first := sha256.Sum256(data)
		secnd := sha256.Sum256(first[:])

		return secnd[:]
	}

	tree := CreateTree(doubleSha256, stuff)

	fmt.Println(tree.AsString(hex.EncodeToString, 0))
}
