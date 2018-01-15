package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"bytes"
	"log"
	"io/ioutil"
	"fmt"
	"os"
	"encoding/hex"
	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x01)
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallet() *Wallet {
	privKey, pubKey := newKeyPair()
	wallet := Wallet{privKey, pubKey}
	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {

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

	addrBytes, _ := hex.DecodeString(addr)
	fullPayload := Base58Decode(addrBytes)

	fmt.Println(addr)
	fmt.Println(len(fullPayload))
	return fullPayload[1:len(fullPayload)-addressChecksumLen]
}

func checkSum(hashPubKey []byte) []byte {

	firstSum := sha256.Sum256(hashPubKey)
	sum256 := sha256.Sum256(firstSum[:])
	return sum256[:addressChecksumLen]

}

func HashPubKey(pubKey []byte) []byte {

	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

func (w *Wallet) SaveToFile(fileName string) bool {

	if _, err := os.Stat(fileName); err == nil {
		fmt.Printf("%s is exist.\n", fileName)
		return false
	}

	var buf bytes.Buffer
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&buf)
	encoder.Encode(w)

	ioutil.WriteFile(fileName, buf.Bytes(), 0644)
	return true
}

func ReadWalletFromFile(fileName string) (*Wallet, error) {

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		fmt.Printf("%s is not exist.\n", fileName)
		return nil, err
	}

	gob.Register(elliptic.P256())

	content, _ := ioutil.ReadFile(fileName)
	buf := bytes.NewReader(content)

	decoder := gob.NewDecoder(buf)
	var wallet Wallet
	decoder.Decode(&wallet)
	return &wallet, nil
}

// 返回 16 进制的 private key public key
func (w *Wallet) String() string {

	var buf bytes.Buffer
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(w.PrivateKey)

	if err != nil {
		log.Panic(err)
	}

	privKeyBytes := buf.Bytes()

	fmt.Println(w.PrivateKey)

	res := fmt.Sprintf("Public  Key: %s\n", hex.EncodeToString(w.PublicKey))
	res += fmt.Sprintf("Private Key: %s\n", hex.EncodeToString(privKeyBytes))

	return res;
}
