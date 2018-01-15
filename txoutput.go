package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TXOutput struct {
	Value      int    // 一定量的比特币
	PubKeyHash []byte // 公钥
}

func NewTxOutput(amount int, addr string) TXOutput {
	output := TXOutput{amount, nil}

	output.Lock(addr)
	return output
}

// 是否是转到 PubKeyHash 下
func (out *TXOutput) IsLockedWith(PubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, PubKeyHash) == 0
}

func (out *TXOutput) Lock(addr string) {
	pubKeyHash := GetPubKeyHashFromAddr(addr)
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(out)

	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func DeserializeOutput(b []byte) *TXOutput {

	var output TXOutput
	reader := bytes.NewReader(b)
	decoder := gob.NewDecoder(reader)

	err := decoder.Decode(&output)
	if err != nil {
		log.Panic(err)
	}

	return &output
}

type TXOutputs map[int]TXOutput

func NewTxOutputs() TXOutputs {
	return TXOutputs{}
}

func (outs TXOutputs) delete(i int) {

	delete(outs, i)
}

func (outs *TXOutputs) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(outs)

	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func DeserializeOutputs(b []byte) TXOutputs {

	var outputs TXOutputs
	reader := bytes.NewReader(b)
	decoder := gob.NewDecoder(reader)

	err := decoder.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
