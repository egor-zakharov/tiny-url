package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello, world!")
	os.Exit(0) // want "can't use osExit with exit code"
}
