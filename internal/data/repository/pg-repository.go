package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
)

var _ entities.BannerRepository = (*PGRepo)(nil)
var _ entities.SlotRepository = (*PGRepo)(nil)
var _ entities.PageRepository = (*PGRepo)(nil)
var _ entities.ActionRepository = (*PGRepo)(nil)
var _ entities.GroupRepository = (*PGRepo)(nil)

type PGRepo struct {
	conn *gorm.DB
}

func NewPGRepo(db *gorm.DB, isDebug bool) *PGRepo {
	if isDebug {
		db = db.Debug()
		db.LogMode(true)
	}
	return &PGRepo{conn: db}
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
	repoBanners, err := r.getRepoBanners(pageURL, slotInnerID, 0)
	if err != nil {
		return nil, err
	}

	for _, repoBanner := range repoBanners {
		banners = append(banners, repoBanner.Banner)
	}

	return
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
	bannerSlot, err := r.getRepoBannerSlot(pageURL, slotInnerID, bannerInnerID)
	if err != nil {
		return nil, err
	}

	actions = make(map[entities.Group]entities.Action)
	var events []*BannerEvent
	// get events.
	r.conn.Model(bannerSlot).Related(&events, "Events")
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
		r.conn.First(group, event.GroupID)
		actions[group.Group] = event.Action
	}

	return
}

func (r *PGRepo) AddSlot(pageURL string, slotInnerID uint, slotDescription string) (err error) {
	// init page.
	page := &Page{
		Page: entities.Page{URL: pageURL},
	}
	// get page or create.
	if err := r.conn.Debug().Where(page).FirstOrCreate(page).Error; err != nil {
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
	if err := r.conn.Debug().Create(slot).Error; err != nil {
		return err
	}

	return nil
}

func (r *PGRepo) AddBannerToSlot(pageURL string, slotInnerID uint, bannerInnerID uint, bannerDescription string) (err error) {
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
	if err := r.conn.Where(banner).FirstOrCreate(banner).Error; err != nil {
		return err
	}
	// update banner with ID.
	banner.BannerSlots = []*BannerSlot{{
		BannerID: banner.ID,
		SlotID:   slot.ID,
	}}
	// update in DB.
	if err := r.conn.Save(banner).Error; err != nil {
		return err
	}

	return nil
}

func (r *PGRepo) AddAction(actionName, pageURL string, slotInnerID, bannerInnerID, userAge uint, userSex string) error {
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
		r.conn.Model(&bannerSlot).Related("Events").Where(&event).FirstOrCreate(&event).Count(&count)
		if count == 0 {
			bannerSlot.Events = []*BannerEvent{event}
			// update in DB.
			if err := r.conn.Save(bannerSlot).Error; err != nil {
				return err
			}

		}
		r.conn.Model(event).Where(event).UpdateColumn(actionName, gorm.Expr(actionName+" + $1", 1))
	}

	return nil
}

/*func (r *PGRepo) AddShowAction(pageURL string, slotInnerID, bannerInnerID, userAge uint, userSex string) error {
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
		r.conn.Model(&bannerSlot).Related("Events").Where(&event).FirstOrCreate(&event).Count(&count)
		if count == 0 {
			bannerSlot.Events = []*BannerEvent{event}
			// update in DB.
			if err := r.conn.Save(bannerSlot).Error; err != nil {
				return err
			}

		}
		r.conn.Model(event).Where(event).UpdateColumn("shows", gorm.Expr("shows + $1", 1))
	}

	return nil
}
*/
func (r *PGRepo) DeleteSlot(pageURL string, slotInnerID uint) error {
	slot, err := r.getRepoSlot(pageURL, slotInnerID)
	if err != nil {
		return err
	}
	// delete banners from slot.
	if err := r.DeleteAllBannersFormSlot(pageURL, slotInnerID); err != nil {
		return err
	}
	// delete slot.
	r.conn.Where(slot).Unscoped().Delete(slot)
	return nil
}

func (r *PGRepo) DeleteBannerFromSlot(pageURL string, slotInnerID, bannerInnerID uint) error {
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
		if err := r.conn.Model(&BannerEvent{}).Where("banner_slot_id=?", bannerSlot.ID).Unscoped().Delete(&BannerEvent{}).Error; err != nil {
			return err
		}
		// delete bannerSlot
		if err := r.conn.Where(bannerSlot).Unscoped().Delete(bannerSlot).Error; err != nil {
			return err
		}
		//delete banner if it doesn't contain at least one relation to slot
		if cnt := r.conn.Model(banner).Association("BannerSlots").Count(); cnt == 0 {
			r.conn.Where(banner).Unscoped().Delete(banner)
		}
	}

	return nil
}

func (r *PGRepo) DeleteAllSlots(pageURL string) error {
	slots, err := r.getRepoSlots(pageURL)
	if err != nil {
		return err
	}

	for _, slot := range slots {
		r.DeleteSlot(pageURL, slot.InnerID)
	}

	return nil
}

func (r *PGRepo) DeleteAllBannersFormSlot(pageURL string, slotInnerID uint) error {
	banners, err := r.GetBannersBySlotID(pageURL, slotInnerID)
	if err != nil {
		return err
	}
	//delete loop
	for _, banner := range banners {
		r.DeleteBannerFromSlot(pageURL, slotInnerID, banner.InnerID)
	}

	return nil
}

func (r *PGRepo) getRepoPage(pageURL string) (*Page, error) {
	// init page.
	page := &Page{
		Page: entities.Page{URL: pageURL},
	}
	// get page.
	if err := r.conn.Where(page).First(page).Error; err != nil {
		return nil, err
	}
	return page, nil
}

func (r *PGRepo) getRepoPages() ([]*Page, error) {
	// init page.
	ps := make([]*Page, 0, 1)
	if err := r.conn.Find(&ps).Error; err != nil {
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
	if err := r.conn.Where(slot).First(slot).Error; err != nil {
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
	if err := r.conn.Model(&page).Related(&slots).Error; err != nil {
		return nil, err
	}
	return slots, nil
}

func (r *PGRepo) getRepoBanner(pageURL string, slotInnerID uint, bannerInnerID uint) (*Banner, error) {
	var banner = &Banner{}
	if err := r.conn.Table("pages").
		Select("banners.*").
		Joins("JOIN slots on pages.id = slots.page_id AND pages.url = ?", pageURL).
		Joins("JOIN banner_slots on banner_slots.slot_id = slots.id AND slots.inner_id = ?", slotInnerID).
		Joins("JOIN banners on banner_slots.banner_id = banners.id AND banners.inner_id = ?", bannerInnerID).
		First(banner).Error; err != nil {
		return nil, err
	}
	return banner, nil
}

func (r *PGRepo) getRepoBanners(pageURL string, slotInnerID uint, bannerInnerID uint) (banners []*Banner, err error) {
	slot, err := r.getRepoSlot(pageURL, slotInnerID)
	if err != nil {
		return nil, err
	}

	// init banners.
	bs := []*BannerSlot{}
	// get banners.
	if err := r.conn.Model(slot).Related(&bs, "BannerSlots").Error; err != nil {
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
		r.conn.Where(banner).Find(banner).Count(&count)
		// if banner exist
		if count != 0 {
			banners = append(banners, banner)
		}
	}

	return
}

func (r *PGRepo) getRepoBannerSlot(pageURL string, slotInnerID uint, bannerInnerID uint) (*BannerSlot, error) {
	var bannerSlot = &BannerSlot{}
	if err := r.conn.Table("pages").
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
	if err := r.conn.Model(&Group{}).Where("min_age<$1 AND max_age>=$1 AND sex=$2", userAge, userSex).First(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

func (r *PGRepo) getRepoGroups() (groups []Group, err error) {
	if err := r.conn.Find(&groups).Error; err != nil {
		return nil, err
	}
	return
}
