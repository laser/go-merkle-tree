package main

import (
	"fmt"
	"strings"
	"testing"
)

func trimNewlines(str string) string {
	return strings.Trim(str, "\n")
}

func expectEqual(t *testing.T, actual string, expected string) {
	if trimNewlines(actual) != expected {
		fmt.Println(fmt.Sprintf("=====ACTUAL======\n\n%s\n\n=====EXPECTED======\n\n%s\n", actual, expected))
		t.Fail()
	}
}

func id(strbytes []byte) []byte {
	return strbytes
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

func TestMerkleTree(t *testing.T) {
	t.Run("easy tree - just one level (the root) of nodes", func(t *testing.T) {
		stuff := [][]byte{[]byte("alpha"), []byte("beta")}
		tree := CreateTree(id, stuff)

		expectEqual(t, tree.AsString(str, 0), givenTwoBlocks)
	})

	t.Run("two levels of nodes", func(t *testing.T) {
		stuff := [][]byte{[]byte("alpha"), []byte("beta"), []byte("kappa"), []byte("gamma")}
		tree := CreateTree(id, stuff)

		expectEqual(t, tree.AsString(str, 0), givenFourBlocks)
	})

	t.Run("one block - one level", func(t *testing.T) {
		stuff := [][]byte{[]byte("alpha")}
		tree := CreateTree(id, stuff)

		expectEqual(t, tree.AsString(str, 0), givenOneBlock)
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
		stuff := [][]byte{[]byte("alpha"), []byte("beta"), []byte("kappa")}
		tree := CreateTree(id, stuff)

		expectEqual(t, tree.AsString(str, 0), givenThreeBlocks)
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
		stuff := [][]byte{[]byte("alpha"), []byte("beta"), []byte("kappa"), []byte("gamma"), []byte("epsilon"), []byte("omega")}
		tree := CreateTree(id, stuff)

		expectEqual(t, tree.AsString(str, 0), givenSixBlocks)
	})
}
