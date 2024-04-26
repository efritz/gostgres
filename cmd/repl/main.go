package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/efritz/gostgres/internal/engine"
	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/sample"
	"github.com/efritz/gostgres/internal/serialization"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	l, err := readline.NewEx(&readline.Config{
		Prompt:            "gostgres \033[32m❯\033[0m ",
		HistoryFile:       "/tmp/gostgres.tmp",
		HistorySearchFold: true,
	})
	if err != nil {
		return err
	}
	defer l.Close()

	tables, err := sample.CreateSampleTables("tests/")
	if err != nil {
		return err
	}

	functions := functions.NewFunctionspace()
	functions.SetFunction("now", func(args []any) (any, error) { return time.Now(), nil })

	engine := engine.NewEngine(tables, functions)

	log.SetOutput(l.Stderr())
loop:
	for {
		line, err := l.Readline()
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)

		switch {
		case line == "":
			continue

		case line == "exit":
			break loop

		default:
			if err := handleQuery(engine, line); err != nil {
				fmt.Printf("error: %s\n", err)
			}
		}
	}

	return nil
}

func handleQuery(engine *engine.Engine, input string) (err error) {
	start := time.Now()
	defer func() {
		if err == nil {
			fmt.Printf("Time: %s\n", time.Since(start))
		}
	}()

	rows, err := engine.Query(input)
	if err != nil {
		return err
	}

	if len(rows.Fields) > 0 {
		fmt.Println(serialization.SerializeRowsString(rows))
	}

	return nil
}
