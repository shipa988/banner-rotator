package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
)
var _ entities.BannerRepository=(*PGRepo)(nil)
var _ entities.SlotRepository=(*PGRepo)(nil)
var _ entities.EventRepository=(*PGRepo)(nil)

type PGRepo struct {
	conn *gorm.DB
}

func NewPGRepo(db *gorm.DB) *PGRepo  {
	return &PGRepo{conn:db}
}

func (r *PGRepo) AddEvent(pageURL string, bannerInnerID, bannerID, userAge uint, userSex string) error {
	panic("implement me")
}

func (r *PGRepo) AddSlot(pageURL string, slotInnerID uint, slotDescription string) (err error) {
	/*page:=&Page{
		Page:  entities.Page{URL:pageURL},
		Slots: []*Slot{{
			Slot:    entities.Slot{
				InnerID:     slotInnerID,
				Description: slotDescription,
			},
		}},
	}
	fmt.Println("before",page)
	r.conn.Debug().Create(page)
	r.conn.Debug().Save(page)
	fmt.Println("after",page)
	return nil*/
	page:=&Page{}
	slot:=&Slot{}

	db:=r.conn.Debug().Where(Page{
		Page:  entities.Page{URL:pageURL},
	}).FirstOrCreate(page)
	if db.Error!=nil{
		return db.Error
	}

	db=r.conn.Debug().Where(Slot{
		PageID:  page.ID,
		Slot:    entities.Slot{
			InnerID:     slotInnerID,
			Description: slotDescription,
		},
	}).FirstOrCreate(slot)
	if db.Error!=nil{
		return db.Error
	}

	page=&Page{
		Page:  page.Page,
		Slots: []*Slot{slot},
	}
	
	fmt.Println(page)
	return nil
}

func (r *PGRepo) DeleteSlot(pageURL string, slotInnerID uint) error {
	panic("implement me")
}

func (r *PGRepo) DeleteAllSlots(pageURL string) error {
	panic("implement me")
}

func (r *PGRepo) GetSlotsByPageURL(pageURL string) (slots []entities.Slot, err error) {
	panic("implement me")
}

func (r *PGRepo) AddBanner(pageURL string, slotInnerID uint, bannerInnerID uint, bannerDescription string) (err error) {
	panic("implement me")
}

func (r *PGRepo) DeleteBanner(pageURL string, slotInnerID, bannerInnerID uint) error {
	panic("implement me")
}

func (r *PGRepo) DeleteAllBanners(pageURL string, slotID int) error {
	panic("implement me")
}

func (r *PGRepo) GetBannersBySlotID(pageURL string, slotInnerID int) (banners []entities.Banner, err error) {
	panic("implement me")
}

type Slot struct {
	gorm.Model
	PageID uint `gorm:"UNIQUE_INDEX:innerid_pageid_descr; NOT NULL"`
	entities.Slot
	Banners []*Banner `gorm:"many2many:banner_slots;"`
}

type Banner struct {
	gorm.Model
	entities.Banner
	Slots  []*Slot `gorm:"many2many:banner_slots;"`
	Events []*BannerClick
}

type Page struct {
	gorm.Model
	entities.Page
	Slots []*Slot
}

type Group struct {
	gorm.Model
	entities.Group
	Events []*BannerClick
}

type BannerClick struct {
	gorm.Model
	entities.Click
	BannerID uint
	GroupID  uint
}
