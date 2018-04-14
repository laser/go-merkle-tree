package merkletree

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func bytesToStr(xs []byte) string {
	return string(xs)
}

func bytesToHexStr(xs []byte) string {
	return hex.EncodeToString(xs)[:16]
}

func trimNewlines(str string) string {
	return strings.Trim(str, "\n")
}

func expectStrEqual(t *testing.T, actual string, expected string) {
	if trimNewlines(actual) != expected {
		fmt.Println(fmt.Sprintf("=====ACTUAL======\n\n%s\n\n=====EXPECTED======\n\n%s\n", actual, expected))
		t.Fail()
	}
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
route from omega (leaf) to root:

epsilon + omega = epsilonomega
epsilonomega + muzeta = epsilonomegamuzeta
alphabetakappagamma + epsilonomegamuzeta = alphabetakappagammaepsilonomegamuzeta
`)

func TestCreateMerkleTree(t *testing.T) {
	t.Run("easy tree - just one level (the root) of nodes", func(t *testing.T) {
		blocks := [][]byte{[]byte("alpha"), []byte("beta")}
		tree := NewTree(IdentityHashForTest, blocks)

		expectStrEqual(t, tree.ToString(bytesToStr, 0), givenTwoBlocks)
	})

	t.Run("two levels of nodes", func(t *testing.T) {
		blocks := [][]byte{[]byte("alpha"), []byte("beta"), []byte("kappa"), []byte("gamma")}
		tree := NewTree(IdentityHashForTest, blocks)

		expectStrEqual(t, tree.ToString(bytesToStr, 0), givenFourBlocks)
	})

	t.Run("one block - one level", func(t *testing.T) {
		blocks := [][]byte{[]byte("alpha")}
		tree := NewTree(IdentityHashForTest, blocks)

		expectStrEqual(t, tree.ToString(bytesToStr, 0), givenOneBlock)
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

		expectStrEqual(t, tree.ToString(bytesToStr, 0), givenThreeBlocks)
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

		expectStrEqual(t, tree.ToString(bytesToStr, 0), givenSixBlocks)
	})
}

func TestAuditProof(t *testing.T) {
	t.Run("Tree#CreateProof", func(t *testing.T) {
		blocks := [][]byte{
			[]byte("alpha"),
			[]byte("beta"),
			[]byte("kappa"),
		}

		tree := NewTree(IdentityHashForTest, blocks)
		checksum := tree.checksumFunc([]byte("alpha"))

		proof, err := tree.CreateProof(tree.root.GetChecksum(), checksum)
		if err != nil {
			t.Fail()
		}

		expected := Proof{
			parts: []*ProofPart{{
				isRight:  true,
				checksum: IdentityHashForTest([]byte("beta")),
			}, {
				isRight:  true,
				checksum: IdentityHashForTest([]byte("kappakappa")),
			}},
			target: checksum,
		}

		if !expected.Equals(proof) {
			t.Fail()
		}
	})

	t.Run("Proof#ToString", func(t *testing.T) {
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
		checksum := tree.checksumFunc([]byte("omega"))
		proof, _ := tree.CreateProof(tree.root.GetChecksum(), checksum)

		expectStrEqual(t, proof.ToString(IdentityHashForTest, bytesToStr), proofA)
	})

	t.Run("Tree#VerifyProof", func(t *testing.T) {
		t.Run("valid proof for a two-leaf tree", func(t *testing.T) {
			blocks := [][]byte{
				[]byte("alpha"),
				[]byte("beta"),
			}

			tree := NewTree(IdentityHashForTest, blocks)

			proof := &Proof{
				parts: []*ProofPart{{
					isRight:  true,
					checksum: IdentityHashForTest([]byte("beta")),
				}},
				target: IdentityHashForTest([]byte("alpha")),
			}

			if !tree.VerifyProof(proof) {
				t.Fail()
			}
		})

		t.Run("invalid proof (isRight should be true) for a two-leaf tree", func(t *testing.T) {
			blocks := [][]byte{
				[]byte("alpha"),
				[]byte("beta"),
			}

			tree := NewTree(IdentityHashForTest, blocks)

			proof := &Proof{
				parts: []*ProofPart{{
					isRight:  false,
					checksum: IdentityHashForTest([]byte("beta")),
				}},
				target: IdentityHashForTest([]byte("alpha")),
			}

			if tree.VerifyProof(proof) {
				t.Fail()
			}
		})

		t.Run("invalid proof (wrong sibling) for a two-leaf tree", func(t *testing.T) {
			blocks := [][]byte{
				[]byte("alpha"),
				[]byte("beta"),
			}

			tree := NewTree(IdentityHashForTest, blocks)

			proof := &Proof{
				parts: []*ProofPart{{
					isRight:  true,
					checksum: IdentityHashForTest([]byte("kappa")),
				}},
				target: IdentityHashForTest([]byte("alpha")),
			}

			if tree.VerifyProof(proof) {
				t.Fail()
			}
		})

		t.Run("invalid proof (tree doesn't contain target) for a two-leaf tree", func(t *testing.T) {
			blocks := [][]byte{
				[]byte("alpha"),
				[]byte("beta"),
			}

			tree := NewTree(IdentityHashForTest, blocks)

			proof := &Proof{
				parts: []*ProofPart{{
					isRight:  true,
					checksum: IdentityHashForTest([]byte("beta")),
				}},
				target: IdentityHashForTest([]byte("kappa")),
			}

			if tree.VerifyProof(proof) {
				t.Fail()
			}
		})

		t.Run("valid proof for eight leaf tree", func(t *testing.T) {
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
			checksum := tree.checksumFunc([]byte("alpha"))

			proof, err := tree.CreateProof(tree.root.GetChecksum(), checksum)
			if err != nil {
				t.Fail()
			}

			if !tree.VerifyProof(proof) {
				t.Fail()
			}
		})
	})
}

func TestDocsCreateAndPrintAuditProof(t *testing.T) {
	blocks := [][]byte{
		[]byte("alpha"),
		[]byte("beta"),
		[]byte("kappa"),
	}

	tree := NewTree(Sha256DoubleHash, blocks)
	checksum := tree.checksumFunc([]byte("alpha"))
	proof, _ := tree.CreateProof(tree.root.GetChecksum(), checksum)

	fmt.Println(proof.ToString(Sha256DoubleHash, func(bytes []byte) string {
		return hex.EncodeToString(bytes)[0:16]
	}))

	/*

		output:

		route from aa86be763e41db7e (leaf) to root:

		aa86be763e41db7e + 05e3bc756e005c1b = 65492c0681df09eb
		65492c0681df09eb + 3ae3330dcf932104 = 1add1cfdf5df2841

	*/
}

func TestDocsCreateAndPrintTree(t *testing.T) {
	blocks := [][]byte{
		[]byte("alpha"),
		[]byte("beta"),
		[]byte("kappa"),
	}

	tree := NewTree(Sha256DoubleHash, blocks)

	fmt.Println(tree.ToString(func(bytes []byte) string {
		return hex.EncodeToString(bytes)[0:16]
	}, 0))

	/*

		output:

		(B root: 1add1cfdf5df2841
		  (B root: 65492c0681df09eb
			(L root: aa86be763e41db7e)
			(L root: 05e3bc756e005c1b))
		  (B root: 3ae3330dcf932104
			(L root: 4cc9e99389b5f729)
			(L root: 4cc9e99389b5f729)))

	*/
}

func TestDocsValidateProof(t *testing.T) {
	blocks := [][]byte{
		[]byte("alpha"),
		[]byte("beta"),
		[]byte("kappa"),
	}

	tree := NewTree(Sha256DoubleHash, blocks)

	proof := &Proof{
		parts: []*ProofPart{{
			isRight:  true,
			checksum: Sha256DoubleHash([]byte("beta")),
		}, {
			isRight:  true,
			checksum: Sha256DoubleHash(append(Sha256DoubleHash([]byte("kappa")), Sha256DoubleHash([]byte("kappa"))...)),
		}},
		target: Sha256DoubleHash([]byte("alpha")),
	}

	tree.VerifyProof(proof) // true
}
