package repl

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/efritz/gostgres/internal/relations"
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
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break loop
			} else {
				continue
			}
		} else if err == io.EOF {
			break loop
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

func handleQuery(line string) error {
	start := time.Now()

	relation, err := parseRelation(line)
	if err != nil {
		return fmt.Errorf("failed to parse relation: %s", err)
	}

	var buf bytes.Buffer
	relation.Serialize(&buf, 0)
	fmt.Printf("Query plan:\n\n%s\n", buf.String())

	relation.Optimize()

	buf.Reset()
	relation.Serialize(&buf, 0)
	fmt.Printf("Optimized query plan:\n\n%s\n", buf.String())

	fmt.Printf("Results:\n\n")
	rows, err := relations.ScanRows(relation)
	if err != nil {
		return fmt.Errorf("failed to execute query: %s", err)
	}
	elapsed := time.Since(start)

	displayValues(rows)
	fmt.Printf("\nTime: %s\n", elapsed)
	return nil
}

func parseRelation(text string) (relations.Relation, error) {
	return syntax.Parse(syntax.Lex(text), tables)
}
