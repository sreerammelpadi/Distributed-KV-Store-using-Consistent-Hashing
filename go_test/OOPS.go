package main

import (
	"fmt"
)

type Person struct {
	Name   string
	Height int
	Weight int
}

func (p *Person) Speak(sent string) bool {
	fmt.Printf("%v says: %v", p.Name, sent)
	return true
}

func (p *Person) BioData() {
	b := `
Name: %v,
Height: %v,
Weight: %v
	`
	fmt.Printf(b, p.Name, p.Height, p.Weight)
}

func main() {
	// newp := Person{Name: "Sreeram", Height: 173, Weight: 65}
	var newp Person = Person{Name: "Sreeram", Height: 173, Weight: 65}
	newp.Name = "Ayu"

	spoke := newp.Speak("Hey There\n")
	fmt.Printf("Spoke %v", spoke)
	newp.BioData()

}
