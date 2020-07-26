package multiarms

import (
	"fmt"
	"github.com/shipa988/banner_rotator/cmd/rotator/internal/domain/usecase"
	"math"
	"sync"
)

var _ usecase.NextBannerAlgo = (*UCB1Algo)(nil)
var algoErr = func(page string, slotId uint, groupDescription string) *usecase.AlgoError {
	return &usecase.AlgoError{
		Mess:        fmt.Sprintf("banners for page: %v, slotId: %v, groupDescription: %v not found", page, slotId, groupDescription),
		IsOldSchema: true,
	}
}

type state struct {
	arms    map[uint]*arm
	trys    float64
	nextarm uint
}

type arm struct {
	try    float64
	reward float64
}

type groupName string

type UCB1Algo struct {
	sync.RWMutex
	states map[string]map[uint]map[groupName]*state
}

func (a *UCB1Algo) Init(pages *usecase.Pages) error {
	a.Lock()
	a.states = make(map[string]map[uint]map[groupName]*state)
	pgs := a.states
	for page, slots := range *pages {
		pgs[page.URL] = make(map[uint]map[groupName]*state)
		sls := pgs[page.URL]
		for slot, banners := range slots {
			sls[slot.InnerID] = make(map[groupName]*state)
			grps := sls[slot.InnerID]
			for banner, stats := range banners {
				for group, action := range stats {

					st, ok := grps[groupName(group.Description)]
					if !ok {
						grps[groupName(group.Description)] = &state{
							arms:    make(map[uint]*arm),
							trys:    0,
							nextarm: 0,
						}
						st = grps[groupName(group.Description)]
					}
					st.arms[banner.InnerID] = &arm{
						try:    float64(action.Shows),
						reward: float64(action.Clicks),
					}
					st.trys += float64(action.Shows)
					/*
						a.states[page.URL][slot.InnerID][groupName(group.Description)] = &state*/
				}
			}
			// update nextArm value in slot.
			for _, s := range grps {
				a.setNext(s)
			}
		}
	}
	a.Unlock()
	return nil
}

func (a *UCB1Algo) UpdateTry(pageURL string, slotID, bannerID uint, groupDescription string) (err error) {
	a.Lock()
	defer a.Unlock()
	s, ok := a.states[pageURL][slotID][groupName(groupDescription)]
	if !ok {
		err = algoErr(pageURL, slotID, groupDescription)
		return
	}
	b, ok := s.arms[bannerID]
	if !ok {
		err = algoErr(pageURL, slotID, groupDescription)
		return
	}
	b.try++
	s.trys++
	a.setNext(s)
	return nil
}

func (a *UCB1Algo) UpdateReward(pageURL string, slotID, bannerID uint, groupDescription string) (err error) {
	a.Lock()
	defer a.Unlock()
	s, ok := a.states[pageURL][slotID][groupName(groupDescription)]
	if !ok {
		err = algoErr(pageURL, slotID, groupDescription)
		return
	}
	/*if s==nil{

		s=&state{
			arms:     make(map[uint]*arm),
			trys:    0,
			nextarm: 0,
		}
	}*/
	b, ok := s.arms[bannerID]
	if !ok {
		err = algoErr(pageURL, slotID, groupDescription)
		return
	}
	b.reward++
	s.trys++
	a.setNext(s)
	return nil
}

func NewUCB1Algo() *UCB1Algo {
	return &UCB1Algo{}
}

func (a *UCB1Algo) GetNext(pageURL string, slotID uint, groupDescription string) (id uint, err error) {
	a.RLock()
	defer a.RUnlock()
	if state := a.states[pageURL][slotID][groupName(groupDescription)]; state != nil {
		return state.nextarm, nil
	}
	err = algoErr(pageURL, slotID, groupDescription)

	return 0, err
}

func (a *UCB1Algo) setNext(s *state) {
	max := 0.0
	for id, armState := range s.arms {
		if armState.try == 0 {
			s.nextarm = id
		}
		x := armState.reward / armState.try
		val := x + math.Sqrt(2*math.Log(s.trys)/armState.try)
		if val > max {
			max = val
			s.nextarm = id
		}
	}
	return
}
