package handlers

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
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

func TestNewFactory(t *testing.T) {
	db, _, closeDB := SetupDatabaseMock(t)
	defer closeDB()

	_, rdb, closeRedis := SetupRedisMock(t)
	defer closeRedis()

	factory := NewFactory(db, rdb)
	assert.NotNil(t, factory)
}
