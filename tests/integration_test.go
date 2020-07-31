// +build integration

package tests

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/shipa988/banner_rotator/cmd/rotator/api"
	"github.com/shipa988/banner_rotator/internal/data/logger/zaplogger"
	"github.com/shipa988/banner_rotator/internal/data/repository"
)

const (
	pageURL           = "site.com"
	slotID            = 1
	slotDescription   = "slot_description"
	bannerID          = 1
	bannerDescription = "banner_description"
	userAge           = 31
	userSex           = "man"
)

var (
	validMetadata   = metadata.New(map[string]string{"Cookie": fmt.Sprintf("page_url=%v;otherparam=othervalue", pageURL)})
	invalidMetadata = metadata.New(map[string]string{"Cookie": "otherparam=othervalue"})
)

type Suite struct {
	suite.Suite
	client grpcservice.BannerRotatorServiceClient
	conn   *grpc.ClientConn
	repo   *repository.PGRepo
}
type condition func()

func TestIntegration(t *testing.T) {
	s := new(Suite)
	suite.Run(t, s)
}

func (s *Suite) SetupSuite() {
	//port := os.Getenv("GRPC_PORT")
	//dsn := os.Getenv("DSN")
	port := "4445"
	dsn := "host=localhost port=5432 user=igor password=igor dbname=rotator sslmode=disable"
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	serverAddr := net.JoinHostPort("localhost", port)
	s.conn, _ = grpc.Dial(serverAddr, opts...)
	s.client = grpcservice.NewBannerRotatorServiceClient(s.conn)

	db, err := gorm.Open("postgres", dsn)
	require.Nil(s.T(), err)
	wr := os.Stdout
	logger := zaplogger.NewLogger(wr, false)
	s.repo = repository.NewPGRepo(db, logger, false)
	s.AfterTest("", "")
}

func (s *Suite) AfterTest(_, _ string) {
	headers := validMetadata
	ctx := metadata.NewOutgoingContext(context.Background(), headers)
	s.client.DeleteAllSlots(ctx, &grpcservice.DeleteAllSlotsRequest{})
}

func (s *Suite) TestIntegration_AddSlot() {
	tcases := []struct {
		name          string
		headers       metadata.MD
		preCondition  condition
		request       *grpcservice.RegisterSlotRequest
		postCondition condition
		err           bool
	}{
		{
			name:         "good",
			headers:      validMetadata,
			preCondition: nil,
			request: &grpcservice.RegisterSlotRequest{
				SlotId:          slotID,
				SlotDescription: slotDescription,
			},
			postCondition: func() {
				slots, err := s.repo.GetSlotsByPageURL(pageURL)
				require.Nil(s.T(), err)
				contains := false
				for _, slot := range slots {
					if slot.InnerID == slotID && slot.Description == slotDescription {
						contains = true
					}
				}
				require.True(s.T(), contains)
				s.AfterTest("", "")
			},
			err: false,
		},
		{
			name:    "duplicate",
			headers: validMetadata,
			preCondition: func() {
				err := s.repo.AddSlot(pageURL, slotID, slotDescription)
				require.Nil(s.T(), err)
			},
			request: &grpcservice.RegisterSlotRequest{
				SlotId:          slotID,
				SlotDescription: slotDescription,
			},
			postCondition: func() {
				slots, _ := s.repo.GetSlotsByPageURL(pageURL)
				require.Equal(s.T(), 1, len(slots))
				s.AfterTest("", "")
			},
			err: true,
		},
		{
			name:         "unauth",
			headers:      invalidMetadata,
			preCondition: nil,
			request: &grpcservice.RegisterSlotRequest{
				SlotId:          slotID,
				SlotDescription: slotDescription,
			},
			postCondition: nil,
			err:           true,
		},
		{
			name:         "bad zero value",
			headers:      validMetadata,
			preCondition: nil,
			request: &grpcservice.RegisterSlotRequest{
				SlotId:          0,
				SlotDescription: slotDescription,
			},
			postCondition: func() {
				slots, _ := s.repo.GetSlotsByPageURL(pageURL)
				require.Equal(s.T(), 0, len(slots))
				s.AfterTest("", "")
			},
			err: true,
		},
	}

	for id, tcase := range tcases {
		s.Run(fmt.Sprintf("%d: %v", id, tcase.name), func() {
			if tcase.preCondition != nil {
				tcase.preCondition()
			}
			ctx := metadata.NewOutgoingContext(context.Background(), tcase.headers)
			response, err := s.client.RegisterSlot(ctx, tcase.request)
			if tcase.err {
				require.NotNil(s.T(), err)
				require.Nil(s.T(), response)
			} else {
				require.Nil(s.T(), err)
				require.NotNil(s.T(), response)
			}
			if tcase.postCondition != nil {
				tcase.postCondition()
			}
		})
	}
}

func (s *Suite) TestIntegration_DeleteSlot() {
	tcases := []struct {
		name          string
		headers       metadata.MD
		preCondition  condition
		request       *grpcservice.DeleteSlotRequest
		postCondition condition
		err           bool
	}{
		{
			name:    "good",
			headers: validMetadata,
			preCondition: func() {
				err := s.repo.AddSlot(pageURL, slotID, slotDescription)
				require.Nil(s.T(), err)
			},
			request: &grpcservice.DeleteSlotRequest{
				SlotId: slotID,
			},
			postCondition: func() {
				slots, _ := s.repo.GetSlotsByPageURL(pageURL)
				require.Equal(s.T(), 0, len(slots))
				s.AfterTest("", "")
			},
			err: false,
		},
		{
			name:         "unauth",
			headers:      invalidMetadata,
			preCondition: nil,
			request: &grpcservice.DeleteSlotRequest{
				SlotId: slotID,
			},
			postCondition: nil,
			err:           true,
		},
		{
			name:         "bad zero value",
			headers:      validMetadata,
			preCondition: nil,
			request: &grpcservice.DeleteSlotRequest{
				SlotId: slotID,
			},
			postCondition: func() {
				s.AfterTest("", "")
			},
			err: true,
		},
		{
			name:    "not found",
			headers: validMetadata,
			preCondition: func() {
				err := s.repo.AddSlot(pageURL, slotID, slotDescription)
				require.Nil(s.T(), err)
			},
			request: &grpcservice.DeleteSlotRequest{
				SlotId: 2,
			},
			postCondition: func() {
				slots, _ := s.repo.GetSlotsByPageURL(pageURL)
				require.Equal(s.T(), 1, len(slots))
				s.AfterTest("", "")
			},
			err: true,
		},
	}

	for id, tcase := range tcases {
		s.Run(fmt.Sprintf("%d: %v", id, tcase.name), func() {
			if tcase.preCondition != nil {
				tcase.preCondition()
			}
			ctx := metadata.NewOutgoingContext(context.Background(), tcase.headers)
			response, err := s.client.DeleteSlot(ctx, tcase.request)
			if tcase.err {
				require.NotNil(s.T(), err)
				require.Nil(s.T(), response)
			} else {
				require.Nil(s.T(), err)
				require.NotNil(s.T(), response)
			}
			if tcase.postCondition != nil {
				tcase.postCondition()
			}
		})
	}

}

func (s *Suite) TestIntegration_AddBanner() {
	tcases := []struct {
		name          string
		headers       metadata.MD
		preCondition  condition
		request       *grpcservice.RegisterBannerRequest
		postCondition condition
		err           bool
	}{
		{
			name:         "bad: slot not found",
			headers:      validMetadata,
			preCondition: nil,
			request: &grpcservice.RegisterBannerRequest{
				SlotId:            slotID,
				BannerId:          bannerID,
				BannerDescription: bannerDescription,
			},
			postCondition: func() {
				banners, err := s.repo.GetBannersBySlotID(pageURL, slotID)
				require.NotNil(s.T(), err)
				require.Equal(s.T(), 0, len(banners))
				s.AfterTest("", "")
			},
			err: true,
		},
		{
			name:    "good",
			headers: validMetadata,
			preCondition: func() {
				err := s.repo.AddSlot(pageURL, slotID, slotDescription)
				require.Nil(s.T(), err)
			},
			request: &grpcservice.RegisterBannerRequest{
				SlotId:            slotID,
				BannerId:          bannerID,
				BannerDescription: bannerDescription,
			},
			postCondition: func() {
				banners, err := s.repo.GetBannersBySlotID(pageURL, slotID)
				require.Nil(s.T(), err)
				contains := false
				for _, banner := range banners {
					if banner.InnerID == bannerID && banner.Description == bannerDescription {
						contains = true
					}
				}
				require.True(s.T(), contains)
				s.AfterTest("", "")
			},
			err: false,
		},
		{
			name:    "unauth",
			headers: invalidMetadata,
			request: &grpcservice.RegisterBannerRequest{
				SlotId:            slotID,
				BannerId:          bannerID,
				BannerDescription: bannerDescription,
			},
			err: true,
		},
	}

	for id, tcase := range tcases {
		s.Run(fmt.Sprintf("%d: %v", id, tcase.name), func() {
			if tcase.preCondition != nil {
				tcase.preCondition()
			}
			ctx := metadata.NewOutgoingContext(context.Background(), tcase.headers)
			response, err := s.client.RegisterBanner(ctx, tcase.request)
			if tcase.err {
				require.NotNil(s.T(), err)
				require.Nil(s.T(), response)
			} else {
				require.Nil(s.T(), err)
				require.NotNil(s.T(), response)
			}
			if tcase.postCondition != nil {
				tcase.postCondition()
			}
		})
	}
}

func (s *Suite) TestIntegration_DeleteBanner() {
	tcases := []struct {
		name          string
		headers       metadata.MD
		preCondition  condition
		request       *grpcservice.DeleteBannerRequest
		postCondition condition
		err           bool
	}{
		{
			name:    "good",
			headers: validMetadata,
			preCondition: func() {
				err := s.repo.AddSlot(pageURL, slotID, slotDescription)
				require.Nil(s.T(), err)
				err = s.repo.AddBannerToSlot(pageURL, slotID, bannerID, bannerDescription)
				require.Nil(s.T(), err)
			},
			request: &grpcservice.DeleteBannerRequest{
				SlotId:   slotID,
				BannerId: bannerID,
			},
			postCondition: func() {
				banners, _ := s.repo.GetBannersBySlotID(pageURL, slotID)
				require.Equal(s.T(), 0, len(banners))
				s.AfterTest("", "")
			},
			err: false,
		},
		{
			name:         "unauth",
			headers:      invalidMetadata,
			preCondition: nil,
			request: &grpcservice.DeleteBannerRequest{
				SlotId:   slotID,
				BannerId: bannerID,
			},
			postCondition: nil,
			err:           true,
		},
		{
			name:    "bad zero value",
			headers: validMetadata,
			preCondition: func() {
				err := s.repo.AddSlot(pageURL, slotID, slotDescription)
				require.Nil(s.T(), err)
				err = s.repo.AddBannerToSlot(pageURL, slotID, bannerID, bannerDescription)
				require.Nil(s.T(), err)
			},
			request: &grpcservice.DeleteBannerRequest{
				SlotId:   slotID,
				BannerId: 0,
			},
			postCondition: func() {
				banners, _ := s.repo.GetBannersBySlotID(pageURL, slotID)
				require.Equal(s.T(), 1, len(banners))
				s.AfterTest("", "")
			},
			err: true,
		},
		{
			name:    "not found",
			headers: validMetadata,
			preCondition: func() {
				err := s.repo.AddSlot(pageURL, slotID, slotDescription)
				require.Nil(s.T(), err)
				err = s.repo.AddBannerToSlot(pageURL, slotID, bannerID, bannerDescription)
				require.Nil(s.T(), err)
			},
			request: &grpcservice.DeleteBannerRequest{
				SlotId: 2,
			},
			postCondition: func() {
				slots, _ := s.repo.GetSlotsByPageURL(pageURL)
				require.Equal(s.T(), 1, len(slots))
				s.AfterTest("", "")
			},
			err: true,
		},
	}

	for id, tcase := range tcases {
		s.Run(fmt.Sprintf("%d: %v", id, tcase.name), func() {
			if tcase.preCondition != nil {
				tcase.preCondition()
			}
			ctx := metadata.NewOutgoingContext(context.Background(), tcase.headers)
			response, err := s.client.DeleteBanner(ctx, tcase.request)
			if tcase.err {
				require.NotNil(s.T(), err)
				require.Nil(s.T(), response)
			} else {
				require.Nil(s.T(), err)
				require.NotNil(s.T(), response)
			}
			if tcase.postCondition != nil {
				tcase.postCondition()
			}
		})
	}
}

func (s *Suite) TestIntegration_ClickOnBanner() {
	tcases := []struct {
		name          string
		headers       metadata.MD
		preCondition  condition
		request       *grpcservice.ClickRequest
		postCondition condition
		err           bool
	}{
		{
			name:         "bad: banner not found",
			headers:      validMetadata,
			preCondition: func() { s.AfterTest("", "") },
			request: &grpcservice.ClickRequest{
				SlotId:   slotID,
				BannerId: 2,
				UserAge:  userAge,
				UserSex:  userSex,
			},
			postCondition: func() {
				group, err := s.repo.GetGroup(userAge, userSex)
				require.Nil(s.T(), err)
				// 10 tries - because queue makes delay
				clicks := uint(0)
				for i := 0; i < 10; i++ {
					actions, _ := s.repo.GetActions(pageURL, slotID, bannerID)
					if actions[*group].Clicks != 0 {
						clicks = actions[*group].Clicks
						break
					}
					time.Sleep(time.Second)
				}
				require.Equal(s.T(), uint(0), clicks)
				s.AfterTest("", "")
			},
			err: true,
		},
		{
			name:    "good",
			headers: validMetadata,
			preCondition: func() {
				err := s.repo.AddSlot(pageURL, slotID, slotDescription)
				require.Nil(s.T(), err)
				err = s.repo.AddBannerToSlot(pageURL, slotID, bannerID, bannerDescription)
				require.Nil(s.T(), err)
			},
			request: &grpcservice.ClickRequest{
				SlotId:   slotID,
				BannerId: bannerID,
				UserAge:  userAge,
				UserSex:  userSex,
			},
			postCondition: func() {
				group, err := s.repo.GetGroup(userAge, userSex)
				require.Nil(s.T(), err)
				// 10 tries - because queue makes delay
				clicks := uint(0)
				for i := 0; i < 10; i++ {
					actions, err := s.repo.GetActions(pageURL, slotID, bannerID)
					require.Nil(s.T(), err)
					if actions[*group].Clicks != 0 {
						clicks = actions[*group].Clicks
						break
					}
					time.Sleep(time.Second)
				}
				require.Equal(s.T(), uint(1), clicks)
				s.AfterTest("", "")
			},
			err: false,
		},
		{
			name:    "unauth",
			headers: invalidMetadata,
			request: &grpcservice.ClickRequest{
				SlotId:   slotID,
				BannerId: bannerID,
				UserAge:  userAge,
				UserSex:  userSex,
			},
			err: true,
		},
	}

	for id, tcase := range tcases {
		s.Run(fmt.Sprintf("%d: %v", id, tcase.name), func() {
			if tcase.preCondition != nil {
				tcase.preCondition()
			}
			ctx := metadata.NewOutgoingContext(context.Background(), tcase.headers)
			response, err := s.client.ClickEvent(ctx, tcase.request)
			if tcase.err {
				require.NotNil(s.T(), err)
				require.Nil(s.T(), response)
			} else {
				require.Nil(s.T(), err)
				require.NotNil(s.T(), response)
			}
			if tcase.postCondition != nil {
				tcase.postCondition()
			}
		})
	}
}

func (s *Suite) TestIntegration_GetNextBanner() {
	tcases := []struct {
		name          string
		headers       metadata.MD
		preCondition  condition
		request       *grpcservice.GetNextBannerRequest
		response      *grpcservice.GetNextBannerResponse
		postCondition condition
		err           bool
	}{
		{
			name:         "bad: banner not found",
			headers:      validMetadata,
			preCondition: func() { s.AfterTest("", "") },
			request: &grpcservice.GetNextBannerRequest{
				SlotId:  2,
				UserAge: userAge,
				UserSex: userSex,
			},
			response: &grpcservice.GetNextBannerResponse{
				BannerId: 0,
			},
			postCondition: func() {
				group, err := s.repo.GetGroup(userAge, userSex)
				require.Nil(s.T(), err)
				// 10 tries - because queue makes delay
				shows := uint(0)
				for i := 0; i < 10; i++ {
					actions, _ := s.repo.GetActions(pageURL, slotID, bannerID)
					if actions[*group].Shows != 0 {
						shows = actions[*group].Shows
						break
					}
					time.Sleep(time.Second)
				}
				require.Equal(s.T(), uint(0), shows)
				s.AfterTest("", "")
			},
			err: true,
		},
		{
			name:    "good",
			headers: validMetadata,
			preCondition: func() {
				err := s.repo.AddSlot(pageURL, slotID, slotDescription)
				require.Nil(s.T(), err)
				err = s.repo.AddBannerToSlot(pageURL, slotID, bannerID, bannerDescription)
				require.Nil(s.T(), err)
			},
			request: &grpcservice.GetNextBannerRequest{
				SlotId:  slotID,
				UserAge: userAge,
				UserSex: userSex,
			},
			response: &grpcservice.GetNextBannerResponse{
				BannerId: bannerID,
			},
			postCondition: func() {
				group, err := s.repo.GetGroup(userAge, userSex)
				require.Nil(s.T(), err)
				// 10 tries - because queue makes delay
				shows := uint(0)
				for i := 0; i < 10; i++ {
					actions, err := s.repo.GetActions(pageURL, slotID, bannerID)
					require.Nil(s.T(), err)
					if actions[*group].Shows != 0 {
						shows = actions[*group].Shows
						break
					}
					time.Sleep(time.Second)
				}
				require.Equal(s.T(), uint(1), shows)
				s.AfterTest("", "")
			},
			err: false,
		},
		{
			name:    "unauth",
			headers: invalidMetadata,
			request: &grpcservice.GetNextBannerRequest{
				SlotId:  slotID,
				UserAge: userAge,
				UserSex: userSex,
			},
			err: true,
		},
	}

	for id, tcase := range tcases {
		s.Run(fmt.Sprintf("%d: %v", id, tcase.name), func() {
			if tcase.preCondition != nil {
				tcase.preCondition()
			}
			ctx := metadata.NewOutgoingContext(context.Background(), tcase.headers)
			response, err := s.client.GetNextBanner(ctx, tcase.request)
			if tcase.err {
				require.NotNil(s.T(), err)
			} else {
				require.Nil(s.T(), err)
			}
			require.Equal(s.T(), tcase.response.GetBannerId(), response.GetBannerId())
			if tcase.postCondition != nil {
				tcase.postCondition()
			}
		})
	}
}
