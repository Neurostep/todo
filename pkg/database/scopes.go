package database

import (
	"github.com/jinzhu/gorm"
)

func WithLimit(limit uint32) Scope {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Limit(limit)
	}
}

func WithOffset(offset uint32) Scope {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Offset(offset)
	}
}
