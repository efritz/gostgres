package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/lexing"
	"github.com/efritz/gostgres/internal/syntax/parsing"
	"github.com/efritz/gostgres/internal/tablespace"
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

	tables, err := tablespace.CreateSampleTables("tests/")
	if err != nil {
		return err
	}

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
			if err := handleQuery(tables, line); err != nil {
				fmt.Printf("error: %s\n", err)
			}
		}
	}

	return nil
}

func handleQuery(tables *tablespace.Tablespace, line string) (err error) {
	start := time.Now()
	defer func() {
		if err == nil {
			fmt.Printf("Time: %s\n", time.Since(start))
		}
	}()

	var explain bool
	line, explain = eatExplain(line)

	node, err := parsing.Parse(lexing.Lex(line), tables)
	if err != nil {
		return fmt.Errorf("failed to parse node: %s", err)
	}
	node.Optimize()

	if explain {
		fmt.Println(serialization.SerializePlanString(node))
		return nil
	}

	scanner, err := node.Scanner(scan.ScanContext{
		Tables: tables,
	})
	if err != nil {
		return err
	}
	rows, err := shared.NewRows(node.Fields())
	if err != nil {
		return err
	}
	rows, err = scan.ScanIntoRows(scanner, rows)
	if err != nil {
		return fmt.Errorf("failed to execute query: %s", err)
	}

	fmt.Println(serialization.SerializeRowsString(rows))
	return nil
}

const explainPrefix = "explain "

func eatExplain(line string) (string, bool) {
	if len(line) < len(explainPrefix) || strings.ToLower(line[:len(explainPrefix)]) != explainPrefix {
		return line, false
	}

	return line[len(explainPrefix):], true
}
