package main

import "bytes"

type TXInput struct {
	Txid      []byte // 存储的是之前交易的 ID
	Vout      int    // 该输出在之前那笔交易中所输出的索引
	Signature []byte
	PubKey    []byte
}



// 校验是否是由 当前PubKeyHash 发起的交易
func (in *TXInput) UsesKey(PubKeyHash []byte) bool {
	hashPubKey := HashPubKey(in.PubKey)
	return bytes.Compare(PubKeyHash, hashPubKey) == 0
}





