package repl

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/relations"
	"github.com/efritz/gostgres/internal/syntax"
)

func Start() error {
	// reader := bufio.NewReader(os.Stdin)

	i := 0
loop:
	for {
		i++
		if i > 1 {
			break
		}

		// fmt.Print("> ")
		// text, err := reader.ReadString('\n')
		// if err != nil {
		// 	return err
		// }
		// text := "select * from locations join regions r on (r.region_id=locations.region_id and r.region_id<>5 or false ) where (locations.location_id < 10 or (1 > 3 and 4 > 5)) and r.region_id < 5*5 and (3 < 5 or 5 < 8 - 4)" // TODO: TEMP
		text := "SELECT * FROM (SELECT e.employee_id, e.first_name, e.last_name, m.first_name, m.last_name FROM employees e JOIN employees m ON m.employee_id = e.manager_id WHERE e.employee_id < 10 AND m.employee_id < 15 ORDER BY m.last_name) as s WHERE s.employee_id<=7 and s.employee_id > 1"

		fmt.Printf("> %s\n", text)

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
	return syntax.Parse(syntax.Lex(text), builtinFactories)
}
