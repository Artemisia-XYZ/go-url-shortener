package handlers

import (
	"url-shortener/repositories"
	"url-shortener/usecases"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Factory struct {
	ShortLink *shortLinkHandler
}

func NewFactory(db *gorm.DB, rdb *redis.Client) *Factory {
	shortLinkRepo := repositories.NewShortLinkRepository(db, rdb)
	shortLinkUcase := usecases.NewShortLinkUsecase(shortLinkRepo)
	shortLinkHandler := NewShortLinkHandler(shortLinkUcase)

	return &Factory{
		ShortLink: shortLinkHandler,
	}
}
