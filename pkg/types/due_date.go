package types

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

const DueDateFormat = "2006-01-02"

type DueDate time.Time

func (d *DueDate) UnmarshalJSON(b []byte) error {
	b = b[1 : len(b)-1]
	v, err := time.Parse(DueDateFormat, string(b))
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal due_date")
	}

	*d = DueDate(v)
	return nil
}

func (d DueDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Format(DueDateFormat))
}

func (d *DueDate) Time() *time.Time {
	if d == nil {
		return nil
	}
	t := time.Time(*d)
	return &t
}
