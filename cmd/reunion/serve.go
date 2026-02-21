package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	reunion "github.com/kevin-cantwell/reunion-go"
	"github.com/kevin-cantwell/reunion-go/web"
)

var serveCmd = &cobra.Command{
	Use:   "serve <bundle>",
	Short: "Start web server to explore the family file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, _ := cmd.Flags().GetString("addr")

		path := args[0]
		familyFile, err := reunion.Open(path, nil)
		if err != nil {
			return fmt.Errorf("opening bundle: %w", err)
		}

		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

		srv := web.New(familyFile, logger)

		logger.Info("starting server",
			"addr", addr,
			"persons", len(familyFile.Persons),
			"families", len(familyFile.Families),
			"places", len(familyFile.Places),
			"events", len(familyFile.EventDefinitions),
			"sources", len(familyFile.Sources),
			"notes", len(familyFile.Notes),
		)

		// Watch for bundle changes and reload automatically.
		go func() {
			if err := srv.Watch(path); err != nil {
				logger.Error("file watcher stopped", "err", err)
			}
		}()

		fmt.Fprintf(os.Stderr, "Reunion Explorer running at http://localhost%s\n", addr)
		return srv.ListenAndServe(addr)
	},
}

func init() {
	serveCmd.Flags().StringP("addr", "a", ":8080", "Listen address")
}
