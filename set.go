package main

// 使用 Map 来实现 Set, 相当于就是给 Map 取了个别名
type Set map[string]bool

func NewSet() Set {
	// return Set{}
	return make(map[string]bool)
}

func (set Set) add(s string) {
	set[s] = true
}

func (set Set) delete(s string) {
	delete(set, s)
}

func (set Set) contains(s string) bool {
	return set[s]
}


type IntSet map[int]bool

func NewIntSet() IntSet {
	return IntSet{}
}

func (set IntSet) add(s int) {

	if set == nil {  // 这种骚操作都可以 ???
		set = NewIntSet()
	}

	set[s] = true
}

func (set IntSet) delete(s int) {
	delete(set, s)
}

func (set IntSet) contains(s int) bool {
	return set[s]
}
