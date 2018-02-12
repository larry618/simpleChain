package main

import (
	"fmt"
	"net"
	"log"
	"bytes"
	"io"
	"io/ioutil"
	"encoding/hex"
)

const (
	centralNode = "localhost:3000"
	protocol    = "tcp"
	nodeVersion = 1
)

var (
	knownNodes    = NewSet().Add(centralNode)     // 当前网络中的已知节点地址
	nodeAddress   string                          // 网络地址
	walletAddress string                          // 钱包地址
	bc            *BlockChain                     // 当前节点的区块
	blockMemPool  = make(map[string]*Block)       // 临时存储收到的区块
	txMemPool     = make(map[string]*Transaction) // 临时存储收到的交易
)

//  网络中的数据包
type packet struct {
	Version     int
	Command     string
	SourAddress string
	DestAddress string
	Data        []byte
}

// 向别人展示我有的区块或交易的id列表  dataList
type inv struct {
	// inv => inventory
	Type  string
	Items [][]byte // IDs
}

// 向别人获取一个区块或交易
type getData struct {
	Type string
	Item []byte // ID
}

func StartServer(nodeId, addr string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeId)
	walletAddress = addr
	bc = NewBlockChain(walletAddress, nodeId)
	if nodeAddress != centralNode {
		//knownNodes = append(knownNodes, nodeAddress)
		sendBlockHeight(centralNode)
	}

	listener, err := net.Listen(protocol, nodeAddress)

	if err != nil {
		log.Panic(err)
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Panic(err)
		}
		go handleConn(conn) // 每次请求产生一个线程
	}

}

func handleConn(conn net.Conn) {

	reqData, err := ioutil.ReadAll(conn)

	if err != nil {
		log.Panic(err)
	}

	packet := &packet{}
	GobDecode(reqData, packet)
	command := packet.Command
	//Version := packet.Version

	switch command {
	case "blockHeight":
		// 处理请求
		handleBlockHeightReq(packet)
	case "getBlocks":
		// 处理请求
		handleGetBlocksReq(packet)
	case "getData":
		// 处理请求
		handleGetDataReq(packet)
	case "inv":
		// 处理回复
		handleReceivedInv(packet)
	case "block":
		// 处理回复
		handleReceivedBlock(packet)
	default:
		fmt.Println("Unknown Command")
	}
}

// 处理收到一个块
func handleReceivedBlock(packet *packet) {

	var block *Block
	GobDecode(packet.Data, block)

	/*
	收到一个区块时处理的步骤
	  1. 看收到的区块中的高度 和 本地链中区块做对比
	  2. 如果高度相差一 并且 prevHash 的值正确, 开始验证区块数据, 加入到本地区块链中
	  3. 如我我的高度比他的高, 直接丢弃收到的区块
	  4. 他的比我的高
	*/

	otherHeight := block.Height

	myBlock := bc.GetLastBlock()

	myHeight := myBlock.Height

	heightDiff := otherHeight - myHeight

	// heightDiff == 0(区块和我一样多) 直接丢去收到的区块

	if heightDiff < 0 { // 区块比我少
		sendBlockHeight(packet.SourAddress)

	} else if heightDiff == 1 { // 刚好合适, 收到下一个区块
		if verifyBlock(block) {
			bc.AddBlock(block)

		}

	} else if heightDiff > 1 { // 区块比我多
		sendGetBlocks(heightDiff + 6)
	}

}

func organizeBlockMemPool() {

}

func verifyBlock(block *Block) bool {
	return true
}

// 获取区块的ID列表
func handleGetBlocksReq(req *packet) {
	var amount int64 // 要获取的区块的数量
	GobDecode(req.Data, &amount)
	hashes := bc.GetBlocksHash(amount)

	sendInv(req.SourAddress, &inv{"block", hashes})
}

// 处理 获取一个区块的数据或获取一笔交易的请求
func handleGetDataReq(req *packet) {

	getData := getData{}
	GobDecode(req.Data, getData)

	item := getData.Item
	switch getData.Type {
	case "block":
		sendBlock(req.SourAddress, item)
	case "tx":
		// 会有这种请求吗?!!
	}
}

// 我收到了一个response, 这个response展示了区块ID或交易ID的列表
func handleReceivedInv(packet *packet) {

	inv := &inv{}
	GobDecode(packet.Data, inv)

	fmt.Printf("Recevied inventory with %d %s\n", len(inv.Items), inv.Type)

	items := inv.Items
	switch inv.Type {

	case "block":
		// get block data from different node
		sendGetBlocksData(items)
	case "tx":

	default:
		fmt.Println("Unknown type")
	}
}

func sendGetBlocksData(hashes [][]byte) {

	for _, hash := range hashes {

		sendGetDataReq("block", hash)  // 每次随机选取一个已知节点发请求

	}
}

// 随机选取一个已知节点发请求
func sendGetDataReq(command string, date []byte) {

	destAddr := GetRandomNodeAddr()

	if destAddr == "" {
		log.Panic("No available node")
	}

	err := sendNetworkPacket(buildNetworkPacket(destAddr, command, date))

	if err != nil {
		knownNodes.Delete(destAddr)
		if knownNodes.IsEmpty() {
			log.Panic("No available node")
		}

		sendGetDataReq(command, date)
	}
}

func sendBlock(destAddr string, hash []byte) {

	b := bc.GetBlock(hash)
	sendNetworkPacket(buildNetworkPacket(destAddr, "block", b.Serialize()))
}

func sendInv(destAddr string, inv *inv) {
	sendNetworkPacket(buildNetworkPacket(destAddr, "inv", inv))
}

func handleBlockHeightReq(packet *packet) {
	var height int64
	GobDecode(packet.Data, &height)

	myHeight := bc.GetBestHeight()

	if myHeight > height {
		sendBlockHeight(nodeAddress) // 把我的区块高度发给他
	} else if myHeight < height {
		sendGetBlocks(height-myHeight+6) // 如果比我的多, 请求他的区块
	}

	if !knownNodes.Contains(packet.SourAddress) {
		knownNodes.Add(packet.SourAddress)
	}
}

// 根据手续费排序
func sortTx() []*Transaction {

	var s []*Transaction

	for id, tx := range txMemPool {
		if bc.VerifyTx(tx) {

		} else {
			delete(txMemPool, id)
		}
	}
}

func GetRandomNodeAddr() string {
	return knownNodes.GetRandomElement()
}

func sendGetBlocks(amount int64) {

	destAddr := GetRandomNodeAddr()

	if destAddr == "" {
		log.Panic("No available node")
	}
	sendNetworkPacket(buildNetworkPacket(destAddr, "getBlocks", amount))
}

// 用于交换两个节点间的区块高度
func sendBlockHeight(destAddr string) {

	height := bc.GetBestHeight()
	packet := buildNetworkPacket(destAddr, "blockHeight", height)

	sendNetworkPacket(packet)

	// 更新 Blockchain 的时候要加同步锁

}

func buildNetworkPacket(destAddr, command string, e interface{}) *packet {
	return &packet{nodeVersion, command, nodeAddress, destAddr, GobEncode(e)}
}

func sendNetworkPacket(packet *packet) error {
	return sendData(packet.DestAddress, GobEncode(packet))
}

func sendData(addr string, data []byte) error {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		knownNodes.Delete(addr)
		return err
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}

	return err
}
