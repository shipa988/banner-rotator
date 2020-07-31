package entities

type Slot struct {
	InnerID     uint `gorm:"UNIQUE_INDEX:innerid_pageid; NOT NULL"`
	Description string
}

type SlotRepository interface {
	AddSlot(pageURL string, slotInnerID uint, slotDescription string) error
	DeleteSlot(pageURL string, slotInnerID uint) error
	DeleteAllSlots(pageURL string) error
	GetSlotsByPageURL(pageURL string) (slots []Slot, err error)
}
