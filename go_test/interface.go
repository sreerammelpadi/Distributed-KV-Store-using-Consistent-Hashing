package main

import "fmt"

type list struct {
	vals []interface{}
}

func (l *list) append(x interface{}) {
	l.vals = append(l.vals, x)
}

func main() {
	var x interface{} // empty interfaces hold any type
	x = 10
	fmt.Println(x)

	x = "hi"
	fmt.Println(x)

	x = 5.01
	fmt.Println(x)

	// like a python list

	l := list{vals: []interface{}{1, "two", 3.0}}

	l.append(2)

	for _, v := range l.vals {
		fmt.Println(v)
	}

}
