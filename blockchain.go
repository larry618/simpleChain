package main

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"fmt"
	"bytes"
	"crypto/ecdsa"
	"time"
)

// 常量只能是字符串、布尔和数字三种类型。
const (
	originDbFile    = "blockchain_%s.db"
	blocksBucketStr = "asdf"
)

var (
	blocksBucket = []byte(blocksBucketStr)
	tipKey       = []byte("l")
	dbFile       string
)

type BlockChain struct {
	//blocks []*Block
	tip     []byte   // 最后一个区块的 hash
	db      *bolt.DB // 存储 区块的数据库
	utxoSet *UTXOSet
}

func (bc *BlockChain) GetBestHeight() int64 {

	var height int64
	bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucket)
		blockBytes := bucket.Get(bc.tip)

		block := DeserializeBlock(blockBytes)

		height = block.Height
		return nil
	})

	return height
}

func (bc *BlockChain) GetLastBlock() *Block {
	return bc.GetBlock(bc.tip)
}

func (bc *BlockChain) GetBlock(hash []byte) *Block {

	var block *Block
	bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucket)

		blockData := bucket.Get(hash)
		block = DeserializeBlock(blockData)
		return nil
	})

	return block
}

func (bc *BlockChain) GetBlocksHash(amount int64) [][]byte {

	var hashes [][]byte
	iterator := bc.Iterator()

	for iterator.HasNext() && amount > 0 {
		amount--
		block := iterator.Next()
		hashes = append(hashes, block.Hash)
	}

	return hashes

}

func (bc *BlockChain) MiningBlock(txs []*Transaction) {

	var height int64
	bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucket)

		blockData := bucket.Get(bc.tip)

		lastBlock := DeserializeBlock(blockData)

		height = lastBlock.Height

		return nil
	})

	newBlock := NewBlock(txs, bc.tip, height+1)
	bc.AddBlock(newBlock)
}

func (bc *BlockChain) AddBlock(newBlock *Block) {
	bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucket)
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		err = bucket.Put(tipKey, newBlock.Hash)
		return err
	})
	bc.tip = newBlock.Hash
	bc.utxoSet.Update(newBlock)
}

// address: 用于接受创世区块的奖励
func NewBlockChain(address, nodeId string) *BlockChain {

	var tip []byte
	dbFile = fmt.Sprintf(originDbFile, nodeId)
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
	bc.utxoSet.Reindex()

	return bc
}

func LoadBlockChain(nodeId string) *BlockChain {
	var tip []byte

	dbFile = fmt.Sprintf(originDbFile, nodeId)
	db, _ := bolt.Open(dbFile, 0666, nil)
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		tip = bucket.Get(tipKey)
		return nil
	})

	bc := &BlockChain{tip, db, nil}
	bc.utxoSet = NewUTXOSet(bc)
	bc.utxoSet.Reindex()

	return bc
}

// 挖矿
func (bc *BlockChain) Mining(txs []*Transaction, addr string) {

	tx := NewCoinBaseTX(addr, "")
	fmt.Printf("%s is mining...\n", addr)
	txs = append(txs, tx)

	fmt.Printf("当前区块的交易数量:%d\n", len(txs))
	for i, tx := range txs {
		fmt.Println(i)
		if bc.VerifyTx(tx) == false {
			log.Panic("ERROR: Invalid transcation")
		}
	}

	bc.MiningBlock(txs)
}

func (bc *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{bc.tip, bc.db}
}

// 查看余额的时候调用
func (bc *BlockChain) FindUTXO(pubKeyHash []byte) []TXOutput {
	return bc.utxoSet.FindUTXO(pubKeyHash)
}

// find all utxo for build utxo_set when ever a blockchain is been created
func (bc *BlockChain) FindAllUTXOs() map[string]TXOutputs {

	spendTxOutputs := make(map[string]IntSet) // 已花费的output  key: 交易ID, value: 当前交易的所有花费了的output的 索引合集
	utxos := make(map[string]TXOutputs)       // 未花费的output

	it := bc.Iterator()

	for it.HasNext() {
		block := it.Next()

		for _, tx := range block.Transactions {

			// 统计当前交易中已花费的别的交易中的output
			for _, in := range tx.Vin {
				prevTxID := hex.EncodeToString(in.Txid)

				if spendTxOutputs[prevTxID] == nil {
					spendTxOutputs[prevTxID] = NewIntSet()
				}

				spendTxOutputs[prevTxID].Add(in.Vout)
			}

			// 统计当前交易中的所有未被花费的交易
			curtTxID := hex.EncodeToString(tx.ID)
			outs := NewTxOutputs()
			for outIdx, out := range tx.Vout {
				if spendTxOutputs[curtTxID].Contains(outIdx) == false { // 当前output未被花费
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

	if amount <= 0 {
		log.Panic("ERROR: Wrong amount")
		return nil
	}

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
		addr := hex.EncodeToString(from.GetAddress())
		outputs = append(outputs, NewTxOutput(acc-amount, addr))
	}

	outputs = append(outputs, NewTxOutput(amount, to))

	tx := Transaction{nil, inputs, outputs, time.Now().UnixNano()}

	bc.SignTx(&tx, from.PrivateKey) // 签名交易
	tx.Hash()

	return &tx
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

	if tx.IsCoinbase() {
		return true
	}
	prevTxs := bc.getPrevTxs(tx)
	res := tx.Verify(prevTxs)
	return res
}

//  返回  >= amount 数量的 UTXOs
func (bc *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	return bc.utxoSet.FindSpendableOutputs(pubKeyHash, amount)
}
