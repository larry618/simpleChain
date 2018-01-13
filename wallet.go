package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto"
)

const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallet() *Wallet {
	privKey, pubKey := NewKeyPair()
	return &Wallet{privKey, pubKey}
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {

	curve := elliptic.P256()
	privateKey, _ := ecdsa.GenerateKey(curve, rand.Reader)

	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...) // 两个slice拼接在一起

	return *privateKey, publicKey

}

func (w *Wallet) GetAddress() []byte {
	hashPubKey := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, hashPubKey...)
	checkSum := checkSum(versionedPayload)

	fullPayload := append(versionedPayload, checkSum...)
	return Base58Encode(fullPayload)
}

func GetPubKeyHashFromAddr(addr string) []byte {
	fullPayload := Base58Decode([]byte(addr))

	return fullPayload[1:len(fullPayload)-addressChecksumLen]
}

func checkSum(hashPubKey []byte) []byte {

	firstSum := sha256.Sum256(hashPubKey)
	return sha256.Sum256(firstSum[:])[:addressChecksumLen]

}

func HashPubKey(pubKey []byte) []byte {
	sum256 := sha256.Sum256(pubKey)

	ripHasher := crypto.RIPEMD160.New()
	ripHasher.Write(sum256[:])
	sum := ripHasher.Sum(nil)

	return sum

}
