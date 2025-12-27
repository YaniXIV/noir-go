package fs

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestNargoParse(t *testing.T) {
	_, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Current directory:", dir)
	foo := parseNargo(".")
	for i := range foo.Dependencies {
		j := foo.Dependencies[i]
		v, ok := j["git"]
		if ok {
			fmt.Printf("link: (%s)\n", v)
		} else {
			v, ok = j["path"]
			if ok {
				fmt.Printf("File Path: (%v)\n", v)
			} else {
				panic("Could not resolve dependency type")
			}
		}

	}
	//fmt.Println(foo.Dependencies)

}
