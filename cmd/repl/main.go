package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/efritz/gostgres/internal/execution/engine"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/sample"
	"github.com/efritz/gostgres/internal/syntax/parsing"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

type options struct {
	displayExpandedResults bool
	debug                  bool
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

	opts := options{}
	engine := engine.NewDefaultEngine()
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

			case "\\x":
				opts.displayExpandedResults = !opts.displayExpandedResults
				continue

			case "debug":
				opts.debug = !opts.debug
				continue

			case "load sample":
				if err := sample.LoadPagilaSampleSchemaAndData(engine); err != nil {
					return err
				}
				continue

			case "load sample schema":
				if err := sample.LoadPagilaSampleSchema(engine); err != nil {
					return err
				}
				continue

			case "load sample data":
				if err := sample.LoadPagilaSampleData(engine); err != nil {
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
				if err := handleQuery(engine, opts, line); err != nil {
					fmt.Printf("error: %s\n", err)
				}
			}

			buffer = strings.Join(parts, "\n")
		}
	}

	return nil
}

func handleQuery(engine *engine.Engine, opts options, input string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("query execution panic: %v\n%s", r, string(debug.Stack()))
		}
	}()

	start := time.Now()
	defer func() {
		if err == nil {
			fmt.Printf("Time: %s\n", time.Since(start))
		}
	}()

	rows, err := engine.QueryRows(protocol.Request{
		Query: input,
		Debug: opts.debug,
	})
	if err != nil {
		return err
	}

	if len(rows.Fields) > 0 {
		if opts.displayExpandedResults {
			fmt.Println(serialization.SerializeRowsExpanded(rows))
		} else {
			fmt.Println(serialization.SerializeRows(rows))
		}
	}

	return nil
}
