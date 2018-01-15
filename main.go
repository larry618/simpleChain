package main

import (
	"fmt"
	"encoding/hex"
)

func main() {
	cli := CLI{}
	cli.run()
	//testPubKeyHash()
}

func test() {
	hehe := []byte{}
	fmt.Println(hehe)        // []
	fmt.Println(hehe == nil) // false
	fmt.Println(len(hehe))   // 0
}

func testPubKeyHash() {
	wallet := NewWallet()
	hashPubKey := HashPubKey(wallet.PublicKey)
	fmt.Printf("hashPubKey    %x\n", hashPubKey)
	address := wallet.GetAddress()
	addrStr := hex.EncodeToString(address)
	fmt.Printf("your address is %s\n", addrStr)

	pubKeyHashFromAddr := GetPubKeyHashFromAddr(addrStr)


	fmt.Printf("pubKeyHashFromAddr    %x\n", pubKeyHashFromAddr)
}
