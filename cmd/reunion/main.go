package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	reunion "github.com/kevin-cantwell/reunion-go"
	"github.com/kevin-cantwell/reunion-go/model"
	_ "github.com/kevin-cantwell/reunion-go/parser" // register v14 parser
)

const usage = `Usage: reunion <command> <bundle> [options]

Commands:
  json                  Dump full FamilyFile as JSON
  stats                 Summary counts (persons, families, places, etc.)
  persons               List all persons (--surname to filter)
  person <id>           Detail view for a person (--json for JSON)
  couples               List all couples
  places                List all places
  events                List all event type definitions
  search <query>        Search person names
  ancestors <id>        Walk ancestor tree (--generations N)
  descendants <id>      Walk descendant tree (--generations N)
  summary <id>          Per-person stats (--json for JSON)
  treetops <id>         List terminal ancestors
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	if cmd == "help" || cmd == "-h" || cmd == "--help" {
		fmt.Print(usage)
		return
	}

	// Commands that need a bundle path
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "error: missing bundle path\n\n%s", usage)
		os.Exit(1)
	}
	bundlePath := os.Args[2]

	// Parse remaining flags
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	surname := fs.String("surname", "", "Filter by surname")
	generations := fs.Int("generations", 10, "Max generation depth")
	asJSON := fs.Bool("json", false, "Output as JSON")

	// Determine where extra args start
	extraArgs := os.Args[3:]
	fs.Parse(extraArgs)

	// Load bundle
	ff, idx := loadBundle(bundlePath)

	switch cmd {
	case "json":
		cmdJSON(ff)

	case "stats":
		cmdStats(ff, idx)

	case "persons":
		cmdPersons(ff, idx, *surname)

	case "person":
		id := requireIDArg(fs, "person")
		cmdPerson(ff, idx, id, *asJSON)

	case "couples":
		cmdCouples(ff, idx)

	case "places":
		cmdPlaces(ff)

	case "events":
		cmdEvents(ff)

	case "search":
		query := fs.Arg(0)
		if query == "" {
			fmt.Fprintf(os.Stderr, "error: search requires a query\n")
			os.Exit(1)
		}
		cmdSearch(ff, query)

	case "ancestors":
		id := requireIDArg(fs, "ancestors")
		cmdAncestors(idx, id, *generations)

	case "descendants":
		id := requireIDArg(fs, "descendants")
		cmdDescendants(idx, id, *generations)

	case "summary":
		id := requireIDArg(fs, "summary")
		cmdSummary(idx, id, *asJSON)

	case "treetops":
		id := requireIDArg(fs, "treetops")
		cmdTreetops(idx, id)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", cmd, usage)
		os.Exit(1)
	}
}

func loadBundle(path string) (*model.FamilyFile, *Index) {
	ff, err := reunion.Open(path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening bundle: %v\n", err)
		os.Exit(1)
	}
	idx := BuildIndex(ff)
	return ff, idx
}

func requireIDArg(fs *flag.FlagSet, cmd string) uint32 {
	arg := fs.Arg(0)
	if arg == "" {
		fmt.Fprintf(os.Stderr, "error: %s requires a person ID\n", cmd)
		os.Exit(1)
	}
	n, err := strconv.ParseUint(arg, 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid ID %q\n", arg)
		os.Exit(1)
	}
	return uint32(n)
}
