package entities

type BannerRepository interface {
	AddBanner(pageURL string, slotInnerID uint, bannerInnerID uint,bannerDescription string)  error
	DeleteBanner(pageURL string, slotInnerID, bannerInnerID uint) error
	DeleteAllBanners(pageURL string, slotID int) error
	GetBannersBySlotID(pageURL string, slotInnerID int) (banners []Banner, err error)

}
type EventRepository interface {
	AddEvent(pageURL string, bannerInnerID,bannerID,userAge uint,userSex string) error
}

type SlotRepository interface {
	AddSlot(pageURL string, slotInnerID uint,slotDescription string)  error
	DeleteSlot(pageURL string, slotInnerID uint) error
	DeleteAllSlots(pageURL string) error
	GetSlotsByPageURL(pageURL string) (slots []Slot, err error)
}
