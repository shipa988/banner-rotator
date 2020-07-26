package entities

import "time"

type Event struct {
	EventType                 string
	DT                        time.Time
	PageURL                   string
	SlotID, BannerID, UserAge uint
	UserSex                   string
}

type EventQueue interface {
	Pull(events chan<- Event) error
	Push(Event) error
}
