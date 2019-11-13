package todo

import (
	"time"
)

type Todo struct {
	ID      uint      `gorm:"primary_key"`
	Title   string    `gorm:"title"`
	DueDate time.Time `gorm:"due_date"`
	Done    bool      `gorm:"done"`
}

func (t Todo) TableName() string {
	return "todos"
}
