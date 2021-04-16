package repl

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/syntax"
)

func Start() error {
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
			if err := handleQuery(line); err != nil {
				fmt.Printf("error: %s\n", err)
			}
		}
	}

	return nil
}

func handleQuery(line string) (err error) {
	start := time.Now()
	defer func() {
		if err == nil {
			fmt.Printf("Time: %s\n", time.Since(start))
		}
	}()

	var explain bool
	line, explain = eatExplain(line)

	node, err := syntax.Parse(syntax.Lex(line), tables)
	if err != nil {
		return fmt.Errorf("failed to parse node: %s", err)
	}
	node.Optimize()

	if explain {
		var buf bytes.Buffer
		serializePlan(&buf, node)
		fmt.Println(buf.String())
		return nil
	}

	rows, err := nodes.ScanRows(node)
	if err != nil {
		return fmt.Errorf("failed to execute query: %s", err)
	}

	var buf bytes.Buffer
	serializeRows(&buf, rows)
	fmt.Println(buf.String())
	return nil
}

const explainPrefix = "explain "

func eatExplain(line string) (string, bool) {
	if len(line) < len(explainPrefix) || strings.ToLower(line[:len(explainPrefix)]) != explainPrefix {
		return line, false
	}

	return line[len(explainPrefix):], true
}
