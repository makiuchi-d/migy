package migrations

import (
	"iter"
	"time"

	"github.com/jmoiron/sqlx"
)

type Status struct {
	*Migration
	Applied time.Time
	DBTitle string // mismatched title
}

func (s Status) IsApplied() bool {
	return s.Applied != time.Time{}
}

type History struct {
	Id      int       `db:"id"`
	Applied time.Time `db:"applied"`
	Title   string    `db:"title"`
}

type Histories []History

func LoadHistories(db *sqlx.DB) (Histories, error) {
	const sql = "SELECT id, applied, title FROM _migrations ORDER BY id"
	var recs []History
	err := db.Select(&recs, sql)
	return recs, err
}

func (hs Histories) CurrentNum() int {
	if len(hs) == 0 {
		return -1
	}
	return hs[len(hs)-1].Id
}

func BuildStatus(migs Migrations, hists []History) iter.Seq[Status] {
	return func(yield func(Status) bool) {
		var i, j int
		for i < len(migs) && j < len(hists) {
			m, h := migs[i], &hists[j]

			if m.Number < h.Id {
				if !yield(status(m, nil)) {
					return
				}
				i++
				continue
			}
			if m.Number > h.Id {
				if !yield(status(nil, h)) {
					return
				}
				j++
				continue
			}
			if !yield(status(m, h)) {
				return
			}
			i++
			j++
		}
		for ; i < len(migs); i++ {
			if !yield(status(migs[i], nil)) {
				return
			}
		}
		for ; j < len(hists); j++ {
			if !yield(status(nil, &hists[j])) {
				return
			}
		}
	}
}

func status(m *Migration, h *History) Status {
	if h == nil {
		return Status{
			Migration: m,
		}
	}

	if m == nil {
		return Status{
			Migration: &Migration{
				Number: h.Id,
				Title:  h.Title,
			},
			Applied: h.Applied,
		}
	}

	s := Status{
		Migration: m,
		Applied:   h.Applied,
	}
	if m.Title != h.Title {
		s.DBTitle = h.Title
	}
	return s
}
