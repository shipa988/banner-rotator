package usecase

import (
	"errors"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
)

var _ Rotator= (*RotatorInteractor)(nil)

type RotatorInteractor struct {
	EventRepo entities.EventRepository
	SlotRepo entities.SlotRepository
	BannerRepo entities.BannerRepository
}

func NewRotatorInteractor(repo interface{}) (*RotatorInteractor,error)  {
	re,eok:= repo.(entities.EventRepository)
	rs,sok:= repo.(entities.SlotRepository)
	rb,bok:= repo.(entities.BannerRepository)
	if !eok || !sok || !bok{
		return nil,errors.New("repository must be implement entities.EventRepository,entities.SlotRepository,entities.BannerRepository")
	}
	return &RotatorInteractor{
		EventRepo:  re,
		SlotRepo:   rs,
		BannerRepo: rb,
	},nil

}

func (r *RotatorInteractor) AddSlot(pageURL string, slotID uint, slotDescription string) error {
	return r.SlotRepo.AddSlot(pageURL,slotID,slotDescription)
}

func (r *RotatorInteractor) DeleteSlot(pageURL string, slotID int) error {
	panic("implement me")
}

func (r *RotatorInteractor) DeleteAllSlots(pageURL string) error {
	panic("implement me")
}

func (r *RotatorInteractor) AddBanner(pageURL string, slotID int, bannerID int, bannerDescription string) error {
	panic("implement me")
}

func (r *RotatorInteractor) DeleteBanner(pageURL string, slotID, bannerID int) error {
	panic("implement me")
}

func (r *RotatorInteractor) DeleteAllBanners(pageURL string, slotID int) error {
	panic("implement me")
}

func (r *RotatorInteractor) ClickByBanner(pageURL string, slotID, bannerID, userAge int, userSex string) error {
	panic("implement me")
}

func (r *RotatorInteractor) GetNextBanner(pageURL string, slotID, userAge int, userSex string) (bannerID string, err error) {
	panic("implement me")
}

func (r *RotatorInteractor) GetBannersBySlotID(pageURL string, slotID int) (banners []entities.Banner, err error) {
	panic("implement me")
}

func (r *RotatorInteractor) GetSlotsByPageURL(pageURL string) (slots []entities.Slot, err error) {
	panic("implement me")
}





