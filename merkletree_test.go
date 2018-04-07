package merkletree

import (
	"fmt"
	"strings"
	"testing"
)

func trimNewlines(str string) string {
	return strings.Trim(str, "\n")
}

func expectStrEqual(t *testing.T, actual string, expected string) {
	if trimNewlines(actual) != expected {
		fmt.Println(fmt.Sprintf("=====ACTUAL======\n\n%s\n\n=====EXPECTED======\n\n%s\n", actual, expected))
		t.Fail()
	}
}

func str(strbytes []byte) string {
	return string(strbytes)
}

var givenOneBlock = trimNewlines(`
(B root: alphaalpha 
  (L root: alpha) 
  (L root: alpha))
`)

var givenFourBlocks = trimNewlines(`
(B root: alphabetakappagamma 
  (B root: alphabeta 
    (L root: alpha) 
    (L root: beta)) 
  (B root: kappagamma 
    (L root: kappa) 
    (L root: gamma)))
`)

var givenTwoBlocks = trimNewlines(`
(B root: alphabeta 
  (L root: alpha) 
  (L root: beta))
`)

var givenThreeBlocks = trimNewlines(`
(B root: alphabetakappakappa 
  (B root: alphabeta 
    (L root: alpha) 
    (L root: beta)) 
  (B root: kappakappa 
    (L root: kappa) 
    (L root: kappa)))
`)

var givenSixBlocks = trimNewlines(`
(B root: alphabetakappagammaepsilonomegaepsilonomega 
  (B root: alphabetakappagamma 
    (B root: alphabeta 
      (L root: alpha) 
      (L root: beta)) 
    (B root: kappagamma 
      (L root: kappa) 
      (L root: gamma))) 
  (B root: epsilonomegaepsilonomega 
    (B root: epsilonomega 
      (L root: epsilon) 
      (L root: omega)) 
    (B root: epsilonomega 
      (L root: epsilon) 
      (L root: omega))))
`)

var proofA = trimNewlines(`
epsilon + omega = epsilonomega
epsilonomega + muzeta = epsilonomegamuzeta
alphabetakappagamma + epsilonomegamuzeta = alphabetakappagammaepsilonomegamuzeta
`)

func TestCreateMerkleTree(t *testing.T) {
	t.Run("easy tree - just one level (the root) of nodes", func(t *testing.T) {
		blocks := [][]byte{[]byte("alpha"), []byte("beta")}
		tree := NewTree(IdentityHashForTest, blocks)

		expectStrEqual(t, tree.ToString(str, 0), givenTwoBlocks)
	})

	t.Run("two levels of nodes", func(t *testing.T) {
		blocks := [][]byte{[]byte("alpha"), []byte("beta"), []byte("kappa"), []byte("gamma")}
		tree := NewTree(IdentityHashForTest, blocks)

		expectStrEqual(t, tree.ToString(str, 0), givenFourBlocks)
	})

	t.Run("one block - one level", func(t *testing.T) {
		blocks := [][]byte{[]byte("alpha")}
		tree := NewTree(IdentityHashForTest, blocks)

		expectStrEqual(t, tree.ToString(str, 0), givenOneBlock)
	})

	/*

				duplicate a leaf

		            123{3}
				 /        \
			   12          3{3}
			 /    \      /    \
			1      2    3      {3}

	*/
	t.Run("duplicate a leaf to keep the binary tree balanced", func(t *testing.T) {
		blocks := [][]byte{[]byte("alpha"), []byte("beta"), []byte("kappa")}
		tree := NewTree(IdentityHashForTest, blocks)

		expectStrEqual(t, tree.ToString(str, 0), givenThreeBlocks)
	})

	/*

			          duplicate a node

		                123456{56}
		          /                    \
		        1234                  56{56}
		     /        \              /      \
		   12          34          56        {56}
		 /    \      /    \      /    \     /    \
		1      2    3      4    5      6  {5}    {6}

	*/
	t.Run("duplicate a branch to keep the tree balanced", func(t *testing.T) {
		blocks := [][]byte{[]byte("alpha"), []byte("beta"), []byte("kappa"), []byte("gamma"), []byte("epsilon"), []byte("omega")}
		tree := NewTree(IdentityHashForTest, blocks)

		expectStrEqual(t, tree.ToString(str, 0), givenSixBlocks)
	})
}

func TestConsistencyProof(t *testing.T) {
	blocksA := [][]byte{[]byte("alpha"), []byte("beta"), []byte("kappa")}
	treeA := NewTree(IdentityHashForTest, blocksA)

	blocksB := [][]byte{[]byte("alpha"), []byte("beta")}
	treeB := NewTree(IdentityHashForTest, blocksB)

	t.Run("Tree#Verify presented Merkle root (checksum) does not match receiving tree", func(t *testing.T) {
		if treeB.Verify(treeA.root.GetChecksum(), []byte("beta")) {
			s := "should have failed (different checksums) - treeA: %s treeB: %s"
			t.Fatalf(s, treeA.root.GetChecksum(), treeB.root.GetChecksum())
		}
	})

	t.Run("Tree#Verify presented leaf checksum does not exist in receiving tree", func(t *testing.T) {
		if treeB.Verify(treeB.root.GetChecksum(), []byte("kappa")) {
			t.Fatal("should have failed (leaf shouldn't exist in receiver's cache")
		}
	})

	t.Run("Tree#GetProofString", func(t *testing.T) {
		blocks := [][]byte{
			[]byte("alpha"),
			[]byte("beta"),
			[]byte("kappa"),
			[]byte("gamma"),
			[]byte("epsilon"),
			[]byte("omega"),
			[]byte("mu"),
			[]byte("zeta"),
		}

		tree := NewTree(IdentityHashForTest, blocks)
		checksum := treeA.checksumFunc([]byte("omega"))

		f := func(xs []byte) string {
			return string(xs)
		}

		expectStrEqual(t, tree.GetProofString(tree.root.GetChecksum(), checksum, f), proofA)
	})
}
