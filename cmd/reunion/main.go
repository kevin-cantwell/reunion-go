package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	reunion "github.com/kevin-cantwell/reunion-go"
	"github.com/kevin-cantwell/reunion-go/model"
	_ "github.com/kevin-cantwell/reunion-go/parser" // register v14 parser
)

var (
	ff  *model.FamilyFile
	idx *Index
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "reunion <command> <bundle>",
	Short: "Explore Reunion 14 family files",
	Long:  "A CLI for parsing and exploring Reunion 14 genealogy bundles (.familyfile14).",
}

// persistentPreRunLoad is a helper that loads the bundle from args[0].
// Commands that need the bundle set this as their PersistentPreRunE.
func loadBundleFromArgs(cmd *cobra.Command, args []string) error {
	path := args[0]
	var err error
	ff, err = reunion.Open(path, nil)
	if err != nil {
		return fmt.Errorf("opening bundle: %w", err)
	}
	idx = BuildIndex(ff)
	return nil
}

func parseIDArg(args []string, pos int) (uint32, error) {
	if pos >= len(args) {
		return 0, fmt.Errorf("missing person ID argument")
	}
	n, err := strconv.ParseUint(args[pos], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid ID %q: %w", args[pos], err)
	}
	return uint32(n), nil
}

func init() {
	rootCmd.AddCommand(jsonCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(personsCmd)
	rootCmd.AddCommand(personCmd)
	rootCmd.AddCommand(couplesCmd)
	rootCmd.AddCommand(placesCmd)
	rootCmd.AddCommand(eventsCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(ancestorsCmd)
	rootCmd.AddCommand(descendantsCmd)
	rootCmd.AddCommand(summaryCmd)
	rootCmd.AddCommand(treetopsCmd)
}

// --- json ---

var jsonCmd = &cobra.Command{
	Use:   "json <bundle>",
	Short: "Dump full FamilyFile as JSON",
	Args:  cobra.ExactArgs(1),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdJSON(ff)
	},
}

// --- stats ---

var statsCmd = &cobra.Command{
	Use:   "stats <bundle>",
	Short: "Summary counts (persons, families, places, etc.)",
	Args:  cobra.ExactArgs(1),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdStats(ff, idx)
		return nil
	},
}

// --- persons ---

var personsCmd = &cobra.Command{
	Use:   "persons <bundle>",
	Short: "List all persons",
	Args:  cobra.ExactArgs(1),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		surname, _ := cmd.Flags().GetString("surname")
		cmdPersons(ff, idx, surname)
		return nil
	},
}

func init() {
	personsCmd.Flags().String("surname", "", "Filter by surname")
}

// --- person ---

var personCmd = &cobra.Command{
	Use:   "person <bundle> <id>",
	Short: "Detail view for a person",
	Args:  cobra.ExactArgs(2),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args, 1)
		if err != nil {
			return err
		}
		asJSON, _ := cmd.Flags().GetBool("json")
		return cmdPerson(ff, idx, id, asJSON)
	},
}

func init() {
	personCmd.Flags().Bool("json", false, "Output as JSON")
}

// --- couples ---

var couplesCmd = &cobra.Command{
	Use:   "couples <bundle>",
	Short: "List all couples",
	Args:  cobra.ExactArgs(1),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdCouples(ff, idx)
		return nil
	},
}

// --- places ---

var placesCmd = &cobra.Command{
	Use:   "places <bundle>",
	Short: "List all places",
	Args:  cobra.ExactArgs(1),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdPlaces(ff)
		return nil
	},
}

// --- events ---

var eventsCmd = &cobra.Command{
	Use:   "events <bundle>",
	Short: "List all event type definitions",
	Args:  cobra.ExactArgs(1),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdEvents(ff)
		return nil
	},
}

// --- search ---

var searchCmd = &cobra.Command{
	Use:   "search <bundle> <query>",
	Short: "Search person names",
	Args:  cobra.ExactArgs(2),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdSearch(ff, args[1])
		return nil
	},
}

// --- ancestors ---

var ancestorsCmd = &cobra.Command{
	Use:   "ancestors <bundle> <id>",
	Short: "Walk ancestor tree",
	Args:  cobra.ExactArgs(2),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args, 1)
		if err != nil {
			return err
		}
		gen, _ := cmd.Flags().GetInt("generations")
		return cmdAncestors(idx, id, gen)
	},
}

func init() {
	ancestorsCmd.Flags().IntP("generations", "g", 10, "Max generation depth")
}

// --- descendants ---

var descendantsCmd = &cobra.Command{
	Use:   "descendants <bundle> <id>",
	Short: "Walk descendant tree",
	Args:  cobra.ExactArgs(2),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args, 1)
		if err != nil {
			return err
		}
		gen, _ := cmd.Flags().GetInt("generations")
		return cmdDescendants(idx, id, gen)
	},
}

func init() {
	descendantsCmd.Flags().IntP("generations", "g", 10, "Max generation depth")
}

// --- summary ---

var summaryCmd = &cobra.Command{
	Use:   "summary <bundle> <id>",
	Short: "Per-person stats (spouses, ancestors, descendants, treetops, surnames)",
	Args:  cobra.ExactArgs(2),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args, 1)
		if err != nil {
			return err
		}
		asJSON, _ := cmd.Flags().GetBool("json")
		return cmdSummary(idx, id, asJSON)
	},
}

func init() {
	summaryCmd.Flags().Bool("json", false, "Output as JSON")
}

// --- treetops ---

var treetopsCmd = &cobra.Command{
	Use:   "treetops <bundle> <id>",
	Short: "List terminal ancestors (persons with no parents)",
	Args:  cobra.ExactArgs(2),
	PreRunE: loadBundleFromArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args, 1)
		if err != nil {
			return err
		}
		return cmdTreetops(idx, id)
	},
}
