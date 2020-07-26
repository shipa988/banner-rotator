package usecase

import "github.com/shipa988/banner_rotator/internal/domain/entities"

type AlgoError struct {
	Mess        string
	IsOldSchema bool
}

func (a AlgoError) Temporary() bool {
	return a.IsOldSchema == true
}
func (a AlgoError) Error() string {
	return a.Mess
}

type GroupStats map[entities.Group]entities.Action
type Banners map[entities.Banner]GroupStats
type Slots map[entities.Slot]Banners
type Pages map[entities.Page]Slots

type NextBannerAlgo interface {
	GetNext(pageURL string, slotID uint, groupDescription string) (id uint, err error)
	UpdateTry(pageURL string, slotID, bannerID uint, groupDescription string) error
	UpdateReward(pageURL string, slotID, bannerID uint, groupDescription string) error
	Init(pages *Pages) error
}
