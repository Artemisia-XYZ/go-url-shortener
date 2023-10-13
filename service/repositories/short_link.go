package repositories

import (
	"context"
	"time"
	"url-shortener/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const cacheDestPrefix = "dest_slash_"

type shortLinkRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewShortLinkRepository(db *gorm.DB, rdb *redis.Client) *shortLinkRepository {
	return &shortLinkRepository{db, rdb}
}

func (r *shortLinkRepository) Create(shortLink *models.ShortLink) error {
	return r.db.Create(shortLink).Error
}

func (r *shortLinkRepository) FindBySlashCode(slashCode string) (*models.ShortLink, error) {
	shortLink := &models.ShortLink{}
	if err := r.db.Where("slash_code = ?", slashCode).First(shortLink).Error; err != nil {
		return nil, err
	}
	return shortLink, nil
}

func (r *shortLinkRepository) IncrementVisitor(slashCode string, visitors int) error {
	return r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&models.ShortLink{}).
		Where("slash_code = ?", slashCode).
		UpdateColumn("visitors", gorm.Expr("visitors + ?", visitors)).
		Error
}

func (r *shortLinkRepository) SetShortLinkCache(slashCode string, dest string, exp time.Duration) error {
	return r.rdb.Set(context.Background(), cacheDestPrefix+slashCode, dest, exp).Err()
}

func (r *shortLinkRepository) FindShortLinkCache(slashCode string) (string, error) {
	dest, err := r.rdb.Get(context.Background(), cacheDestPrefix+slashCode).Result()
	if err != nil {
		return "", err
	}
	return dest, nil
}
