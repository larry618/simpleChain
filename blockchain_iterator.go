package main

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockChainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (bci *BlockChainIterator) Next() *Block {

	var block *Block
	err := bci.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucket)
		encodedBlock := bucket.Get(bci.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bci.currentHash = block.PrevBlockHash
	return block
}

func (bci *BlockChainIterator) HasNext() bool {
	return len(bci.currentHash) != 0
}
