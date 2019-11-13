package todo

import (
	db "github.com/Neurostep/todo/pkg/database"
	"github.com/jinzhu/gorm"
)

func withTodoID(ID uint) db.Scope {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id = ?", ID)
	}
}

func buildPaginatedScope(pg PaginateTodos) []db.Scope {
	res := []db.Scope{}

	if pg.Offset != 0 {
		res = append(res, db.WithOffset(pg.Offset))
	}

	if pg.Limit != 0 {
		res = append(res, db.WithLimit(pg.Limit))
	}

	return res
}
