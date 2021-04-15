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

	// i := 0
loop:
	for {
		// i++
		// if i > 1 {
		// 	break
		// }

		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		// text := `
		// 	select *
		// 	from locations
		// 	join regions r
		// 	on (r.region_id = locations.region_id and r.region_id <> 5 or false)
		// 	where
		// 		(
		// 			locations.location_id < 10 or
		// 			(1 > 3 and 4 > 5)
		// 		) and
		// 		r.region_id < 5 * 5 and
		// 		(3 < 5 or 5 < 8 - 4)
		// `
		// // text := "INSERT INTO regions VALUES (4, 'foo'), (5, 'bar'), (6, 'baz')"
		// fmt.Printf("> %s\n", text)

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if text == "exit" {
			break loop
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

	return nil
}

func parseRelation(text string) (relations.Relation, error) {
	return syntax.Parse(syntax.Lex(text), tables)
}
