package main

import (
	"flag"
	"os"
	"fmt"
	"encoding/hex"
)

type CLI struct {
	//bc *BlockChain
}

func addAddrCmdFlag(cmd *flag.FlagSet) *string {
	addrData := cmd.String("addr", "", "")
	return addrData
}

func (cli *CLI) run() {
	addBlockCmd := flag.NewFlagSet("addBlock", flag.ExitOnError)                 // 添加区块
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)             // 打印
	createBlockChainCmd := flag.NewFlagSet("createBlockChain", flag.ExitOnError) // 创建链
	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)             // 查看余额
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)                         // 挖矿
	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)         // 创建钱包
	sendCmd := flag.NewFlagSet("sendNetworkPacket", flag.ExitOnError)                         // 转账

	//addBlockData := addBlockCmd.String("Data", "", "add the fucking Data to a new block")
	createBlockChainAddr := addAddrCmdFlag(createBlockChainCmd)
	printChainAddr := addAddrCmdFlag(printChainCmd)
	getBalanceAddr := addAddrCmdFlag(getBalanceCmd)
	mineAddr := addAddrCmdFlag(mineCmd)

	fromAddr := sendCmd.String("from", "", "")
	toAddr := sendCmd.String("to", "", "")
	sendAmount := sendCmd.Int("amount", 0, "")

	switch os.Args[1] {
	case "addBlock":
		addBlockCmd.Parse(os.Args[2:])

	case "printChain":
		printChainCmd.Parse(os.Args[2:])

	case "createBlockChain":
		createBlockChainCmd.Parse(os.Args[2:])

	case "getBalance":
		getBalanceCmd.Parse(os.Args[2:])

	case "mine":
		mineCmd.Parse(os.Args[2:])

	case "createWallet":
		createWalletCmd.Parse(os.Args[2:])

	case "sendNetworkPacket":
		sendCmd.Parse(os.Args[2:])

	default:
		fmt.Println("error")
		os.Exit(1)
	}

	switch {
	case createBlockChainCmd.Parsed():
		if len(*createBlockChainAddr) == 0 {
			createBlockChainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createBlockChainAddr)

	case printChainCmd.Parsed():
		if len(*printChainAddr) == 0 {
			printChainCmd.Usage()
			os.Exit(1)
		}

		cli.printChain(*printChainAddr)

	case getBalanceCmd.Parsed():
		cli.getBalance(*getBalanceAddr)

	case createWalletCmd.Parsed():
		cli.createWallet()

	case mineCmd.Parsed():
		cli.mine(*mineAddr)

	case sendCmd.Parsed():
		cli.send(*fromAddr, *toAddr, *sendAmount)
	}
}

func (cli *CLI) mine(addr string) {
	bc := NewBlockChain(addr)
	defer bc.db.Close()
	bc.Mining(nil, addr)
}

func (cli *CLI) createWallet() {
	wallet := NewWallet()
	address := wallet.GetAddress()

	addrStr := hex.EncodeToString(address)
	wallet.SaveToFile(addrStr)
	fmt.Printf("your address is %s\n", addrStr)
}

func (cli *CLI) createBlockChain(addr string) {
	bc := NewBlockChain(addr)
	defer bc.db.Close()
}

func (cli *CLI) printChain(addr string) {

	bc := NewBlockChain(addr)
	defer bc.db.Close()

	iterator := bc.Iterator()
	for iterator.HasNext() {
		block := iterator.Next()
		fmt.Println(block)
	}
}

func (cli *CLI) getBalance(addr string) {
	bc := NewBlockChain(addr)
	defer bc.db.Close()

	balance := bc.GetBalance(addr)

	fmt.Printf("Balance of '%s': %d\n", addr, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.db.Close()
	wallet, _ := ReadWalletFromFile(from)
	tx := bc.NewUTXOTransaction(wallet, to, amount)

	fmt.Println(tx.IsCoinbase())

	bc.Mining([]*Transaction{tx}, from)
	fmt.Println("Success!")
}
