//nolint: dupl
package repository

import (
	"context"
	"fmt"

	"github.com/jinzhu/gorm"
	// used by gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"

	"github.com/shipa988/banner_rotator/internal/data/logger"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
)

const ErrZeroValue = "params has zerovalues"

var _ entities.BannerRepository = (*PGRepo)(nil)
var _ entities.SlotRepository = (*PGRepo)(nil)
var _ entities.PageRepository = (*PGRepo)(nil)
var _ entities.ActionRepository = (*PGRepo)(nil)
var _ entities.GroupRepository = (*PGRepo)(nil)

type PGRepo struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewPGRepo(db *gorm.DB, logger logger.Logger, isDebug bool) *PGRepo {
	if isDebug {
		db = db.Debug()
		db.LogMode(true)
		db.SetLogger(logger)
	}
	return &PGRepo{db: db, logger: logger}
}

func (r *PGRepo) CreateDB() {
	r.logger.Log(context.Background(), "db creating...")
	//Миграция базы данных
	if err := r.db.AutoMigrate(&Banner{}, &Slot{}, &Page{}, &Group{}, &BannerEvent{}, &BannerSlot{}).Error; err != nil {
		r.logger.Log(context.Background(), errors.Wrapf(err, "can't create db"))
	}
	groups := []*entities.Group{
		{
			Description: "young man",
			Sex:         "man",
			MinAge:      0,
			MaxAge:      40,
		},
		{
			Description: "young women",
			Sex:         "women",
			MinAge:      0,
			MaxAge:      40,
		},
		{
			Description: "middle-age man",
			Sex:         "man",
			MinAge:      41,
			MaxAge:      60,
		},
		{
			Description: "middle-age women",
			Sex:         "women",
			MinAge:      41,
			MaxAge:      60,
		},
		{
			Description: "old man",
			Sex:         "man",
			MinAge:      61,
			MaxAge:      150,
		},
		{
			Description: "old women",
			Sex:         "women",
			MinAge:      61,
			MaxAge:      150,
		},
		{
			Description: "unknown age-sex group",
			Sex:         "unknown",
			MinAge:      0,
			MaxAge:      0,
		},
	}

	for _, group := range groups {
		if err := r.db.Save(group).Error; err != nil {
			r.logger.Log(context.Background(), errors.Wrapf(err, "can't add group to db"))
		}
	}
	r.logger.Log(context.Background(), "db creation complete")
}

func (r *PGRepo) DeleteDB() {
	r.logger.Log(context.Background(), "db deleting...")
	if err := r.db.DropTableIfExists(&Banner{}, &Slot{}, &Page{}, &Group{}, &BannerEvent{}, &BannerSlot{}).Error; err != nil {
		r.logger.Log(context.Background(), errors.Wrapf(err, "can't delete db"))
	}
	r.logger.Log(context.Background(), "db deleting complete...")
}

func (r *PGRepo) GetPages() (pages []entities.Page, err error) {
	ps, err := r.getRepoPages()
	if err != nil {
		return nil, err
	}

	for _, page := range ps {
		pages = append(pages, page.Page)
	}

	return
}

func (r *PGRepo) GetSlotsByPageURL(pageURL string) (slots []entities.Slot, err error) {
	if err := validateZeroParam(pageURL); err != nil {
		return nil, err
	}
	sls, err := r.getRepoSlots(pageURL)
	if err != nil {
		return nil, err
	}

	for _, slot := range sls {
		slots = append(slots, slot.Slot)
	}

	return
}

func (r *PGRepo) GetBannersBySlotID(pageURL string, slotInnerID uint) (banners []entities.Banner, err error) {
	if err := validateZeroParam(pageURL, slotInnerID); err != nil {
		return nil, err
	}
	repoBanners, err := r.getRepoBanners(pageURL, slotInnerID, 0)
	if err != nil {
		return nil, err
	}

	for _, repoBanner := range repoBanners {
		banners = append(banners, repoBanner.Banner)
	}

	return
}

func (r *PGRepo) GetGroup(userAge uint, userSex string) (group *entities.Group, err error) {
	if err := validateZeroParam(userAge, userSex); err != nil {
		return nil, err
	}
	repoGroup, err := r.getRepoGroup(userAge, userSex)
	if err != nil {
		return nil, err
	}
	return &repoGroup.Group, nil
}

func (r *PGRepo) GetGroups() (groups []entities.Group, defaultGroupDescription string, err error) {
	gs, err := r.getRepoGroups()
	if err != nil {
		return nil, "", err
	}

	for _, g := range gs {
		if g.MinAge == 0 && g.MaxAge == 0 {
			defaultGroupDescription = g.Description
		}
		groups = append(groups, g.Group)
	}

	return
}

func (r *PGRepo) GetActions(pageURL string, slotInnerID, bannerInnerID uint) (actions map[entities.Group]entities.Action, err error) {
	if err := validateZeroParam(pageURL, slotInnerID, bannerInnerID); err != nil {
		return nil, err
	}
	bannerSlot, err := r.getRepoBannerSlot(pageURL, slotInnerID, bannerInnerID)
	if err != nil {
		return nil, err
	}

	actions = make(map[entities.Group]entities.Action)
	var events []*BannerEvent
	// get events.
	r.db.Model(bannerSlot).Related(&events, "Events")
	// get all groups.
	groups, _, err := r.GetGroups()
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		actions[group] = entities.Action{
			Clicks: 0,
			Shows:  0,
		}
	}
	// to fill in groups-actions
	for _, event := range events {
		group := &Group{}
		r.db.First(group, event.GroupID)
		actions[group.Group] = event.Action
	}

	return
}

func (r *PGRepo) AddSlot(pageURL string, slotInnerID uint, slotDescription string) (err error) {
	if err := validateZeroParam(pageURL, slotInnerID, slotDescription); err != nil {
		return err
	}
	// init page.
	page := &Page{
		Page: entities.Page{URL: pageURL},
	}
	// get page or create.
	if err := r.db.Where(page).FirstOrCreate(page).Error; err != nil {
		return err
	}
	//init slot.
	slot := &Slot{
		PageID: page.ID,
		Slot: entities.Slot{
			InnerID:     slotInnerID,
			Description: slotDescription,
		},
	}
	// create slot if not exist.
	if err := r.db.Create(slot).Error; err != nil {
		return err
	}

	return nil
}

func (r *PGRepo) AddBannerToSlot(pageURL string, slotInnerID uint, bannerInnerID uint, bannerDescription string) (err error) {
	if err := validateZeroParam(pageURL, slotInnerID, bannerInnerID, bannerDescription); err != nil {
		return err
	}
	slot, err := r.getRepoSlot(pageURL, slotInnerID)
	if err != nil {
		return err
	}
	// init banner
	banner := &Banner{
		Banner: entities.Banner{
			InnerID:     bannerInnerID,
			Description: bannerDescription,
		},
	}
	// get banner if exist, if not-create
	if err := r.db.Where(banner).FirstOrCreate(banner).Error; err != nil {
		return err
	}
	// update banner with ID.
	banner.BannerSlots = []*BannerSlot{{
		BannerID: banner.ID,
		SlotID:   slot.ID,
	}}
	// update in DB.
	if err := r.db.Save(banner).Error; err != nil {
		return err
	}

	return nil
}

func (r *PGRepo) AddClickAction(pageURL string, slotInnerID, bannerInnerID, userAge uint, userSex string) error {
	if err := validateZeroParam(pageURL, slotInnerID, bannerInnerID, userAge, userSex); err != nil {
		return err
	}
	bannerSlot, err := r.getRepoBannerSlot(pageURL, slotInnerID, bannerInnerID)
	if err != nil {
		return err
	}
	// init group
	group, err := r.getRepoGroup(userAge, userSex)
	if err != nil {
		return err
	}
	if group.ID != 0 {
		var event = &BannerEvent{
			BannerSlotID: bannerSlot.ID,
			GroupID:      group.ID,
		}

		count := 0
		r.db.Model(&bannerSlot).Related("Events").Where(&event).FirstOrCreate(&event).Count(&count)
		if count == 0 {
			bannerSlot.Events = []*BannerEvent{event}
			// update in DB.
			if err := r.db.Save(bannerSlot).Error; err != nil {
				return err
			}
		}
		r.db.Model(event).Where(event).UpdateColumn("clicks", gorm.Expr("clicks + $1", 1))
	}

	return nil
}

func (r *PGRepo) AddShowAction(pageURL string, slotInnerID, bannerInnerID, userAge uint, userSex string) error {
	if err := validateZeroParam(pageURL, slotInnerID, bannerInnerID, userAge, userSex); err != nil {
		return err
	}
	bannerSlot, err := r.getRepoBannerSlot(pageURL, slotInnerID, bannerInnerID)
	if err != nil {
		return err
	}
	// init group
	group, err := r.getRepoGroup(userAge, userSex)
	if err != nil {
		return err
	}
	if group.ID != 0 {
		var event = &BannerEvent{
			BannerSlotID: bannerSlot.ID,
			GroupID:      group.ID,
		}

		count := 0
		r.db.Model(&bannerSlot).Related("Events").Where(&event).FirstOrCreate(&event).Count(&count)
		if count == 0 {
			bannerSlot.Events = []*BannerEvent{event}
			// update in DB.
			if err := r.db.Save(bannerSlot).Error; err != nil {
				return err
			}
		}
		r.db.Model(event).Where(event).UpdateColumn("shows", gorm.Expr("shows + $1", 1))
	}

	return nil
}

func (r *PGRepo) DeleteSlot(pageURL string, slotInnerID uint) error {
	if err := validateZeroParam(pageURL, slotInnerID); err != nil {
		return err
	}
	slot, err := r.getRepoSlot(pageURL, slotInnerID)
	if err != nil {
		return err
	}
	// delete banners from slot.
	if err := r.DeleteAllBannersFormSlot(pageURL, slotInnerID); err != nil {
		return err
	}
	// delete slot.
	r.db.Where(slot).Unscoped().Delete(slot)
	return nil
}

func (r *PGRepo) DeleteBannerFromSlot(pageURL string, slotInnerID, bannerInnerID uint) error {
	if err := validateZeroParam(pageURL, slotInnerID, slotInnerID, bannerInnerID); err != nil {
		return err
	}
	banners, err := r.getRepoBanners(pageURL, slotInnerID, bannerInnerID)
	if err != nil {
		return err
	}

	for _, banner := range banners {
		bannerSlot, err := r.getRepoBannerSlot(pageURL, slotInnerID, bannerInnerID)
		if err != nil {
			return err
		}
		// delete bannerSlot Events
		if err := r.db.Model(&BannerEvent{}).Where("banner_slot_id=?", bannerSlot.ID).Unscoped().Delete(&BannerEvent{}).Error; err != nil {
			return err
		}
		// delete bannerSlot
		if err := r.db.Where(bannerSlot).Unscoped().Delete(bannerSlot).Error; err != nil {
			return err
		}
		//delete banner if it doesn't contain at least one relation to slot
		if cnt := r.db.Model(banner).Association("BannerSlots").Count(); cnt == 0 {
			r.db.Where(banner).Unscoped().Delete(banner)
		}
	}

	return nil
}

func (r *PGRepo) DeleteAllSlots(pageURL string) error {
	if err := validateZeroParam(pageURL); err != nil {
		return err
	}
	slots, err := r.getRepoSlots(pageURL)
	if err != nil {
		return err
	}

	for _, slot := range slots {
		if err := r.DeleteSlot(pageURL, slot.InnerID); err != nil {
			return err
		}
	}

	return nil
}

func (r *PGRepo) DeleteAllBannersFormSlot(pageURL string, slotInnerID uint) error {
	if err := validateZeroParam(pageURL, slotInnerID); err != nil {
		return err
	}
	banners, err := r.GetBannersBySlotID(pageURL, slotInnerID)
	if err != nil {
		return err
	}
	//delete loop
	for _, banner := range banners {
		if err := r.DeleteBannerFromSlot(pageURL, slotInnerID, banner.InnerID); err != nil {
			return err
		}
	}

	return nil
}

func (r *PGRepo) getRepoPage(pageURL string) (*Page, error) {
	// init page.
	page := &Page{
		Page: entities.Page{URL: pageURL},
	}
	// get page.
	if err := r.db.Where(page).First(page).Error; err != nil {
		return nil, err
	}
	return page, nil
}

func (r *PGRepo) getRepoPages() ([]*Page, error) {
	// init page.
	ps := make([]*Page, 0, 1)
	if err := r.db.Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (r *PGRepo) getRepoSlot(pageURL string, slotInnerID uint) (*Slot, error) {
	// get page.
	page, err := r.getRepoPage(pageURL)
	if err != nil {
		return nil, err
	}
	slot := &Slot{
		PageID: page.ID,
		Slot: entities.Slot{
			InnerID: slotInnerID,
		},
	}
	//get slot.
	if err := r.db.Where(slot).First(slot).Error; err != nil {
		return nil, err
	}
	return slot, nil
}

func (r *PGRepo) getRepoSlots(pageURL string) ([]*Slot, error) {
	// init slots.
	var slots []*Slot
	// get page.
	page, err := r.getRepoPage(pageURL)
	if err != nil {
		return nil, err
	}
	// get slots.
	if err := r.db.Model(&page).Related(&slots).Error; err != nil {
		return nil, err
	}
	return slots, nil
}

func (r *PGRepo) getRepoBanners(pageURL string, slotInnerID uint, bannerInnerID uint) (banners []*Banner, err error) {
	slot, err := r.getRepoSlot(pageURL, slotInnerID)
	if err != nil {
		return nil, err
	}

	// init banners.
	bs := []*BannerSlot{}
	// get banners.
	if err := r.db.Model(slot).Related(&bs, "BannerSlots").Error; err != nil {
		return nil, err
	}
	//repo banner
	for _, bannerSlot := range bs {
		banner := &Banner{
			Model: gorm.Model{
				ID: bannerSlot.BannerID,
			},
			Banner: entities.Banner{
				InnerID: bannerInnerID,
			},
		}
		count := 0
		r.db.Where(banner).Find(banner).Count(&count)
		// if banner exist
		if count != 0 {
			banners = append(banners, banner)
		}
	}

	return
}

func (r *PGRepo) getRepoBannerSlot(pageURL string, slotInnerID uint, bannerInnerID uint) (*BannerSlot, error) {
	var bannerSlot = &BannerSlot{}
	if err := r.db.Table("pages").
		Select("banner_slots.*").
		Joins("JOIN slots on pages.id = slots.page_id AND pages.url = ?", pageURL).
		Joins("JOIN banner_slots on banner_slots.slot_id = slots.id AND slots.inner_id = ?", slotInnerID).
		Joins("JOIN banners on banner_slots.banner_id = banners.id AND banners.inner_id = ?", bannerInnerID).
		First(bannerSlot).Error; err != nil {
		return nil, err
	}
	return bannerSlot, nil
}

func (r *PGRepo) getRepoGroup(userAge uint, userSex string) (*Group, error) {
	var group = &Group{}
	if err := r.db.Model(&Group{}).Where("min_age<$1 AND max_age>=$1 AND sex=$2", userAge, userSex).First(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

func (r *PGRepo) getRepoGroups() (groups []Group, err error) {
	if err := r.db.Find(&groups).Error; err != nil {
		return nil, err
	}
	return
}

func validateZeroParam(params ...interface{}) error {
	for _, param := range params {
		switch v := param.(type) {
		case string:
			if v == "" {
				return fmt.Errorf(ErrZeroValue)
			}
		case uint:
			if v == 0 {
				return fmt.Errorf(ErrZeroValue)
			}
		}
	}
	return nil
}
