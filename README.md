# go-merkle-tree

> A Merkle Tree, implemented in Golang

Many people have written many things about Merkle Trees. For a good overview (uses, characteristics, etc.), read Marc
Clifton's [_Understanding Merkle Trees - Why use them, who uses them, and how to use them_][1].

## Warning

*Warning: This is alpha software.*

## Usage

### Construction

```go
data := [][]byte{[]byte("alpha"), []byte("beta"), []byte("kappa")}

tree := CreateTree(Sha256DoubleHash, data)

fmt.Println(tree.AsString(hex.EncodeToString, 0))

/*

output:

(B root: 1add1cfdf5df28414b715199a740f80b7f559bd558a3f0c0186e60149ee86620
  (B root: 65492c0681df09eb403160136bb648de17f67bd8efc441467c0fc23b8d2950e9
    (L root: aa86be763e41db7eaae266afc79ab46d02343c5d3b05da171d351afbd25c1525)
    (L root: 05e3bc756e005c1bc5e4daf8a3da95d435af52476b0a0e6d52e719a2b1e3434a))
  (B root: 3ae3330dcf932104d42b75b4da386896a628926f411737b34430fa65e526824d
    (L root: 4cc9e99389b5f729cbef6fe79e97a6f562841a2852e25e508e3bd06ce0de9c26)
    (L root: 4cc9e99389b5f729cbef6fe79e97a6f562841a2852e25e508e3bd06ce0de9c26)))

*/
```

### Get Audit Proof

```go

blocks := [][]byte{
    []byte("alpha"),
    []byte("beta"),
    []byte("kappa"),
}

tree := NewTree(IdentityHashForTest, blocks)
checksum := tree.checksumFunc([]byte("alpha"))

proof, e := tree.GetProof(tree.root.GetChecksum(), checksum)
```

### Print Audit Proof

```go
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
proof, _ := tree.GetProof(tree.root.GetChecksum(), checksum)
toStr := func(xs []byte) string { return string(xs) }

/*

output:

epsilon + omega = epsilonomega
epsilonomega + muzeta = epsilonomegamuzeta
alphabetakappagamma + epsilonomegamuzeta = alphabetakappagammaepsilonomegamuzeta

*/
```

[1]: https://www.codeproject.com/Articles/1176140/Understanding-Merkle-Trees-Why-use-them-who-uses-t