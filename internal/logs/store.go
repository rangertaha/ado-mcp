// SPDX-License-Identifier: MIT

package logs

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// dbFileName is the SQLite database file created within the data directory.
const dbFileName = "ado-mcp.db"

// DBPath returns the work-log database path within a data directory.
func DBPath(dataDir string) string { return filepath.Join(dataDir, dbFileName) }

// type Project struct {
// 	ID    uint   `gorm:"primaryKey" json:"id"`
// 	UID   string `json:"uid"` // unique identifier for the project
// 	Name  string `json:"name"`
// 	Alias string `json:"alias"` // short name or abbreviation
// 	URL   string `json:"url"`   // product home page
// }

// type Service struct {
// 	ID          uint   `gorm:"primaryKey" json:"id"`
// 	Name        string `json:"name"`
// 	Label       string `json:"label"`
// 	Description string `json:"description"`

// 	// Code
// 	Repo    string `json:"repository"` // source code repository URL
// 	Branch  string `json:"branch"`     // source code branch
// 	Version string `json:"version"`
// 	Build   string `json:"build"`
// 	Commit  string `json:"commit"`

// 	// Tags
// 	Environment string `json:"environment"`
// 	Hostname    string `json:"hostname"`
// 	Project     string `json:"project"`

// 	CreatedAt time.Time `json:"created"`
// }

// type Record struct {
// 	ID      uint      `gorm:"primaryKey" json:"id"`
// 	Start   time.Time `gorm:"index" json:"date"` // time
// 	End     time.Time `json:"end"`               // time
// 	Summary string    `json:"summary"`           // what was done.
// 	Tasks   []Task    `gorm:"foreignKey:RecordID" json:"tasks"`
// }

// type Task struct {
// 	ID          uint    `gorm:"primaryKey" json:"id"`
// 	Timestamp   string  `gorm:"index" json:"timestamp"`
// 	Description string  `json:"description"`
// 	RecordID    uint    `json:"recordId"`
// 	Service     Service `gorm:"foreignKey:ServiceID" json:"service"`
// 	ServiceID   uint    `json:"serviceId"`
// }

// Entry is a single daily work-log entry: a note of what was worked on, which
// Claude can turn into an Azure DevOps work item (ticket) and a 7pace time
// entry. The TicketCreated/HoursLogged flags track that synchronization so the
// same entry is not double-processed.
type Entry struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Date    string `gorm:"index" json:"date"` // work date, YYYY-MM-DD
	Summary string `json:"summary"`           // what was done
	Minutes int    `json:"minutes,omitempty"` // time spent, in minutes
	Project string `json:"project,omitempty"` // Azure DevOps project
	// WorkItemID links this entry to an Azure DevOps work item (0 if none yet).
	WorkItemID    int       `json:"workItemId,omitempty"`
	Activity      string    `json:"activity,omitempty"` // 7pace activity type
	Tags          string    `json:"tags,omitempty"`
	TicketCreated bool      `json:"ticketCreated"` // a work item was created from this entry
	HoursLogged   bool      `json:"hoursLogged"`   // time was logged to 7pace for this entry
	CreatedAt     time.Time `json:"createdAt"`

	UpdatedAt time.Time `json:"updatedAt"`
}

// Store is a GORM-backed SQLite store for daily work-log entries.
type Store struct {
	db *gorm.DB
}

// Open opens (creating if needed) the SQLite database at path and migrates the
// schema. GORM's own logger is silenced so it never writes to stdout, which the
// MCP protocol owns.
func Open(path string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		return nil, fmt.Errorf("opening database %s: %w", path, err)
	}
	if err := db.AutoMigrate(&Entry{}); err != nil {
		return nil, fmt.Errorf("migrating schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Add inserts a new work-log entry and returns it (with its assigned ID).
func (s *Store) Add(e *Entry) (*Entry, error) {
	if err := s.db.Create(e).Error; err != nil {
		return nil, err
	}
	return e, nil
}

// Get returns a single entry by ID.
func (s *Store) Get(id uint) (*Entry, error) {
	var e Entry
	if err := s.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

// ListFilter scopes a List query. Empty fields are ignored.
type ListFilter struct {
	Date     string // exact work date, YYYY-MM-DD
	From     string // inclusive lower bound, YYYY-MM-DD
	To       string // inclusive upper bound, YYYY-MM-DD
	Limit    int    // max rows (default 200)
	Unlogged bool   // only entries not yet logged to 7pace
}

// scope applies a ListFilter's where-clauses to a query.
func (s *Store) scope(f ListFilter) *gorm.DB {
	q := s.db.Model(&Entry{})
	if f.Date != "" {
		q = q.Where("date = ?", f.Date)
	}
	if f.From != "" {
		q = q.Where("date >= ?", f.From)
	}
	if f.To != "" {
		q = q.Where("date <= ?", f.To)
	}
	if f.Unlogged {
		q = q.Where("hours_logged = ?", false)
	}
	return q
}

// List returns work-log entries matching the filter, newest date first.
func (s *Store) List(f ListFilter) ([]Entry, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 200
	}
	var out []Entry
	if err := s.scope(f).Order("date desc, id desc").Limit(limit).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// Bucket is an aggregated total for a date or project.
type Bucket struct {
	Key     string `json:"key" jsonschema:"the date or project this total is for"`
	Entries int    `json:"entries" jsonschema:"number of entries"`
	Minutes int    `json:"minutes" jsonschema:"total minutes"`
}

// Summary aggregates work-log entries over a date range.
type Summary struct {
	Entries        int      `json:"entries" jsonschema:"total entries"`
	TotalMinutes   int      `json:"totalMinutes" jsonschema:"total minutes across entries"`
	TotalHours     float64  `json:"totalHours" jsonschema:"total hours (minutes / 60)"`
	TicketsCreated int      `json:"ticketsCreated" jsonschema:"entries with a ticket created"`
	HoursLogged    int      `json:"hoursLogged" jsonschema:"entries whose hours are logged to 7pace"`
	ByDate         []Bucket `json:"byDate" jsonschema:"per-day totals, newest first"`
	ByProject      []Bucket `json:"byProject" jsonschema:"per-project totals, most minutes first"`
}

// Summarize aggregates entries matching the filter by day and by project.
func (s *Store) Summarize(f ListFilter) (*Summary, error) {
	var entries []Entry
	if err := s.scope(f).Order("date desc").Find(&entries).Error; err != nil {
		return nil, err
	}

	sum := &Summary{}
	dates := map[string]*Bucket{}
	var dateOrder []string
	projects := map[string]*Bucket{}

	for _, e := range entries {
		sum.Entries++
		sum.TotalMinutes += e.Minutes
		if e.TicketCreated {
			sum.TicketsCreated++
		}
		if e.HoursLogged {
			sum.HoursLogged++
		}

		d := dates[e.Date]
		if d == nil {
			d = &Bucket{Key: e.Date}
			dates[e.Date] = d
			dateOrder = append(dateOrder, e.Date) // entries are date-desc ordered
		}
		d.Entries++
		d.Minutes += e.Minutes

		pkey := e.Project
		if pkey == "" {
			pkey = "(none)"
		}
		p := projects[pkey]
		if p == nil {
			p = &Bucket{Key: pkey}
			projects[pkey] = p
		}
		p.Entries++
		p.Minutes += e.Minutes
	}
	sum.TotalHours = float64(sum.TotalMinutes) / 60

	for _, d := range dateOrder {
		sum.ByDate = append(sum.ByDate, *dates[d])
	}
	for _, p := range projects {
		sum.ByProject = append(sum.ByProject, *p)
	}
	sort.Slice(sum.ByProject, func(i, j int) bool {
		if sum.ByProject[i].Minutes != sum.ByProject[j].Minutes {
			return sum.ByProject[i].Minutes > sum.ByProject[j].Minutes
		}
		return sum.ByProject[i].Key < sum.ByProject[j].Key
	})
	return sum, nil
}

// Update applies non-empty fields from changes to the entry with the given ID
// and returns the updated entry. Booleans are always applied.
func (s *Store) Update(id uint, changes map[string]any) (*Entry, error) {
	if len(changes) > 0 {
		if err := s.db.Model(&Entry{}).Where("id = ?", id).Updates(changes).Error; err != nil {
			return nil, err
		}
	}
	return s.Get(id)
}

// Delete removes an entry by ID. It returns gorm.ErrRecordNotFound when no entry
// has that ID, so callers don't report a successful delete for a row that never
// existed (GORM itself treats a zero-row delete as success).
func (s *Store) Delete(id uint) error {
	res := s.db.Delete(&Entry{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
