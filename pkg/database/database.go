package database

import (
	"github.com/go-kit/kit/log"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type Config struct {
	Address string
}

type Scope = func(*gorm.DB) *gorm.DB

func New(cfg Config, logger log.Logger) (*gorm.DB, error) {
	db, err := gorm.Open("postgres", cfg.Address)
	if err != nil {
		return nil, err
	}
	return db.LogMode(true), nil
}
