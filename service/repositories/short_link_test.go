package repositories

import (
	"context"
	"errors"
	"testing"
	"time"
	"url-shortener/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupDatabaseMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("8.1.0"))
	instance, err := gorm.Open(mysql.New(mysql.Config{Conn: db}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Errorf("can't connect to mock database: %v", err)
	}

	return instance, mock, func() { db.Close() }
}

func SetupRedisMock(t *testing.T) (*miniredis.Miniredis, *redis.Client, func()) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cleanup := func() {
		mr.Close()
		rdb.Close()
	}

	return mr, rdb, cleanup
}

func TestNewShortLinkRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, closeDB := SetupDatabaseMock(t)
	defer closeDB()

	_, rdb, closeRedis := SetupRedisMock(t)
	defer closeRedis()

	repo := NewShortLinkRepository(db, rdb)

	assert.NotNil(t, repo.db)
	assert.NotNil(t, repo.rdb)
}

func TestShortLinkCreate(t *testing.T) {
	db, mock, closeDB := SetupDatabaseMock(t)
	defer closeDB()

	mockData := struct {
		shortLink *models.ShortLink
		err       error
	}{
		shortLink: &models.ShortLink{
			SlashCode:   "example",
			Destination: "https://example.com",
			Visitors:    0,
		},
		err: errors.New("error"),
	}

	tests := []struct {
		name        string
		setup       func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `short_links`").
					WithArgs(
						sqlmock.AnyArg(),
						mockData.shortLink.SlashCode,
						mockData.shortLink.Destination,
						mockData.shortLink.Visitors,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name: "error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `short_links`").
					WithArgs(
						sqlmock.AnyArg(),
						mockData.shortLink.SlashCode,
						mockData.shortLink.Destination,
						mockData.shortLink.Visitors,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnError(mockData.err)
				mock.ExpectRollback()
			},
			expectedErr: mockData.err,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(mock)
			repo := &shortLinkRepository{db: db}
			err := repo.Create(mockData.shortLink)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShortLinkFindBySlashCode(t *testing.T) {
	db, mock, closeDB := SetupDatabaseMock(t)
	defer closeDB()

	mockData := struct {
		shortLink *models.ShortLink
		query     string
		err       error
	}{
		shortLink: &models.ShortLink{
			ID:          uuid.New(),
			SlashCode:   "foo",
			Destination: "https://example.com",
			Visitors:    0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		query: "SELECT (.+) FROM `short_links` WHERE slash_code = ?",
		err:   errors.New("error"),
	}

	tests := []struct {
		name        string
		setup       func(mock sqlmock.Sqlmock)
		expected    *models.ShortLink
		expectedErr error
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(mockData.query).
					WithArgs(mockData.shortLink.SlashCode).
					WillReturnRows(sqlmock.NewRows([]string{"id", "slash_code", "destination", "visitors", "created_at", "updated_at"}).
						AddRow(
							mockData.shortLink.ID,
							mockData.shortLink.SlashCode,
							mockData.shortLink.Destination,
							mockData.shortLink.Visitors,
							mockData.shortLink.CreatedAt,
							mockData.shortLink.UpdatedAt,
						))
			},
			expected: mockData.shortLink,
		}, {
			name: "not found",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(mockData.query).
					WithArgs(mockData.shortLink.SlashCode).
					WillReturnRows(sqlmock.NewRows([]string{}))
			},
			expectedErr: gorm.ErrRecordNotFound,
		}, {
			name: "error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(mockData.query).
					WithArgs(mockData.shortLink.SlashCode).
					WillReturnError(mockData.err)
			},
			expectedErr: mockData.err,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(mock)
			repo := &shortLinkRepository{db: db}
			res, err := repo.FindBySlashCode(mockData.shortLink.SlashCode)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestShortLinkIncrementVisitor(t *testing.T) {
	db, mock, closeDB := SetupDatabaseMock(t)
	defer closeDB()

	mockData := struct {
		slashCode string
		query     string
		err       error
	}{
		slashCode: "foo",
		query:     "UPDATE `short_links`",
		err:       errors.New("error"),
	}

	tests := []struct {
		name        string
		setup       func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(mockData.query).
					WithArgs(1, mockData.slashCode).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name: "not found",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(mockData.query).
					WithArgs(1, mockData.slashCode).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			expectedErr: gorm.ErrRecordNotFound,
		}, {
			name: "error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(mockData.query).
					WithArgs(1, mockData.slashCode).
					WillReturnError(mockData.err)
				mock.ExpectRollback()
			},
			expectedErr: mockData.err,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(mock)
			repo := &shortLinkRepository{db: db}
			err := repo.IncrementVisitor(mockData.slashCode, 1)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShortLinkSetShortLinkCache(t *testing.T) {
	tests := []struct {
		name       string
		slashCode  string
		dest       string
		expiration time.Duration
		setupErr   func(mr *miniredis.Miniredis)
	}{
		{
			name:       "success",
			slashCode:  "foo",
			dest:       "www.example.com",
			expiration: 1 * time.Second,
		}, {
			name:       "invalid expiration",
			slashCode:  "foo",
			dest:       "www.example.com",
			expiration: -1 * time.Second,
			setupErr: func(mr *miniredis.Miniredis) {
				mr.SetError("invalid expiration")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr, rdb, cleanup := SetupRedisMock(t)
			defer cleanup()

			if tt.setupErr != nil {
				tt.setupErr(mr)
			}

			repo := &shortLinkRepository{rdb: rdb}
			err := repo.SetShortLinkCache(tt.slashCode, tt.dest, tt.expiration)

			if tt.setupErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				dest, err := rdb.Get(context.Background(), cacheDestPrefix+tt.slashCode).Result()
				assert.NotEmpty(t, dest)
				assert.NoError(t, err)
			}
		})
	}
}

func TestShortLinkFindShortLinkCache(t *testing.T) {
	tests := []struct {
		name        string
		slashCode   string
		setup       func(rdb *redis.Client, key string, val string)
		expected    string
		setupErr    func(mr *miniredis.Miniredis)
		expectedErr error
	}{
		{
			name:      "success",
			slashCode: "foo",
			setup: func(rdb *redis.Client, key string, val string) {
				rdb.Set(context.Background(), cacheDestPrefix+key, val, 1*time.Second).Err()
			},
			expected: "bar",
		}, {
			name:        "not found",
			slashCode:   "foo",
			expectedErr: redis.Nil,
		}, {
			name:      "error",
			slashCode: "foo",
			setupErr: func(mr *miniredis.Miniredis) {
				mr.SetError("error")
			},
			expectedErr: errors.New("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr, rdb, cleanup := SetupRedisMock(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(rdb, tt.slashCode, tt.expected)
			}

			if tt.setupErr != nil {
				tt.setupErr(mr)
			}

			repo := &shortLinkRepository{rdb: rdb}
			dest, err := repo.FindShortLinkCache(tt.slashCode)

			if tt.expectedErr != nil {
				assert.Empty(t, dest)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NotEmpty(t, dest)
				assert.NoError(t, err)
			}
		})
	}
}
