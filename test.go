package main

import (
	"fmt"
)

type Name interface {
	Name() string
}

type A struct {
}

func (self A) say() {
	println(self.Name())
}

func (self A) sayReal(child Name) {
	fmt.Println(child.Name())
}

func (self A) Name() string {
	return "I'm A"
}

type B struct {
	A
}

func (self B) Name() string {
	return "I'm B"
}

type C struct {
	B
}

func main() {
	b := B{}
	b.say()                     //I'm A
	fmt.Println(Name(b).Name()) //I'm B

	c := C{}
	c.say()                     //I'm A
	fmt.Println(Name(c).Name()) //I'm A
}
