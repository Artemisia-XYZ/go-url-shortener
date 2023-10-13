package domain

import (
	"time"
	"url-shortener/models"
)

type ShortLinkRepository interface {
	Create(shortLink *models.ShortLink) error
	FindBySlashCode(slashCode string) (*models.ShortLink, error)
	IncrementVisitor(slashCode string, visitors int) error

	SetShortLinkCache(slashCode string, dest string, exp time.Duration) error
	FindShortLinkCache(slashCode string) (string, error)
}

type CreateShortLinkRequest struct {
	SlashCode   string `json:"slash_code" validate:"max=12"`
	Destination string `json:"destination" validate:"required,url,max=512"`
}

type ShortLinkUsecase interface {
	CreateShortLink(req *CreateShortLinkRequest) (*models.ShortLink, error)
	FindBySlashCode(slashCode string) (*models.ShortLink, error)
	Redirect(slashCode string) (string, error)
}
