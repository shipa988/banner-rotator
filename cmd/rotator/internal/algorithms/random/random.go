package random

import (
	"github.com/shipa988/banner_rotator/cmd/rotator/internal/domain/usecase"
)

var _ usecase.NextBannerAlgo = (*Randomizer)(nil)

type Randomizer struct {
	pages usecase.Pages
}

func (r Randomizer) GetNext(pageURL string, slotID uint, groupDescription string) (id uint, err error) {
	panic("implement me")
}

func (r Randomizer) UpdateTry(pageURL string, slotID, bannerID uint, groupDescription string) error {
	panic("implement me")
}

func (r Randomizer) UpdateReward(pageURL string, slotID, bannerID uint, groupDescription string) error {
	panic("implement me")
}

func (r Randomizer) Init(pages *usecase.Pages) error {
	panic("implement me")
}

func NewRandomizer() *Randomizer {
	return &Randomizer{}
}
