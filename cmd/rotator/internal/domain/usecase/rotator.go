package usecase

import "github.com/shipa988/banner_rotator/internal/domain/entities"

type Rotator interface {
	AddSlot(pageURL string, slotID uint, slotDescription string) error
	DeleteSlot(pageURL string, slotID uint) error
	DeleteAllSlots(pageURL string) error
	GetSlotsByPageURL(pageURL string) (slots []entities.Slot, err error)

	AddBannerToSlot(pageURL string, slotID uint, bannerID uint, bannerDescription string) error
	DeleteBannerFromSlot(pageURL string, slotID, bannerID uint) error
	DeleteAllBannersFormSlot(pageURL string, slotID uint) error
	GetBannersBySlotID(pageURL string, slotID uint) (banners []entities.Banner, err error)

	ClickByBanner(pageURL string, slotID, bannerID, userAge uint, userSex string) error
	GetNextBanner(pageURL string, slotID, userAge uint, userSex string) (bannerID uint, err error)
	Init() error

	GetPageStat(pageUrl string) (Slots, error)
}
