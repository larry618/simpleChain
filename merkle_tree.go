package main

import (
	"crypto/sha256"
)

type MerkleTree struct {
	Root *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {

	node := MerkleNode{left, right, nil}

	var sum256 [32]byte
	if left == nil && right == nil {
		sum256 = sha256.Sum256(data)

	} else {
		prevHash := append(left.Data, right.Data...)
		sum256 = sha256.Sum256(prevHash)
	}
	node.Data = sum256[:]

	return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {

	l := len(data)
	if l%2 != 0 {
		data = append(data, data[l-1])
	}

	var bottomNodes []*MerkleNode

	for _, b := range data {
		bottomNodes = append(bottomNodes, NewMerkleNode(nil, nil, b))
	}

	root := buildMerkleTreeHelper(bottomNodes)
	return &MerkleTree{root}

}

func buildMerkleTreeHelper(nodes []*MerkleNode) *MerkleNode {

	if len(nodes) == 1 {
		return nodes[0]
	}

	l := len(nodes)

	leftRoot := buildMerkleTreeHelper(nodes[:l/2])
	rightRoot := buildMerkleTreeHelper(nodes[l/2:])

	root := NewMerkleNode(leftRoot, rightRoot, nil)

	return root
}
