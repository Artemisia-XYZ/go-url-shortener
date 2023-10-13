package usecases

import (
	"errors"
	"sync"
	"time"
	"url-shortener/domain"
	"url-shortener/helpers"
	"url-shortener/logs"
	"url-shortener/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	maxAttempts = 3
	slashLength = 6
)

var (
	ErrUnexpected        = errors.New("unexpected error")
	ErrCreateShortLink   = errors.New("create short link failed")
	ErrGenerateSlashCode = errors.New("generate slash code failed")
	ErrSlashCodeExists   = errors.New("slash code exists already")
)

type visitorQueue struct {
	isRunning bool
	order     []string
	counts    map[string]int
	mu        sync.Mutex
}

type shortLinkUsecase struct {
	shortLinkRepo domain.ShortLinkRepository
	visitorQueue  *visitorQueue
}

func NewShortLinkUsecase(shortLinkRepo domain.ShortLinkRepository) *shortLinkUsecase {
	visitorQueue := &visitorQueue{
		counts: make(map[string]int),
	}

	return &shortLinkUsecase{shortLinkRepo, visitorQueue}
}

func (u *shortLinkUsecase) CreateShortLink(req *domain.CreateShortLinkRequest) (*models.ShortLink, error) {
	shortLink := &models.ShortLink{
		ID:          uuid.New(),
		Destination: req.Destination,
	}

	if req.SlashCode == "" {
		shortLink.SlashCode = u.generateSlashCode()
		if shortLink.SlashCode == "" {
			return nil, ErrGenerateSlashCode
		}
	} else {
		err := u.checkSlashCodeExist(req.SlashCode)
		if err != nil {
			return nil, err
		}
		shortLink.SlashCode = req.SlashCode
	}

	if err := u.shortLinkRepo.Create(shortLink); err != nil {
		logs.Error(err.Error())
		return nil, ErrCreateShortLink
	}

	return shortLink, nil
}

func (u *shortLinkUsecase) FindBySlashCode(slashCode string) (*models.ShortLink, error) {
	shortLink, err := u.shortLinkRepo.FindBySlashCode(slashCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		logs.Error(err.Error())
		return nil, ErrUnexpected
	}

	return shortLink, nil
}

func (u *shortLinkUsecase) Redirect(slashCode string) (string, error) {
	dest, err := u.shortLinkRepo.FindShortLinkCache(slashCode)
	if err == nil {
		go u.incrementVisitorEnqueue(slashCode)
		return dest, nil
	}

	shortLink, err := u.shortLinkRepo.FindBySlashCode(slashCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", err
		}
		logs.Error(err.Error())
		return "", ErrUnexpected
	}

	go u.setShortLinkCache(slashCode, shortLink.Destination, 3*time.Hour)
	go u.incrementVisitorEnqueue(slashCode)

	return shortLink.Destination, nil
}

func (u *shortLinkUsecase) generateSlashCode() string {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		slashCode := helpers.StrRandom(slashLength)
		_, err := u.shortLinkRepo.FindBySlashCode(slashCode)
		if err == gorm.ErrRecordNotFound {
			return slashCode
		}
	}
	return ""
}

func (u *shortLinkUsecase) checkSlashCodeExist(slashCode string) error {
	_, err := u.shortLinkRepo.FindBySlashCode(slashCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		logs.Error(err.Error())
		return ErrUnexpected
	}
	return ErrSlashCodeExists
}

func (u *shortLinkUsecase) setShortLinkCache(slashCode string, destination string, duration time.Duration) {
	err := u.shortLinkRepo.SetShortLinkCache(slashCode, destination, duration)
	if err != nil {
		logs.Error(err.Error())
	}
}

func (u *shortLinkUsecase) incrementVisitorEnqueue(slashCode string) {
	u.visitorQueue.mu.Lock()
	defer u.visitorQueue.mu.Unlock()

	if _, exist := u.visitorQueue.counts[slashCode]; !exist {
		u.visitorQueue.counts[slashCode] = 1
		u.visitorQueue.order = append(u.visitorQueue.order, slashCode)
		if !u.visitorQueue.isRunning {
			u.visitorQueue.isRunning = true
			go u.incrementVisitorQueueWorker()
		}
	} else {
		u.visitorQueue.counts[slashCode] += 1
	}
}

func (u *shortLinkUsecase) incrementVisitorQueueWorker() {
	for {
		u.visitorQueue.mu.Lock()
		if len(u.visitorQueue.counts) == 0 {
			u.visitorQueue.isRunning = false
			u.visitorQueue.mu.Unlock()
			return
		}

		code := u.visitorQueue.order[0]
		u.visitorQueue.order = u.visitorQueue.order[1:]

		visitors := u.visitorQueue.counts[code]
		delete(u.visitorQueue.counts, code)
		u.visitorQueue.mu.Unlock()

		err := u.shortLinkRepo.IncrementVisitor(code, visitors)
		if err != nil {
			logs.Error(err.Error())
		}
	}
}
