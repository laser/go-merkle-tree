package merkletree

import (
	"bytes"
	"fmt"
	"gx/ipfs/QmVmDhyTTUcQXFD1rRQ64fGLMSAoaQvNH3hwuaCFAPq2hy/errors"
	"math"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// TYPES
////////

type Tree struct {
	root         Node
	rows         [][]Node
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

type ProofPart struct {
	isRight  bool
	checksum []byte
}

type Proof struct {
	parts  []*ProofPart
	target []byte
	root   []byte
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

	return &Tree{
		checksumFunc: sumFunc,
		rows:         rows,
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

func (m Proof) ToString(f func([]byte) string) string {
	var lines []string

	parts := m.parts
	if len(parts) == 0 {
		return "" // checksums don't match up with receiver
	}

	lines = append(lines, fmt.Sprintf("route from %s (leaf) to %s (root):", f(m.target), f(m.root)))
	lines = append(lines, "")

	var prev = m.target
	var curr []byte
	for i := 0; i < len(parts); i++ {
		if parts[i].isRight {
			curr = append(prev, parts[i].checksum...)
			lines = append(lines, fmt.Sprintf("%s + %s = %s", f(prev), f(parts[i].checksum), f(curr)))
		} else {
			curr = append(parts[i].checksum, prev...)
			lines = append(lines, fmt.Sprintf("%s + %s = %s", f(parts[i].checksum), f(prev), f(curr)))
		}
		prev = curr
	}

	return strings.Join(lines, "\n")
}

func (t *Tree) GetProof(rootChecksum []byte, leafChecksum []byte) (*Proof, error) {
	var parts []*ProofPart

	if !bytes.Equal(rootChecksum, t.root.GetChecksum()) {
		return nil, errors.New("root checksums don't match")
	}

	index := -1
	for i := 0; i < len(t.rows[0]); i++ {
		if bytes.Equal(leafChecksum, t.rows[0][i].GetChecksum()) {
			index = i
			break
		}
	}

	if index == -1 {
		return nil, errors.New("target not found in receiver")
	}

	for i := 0; i < len(t.rows)-1; i++ {
		if index%2 == 1 {
			// is right, so go back one to get left
			parts = append(parts, &ProofPart{
				isRight:  false,
				checksum: t.rows[i][index-1].GetChecksum(),
			})
		} else {
			var checksum []byte
			if (index + 1) < len(t.rows[i]) {
				checksum = t.rows[i][index+1].GetChecksum()
			} else {
				checksum = t.rows[i][index].GetChecksum()
			}

			// is left, so go one forward to get hash pair
			parts = append(parts, &ProofPart{
				isRight:  true,
				checksum: checksum,
			})
		}

		index = int(math.Floor(float64(index / 2)))
	}

	return &Proof{
		parts:  parts,
		target: leafChecksum,
		root:   rootChecksum,
	}, nil
}

func (t *Tree) ToString(f func([]byte) string, n int) string {
	return t.root.ToString(f, n)
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
