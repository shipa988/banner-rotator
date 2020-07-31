package repository

import (
	"database/sql"
	"database/sql/driver"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/shipa988/banner_rotator/internal/data/logger/zaplogger"
)

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

type Suite struct {
	suite.Suite
	DB         *gorm.DB
	mock       sqlmock.Sqlmock
	fakedt     AnyTime
	repository *PGRepo
}

func TestRepo(t *testing.T) {
	s := new(Suite)
	suite.Run(t, s)
}

func (s *Suite) SetupSuite() {
	var (
		db  *sql.DB
		err error
	)
	db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)
	s.DB, err = gorm.Open("postgres", db)
	require.NoError(s.T(), err)
	wr := os.Stdout
	logger := zaplogger.NewLogger(wr, false)
	s.repository = NewPGRepo(s.DB, logger, false)
	s.fakedt = AnyTime{}
}

func (s *Suite) TestPGRepo_AddSlot() {
	var (
		url   = "site.com"
		id    = 1
		descr = "111"
	)
	expectedRows := sqlmock.NewRows([]string{"url", "id"}).AddRow(url, id)
	expectedRowsins := sqlmock.NewRows([]string{"id"}).AddRow(id)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pages"`)).WithArgs(url).WillReturnRows(expectedRows)
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "slots"`)).WithArgs(AnyTime{}, AnyTime{}, nil, id, id, descr).WillReturnRows(expectedRowsins)
	s.mock.ExpectCommit()
	err := s.repository.AddSlot(url, uint(id), descr)
	require.NoError(s.T(), err)
}

// pain pain pain
/*func (s *Suite) TestPGRepo_AddBanner() {
	var (
		url   = "site.com"
		id    = 1
		descr = "111"
	)
	expectedPages := sqlmock.NewRows([]string{"url", "id"}).AddRow(url, id)
	expectedSlots := sqlmock.NewRows([]string{"id","id"}).AddRow(id,id)
	expectedBanners := sqlmock.NewRows([]string{"id"})
	expectedResult := sqlmock.NewRows([]string{"id"}).AddRow(id)
	expectedResult1 := sqlmock.NewRows([]string{})
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pages"`)).WithArgs(url).WillReturnRows(expectedPages)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "slots"`)).WithArgs(id,id).WillReturnRows(expectedSlots)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "banners"`)).WithArgs(id,descr).WillReturnRows(expectedBanners)
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "banners"`)).WithArgs(AnyTime{}, AnyTime{}, nil, id, descr).WillReturnRows(expectedResult)
	s.mock.ExpectCommit()
	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "banners"`)).WithArgs(AnyTime{}, AnyTime{}, nil, id, descr,id).WillReturnResult(sqlmock.NewResult(1, 0))
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "banner_slots"`)).WithArgs(AnyTime{}, AnyTime{}, nil, id, id).WillReturnRows(expectedResult1)
	err := s.repository.AddBannerToSlot(url, uint(id), uint(id), descr)
	require.NoError(s.T(), err)
}*/

func (s *Suite) AfterTest(_, _ string) {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}
