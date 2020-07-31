package entities

import "fmt"

func ErrBannerExist(slotID, bannerID uint, pageURL, bannerDescription string) error {
	return fmt.Errorf("Banner for slot %v on page %v  with id %v or description %v exist", slotID, pageURL, bannerID, bannerDescription)
}
