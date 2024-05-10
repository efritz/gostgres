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
	"github.com/efritz/gostgres/internal/sequence"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/syntax/parsing"
	"github.com/efritz/gostgres/internal/table"
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

	log.SetOutput(l.Stderr())

	tables := table.NewTablespace()
	sequences := sequence.NewSequencespace()
	functions := functions.NewDefaultFunctionspace()
	engine := engine.NewEngine(tables, sequences, functions)

	buffer := ""
loop:
	for {
		line, err := l.Readline()
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)

		if buffer == "" {
			switch line {
			case "exit":
				break loop

			case "load sample":
				if err := sample.LoadPagilaSampleSchemaAndData(engine); err != nil {
					return err
				}

				continue
			}
		}

		if buffer != "" {
			buffer += "\n"
		}
		buffer += line

		if buffer != "" {
			parts := parsing.SplitStatements(buffer)
			for len(parts) > 0 && parts[0][len(parts[0])-1] == ';' {
				line := parts[0]
				parts = parts[1:]
				if err := handleQuery(engine, line); err != nil {
					fmt.Printf("error: %s\n", err)
				}
			}

			buffer = strings.Join(parts, "\n")
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
