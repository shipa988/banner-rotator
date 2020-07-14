package entities

type Slot struct {
	InnerID uint	`gorm:"UNIQUE_INDEX:innerid_pageid_descr; NOT NULL"`
	Description string	`gorm:"UNIQUE_INDEX:innerid_pageid_descr; NOT NULL"`
}
