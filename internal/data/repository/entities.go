package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
)

type Slot struct {
	gorm.Model
	PageID uint `gorm:"UNIQUE_INDEX:innerid_pageid; NOT NULL"`
	entities.Slot
	BannerSlots []*BannerSlot //`gorm:"many2many:banner_slots;"`
}
type BannerSlot struct {
	gorm.Model
	BannerID uint `gorm:"UNIQUE_INDEX:BannerID_SlotID; NOT NULL"`
	SlotID   uint `gorm:"UNIQUE_INDEX:BannerID_SlotID; NOT NULL"`
	Events   []*BannerEvent
}

type Banner struct {
	gorm.Model
	entities.Banner
	BannerSlots []*BannerSlot //`gorm:"many2many:banner_slots;"`
}

type Page struct {
	gorm.Model
	entities.Page
	Slots []*Slot
}

type Group struct {
	gorm.Model
	entities.Group
	Events []*BannerEvent
}

type BannerEvent struct {
	gorm.Model
	entities.Action
	BannerSlotID uint `gorm:"UNIQUE_INDEX:BannerSlotID_GroupID; NOT NULL"`
	GroupID      uint `gorm:"UNIQUE_INDEX:BannerSlotID_GroupID; NOT NULL"`
}
