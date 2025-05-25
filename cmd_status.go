package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	"github.com/makiuchi-d/migy/migrations"
)

var statusCheck bool

var cmdStatus = &cobra.Command{
	Use:   "status [flags] [dsn|dumpfile]",
	Short: "Show the status of each migration",
	Long: `Show the status of each migration.
Compares migration files with the given database or dump file
and summarizes their current status.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return showStatus(nil, targetDir)
		}

		db, err := openDsnOrDumpfile(args[0])
		if err != nil {
			return err
		}

		return showStatus(db, targetDir)
	},
}

func init() {
	cmd.AddCommand(cmdStatus)
	cmdStatus.Flags().BoolVarP(&statusCheck, "check", "c", false, "check migration reversibility")
}

func showStatus(db *sqlx.DB, dir string) error {
	var hists []migrations.History
	if db != nil {
		var err error
		hists, err = migrations.LoadHistories(db)
		if err != nil {
			return err
		}
	}

	migs, err := migrations.Load(dir)
	if err != nil {
		return err
	}

	for st := range migrations.BuildStatus(migs, hists) {
		fmt.Println(formatStatus(st))
	}

	return nil
}

func formatStatus(st migrations.Status) string {
	var b []byte
	b = fmt.Appendf(b, "%06d\t", st.Number)

	if st.UpDown {
		b = fmt.Append(b, "⏫⏬")
	} else {
		b = fmt.Append(b, "　　")
	}
	if st.Snapshot {
		b = fmt.Append(b, "⏺\t")
	} else {
		b = fmt.Append(b, "　\t")
	}

	if st.IsApplied() {
		b = st.Applied.AppendFormat(b, "✅2006-01-02 15:04:05\t")
	} else {
		b = fmt.Append(b, "　0000-00-00 --:--:--\t")
	}

	if st.DBTitle == "" {
		b = fmt.Appendf(b, "%q", st.Title)
	} else {
		b = fmt.Appendf(b, "⚠%q DB:%q", st.Title, st.DBTitle)
	}

	return string(b)
}
