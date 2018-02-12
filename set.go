package main

import "fmt"

type Set interface {
	Add(e interface{}) Set
	Delete(e interface{}) Set
	Contains(e interface{}) bool
	Length() int
	GetRandomElement() interface{}
	IsEmpty() bool
}

// 使用 Map 来实现 StringSet, 相当于就是给 Map 取了个别名
type StringSet map[string]bool

func NewSet() StringSet {
	// return StringSet{}
	return make(map[string]bool)
}

func (set StringSet) Add(s string) StringSet {
	set[s] = true
	return set
}

func (set StringSet) Delete(s string) StringSet {
	delete(set, s)
	return set
}

func (set StringSet) Contains(s string) bool {
	return set[s]
}

func (set StringSet) Length() int {
	return len(set)
}

func (set StringSet) GetRandomElement() string {

	for value, _ := range set {
		return value
	}
	return ""
}

func (set StringSet) IsEmpty() bool {
	return set.Length() == 0
}

func (set StringSet) Iterator() func() (string, bool) {

	slice := set.ToSlice()
	index := 0

	return func() (value string, hasNest bool) {
		value = slice[index]
		index++
		hasNest = index < set.Length()
		return
	}
}

func (set StringSet) ToSlice() []string {

	var slice []string
	for value, _ := range set {
		slice = append(slice, value)
	}
	return slice
}

type IntSet map[int]bool

func NewIntSet() IntSet {
	return IntSet{}
}

func (set IntSet) Add(s int) IntSet {
	set[s] = true
	return set
}

func (set IntSet) Delete(s int) IntSet {
	delete(set, s)
	return set
}

func (set IntSet) Contains(s int) bool {
	return set[s]
}

func (set IntSet) Length() int {
	return len(set)
}

func (set IntSet) ToSlice() []int {

	var slice []int
	for value, _ := range set {
		slice = append(slice, value)
	}
	return slice
}

func (set IntSet) Iterator() func() (int, bool) {

	slice := set.ToSlice()
	index := 0

	return func() (value int, hasNest bool) {
		value = slice[index]
		index++
		hasNest = index < set.Length()
		return
	}
}

func (set IntSet) IsEmpty() bool {
	return set.Length() == 0
}


func setTest() {
	set := NewSet()

	set.Add("hehe")
	set.Add("heheda")
	set.Add("nihao")
	set.Add("hehe")

	fmt.Println(set.Length())

	//fmt.Println(set)
	//
	//for key, value := range set {
	//	fmt.Println(key, value)
	//}

	iterator := set.Iterator()
	for {
		value, hasNext := iterator()
		fmt.Println(value)

		if !hasNext {
			break
		}
	}
}