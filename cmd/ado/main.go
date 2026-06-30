// SPDX-License-Identifier: MIT

// Command ado is the Azure DevOps command-line tool. It runs the Model Context
// Protocol server (`ado mcp`), checks connectivity (`ado test`), and manages the
// local work-log journal (`ado log ...`).
//
// Azure DevOps configuration is read from the environment (see package config).
// The `mcp` command communicates over stdio, the transport expected by MCP
// clients such as Claude Desktop/Code and Cursor.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/urfave/cli/v3"

	"github.com/rangertaha/ado-mcp/internal"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/ado/profile"
	"github.com/rangertaha/ado-mcp/internal/app"
	"github.com/rangertaha/ado-mcp/internal/config"
	"github.com/rangertaha/ado-mcp/internal/logs"
)

func main() {
	cmd := &cli.Command{
		Name:    "ado",
		Usage:   "Azure DevOps from the command line and as an MCP server",
		Version: internal.Version(),
		// A bare `ado` (no subcommand) runs the MCP server, matching the old
		// ado-mcp binary so existing MCP client configs keep working.
		Action: runMCP,
		Commands: []*cli.Command{
			mcpCommand(),
			testCommand(),
			logCommand(),
		},
		// Print errors ourselves (once) so the MCP stdio stream is never touched.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "ado: %v\n", err)
		os.Exit(1)
	}
}

// mcpCommand runs the MCP server over stdio.
func mcpCommand() *cli.Command {
	return &cli.Command{
		Name:   "mcp",
		Usage:  "Run the MCP server over stdio (for Claude Desktop/Code, Cursor, ...)",
		Action: runMCP,
	}
}

// runMCP assembles and serves the MCP server over stdio. It backs both the
// `mcp` subcommand and a bare `ado` invocation.
func runMCP(ctx context.Context, _ *cli.Command) error {
	// Load a local .env file (if present) before reading configuration, so
	// credentials can live in a dotenv file for local development. Real shell
	// variables take precedence over the file.
	if err := config.LoadEnvFile(config.EnvFile); err != nil {
		log.Printf("ado: reading %s: %v", config.EnvFile, err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration error:\n%w", err)
	}

	ver := internal.Version()
	srv, cleanup, err := app.Assemble(cfg, ver)
	if err != nil {
		return err
	}
	defer cleanup()

	log.Printf("ado-mcp %s starting: %d tools, %d prompts across toolsets %v (read-only=%v, home=%s)",
		ver, srv.ToolCount(), srv.PromptCount(), srv.Toolsets(), cfg.ReadOnly, cfg.HomeDir)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return srv.Run(ctx, &mcp.StdioTransport{})
}

// testCommand verifies connectivity and credentials against Azure DevOps.
func testCommand() *cli.Command {
	return &cli.Command{
		Name:  "test",
		Usage: "Test connectivity and credentials against Azure DevOps",
		Action: func(ctx context.Context, _ *cli.Command) error {
			if err := config.LoadEnvFile(config.EnvFile); err != nil {
				log.Printf("ado: reading %s: %v", config.EnvFile, err)
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("configuration error:\n%w", err)
			}

			clients, err := ado.NewClients(cfg.OrgURL, cfg.PAT)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			me, err := profile.Check(ctx, clients)
			if err != nil {
				return fmt.Errorf("connecting to %s: %w", cfg.OrgURL, err)
			}

			fmt.Printf("OK  connected to %s\n", cfg.OrgURL)
			who := me.DisplayName
			if me.EmailAddress != "" {
				who = fmt.Sprintf("%s <%s>", who, me.EmailAddress)
			}
			fmt.Printf("    authenticated as %s\n", who)
			fmt.Printf("    read-only=%v\n", cfg.ReadOnly)
			return nil
		},
	}
}

// logCommand groups the local work-log journal subcommands.
func logCommand() *cli.Command {
	return &cli.Command{
		Name:  "log",
		Usage: "Manage the local work-log journal",
		Commands: []*cli.Command{
			logAddCommand(),
			logListCommand(),
			logUpdateCommand(),
			logDeleteCommand(),
		},
	}
}

func logAddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a work-log entry",
		ArgsUsage: "<summary...>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "date", Usage: "work date, YYYY-MM-DD (default today)"},
			&cli.IntFlag{Name: "minutes", Aliases: []string{"m"}, Usage: "time spent, in minutes"},
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "Azure DevOps project"},
			&cli.IntFlag{Name: "work-item", Aliases: []string{"w"}, Usage: "linked work item ID"},
			&cli.StringFlag{Name: "activity", Aliases: []string{"a"}, Usage: "7pace activity type"},
			&cli.StringFlag{Name: "tags", Usage: "comma-separated tags"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			summary := strings.TrimSpace(strings.Join(cmd.Args().Slice(), " "))
			if summary == "" {
				return fmt.Errorf(`a summary is required, e.g. ado log add "fixed the deploy pipeline"`)
			}
			date := cmd.String("date")
			if date == "" {
				date = time.Now().Format("2006-01-02")
			}

			store, closeStore, err := openStore()
			if err != nil {
				return err
			}
			defer closeStore()

			e, err := store.Add(&logs.Entry{
				Date:       date,
				Summary:    summary,
				Minutes:    cmd.Int("minutes"),
				Project:    cmd.String("project"),
				WorkItemID: cmd.Int("work-item"),
				Activity:   cmd.String("activity"),
				Tags:       cmd.String("tags"),
			})
			if err != nil {
				return err
			}
			fmt.Printf("added #%d  %s  %s\n", e.ID, e.Date, e.Summary)
			return nil
		},
	}
}

func logListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List work-log entries (newest first)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "date", Usage: "exact work date, YYYY-MM-DD"},
			&cli.StringFlag{Name: "from", Usage: "inclusive lower bound, YYYY-MM-DD"},
			&cli.StringFlag{Name: "to", Usage: "inclusive upper bound, YYYY-MM-DD"},
			&cli.IntFlag{Name: "limit", Value: 200, Usage: "maximum rows to return"},
			&cli.BoolFlag{Name: "unlogged", Usage: "only entries not yet logged to 7pace"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			store, closeStore, err := openStore()
			if err != nil {
				return err
			}
			defer closeStore()

			entries, err := store.List(logs.ListFilter{
				Date:     cmd.String("date"),
				From:     cmd.String("from"),
				To:       cmd.String("to"),
				Limit:    cmd.Int("limit"),
				Unlogged: cmd.Bool("unlogged"),
			})
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Println("no entries")
				return nil
			}
			for _, e := range entries {
				project := e.Project
				if project == "" {
					project = "-"
				}
				flags := dash(e.TicketCreated, "T") + dash(e.HoursLogged, "H")
				fmt.Printf("#%-4d %s  %4dm  %-14s [%s]  %s\n",
					e.ID, e.Date, e.Minutes, project, flags, e.Summary)
			}
			return nil
		},
	}
}

func logUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update fields of a work-log entry",
		ArgsUsage: "<id>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "summary", Aliases: []string{"s"}, Usage: "what you worked on"},
			&cli.StringFlag{Name: "date", Usage: "work date, YYYY-MM-DD"},
			&cli.IntFlag{Name: "minutes", Aliases: []string{"m"}, Usage: "time spent, in minutes"},
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "Azure DevOps project"},
			&cli.IntFlag{Name: "work-item", Aliases: []string{"w"}, Usage: "linked work item ID"},
			&cli.StringFlag{Name: "activity", Aliases: []string{"a"}, Usage: "7pace activity type"},
			&cli.StringFlag{Name: "tags", Usage: "comma-separated tags"},
			&cli.BoolFlag{Name: "ticket-created", Usage: "mark that a work item was created"},
			&cli.BoolFlag{Name: "hours-logged", Usage: "mark that hours were logged to 7pace"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			id, err := parseID(cmd.Args().First())
			if err != nil {
				return err
			}

			changes := map[string]any{}
			if cmd.IsSet("summary") {
				changes["summary"] = cmd.String("summary")
			}
			if cmd.IsSet("date") {
				changes["date"] = cmd.String("date")
			}
			if cmd.IsSet("minutes") {
				changes["minutes"] = cmd.Int("minutes")
			}
			if cmd.IsSet("project") {
				changes["project"] = cmd.String("project")
			}
			if cmd.IsSet("work-item") {
				changes["work_item_id"] = cmd.Int("work-item")
			}
			if cmd.IsSet("activity") {
				changes["activity"] = cmd.String("activity")
			}
			if cmd.IsSet("tags") {
				changes["tags"] = cmd.String("tags")
			}
			if cmd.IsSet("ticket-created") {
				changes["ticket_created"] = cmd.Bool("ticket-created")
			}
			if cmd.IsSet("hours-logged") {
				changes["hours_logged"] = cmd.Bool("hours-logged")
			}
			if len(changes) == 0 {
				return fmt.Errorf("nothing to update; pass at least one field flag (see ado log update --help)")
			}

			store, closeStore, err := openStore()
			if err != nil {
				return err
			}
			defer closeStore()

			e, err := store.Update(id, changes)
			if err != nil {
				return err
			}
			fmt.Printf("updated #%d  %s  %s\n", e.ID, e.Date, e.Summary)
			return nil
		},
	}
}

func logDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Aliases:   []string{"rm"},
		Usage:     "Delete a work-log entry",
		ArgsUsage: "<id>",
		Action: func(_ context.Context, cmd *cli.Command) error {
			id, err := parseID(cmd.Args().First())
			if err != nil {
				return err
			}
			store, closeStore, err := openStore()
			if err != nil {
				return err
			}
			defer closeStore()

			if err := store.Delete(id); err != nil {
				return err
			}
			fmt.Printf("deleted #%d\n", id)
			return nil
		},
	}
}

// openStore resolves the local data directory and opens the work-log journal.
// It does not require Azure DevOps credentials.
func openStore() (*logs.Store, func(), error) {
	_, data, err := config.Dirs()
	if err != nil {
		return nil, nil, err
	}
	store, err := app.OpenJournal(&config.Config{DataDir: data})
	if err != nil {
		return nil, nil, err
	}
	return store, func() { _ = store.Close() }, nil
}

// parseID parses a positional work-log entry ID.
func parseID(s string) (uint, error) {
	if strings.TrimSpace(s) == "" {
		return 0, fmt.Errorf("an entry ID is required, e.g. ado log delete 12")
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid entry ID %q", s)
	}
	return uint(n), nil
}

// dash returns the marker when set, else "-".
func dash(set bool, marker string) string {
	if set {
		return marker
	}
	return "-"
}
