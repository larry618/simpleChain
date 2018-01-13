package main

import (
	"fmt"
	"bytes"
	"encoding/gob"
	"crypto/sha256"
	"crypto/ecdsa"
	"encoding/hex"
	"log"
	"crypto/rand"
	"math/big"
	"crypto/elliptic"
)

const subsidy = 10 // 是挖出新块的奖励金

// 一笔交易由一些输入（input）和输出（output）组合而来
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}



func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reword to '%s'", to)
	}



	txInput := TXInput{[]byte{}, -1, nil, nil}
	pubKeyHash := GetPubKeyHashFromAddr(to)
	txOutput := TXOutput{subsidy, pubKeyHash}

	tx := Transaction{nil, []TXInput{txInput}, []TXOutput{txOutput}}
	tx.Hash()
	return &tx
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTxs map[string]*Transaction) {

	for _, in := range tx.Vin {

		txID := hex.EncodeToString(in.Txid)
		if prevTxs[txID] == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}

	}

	copyTx := tx.TrimmedCopy()

	for inIdx, in := range copyTx.Vin {
		in.Signature = nil
		txID := hex.EncodeToString(in.Txid)
		in.PubKey = prevTxs[txID].Vout[in.Vout].PubKeyHash
		copyTx.Hash()

		in.PubKey = nil // 是使用tx来求hash得

		r, s, _ := ecdsa.Sign(rand.Reader, &privKey, copyTx.ID)
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inIdx].Signature = signature
	}
}

func (tx *Transaction) Verify(prevTxs map[string]*Transaction) bool {
	curve := elliptic.P256()
	copyTx := tx.TrimmedCopy()

	for inIdx, in := range tx.Vin {
		// 验证每一个 in 的 signature 是否都是合理的

		txID := hex.EncodeToString(in.Txid)

		copyIn := copyTx.Vin[inIdx]

		copyIn.Signature = nil
		copyIn.PubKey = prevTxs[txID].Vout[in.Vout].PubKeyHash
		copyTx.Hash()
		copyIn.PubKey = nil

		sign := in.Signature
		sLen := len(sign)
		r := big.Int{}
		s := big.Int{}

		r.SetBytes(sign[:(sLen / 2)])
		s.SetBytes(sign[(sLen / 2):])

		pubKey := in.PubKey
		keyLen := len(pubKey)
		x := big.Int{}
		y := big.Int{}

		x.SetBytes(pubKey[:(keyLen / 2)])
		y.SetBytes(pubKey[(keyLen / 2):])

		rowPubKey := ecdsa.PublicKey{curve, &x, &y}

		if ecdsa.Verify(&rowPubKey, copyTx.ID, &r, &s) == false {
			return false
		}

	}

	return true
}

func (tx *Transaction) TrimmedCopy() *Transaction {

	var inputs []TXInput
	var outputs []TXOutput

	for _, in := range tx.Vin {
		inputs = append(inputs, TXInput{in.Txid, in.Vout, nil, nil})
	}

	for _, out := range tx.Vout {
		outputs = append(outputs, TXOutput{out.Value, out.PubKeyHash})
	}

	return &Transaction{nil, inputs, outputs}
}

func (tx *Transaction) Hash() {

	var buffer bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&buffer)

	encoder.Encode(tx)

	hash = sha256.Sum256(buffer.Bytes())
	tx.ID = hash[:]
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}
