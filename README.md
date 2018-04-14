# go-merkle-tree

> A Bitcoin Merkle Tree, implemented in Go

Many people have written many things about Merkle Trees. For a good overview (uses, characteristics, etc.), read Marc
Clifton's [_Understanding Merkle Trees - Why use them, who uses them, and how to use them_][1].

## Warning

*This is alpha software.*

## Acknowledgements

This implementation was inspired by:

- [Marc Clifton's _Understanding Merkle Trees - Why use them, who uses them, and how to use them_][1]
- [Miguel Mota's merkle-tree][2] (in particular: proof generation)

## Usage

### Construction

```go
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
```

### Create and Print Audit Proof

```go
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
```

### Verify Audit Proof

```go
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
```

[1]: https://www.codeproject.com/Articles/1176140/Understanding-Merkle-Trees-Why-use-them-who-uses-t
[2]: https://github.com/miguelmota/merkle-tree
