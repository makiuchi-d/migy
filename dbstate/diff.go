package dbstate

import (
	"fmt"
	"iter"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/anydiff"
)

func Diff(db *sqlx.DB, ss *Snapshot) (string, error) {
	var sb strings.Builder

	if err := diffTables(&sb, db, ss); err != nil {
		return "", err
	}
	if err := diffProcedures(&sb, db, ss); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func diffTables(sb *strings.Builder, db *sqlx.DB, ss *Snapshot) error {

	tbls, err := GetTables(db)
	if err != nil {
		return err
	}
	checked := make(map[string]struct{}, len(tbls))
	for _, tbl := range tbls {
		checked[tbl.Name] = struct{}{}

		sstbl, ok := ss.Tables[tbl.Name]
		if !ok {
			fmt.Fprintf(sb, "unexpected %q table found\n", tbl.Name)
			continue
		}
		create := strings.Split(tbl.Create, "\n")
		expcreate := strings.Split(sstbl.Create, "\n")

		edit := anydiff.Diff(expcreate, create, anydiff.Cmp)
		if edit.Distance() != 0 {
			fmt.Fprintf(sb, "create table %q differs:\n", tbl.Name)
			diffString(sb, edit, expcreate, create)
			continue
		}

		// todo: diffRecords
	}
	for name := range ss.Tables {
		if _, ok := checked[name]; !ok {
			fmt.Fprintf(sb, "missing %q table\n", name)
		}
	}

	return nil
}

func diffProcedures(sb *strings.Builder, db *sqlx.DB, ss *Snapshot) error {
	procs, err := GetProcedures(db)
	if err != nil {
		return nil
	}

	checked := make(map[string]struct{}, len(procs))
	for _, proc := range procs {
		checked[proc.Name] = struct{}{}

		ssproc, ok := ss.Procedures[proc.Name]
		if !ok {
			fmt.Fprintf(sb, "unexpected %q stored procedure found\n", proc.Name)
			continue
		}
		create := strings.Split(proc.Create, "\n")
		expcreate := strings.Split(ssproc.Create, "\n")

		edit := anydiff.Diff(expcreate, create, anydiff.Cmp)
		if edit.Distance() == 0 {
			continue
		}

		fmt.Fprintf(sb, "stored procedure %q differs:\n", proc.Name)
		diffString(sb, edit, expcreate, create)
	}

	for name := range ss.Procedures {
		if _, ok := checked[name]; !ok {
			fmt.Fprintf(sb, "missing %q stored procedure", name)
		}
	}

	return nil
}

func diffString[A, B any](sb *strings.Builder, edit anydiff.Edit, a []A, b []B) {

	next, stop := iter.Pull(func(yield func(string) bool) {
		i, j := 0, 0
		for _, op := range edit {
			switch op {
			case anydiff.Deletion:
				if !yield(fmt.Sprintf("-%v\n", a[i])) {
					return
				}
				i++
			case anydiff.Addition:
				if !yield(fmt.Sprintf("+%v\n", b[j])) {
					return
				}
				j++
			case anydiff.Keep:
				if !yield(fmt.Sprintf(" %v\n", b[j])) {
					return
				}
				i++
				j++
			}
		}
	})
	defer stop()

	cnt := 0

	buf, ok := next()
	if !ok {
		return
	}
	if buf[0] != ' ' {
		cnt = 2
	}

	for {
		s, ok := next()
		if !ok {
			break
		}
		if s[0] != ' ' {
			cnt = 3
		}

		if cnt > 0 {
			sb.WriteString(buf)
			cnt--
		} else if cnt == 0 {
			sb.WriteString("...\n")
			cnt--
		}

		buf = s
	}

	if cnt >= 1 {
		sb.WriteString(buf)
	}
	if cnt == 0 {
		sb.WriteString("...\n")
	}
}
