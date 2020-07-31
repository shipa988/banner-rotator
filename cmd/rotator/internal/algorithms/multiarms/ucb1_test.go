//nolint: funlen
package multiarms

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shipa988/banner_rotator/cmd/rotator/internal/domain/usecase"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
)

const (
	pageURL          = "mysite.com"
	slotID           = 1
	groupDescription = "old man"
)

func TestNewUCB1Algo(t *testing.T) {
	pages := initPages()

	t.Run("Init", func(t *testing.T) {
		algo := NewUCB1Algo()
		err := algo.Init(pages)
		require.Nil(t, err)
		expStates := initStates()

		require.Equal(t, expStates, algo.states)
	})

	t.Run("GetNext", func(t *testing.T) {
		algo := NewUCB1Algo()
		err := algo.Init(pages)
		require.Nil(t, err)
		expStates := initStates()

		expNext := expStates[pageURL][slotID][groupDescription].nextarm

		next, err := algo.GetNext(pageURL, slotID, groupDescription)
		require.Nil(t, err)
		require.Equal(t, expNext, next)
	})

	t.Run("UpdateTry", func(t *testing.T) {
		algo := NewUCB1Algo()
		err := algo.Init(pages)
		require.Nil(t, err)
		expStates := initStates()

		expStates[pageURL][slotID][groupDescription].arms[1].try++
		expStates[pageURL][slotID][groupDescription].trys++

		expArm1Try := expStates[pageURL][slotID][groupDescription].arms[1].try
		expAllTryes := expStates[pageURL][slotID][groupDescription].trys
		expNext := getMaxArm(expStates[pageURL][slotID][groupDescription].arms)

		err = algo.UpdateTry(pageURL, slotID, 1, groupDescription)

		require.Nil(t, err)
		require.Equal(t, expNext, algo.states[pageURL][slotID][groupDescription].nextarm)
		require.Equal(t, expArm1Try, algo.states[pageURL][slotID][groupDescription].arms[1].try)
		require.Equal(t, expAllTryes, algo.states[pageURL][slotID][groupDescription].trys)
	})

	t.Run("UpdateReward", func(t *testing.T) {
		algo := NewUCB1Algo()
		err := algo.Init(pages)
		require.Nil(t, err)
		expStates := initStates()

		expStates[pageURL][slotID][groupDescription].arms[1].reward++
		expArm1Reward := expStates[pageURL][slotID][groupDescription].arms[1].reward
		expNext := getMaxArm(expStates[pageURL][slotID][groupDescription].arms)

		err = algo.UpdateReward(pageURL, slotID, 1, groupDescription)

		require.Nil(t, err)
		require.Equal(t, expNext, algo.states[pageURL][slotID][groupDescription].nextarm)
		require.Equal(t, expArm1Reward, algo.states[pageURL][slotID][groupDescription].arms[1].reward)
		require.Equal(t, expStates[pageURL][slotID][groupDescription].trys, algo.states[pageURL][slotID][groupDescription].trys)
	})

	t.Run("When Clicking on Banner often-this banner shows often, but another banners also should be show", func(t *testing.T) {
		algo := NewUCB1Algo()
		err := algo.Init(pages)
		require.Nil(t, err)
		clicks := 60.0
		var expNext uint = 1

		//clicking on banner 1
		for i := 0; i < int(clicks); i++ {
			err := algo.UpdateReward(pageURL, slotID, expNext, groupDescription)
			require.Nil(t, err)
		}

		//show banners 200 iterations
		nexts := make(map[uint]int)
		for i := 0; i < 200; i++ {
			next, err := algo.GetNext(pageURL, slotID, groupDescription)
			require.Nil(t, err)
			err = algo.UpdateTry(pageURL, slotID, next, groupDescription)
			require.Nil(t, err)
			nexts[next]++
		}

		require.Greater(t, nexts[expNext], nexts[2], "clicking banner must be in shows top")
		require.Greater(t, nexts[expNext], nexts[3], "clicking banner must be in shows top")
		require.NotEqual(t, 0, nexts[2], "must be at least one show for banner with no 'click-top' id", uint(2))
		require.NotEqual(t, 0, nexts[3], "must be at least one show for banner with no 'click-top' id", uint(3))
	})
}

func getMaxArm(arms map[uint]*arm) uint {
	max := 0.0
	maxID := uint(0)
	tryes := func() (tryes float64) {
		for _, a2 := range arms {
			tryes += a2.try
		}
		return
	}()
	for id, a := range arms {
		x := a.reward / a.try
		val := x + math.Sqrt(2*math.Log(tryes)/a.try)
		if val > max {
			max = val
			maxID = id
		}
	}
	return maxID
}

func initStates() map[string]map[uint]map[groupName]*state {
	a1 := &arm{try: 100, reward: 10}
	a2 := &arm{try: 10, reward: 1}
	a3 := &arm{try: 5, reward: 0}
	arms := make(map[uint]*arm)

	arms[1] = a1
	arms[2] = a2
	arms[3] = a3
	g := make(map[groupName]*state)
	s := make(map[uint]map[groupName]*state)
	expStates := make(map[string]map[uint]map[groupName]*state)

	id := getMaxArm(arms)
	g[groupDescription] = &state{arms: arms, trys: a1.try + a2.try + a3.try, nextarm: id}
	s[slotID] = g
	expStates[pageURL] = s
	return expStates
}

func initPages() *usecase.Pages {
	a1 := entities.Action{
		Clicks: 10,
		Shows:  100,
	}
	a2 := entities.Action{
		Clicks: 1,
		Shows:  10,
	}
	a3 := entities.Action{
		Clicks: 0,
		Shows:  5,
	}
	g := entities.Group{
		Description: groupDescription,
		Sex:         "man",
		MinAge:      60,
		MaxAge:      150,
	}
	b1 := entities.Banner{
		InnerID:     1,
		Description: "1_banner",
	}
	b2 := entities.Banner{
		InnerID:     2,
		Description: "2_banner",
	}
	b3 := entities.Banner{
		InnerID:     3,
		Description: "3_banner",
	}
	s := entities.Slot{
		InnerID:     slotID,
		Description: "1_slot",
	}
	p := entities.Page{URL: pageURL}
	pages := usecase.Pages{}
	pages[p] = map[entities.Slot]usecase.Banners{}
	pages[p][s] = map[entities.Banner]usecase.GroupStats{}
	pages[p][s][b1] = map[entities.Group]entities.Action{}
	pages[p][s][b2] = map[entities.Group]entities.Action{}
	pages[p][s][b3] = map[entities.Group]entities.Action{}
	pages[p][s][b1][g] = a1
	pages[p][s][b2][g] = a2
	pages[p][s][b3][g] = a3
	return &pages
}
