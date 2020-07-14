package entities

type Group struct {
	Description string `gorm:"UNIQUE_INDEX:des_sex; NOT NULL"`
	Sex         string `gorm:"UNIQUE_INDEX:des_sex; NOT NULL"`
	MinAge      int    `gorm:"UNIQUE_INDEX:des_sex; NOT NULL"`
	MaxAge      int    `gorm:"UNIQUE_INDEX:des_sex; NOT NULL"`
}
