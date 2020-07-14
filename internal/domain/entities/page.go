package entities

type Page struct {
	URL string `gorm:"UNIQUE; NOT NULL"`
}