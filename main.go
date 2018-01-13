package main

import "fmt"

func main() {
	cli := CLI{}
	cli.run()
}

func test() {
	hehe := []byte{}
	fmt.Println(hehe)        // []
	fmt.Println(hehe == nil) // false
	fmt.Println(len(hehe))   // 0
}
