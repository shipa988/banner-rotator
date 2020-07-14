package entities

import (
	// used by gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Banner struct {
	InnerID uint
	Description string
}
