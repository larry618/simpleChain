package main

import (
	"github.com/boltdb/bolt"
	"encoding/hex"
	"log"
	"fmt"
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


			// 把所有产生的UTXO添加到set中
			outs := NewTxOutputs()
			for outIdx, out := range tx.Vout {
				outs[outIdx] = out
			}

			fmt.Printf("update添加的 txID %x\n", tx.ID)
			bucket.Put(tx.ID, outs.Serialize())

			if tx.IsCoinbase() {
				continue
			}

			// 删除花费掉的 output
			for _, in := range tx.Vin {
				utxos := bucket.Get(in.Txid) // 当前in.Txid下所有的UTXO

				outs := DeserializeOutputs(utxos)
				delete(outs, in.Vout)

				if len(outs) == 0 {
					bucket.Delete(in.Txid)
				} else {
					bucket.Put(in.Txid, outs.Serialize())
				}

			}

		}

		return nil
	})
}

func (set *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {

	var spendableOutputs = make(map[string][]int) // 同一个 tx 下会有多个转入同一个地址的 output 吗?

	// 同一个 tx 下会有多个转入同一个地址的 output 吗?
	// 如果完全不考虑手续费的话, 一个交易最多有两个output, 一个是买方地址，一个是找零地址.
	// 但是如果考虑进手续费的话, 一个交易最多有3个output, 一个是买方地址，一个是找零地址, 还有挖出区块的人收的手续费的地址;
	// 如果这笔交易是转给 A 地址, 同时 A 挖出了区块, 要收手续费的话, 就会在一个交易中有两个output同时转入 A 地址下.
	accumulate := 0
	db := set.bc.db

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoSetBucket))
		cursor := bucket.Cursor()
		// Cursor represents an iterator that can traverse over all key/value pairs in a bucket in sorted order.

		for key, value := cursor.First(); key != nil && accumulate < amount; key, value = cursor.Next() {
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

		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {

			fmt.Printf("FindUTXO 遍历的: txID %x\n", key)
			outs := DeserializeOutputs(value)
			for _, out := range outs {


				fmt.Printf("FindUTXO的PubKeyHash: %x\n", pubKeyHash)
				fmt.Printf("tx中的PubKeyHash    : %x\n", out.PubKeyHash)

				if out.IsLockedWith(pubKeyHash) {
					outputs = append(outputs, out)
				}
			}
		}
		return nil
	})

	return outputs
}
