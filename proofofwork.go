package main

import (
	"math/big"
	"bytes"
	"crypto/sha256"
	"math"
	"fmt"
)

const targetBits = 20 // 计算出来 Hash 前面有多少位是 0

type ProofOfWork struct {
	block  *Block
	target *big.Int // 用于比较的 Hash
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	// 把 1 左移 256 - targetBits 位, 使之变成以 targetBits 个 0 开头的数字  (比如targetBits为4时 00001000000...000000)
	target = target.Lsh(target, uint(256-targetBits))

	return &ProofOfWork{b, target}
}

// 生成用于挖矿的数据
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.TransactionsHash(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		}, []byte{})

	return data
}

func (pow *ProofOfWork) run() (int, []byte) {

	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	//fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)

	for nonce < math.MaxInt64 {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x", hash)
			break
		} else {
			nonce++
		}
	}

	fmt.Print("\n\n")
	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	nonce := pow.block.Nonce
	data := pow.prepareData(nonce)
	hash := sha256.Sum256(data)

	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1
}
