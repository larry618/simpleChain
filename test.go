package main

import "fmt"

type MyMap map[string]string

func main()  {

	myMap := MyMap{}
	//var myMap MyMap

	myMap["asdf"] = "asdfasd"


	myMap["3ewdfaw"] = "nasdf"

	for key, value := range myMap {

		fmt.Println(key)
		fmt.Println(value)
	}
}
