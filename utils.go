package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"encoding/gob"
)

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func GobEncode(e interface{}) []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(e)

	if err != nil {
		log.Panic(err)
	}

	return buf.Bytes()
}

// 传入反序列化的数据和对象的指针
func GobDecode(data []byte, e interface{}) {
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	err := decoder.Decode(e)
	if err != nil {
		log.Panic(err)
	}
}

func SliceIterator(bytes [][]byte) func() ([]byte, bool) {

	idx := 0
	return func() ([]byte, bool) {
		value := bytes[idx]
		idx++
		hasNext := idx < len(bytes)
		return value, hasNext
	}
}