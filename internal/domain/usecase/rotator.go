package usecase

import "github.com/shipa988/banner-rotator/internal/domain/entities"

type Rotator interface {
	AddSlot(pageURL string, slotID uint, slotDescription string) error
	DeleteSlot(pageURL string, slotID int) error
	DeleteAllSlots(pageURL string) error
	AddBanner(pageURL string, slotID int, bannerID int, bannerDescription string) error
	DeleteBanner(pageURL string, slotID, bannerID int) error
	DeleteAllBanners(pageURL string, slotID int) error

	ClickByBanner(pageURL string, slotID, bannerID, userAge int, userSex string) error
	GetNextBanner(pageURL string, slotID, userAge int, userSex string) (bannerID string, err error)

	GetBannersBySlotID(pageURL string, slotID int) (banners []entities.Banner, err error)
	GetSlotsByPageURL(pageURL string) (slots []entities.Slot, err error)
}
