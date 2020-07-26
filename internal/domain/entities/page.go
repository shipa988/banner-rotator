package entities

type Page struct {
	URL string `gorm:"UNIQUE; NOT NULL"`
}

type PageRepository interface {
	GetPages() (pages []Page, err error)
}
