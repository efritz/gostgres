package repl

import (
	"bufio"
	"bytes"
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
			fmt.Printf("failed to execute query: %s\n\n", err)
			continue
		}

		displayValues(rows)
	}
}

func parseRelation(text string) (relations.Relation, error) {
	builtins := make(map[string]relations.Relation, len(builtinFactories))
	for k, factory := range builtinFactories {
		builtins[k] = factory()
	}

	return syntax.Parse(syntax.Lex(text), builtins)
}
