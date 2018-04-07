package merkletree

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// TYPES
////////

type Tree struct {
	root         Node
	leaves       []*Leaf
	checksumFunc func([]byte) []byte
}

type Node interface {
	GetChecksum() []byte
	ToString(func([]byte) string, int) string
	SetParent(*Branch)
	GetParent() *Branch
}

type Branch struct {
	checksum []byte
	left     Node
	right    Node
	parent   *Branch
}

type Leaf struct {
	checksum []byte
	block    []byte
	parent   *Branch
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// CONSTRUCTORS
///////////////

func NewLeaf(sumFunc func([]byte) []byte, block []byte, parent *Branch) *Leaf {
	return &Leaf{
		checksum: sumFunc(block),
		block:    block,
		parent:   parent,
	}
}

func NewBranch(sumFunc func([]byte) []byte, left Node, right Node, parent *Branch) *Branch {
	return &Branch{
		checksum: sumFunc(append(left.GetChecksum(), right.GetChecksum()...)),
		left:     left,
		right:    right,
		parent:   parent,
	}
}

func NewTree(sumFunc func([]byte) []byte, blocks [][]byte) *Tree {
	levels := int(math.Ceil(math.Log2(float64(len(blocks)+len(blocks)%2))) + 1)

	// represents each row in the tree, where rows[0] is the base and rows[len(rows)-1] is the root
	rows := make([][]Node, levels)

	// build our base of leaves
	for i := 0; i < len(blocks); i++ {
		rows[0] = append(rows[0], NewLeaf(sumFunc, blocks[i], nil))
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

			// yuck
			b := NewBranch(sumFunc, l, r, nil)
			l.SetParent(b)
			r.SetParent(b)

			rows[i] = append(rows[i], b)
		}
	}

	leaves := make([]*Leaf, len(rows[0]))
	for i := 0; i < len(leaves); i++ {
		leaves[i] = rows[0][i].(*Leaf)
	}

	return &Tree{
		checksumFunc: sumFunc,
		leaves:       leaves,
		root:         rows[len(rows)-1][0],
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// METHODS
//////////

func (m *Branch) SetParent(parent *Branch) {
	m.parent = parent
}

func (m *Branch) GetParent() *Branch {
	return m.parent
}

func (m *Branch) GetChecksum() []byte {
	return m.checksum
}

func (m *Branch) ToString(checksumToString func([]byte) string, n int) string {
	c := checksumToString(m.checksum)
	l := m.left.ToString(checksumToString, n+2)
	r := m.right.ToString(checksumToString, n+2)

	return fmt.Sprintf("\n"+indent(n, "(B root: %s %s %s)"), c, l, r)
}

func (m *Leaf) SetParent(parent *Branch) {
	m.parent = parent
}

func (m *Leaf) GetParent() *Branch {
	return m.parent
}

func (m *Leaf) GetChecksum() []byte {
	return m.checksum
}

func (m *Leaf) ToString(f func([]byte) string, n int) string {
	return fmt.Sprintf("\n"+indent(n, "(L root: %s)"), f(m.checksum))
}

type PathPart struct {
	left  []byte
	right []byte
}

func (t *Tree) GetProofForDisplay(rootChecksum []byte, leafChecksum []byte, f func([]byte) string) string {
	var lines []string

	parts := t.AuditPath(rootChecksum, leafChecksum)
	if len(parts) == 0 {
		return "" // checksums don't match up with receiver
	}

	for _, part := range parts {
		l := f(part.left)
		r := f(part.right)
		c := f(t.checksumFunc(append(part.left, part.right...)))
		lines = append(lines, fmt.Sprintf("%s + %s = %s", l, r, c))
	}

	return strings.Join(lines, "\n")
}

// AuditPath returns the path from leaf to root.
func (t *Tree) AuditPath(rootChecksum []byte, leafChecksum []byte) []PathPart {
	var pathParts []PathPart

	if !bytes.Equal(rootChecksum, t.root.GetChecksum()) {
		return pathParts
	}

	found := t.getLeafByChecksum(leafChecksum)
	if found == nil {
		return pathParts
	}

	// start with the immediate parent of the target checksum
	fparent := found.GetParent
	if bytes.Equal(leafChecksum, fparent().left.GetChecksum()) {
		pathParts = append(pathParts, PathPart{
			left:  leafChecksum,
			right: fparent().right.GetChecksum(),
		})
	} else if bytes.Equal(leafChecksum, fparent().right.GetChecksum()) {
		pathParts = append(pathParts, PathPart{
			left:  fparent().left.GetChecksum(),
			right: leafChecksum,
		})
	}

	// once we've computed the checksum of the target + its sibling, work our way up towards the root
	gparent := fparent().GetParent()
	for gparent != nil {
		h := t.checksumFunc(append(pathParts[len(pathParts)-1].left, pathParts[len(pathParts)-1].right...))

		if bytes.Equal(h, gparent.left.GetChecksum()) {
			pathParts = append(pathParts, PathPart{
				left:  h,
				right: gparent.right.GetChecksum(),
			})
		} else if bytes.Equal(h, gparent.right.GetChecksum()) {
			pathParts = append(pathParts, PathPart{
				left:  gparent.left.GetChecksum(),
				right: h,
			})
		}

		gparent = gparent.parent
	}

	return pathParts
}

func (t *Tree) ToString(f func([]byte) string, n int) string {
	return t.root.ToString(f, n)
}

func (t *Tree) getLeafByChecksum(checksum []byte) *Leaf {
	var found *Leaf = nil

	for i := 0; i < len(t.leaves); i++ {
		if bytes.Equal(checksum, t.leaves[i].GetChecksum()) {
			found = t.leaves[i]
		}
	}

	return found
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// HELPERS
//////////

func indent(spaces int, orig string) string {
	str := ""
	for i := 0; i < spaces; i++ {
		str += " "
	}

	return str + orig
}
