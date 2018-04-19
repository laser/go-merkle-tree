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

    (B root: 3d4bd4dd0a71aeb3
      (B root: 8b3ee349b69b427f
        (L root: c246ba39b1c6c18d)
        (L root: 24960c3aab1f4b41))
      (B root: da2f01ea4b9f38ad
        (L root: 37ce7f776537a298)
        (L root: 37ce7f776537a298)))

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
checksum := tree.checksumFunc(true, []byte("alpha"))
proof, _ := tree.CreateProof(checksum)

fmt.Println(proof.ToString(func(bytes []byte) string {
    return hex.EncodeToString(bytes)[0:16]
}))

/*

    output:

    route from c246ba39b1c6c18d (leaf) to root:

    c246ba39b1c6c18d + 24960c3aab1f4b41 = 8b3ee349b69b427f
    8b3ee349b69b427f + da2f01ea4b9f38ad = 3d4bd4dd0a71aeb3

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
        checksum: tree.checksumFunc(true, []byte("beta")),
    }, {
        isRight: true,
        checksum: tree.checksumFunc(
            false,
            append(
                tree.checksumFunc(true, []byte("kappa")),
                tree.checksumFunc(true, []byte("kappa"))...)),
    }},
    target: tree.checksumFunc(true, []byte("alpha")),
}

tree.VerifyProof(proof) // true
```

[1]: https://www.codeproject.com/Articles/1176140/Understanding-Merkle-Trees-Why-use-them-who-uses-t
[2]: https://github.com/miguelmota/merkle-tree
