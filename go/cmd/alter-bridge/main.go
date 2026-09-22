package main

import (
	"fmt"
	"os"

	"github.com/rntgspr/alter-bridge/internal/broker"
)

func main() {
	root, err := broker.EnsureRoot(os.Getenv("ALTER_BRIDGE_ROOT"), os.Getenv("HOME"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("hello world")
	fmt.Println("mailbox root:", root)
}
