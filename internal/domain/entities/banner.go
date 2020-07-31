package entities

type Banner struct {
	InnerID     uint   `gorm:"UNIQUE_INDEX:innerid_description; NOT NULL"`
	Description string `gorm:"UNIQUE_INDEX:innerid_description; NOT NULL"`
}

type BannerRepository interface {
	AddBannerToSlot(pageURL string, slotInnerID uint, bannerInnerID uint, bannerDescription string) error
	DeleteBannerFromSlot(pageURL string, slotInnerID, bannerInnerID uint) error
	DeleteAllBannersFormSlot(pageURL string, slotInnerID uint) error
	GetBannersBySlotID(pageURL string, slotInnerID uint) (banners []Banner, err error)
}
