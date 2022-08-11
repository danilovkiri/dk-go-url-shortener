// Package main is a test package for testing custom code analyzer.
package main

import (
	"fmt"
	"os"
)

func notmain() {
	fmt.Println("Exiting indirectly from  goroutine.")
	os.Exit(0)
}

func main() {
	// implicitly exit from a goroutine which should not raise a flag in custom analyzer
	go notmain()
	// explicitly exit from main() via os.Exit() and raise a flag in custom analyzer
	fmt.Println("Exiting directly from main().")
	os.Exit(0) // want "call to os.Exit in main body"
}
