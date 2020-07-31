package entities

type Group struct {
	Description string `gorm:"UNIQUE_INDEX:des_sex; NOT NULL"`
	Sex         string `gorm:"UNIQUE_INDEX:des_sex; NOT NULL"`
	MinAge      uint   `gorm:"UNIQUE_INDEX:des_sex; NOT NULL"`
	MaxAge      uint   `gorm:"UNIQUE_INDEX:des_sex; NOT NULL"`
}
type GroupRepository interface {
	GetGroups() (groups []Group, defaultGroupDescription string, err error)
	GetGroup(userAge uint, userSex string) (group *Group, err error)
}
type Action struct {
	Clicks uint
	Shows  uint
}

type ActionRepository interface {
	AddClickAction(pageURL string, slotInnerID, bannerInnerID, userAge uint, userSex string) error
	AddShowAction(pageURL string, slotInnerID, bannerInnerID, userAge uint, userSex string) error
	GetActions(pageURL string, slotInnerID, bannerInnerID uint) (clicks map[Group]Action, err error)
}
