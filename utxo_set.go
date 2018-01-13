package main

import (
	"github.com/boltdb/bolt"
	"encoding/hex"
	"log"
)

const utxoSetBucket = "utxoSet"

// 用于存放 区块链中的所有 UTXO
type UTXOSet struct {
	bc *BlockChain
}

func NewUTXOSet(bc *BlockChain) *UTXOSet {
	return &UTXOSet{bc}
}

// 每次生成新区块链的时候调用
func (set *UTXOSet) Reindex() {
	db := set.bc.db
	bucketName := []byte(utxoSetBucket)
	db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(bucketName)
		tx.CreateBucket(bucketName)
		return nil
	})

	utxos := set.bc.FindAllUTXOs()

	db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket(bucketName)

		for txID, outs := range utxos {
			txId, _ := hex.DecodeString(txID)
			err := bucket.Put(txId, outs.Serialize())

			if err != nil {
				log.Panic(err)
				return err
			}
		}

		return nil
	})
}

// 每生成一个区块后 更行UTXO Set
func (set *UTXOSet) Update(b *Block) {

	db := set.bc.db

	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoSetBucket))

		for _, tx := range b.Transactions {

			// 删除花费掉的 output
			for _, in := range tx.Vin {
				utxos := bucket.Get(in.Txid) // 当前in.Txid下所有的UTXO

				outs := DeserializeOutputs(utxos)
				delete(outs, in.Vout)
				bucket.Put(in.Txid, outs.Serialize())
			}

			// 把所有产生的UTXO添加到set中
			outs := NewTxOutputs()
			for outIdx, out := range tx.Vout {
				outs[outIdx] = out
			}

			bucket.Put(tx.ID, outs.Serialize())
		}

		return nil
	})
}

func (set *UTXOSet) FindSpendableOutput(pubKeyHash []byte, amount int) (int, map[string][]int) {
	spendableOutputs := make(map[string][]int)
	accumulate := 0
	db := set.bc.db

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoSetBucket))
		cursor := bucket.Cursor()
		// Cursor represents an iterator that can traverse over all key/value pairs in a bucket in sorted order.

		for key, value := cursor.First(); key != nil && accumulate < amount; cursor.Next() {
			txID := hex.EncodeToString(key)
			outs := DeserializeOutputs(value)

			for outIdx, out := range outs {

				if out.IsLockedWith(pubKeyHash) {

					accumulate += out.Value
					spendableOutputs[txID] = append(spendableOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})

	return accumulate, spendableOutputs
}

func (set *UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	db := set.bc.db
	var outputs []TXOutput

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoSetBucket))
		cursor := bucket.Cursor()
		// Cursor represents an iterator that can traverse over all key/value pairs in a bucket in sorted order.

		for key, value := cursor.First(); key != nil; cursor.Next() {
			outs := DeserializeOutputs(value)
			for _, out := range outs {

				if out.IsLockedWith(pubKeyHash) {
					outputs = append(outputs, out)
				}
			}
		}
		return nil
	})

	return outputs
}
