package main

import (
	"strconv"
	"bytes"
	"crypto/sha256"
	"time"
	"fmt"
	"encoding/gob"
	"log"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), txs, prevBlockHash, []byte{}, 0}

	// block.setHash()

	pow := NewProofOfWork(block)
	nonce, hash := pow.run()
	block.Hash = hash
	block.Nonce = nonce
	return block
}

func newGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// 每笔交易的TXID 进行哈希
func (b *Block) TransactionsHash() []byte {
	var txIds [][]byte
	var txsHash [32]byte

	for _, tx := range b.Transactions {
		txIds = append(txIds, tx.ID)
	}

	txsHash = sha256.Sum256(bytes.Join(txIds, []byte{}))
	return txsHash[:]
}

func (b *Block) String() string {

	var res string
	res += fmt.Sprintf("Prev. hash: %x\n", b.PrevBlockHash)
	res += fmt.Sprintf("TransactionsHash: %x\n", b.TransactionsHash())
	res += fmt.Sprintf("Hash: %x\n", b.Hash)
	pow := NewProofOfWork(b)
	res += fmt.Sprintf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
	res += fmt.Sprintln()
	return res

	//json, _ := json.Marshal(b)
	//return string(json)
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)

	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func Deserialize(b []byte) *Block {

	var block Block
	reader := bytes.NewReader(b)
	decoder := gob.NewDecoder(reader)

	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
