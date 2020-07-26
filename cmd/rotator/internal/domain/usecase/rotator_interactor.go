package usecase

import (
	"github.com/pkg/errors"
	"github.com/shipa988/banner_rotator/internal/data/logger"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
	"time"
)

var _ Rotator = (*RotatorInteractor)(nil)

const (
	ErrAddSlot            = "can't add new slot id: %v, description: %v for page: %v"
	ErrDeleteSlot         = "can't delete slot id: %v for page: %v"
	ErrDeleteSlots        = "can't delete slots for page: %v"
	ErrAddBanner          = "can't add new banner id: %v, description: %v for page: %v, slot id: %v"
	ErrDeleteBanner       = "can't delete banner id: %v for page: %v, slot id: %v"
	ErrDeleteBanners      = "can't delete banners for page: %v, slot id: %v"
	ErrClickOnBanner      = "can't register click event for banner id: %v page: %v, slot id: %v"
	ErrGetBanners         = "can't return banners for page: %v, slot id: %v"
	ErrGetSlots           = "can't return slots for page: %v"
	ErrGetNextBanner      = "can't return next banner for page: %v, slot id: %v"
	ErrGetPageStat        = "can't return click stat for page: %v"
	ErrInitNextBannerAlgo = "can't init banner rotate algorithm when extract %v"
)

type RotatorInteractor struct {
	pageRepo       entities.PageRepository
	slotRepo       entities.SlotRepository
	bannerRepo     entities.BannerRepository
	groupRepo      entities.GroupRepository
	actionRepo     entities.ActionRepository
	eventQueue     entities.EventQueue
	nextBannerAlgo NextBannerAlgo
	userGroups     userGroups
	logger         logger.Logger
}

func NewRotatorInteractor(repo interface{}, queueManager entities.EventQueue, alg NextBannerAlgo, logger logger.Logger) (*RotatorInteractor, error) {
	rp, pok := repo.(entities.PageRepository)
	rs, sok := repo.(entities.SlotRepository)
	rb, bok := repo.(entities.BannerRepository)
	re, eok := repo.(entities.ActionRepository)
	rg, gok := repo.(entities.GroupRepository)

	if !gok || !eok || !sok || !bok || !pok {
		return nil, errors.New("scheme repository should implements entities.GroupRepository,entities.GroupRepository,entities.SlotRepository,entities.BannerRepository,entities.PageRepository")
	}

	return &RotatorInteractor{
		pageRepo:       rp,
		slotRepo:       rs,
		bannerRepo:     rb,
		actionRepo:     re,
		groupRepo:      rg,
		nextBannerAlgo: alg,
		eventQueue:     queueManager,
		logger:         logger,
	}, nil
}

func (r *RotatorInteractor) Init() error {
	if err := r.initNextBannerAlgo(); err != nil {
		return err
	}
	if err := r.initUserGroups(); err != nil {
		return err
	}
	return nil
}

func (r *RotatorInteractor) initNextBannerAlgo() error {
	ps := Pages{}

	pages, err := r.pageRepo.GetPages()
	if err != nil {
		return errors.Wrapf(err, ErrInitNextBannerAlgo, "pages")
	}

	for _, page := range pages {
		sl, err := r.GetPageStat(page.URL)
		if err != nil {
			return errors.Wrap(err, ErrInitNextBannerAlgo)
		}
		ps[page] = sl
	}
	err = r.nextBannerAlgo.Init(&ps)
	if err != nil {
		return errors.Wrap(err, ErrInitNextBannerAlgo)
	}
	return nil
}

func (r *RotatorInteractor) GetPageStat(pageUrl string) (Slots, error) {
	sl := Slots{}
	// get slots.
	slots, err := r.slotRepo.GetSlotsByPageURL(pageUrl)
	if err != nil {
		return nil, errors.Wrapf(err, ErrGetPageStat, "slots")
	}
	// scan slots.
	for _, slot := range slots {
		bn := Banners{}
		sl[slot] = bn
		// get banners.
		banners, err := r.bannerRepo.GetBannersBySlotID(pageUrl, slot.InnerID)
		if err != nil {
			return nil, errors.Wrapf(err, ErrGetPageStat, "banners")
		}
		// scan banners.
		for _, banner := range banners {
			// get events for each group.
			events, err := r.actionRepo.GetActions(pageUrl, slot.InnerID, banner.InnerID)
			if err != nil {
				return nil, errors.Wrapf(err, ErrGetPageStat, "events")
			}
			bn[banner] = events
		}
	}
	return sl, nil
}

func (r *RotatorInteractor) GetBannersBySlotID(pageURL string, slotID uint) (banners []entities.Banner, err error) {
	if banners, err = r.bannerRepo.GetBannersBySlotID(pageURL, slotID); err != nil {
		return nil, errors.Wrapf(err, ErrGetBanners, pageURL, slotID)
	}
	return banners, nil
}

func (r *RotatorInteractor) GetSlotsByPageURL(pageURL string) (slots []entities.Slot, err error) {
	if slots, err = r.slotRepo.GetSlotsByPageURL(pageURL); err != nil {
		return nil, errors.Wrapf(err, ErrGetSlots, pageURL)
	}
	return slots, nil
}

func (r *RotatorInteractor) AddSlot(pageURL string, slotID uint, slotDescription string) error {
	defer r.Init() //todo:updates algorithm information about the database schema:existing pages,slots,banners (but if the rotator service is not one - there must be a relay service of broadcast messages about the database schema changing, or using only one rotator-service for crud operations per current page)

	if err := r.slotRepo.AddSlot(pageURL, slotID, slotDescription); err != nil {
		return errors.Wrapf(err, ErrAddSlot, slotID, slotDescription, pageURL)
	}
	return nil
}

func (r *RotatorInteractor) DeleteSlot(pageURL string, slotID uint) error {
	defer r.Init() //todo:updates algorithm information about the database schema:existing pages,slots,banners (but if the rotator service is not one - there must be a relay service of broadcast messages about the database schema changing, or using only one rotator-service for crud operations per current page)

	if err := r.slotRepo.DeleteSlot(pageURL, slotID); err != nil {
		return errors.Wrapf(err, ErrDeleteSlot, slotID, pageURL)
	}
	return nil
}

func (r *RotatorInteractor) DeleteAllSlots(pageURL string) error {
	defer r.Init() //todo:updates algorithm information about the database schema:existing pages,slots,banners (but if the rotator service is not one - there must be a relay service of broadcast messages about the database schema changing, or using only one rotator-service for crud operations per current page)

	if err := r.slotRepo.DeleteAllSlots(pageURL); err != nil {
		return errors.Wrapf(err, ErrDeleteSlots, pageURL)
	}
	return nil
}

func (r *RotatorInteractor) AddBannerToSlot(pageURL string, slotID uint, bannerID uint, bannerDescription string) error {
	defer r.Init() //todo:updates algorithm information about the database schema:existing pages,slots,banners (but if the rotator service is not one - there must be a relay service of broadcast messages about the database schema changing, or using only one rotator-service for crud operations per current page)

	if err := r.bannerRepo.AddBannerToSlot(pageURL, slotID, bannerID, bannerDescription); err != nil {
		return errors.Wrapf(err, ErrAddBanner, bannerID, bannerDescription, pageURL, slotID)
	}
	return nil
}

func (r *RotatorInteractor) DeleteBannerFromSlot(pageURL string, slotID, bannerID uint) error {
	defer r.Init() //todo:updates algorithm information about the database schema:existing pages,slots,banners (but if the rotator service is not one - there must be a relay service of broadcast messages about the database schema changing, or using only one rotator-service for crud operations per current page)

	if err := r.bannerRepo.DeleteBannerFromSlot(pageURL, slotID, bannerID); err != nil {
		return errors.Wrapf(err, ErrDeleteBanner, bannerID, pageURL, slotID)
	}
	return nil
}

func (r *RotatorInteractor) DeleteAllBannersFormSlot(pageURL string, slotID uint) error {
	defer r.Init() //todo:updates algorithm information about the database schema:existing pages,slots,banners (but if the rotator service is not one - there must be a relay service of broadcast messages about the database schema changing, or using only one rotator-service for crud operations per current page)

	if err := r.bannerRepo.DeleteAllBannersFormSlot(pageURL, slotID); err != nil {
		return errors.Wrapf(err, ErrDeleteBanners, pageURL, slotID)
	}
	return nil
}

func (r *RotatorInteractor) ClickByBanner(pageURL string, slotID, bannerID, userAge uint, userSex string) error {
	groupDescription := r.userGroups.findGroup(userAge, userSex)
	err := r.nextBannerAlgo.UpdateReward(pageURL, slotID, bannerID, groupDescription)
	if e, ok := err.(*AlgoError); e != nil && ok && e.Temporary() {
		r.Init() //update schema
		err := r.nextBannerAlgo.UpdateReward(pageURL, slotID, bannerID, groupDescription)
		if err != nil {
			return errors.Wrapf(err, ErrClickOnBanner, bannerID, pageURL, slotID)
		}
	}
	e := entities.Event{
		EventType: "click",
		DT:        time.Now(),
		PageURL:   pageURL,
		SlotID:    slotID,
		BannerID:  bannerID,
		UserAge:   userAge,
		UserSex:   userSex,
	}
	go func() {
		if err := r.eventQueue.Push(e); err != nil {
			r.logger.Log(nil, errors.Wrapf(err, ErrClickOnBanner, bannerID, pageURL, slotID))
		}
	}()
	return nil
}

func (r *RotatorInteractor) GetNextBanner(pageURL string, slotID, userAge uint, userSex string) (bannerID uint, err error) {
	groupDescription := r.userGroups.findGroup(userAge, userSex)
	bannerID, err = r.nextBannerAlgo.GetNext(pageURL, slotID, groupDescription)
	if e, ok := err.(*AlgoError); e != nil && ok && e.Temporary() {
		r.Init() //update schema
		bannerID, err = r.nextBannerAlgo.GetNext(pageURL, slotID, groupDescription)
		if err != nil {
			return 0, errors.Wrapf(err, ErrGetNextBanner, bannerID, pageURL, slotID)
		}
	}
	err = r.nextBannerAlgo.UpdateTry(pageURL, slotID, bannerID, groupDescription)
	if e, ok := err.(*AlgoError); e != nil && ok && e.Temporary() {
		r.Init() //update schema
		err := r.nextBannerAlgo.UpdateTry(pageURL, slotID, bannerID, groupDescription)
		if err != nil {
			return 0, errors.Wrapf(err, ErrGetNextBanner, bannerID, pageURL, slotID)
		}
	}
	e := entities.Event{
		EventType: "show",
		DT:        time.Now(),
		PageURL:   pageURL,
		SlotID:    slotID,
		BannerID:  bannerID,
		UserAge:   userAge,
		UserSex:   userSex,
	}
	go func() {
		if err := r.eventQueue.Push(e); err != nil {
			r.logger.Log(nil, errors.Wrapf(err, ErrGetNextBanner, bannerID, pageURL, slotID))
		}
	}()

	return bannerID, nil
}

func (r *RotatorInteractor) initUserGroups() (err error) {
	groups, defaultGroupDescription, err := r.groupRepo.GetGroups()
	if err != nil {
		return err
	}
	r.userGroups = userGroups{
		groups:             groups,
		defaultDescription: defaultGroupDescription,
	}
	return
}

type userGroups struct {
	groups             []entities.Group
	defaultDescription string
}

func (ug *userGroups) findGroup(userAge uint, userSex string) (groupDescription string) {
	groupDescription = ug.defaultDescription
	for _, group := range ug.groups {
		if group.Sex == userSex && group.MinAge <= userAge && group.MaxAge >= userAge {
			return group.Description
		}
	}
	return
}
