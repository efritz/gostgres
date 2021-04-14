package repl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/efritz/gostgres/internal/relations"
	"github.com/efritz/gostgres/internal/syntax"
)

func Start() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if text == "exit" {
			return nil
		}

		relation, err := parseRelation(text)
		if err != nil {
			fmt.Printf("failed to parse relation: %s\n\n", err)
			continue
		}

		rows, err := relations.ScanRows(relation)
		if err != nil {
			fmt.Printf("failed to execute query: %s\n\n", err)
			continue
		}

		displayValues(rows)
	}
}

func parseRelation(text string) (relations.Relation, error) {
	return syntax.Parse(syntax.Lex(text), builtins)
}
