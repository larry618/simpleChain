package main

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"fmt"
	"errors"
	"bytes"
	"crypto/ecdsa"
)

// 常量只能是字符串、布尔和数字三种类型。
const (
	dbFile          = "aaa"
	blocksBucketStr = "asdf"
)

var (
	blocksBucket = []byte(blocksBucketStr)
	tipKey       = []byte("l")
)

type BlockChain struct {
	//blocks []*Block
	tip []byte   // 最后一个区块的 hash
	db  *bolt.DB // 存储 区块的数据库
	utxoSet *UTXOSet
}

func (bc *BlockChain) AddBlock(txs []*Transaction) {
	newBlock := NewBlock(txs, bc.tip)

	bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucket)
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		err = bucket.Put(tipKey, newBlock.Hash)
		return err
	})
	bc.tip = newBlock.Hash
}

// address: 用于接受创世区块的奖励
func NewBlockChain(address string) *BlockChain {
	//return &BlockChain{[]*Block{newGenesisBlock()}}

	var tip []byte
	db, err := bolt.Open(dbFile, 0666, nil)
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))

		if bucket == nil {

			bucket, err := tx.CreateBucket(blocksBucket)
			if err != nil {
				log.Panic(err)
			}

			coinBaseTX := NewCoinBaseTX(address, "Onwards and upwards")
			genesisBlock := newGenesisBlock(coinBaseTX)
			err = bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			err = bucket.Put(tipKey, genesisBlock.Hash)
			tip = genesisBlock.Hash


		} else {
			tip = bucket.Get(tipKey)
		}
		return err
	})

	if err != nil {
		log.Panic(err)
	}

	bc := &BlockChain{tip, db, nil}

	bc.utxoSet = NewUTXOSet(bc)
	return bc
}

// 挖矿
func (bc *BlockChain) Mining(txs []*Transaction, addr string) {

	tx := NewCoinBaseTX(addr, "")
	fmt.Printf("%s is mining...", addr)
	txs = append(txs, tx)
	bc.AddBlock(txs)
}

func (bc *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{bc.tip, bc.db}
}

func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []*Transaction {

	var unspentTXs []*Transaction
	var spendTxs = make(map[string][]int) // key: txid, value: outputIndexs

	iterator := bc.Iterator()
	for iterator.HasNext() {
		block := iterator.Next()
		txs := block.Transactions

		for _, tx := range txs { // 区块中的所有交易
			txId := hex.EncodeToString(tx.ID)

		Outputs:
			for idx, out := range tx.Vout {

				if spendTxs[txId] != nil { // 已花费

					for _, outputIndex := range spendTxs[txId] {
						if idx == outputIndex {
							continue Outputs
						}
					}
				}

				if out.IsLockedWith(pubKeyHash) {
					unspentTXs = append(unspentTXs, tx)
				}

			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						spendTxs[txId] = append(spendTxs[txId], in.Vout)
					}
				}
			}
		}

	}

	return unspentTXs
}


// 查看余额的时候调用
func (bc *BlockChain) FindUTXO(pubKeyHash []byte) []TXOutput {

	txs := bc.FindUnspentTransactions(pubKeyHash)

	var utxos []TXOutput
	for _, tx := range txs {

		for _, out := range tx.Vout {
			if out.IsLockedWith(pubKeyHash) {
				utxos = append(utxos, out)
			}
		}
	}

	return utxos
}

// find all utxo for build utxo_set when ever a blockchain is been created
func (bc *BlockChain) FindAllUTXOs() map[string]TXOutputs {

	spendTxOutputs := make(map[string]IntSet) // 已花费的output  key: 交易ID, value: 当前交易的所有花费了的output的 索引合集
	 utxos := make(map[string]TXOutputs)                   // 未花费的output

	it := bc.Iterator()

	for it.HasNext() {
		block := it.Next()

		for _, tx := range block.Transactions {

			// 统计当前交易中已花费的别的交易中的output
			for _, in := range tx.Vin {
				prevTxID := hex.EncodeToString(in.Txid)
				spendTxOutputs[prevTxID].add(in.Vout)
			}

			// 统计当前交易中的所有未被花费的交易
			curtTxID := hex.EncodeToString(tx.ID)
			outs := NewTxOutputs()
			for outIdx, out := range tx.Vout {
				if spendTxOutputs[curtTxID].contains(outIdx) == false { // 当前output未被花费
					outs[outIdx] = out
				}
			}

			if len(outs) != 0 {
				utxos[curtTxID] = outs
			}
		}
	}

	return utxos
}

func (bc *BlockChain) GetBalance(addr string) int {
	utxos := bc.FindUTXO(GetPubKeyHashFromAddr(addr))

	balance := 0
	for _, utxo := range utxos {
		balance += utxo.Value
	}

	return balance
}

func (bc *BlockChain) NewUTXOTransaction(from *Wallet, to string, amount int) *Transaction {

	// 1. 找到from用户的amount数量的 utxo
	// 2. 添加交易

	acc, spendableOutputs := bc.FindSpendableOutputs(HashPubKey(from.PublicKey), amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput

	for txid, outIdxs := range spendableOutputs {
		txID, _ := hex.DecodeString(txid)

		for _, outIdx := range outIdxs {
			inputs = append(inputs, TXInput{txID, outIdx, nil, from.PublicKey})
		}
	}

	if acc > amount {
		addr := fmt.Sprintf("%s", from.GetAddress())
		outputs = append(outputs, NewTxOutput(acc-amount, addr))
	}

	outputs = append(outputs, NewTxOutput(amount, to))

	tx := Transaction{nil, inputs, outputs}

	bc.SignTx(&tx, from.PrivateKey) // 签名交易
	tx.Hash()

	return nil
}

func (bc *BlockChain) getPrevTxs(tx *Transaction) map[string]*Transaction {
	prevTxs := make(map[string]*Transaction)

	for _, in := range tx.Vin {
		txID := hex.EncodeToString(in.Txid)
		prevTxs[txID] = bc.findTx(in.Txid)
	}

	return prevTxs
}

func (bc *BlockChain) findTx(txId []byte) *Transaction {

	iterator := bc.Iterator()
	for iterator.HasNext() {
		block := iterator.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(txId, tx.ID) == 0 {
				return tx
			}
		}
	}

	return nil

}

func (bc *BlockChain) SignTx(tx *Transaction, key ecdsa.PrivateKey) {

	prevTxs := bc.getPrevTxs(tx)
	tx.Sign(key, prevTxs)
}

func (bc *BlockChain) VerifyTx(tx *Transaction) bool {
	prevTxs := bc.getPrevTxs(tx)
	res := tx.Verify(prevTxs)
	return res

}


//  返回  >= amount 数量的 UTXOs
func (bc *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {

	var spendableOutputs = make(map[string][]int)    // 同一个 tx 下会有多个转入同一个地址的 output 吗?

	// 同一个 tx 下会有多个转入同一个地址的 output 吗?
	// 如果完全不考虑手续费的话, 一个交易最多有两个output, 一个是买方地址，一个是找零地址.
	// 但是如果考虑进手续费的话, 一个交易最多有3个output, 一个是买方地址，一个是找零地址, 还有挖出区块的人收的手续费的地址;
	// 如果这笔交易是转给 A 地址, 同时 A 挖出了区块, 要收手续费的话, 就会在一个交易中有两个output同时转入 A 地址下.

	txs := bc.FindUnspentTransactions(pubKeyHash)
	accumulate := 0

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		for idx, out := range tx.Vout {
			if accumulate < amount && out.IsLockedWith(pubKeyHash) {

				accumulate += out.Value
				spendableOutputs[txID] = append(spendableOutputs[txID], idx)

				if accumulate >= amount {
					return accumulate, spendableOutputs
				}
			}
		}
	}

	return accumulate, spendableOutputs
}
