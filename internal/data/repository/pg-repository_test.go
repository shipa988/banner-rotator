package repository

import (
	"database/sql"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"regexp"
	"testing"
	"time"
)

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

type Suite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	repository *PGRepo
	fakedt     AnyTime
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

	s.DB.LogMode(true)
	s.repository = NewPGRepo(s.DB, false)
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

func (s *Suite) TestPGRepo_DeleteSlot() {
	var (
		url = "site.com"
		id  = 1
	)
	expectedRows := sqlmock.NewRows([]string{"url", "id"}).AddRow(url, id)
	//expectedRowsins := sqlmock.NewRows([]string{"id"}).AddRow(id)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pages"`)).WithArgs(url).WillReturnRows(expectedRows)
	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "slots"`)).WithArgs(id, id).WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()
	err := s.repository.DeleteSlot(url, uint(id))
	require.NoError(s.T(), err)
}

func (s *Suite) TestPGRepo_DeleteAllSlost() {
	var (
		url = "site.com"
		id  = 1
	)
	expectedRows := sqlmock.NewRows([]string{"url", "id"}).AddRow(url, id)
	//expectedRowsins := sqlmock.NewRows([]string{"id"}).AddRow(id)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pages"`)).WithArgs(url).WillReturnRows(expectedRows)
	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "slots"`)).WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()
	err := s.repository.DeleteAllSlots(url)
	require.NoError(s.T(), err)
}

func (s *Suite) TestPGRepo_AddBanner() {
	var (
		url   = "site.com"
		id    = 1
		descr = "111"
	)
	expectedPages := sqlmock.NewRows([]string{"url", "id"}).AddRow(url, id)
	expectedSlots := sqlmock.NewRows([]string{"id", "description"}).AddRow(id, descr)
	expectedBanners := sqlmock.NewRows([]string{"id"}).AddRow(id)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pages"`)).WithArgs(url).WillReturnRows(expectedPages)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "slots"`)).WithArgs(url, id).WillReturnRows(expectedSlots)
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "banners"`)).WithArgs(AnyTime{}, AnyTime{}, nil, id, id, descr).WillReturnRows(expectedBanners)
	s.mock.ExpectCommit()
	err := s.repository.AddBannerToSlot(url, uint(id), uint(id), descr)
	require.NoError(s.T(), err)
}

func (s *Suite) AfterTest(_, _ string) {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}
