package stats

import (
	"encoding/json"
	"strings"
	"time"
)

// Status is a condensed representation of mastodon.Status
type Status struct {
	ID        string         `json:"-"`
	URI       string         `json:"uri" db:"uri" `
	Language  string         `json:"lang" db:"lang"`
	Content   string         `json:"content"`
	TagList   tagList        `json:"tag_list" db:"tag_list"`
	CreatedAt sqliteDatetime `json:"created_at" db:"created_at"`
}

// tagList converts a comma-separated list of tag names to a JSON array.
type tagList string

func (tl tagList) MarshalJSON() ([]byte, error) {
	parts := strings.Split(string(tl), ",")
	return json.Marshal(parts)
}

// sqliteDatetime converts a SQLite text date to a time.Time.
type sqliteDatetime time.Time

func (cv *sqliteDatetime) Scan(src any) error {
	t, err := time.Parse("2006-01-02 15:04:05-07:00", src.(string))

	if err != nil {
		return err
	}

	*cv = sqliteDatetime(t)
	return nil
}

func (cv sqliteDatetime) MarshalJSON() ([]byte, error) {
	return time.Time(cv).MarshalJSON()
}

func (cv sqliteDatetime) String() string {
	return time.Time(cv).String()
}
