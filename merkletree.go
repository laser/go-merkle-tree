package merkletree

import (
	"bytes"
	"gx/ipfs/QmVmDhyTTUcQXFD1rRQ64fGLMSAoaQvNH3hwuaCFAPq2hy/errors"
	"math"
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
	ToString(checksumToStrFunc, int) string
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

type ProofPart struct {
	isRight  bool
	checksum []byte
}

type Proof struct {
	parts  []*ProofPart
	target []byte
	root   []byte
}

type checksumToStrFunc func([]byte) string

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// CONSTRUCTORS
///////////////

func NewLeaf(sumFunc func([]byte) []byte, block []byte) *Leaf {
	return &Leaf{
		checksum: sumFunc(block),
		block:    block,
	}
}

func NewBranch(sumFunc func([]byte) []byte, left Node, right Node) *Branch {
	return &Branch{
		checksum: sumFunc(append(left.GetChecksum(), right.GetChecksum()...)),
		left:     left,
		right:    right,
	}
}

func NewTree(sumFunc func([]byte) []byte, blocks [][]byte) *Tree {
	levels := int(math.Ceil(math.Log2(float64(len(blocks)+len(blocks)%2))) + 1)

	// represents each row in the tree, where rows[0] is the base and rows[len(rows)-1] is the root
	rows := make([][]Node, levels)

	// build our base of leaves
	for i := 0; i < len(blocks); i++ {
		rows[0] = append(rows[0], NewLeaf(sumFunc, blocks[i]))
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
			b := NewBranch(sumFunc, l, r)

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

func (b *Branch) GetChecksum() []byte {
	return b.checksum
}

func (l *Leaf) GetChecksum() []byte {
	return l.checksum
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
