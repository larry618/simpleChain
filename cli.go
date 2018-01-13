package main

import (
	"flag"
	"os"
	"fmt"
)

type CLI struct {
	//bc *BlockChain
}

func addAddrCmdFlag(cmd *flag.FlagSet) *string {
	addrData := cmd.String("addr", "", "")
	return addrData

}

func (cli *CLI) run() {
	addBlockCmd := flag.NewFlagSet("addBlock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createBlockChain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)

	//addBlockData := addBlockCmd.String("data", "", "add the fucking data to a new block")
	createBlockChainData := addAddrCmdFlag(createBlockChainCmd)
	printChainData := addAddrCmdFlag(printChainCmd)
	getBalanceData := addAddrCmdFlag(getBalanceCmd)
	mineData := addAddrCmdFlag(mineCmd)

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

	default:
		fmt.Println("error")
		os.Exit(1)
	}

	//if addBlockCmd.Parsed() {
	//	if len(*addBlockData) == 0 {
	//		addBlockCmd.Usage()
	//		os.Exit(0)
	//	}
	//	cli.addBlock(*addBlockData)
	//}
	//
	//if printChainCmd.Parsed() {
	//	cli.printChain()
	//}

	switch {
	case createBlockChainCmd.Parsed():
		if len(*createBlockChainData) == 0 {
			createBlockChainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createBlockChainData)

		//case addBlockCmd.Parsed():
		//	if len(*addBlockData) == 0 {
		//		addBlockCmd.Usage()
		//		os.Exit(1)
		//	}
		//	cli.addBlock(*addBlockData)

	case printChainCmd.Parsed():
		if len(*printChainData) == 0 {
			printChainCmd.Usage()
			os.Exit(1)
		}

		cli.printChain(*printChainData)

	case getBalanceCmd.Parsed():
		cli.getBalance(*getBalanceData)

	case mineCmd.Parsed():
		cli.mine(*mineData)

	}

}

//func (cli *CLI) addBlock(data string) {
//	cli.bc.AddBlock(data)
//}

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

//func (cli *CLI) mine(addr string) {
//
//	bc := NewBlockChain(addr)
//	defer bc.db.Close()
//
//	bc.Mining(addr)
//}
//

func (cli *CLI) Send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.db.Close()

	tx := bc.NewUTXOTransaction(from, to, amount)
	bc.Mining([]*Transaction{tx}, from)
	fmt.Println("Success!")
}
