package main

import (
	"fmt"
	"os"

	"github.com/efritz/gostgres/internal/repl"
)

func main() {
	if err := repl.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
