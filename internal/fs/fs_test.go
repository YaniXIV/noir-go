package fs

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestNargoParse(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Current directory:", dir)
	foo := parseNargo(".")
	/*
		for i := range foo.Dependencies {
			link := foo.Dependencies[i]

		}
	*/
	fmt.Println(foo.Dependencies)

}
