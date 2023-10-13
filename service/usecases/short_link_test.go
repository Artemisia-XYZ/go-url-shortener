package usecases

import (
	"errors"
	"testing"
	"url-shortener/domain"
	mockDomain "url-shortener/domain/mocks"
	"url-shortener/logs"
	"url-shortener/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func SetupLogger(t *testing.T) func() {
	logs.NewLogger()

	return func() { logs.Close() }
}

func TestNewShortLinkUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mockDomain.NewMockShortLinkRepository(ctrl)
	usecase := NewShortLinkUsecase(mock)

	assert.NotNil(t, usecase.shortLinkRepo)
	assert.NotNil(t, usecase.visitorQueue)
}

func TestShortLinkCreateShortLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	closeLog := SetupLogger(t)
	defer closeLog()

	mockData := struct {
		shortLink *models.ShortLink
		err       error
	}{
		shortLink: &models.ShortLink{
			ID:          uuid.New(),
			SlashCode:   "foo",
			Destination: "https://example.com",
		},
		err: errors.New("error"),
	}

	tests := []struct {
		name        string
		request     *domain.CreateShortLinkRequest
		setup       func(mr *mockDomain.MockShortLinkRepository)
		expected    *models.ShortLink
		expectedErr error
	}{
		{
			name: "success",
			request: &domain.CreateShortLinkRequest{
				Destination: mockData.shortLink.Destination,
			},
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindBySlashCode(gomock.Any()).MaxTimes(maxAttempts).Return(nil, gorm.ErrRecordNotFound)
				mr.EXPECT().Create(gomock.Any()).DoAndReturn(func(shortLink *models.ShortLink) error {
					shortLink.ID = mockData.shortLink.ID
					shortLink.SlashCode = mockData.shortLink.SlashCode
					return nil
				})
			},
			expected: mockData.shortLink,
		}, {
			name: "success with custom slash code",
			request: &domain.CreateShortLinkRequest{
				SlashCode:   mockData.shortLink.SlashCode,
				Destination: mockData.shortLink.Destination,
			},
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(nil, gorm.ErrRecordNotFound)
				mr.EXPECT().Create(gomock.Any()).DoAndReturn(func(shortLink *models.ShortLink) error {
					shortLink.ID = mockData.shortLink.ID
					return nil
				})
			},
			expected: mockData.shortLink,
		}, {
			name: "error",
			request: &domain.CreateShortLinkRequest{
				Destination: mockData.shortLink.Destination,
			},
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(nil, gorm.ErrRecordNotFound).MaxTimes(maxAttempts)
				mr.EXPECT().Create(gomock.Any()).Return(mockData.err)
			},
			expectedErr: ErrCreateShortLink,
		}, {
			name: "error in generateSlashCode()",
			request: &domain.CreateShortLinkRequest{
				Destination: mockData.shortLink.Destination,
			},
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(nil, mockData.err).MinTimes(maxAttempts)
			},
			expectedErr: ErrGenerateSlashCode,
		}, {
			name: "error is checkSlashCodeExist()",
			request: &domain.CreateShortLinkRequest{
				SlashCode:   mockData.shortLink.SlashCode,
				Destination: mockData.shortLink.Destination,
			},
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(nil, mockData.err)
			},
			expectedErr: ErrUnexpected,
		}, {
			name: "error custom slash code is exist",
			request: &domain.CreateShortLinkRequest{
				SlashCode:   mockData.shortLink.SlashCode,
				Destination: mockData.shortLink.Destination,
			},
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(&models.ShortLink{}, nil)
			},
			expectedErr: ErrSlashCodeExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockDomain.NewMockShortLinkRepository(ctrl)
			usecase := NewShortLinkUsecase(mock)
			tt.setup(mock)

			res, err := usecase.CreateShortLink(tt.request)
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

func TestShortLinkFindBySlashCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	closeLog := SetupLogger(t)
	defer closeLog()

	slashCode := "foo"
	tests := []struct {
		name        string
		setup       func(mr *mockDomain.MockShortLinkRepository)
		expected    *models.ShortLink
		expectedErr error
	}{
		{
			name: "success",
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(&models.ShortLink{SlashCode: slashCode}, nil)
			},
			expected: &models.ShortLink{SlashCode: slashCode},
		}, {
			name: "not found",
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedErr: gorm.ErrRecordNotFound,
		}, {
			name: "error",
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(nil, errors.New("error"))
			},
			expectedErr: ErrUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockDomain.NewMockShortLinkRepository(ctrl)
			usecase := NewShortLinkUsecase(mock)
			tt.setup(mock)

			shortLink, err := usecase.FindBySlashCode(slashCode)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, shortLink)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, shortLink)
			}
		})
	}
}

func TestShortLinkRedirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	closeLog := SetupLogger(t)
	defer closeLog()

	mockData := struct {
		shortLink *models.ShortLink
		err       error
	}{
		shortLink: &models.ShortLink{
			SlashCode:   "foo",
			Destination: "https://example.com",
		},
		err: errors.New("error"),
	}

	tests := []struct {
		name        string
		setup       func(mr *mockDomain.MockShortLinkRepository)
		modUcase    func(u *shortLinkUsecase)
		expected    string
		expectedErr error
	}{
		{
			name: "redirect with cache hit",
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindShortLinkCache(gomock.Any()).Return(mockData.shortLink.Destination, nil)
				mr.EXPECT().IncrementVisitor(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expected: mockData.shortLink.Destination,
		}, {
			name: "redirect with cache miss",
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindShortLinkCache(gomock.Any()).Return("", redis.Nil)
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(mockData.shortLink, nil)
				mr.EXPECT().SetShortLinkCache(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mr.EXPECT().IncrementVisitor(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expected: mockData.shortLink.Destination,
		}, {
			name: "redirect no slash code",
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindShortLinkCache(gomock.Any()).Return("", redis.Nil)
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedErr: gorm.ErrRecordNotFound,
		}, {
			name: "error increment vistor",
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindShortLinkCache(gomock.Any()).Return(mockData.shortLink.Destination, nil)
				mr.EXPECT().IncrementVisitor(gomock.Any(), gomock.Any()).Return(mockData.err).AnyTimes()
			},
			expected: mockData.shortLink.Destination,
		}, {
			name: "error setShortLinkCache()",
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindShortLinkCache(gomock.Any()).Return("", redis.Nil)
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(mockData.shortLink, nil)
				mr.EXPECT().SetShortLinkCache(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockData.err).AnyTimes()
				mr.EXPECT().IncrementVisitor(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expected: mockData.shortLink.Destination,
		}, {
			name: "error FindBySlashCode()",
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindShortLinkCache(gomock.Any()).Return("", redis.Nil)
				mr.EXPECT().FindBySlashCode(gomock.Any()).Return(nil, mockData.err)
			},
			expectedErr: ErrUnexpected,
		}, {
			name: "test incrementVisitorEnqueue()",
			setup: func(mr *mockDomain.MockShortLinkRepository) {
				mr.EXPECT().FindShortLinkCache(mockData.shortLink.SlashCode).Return(mockData.shortLink.Destination, nil)
				mr.EXPECT().IncrementVisitor(mockData.shortLink.SlashCode, gomock.Any()).Return(nil).AnyTimes()
			},
			modUcase: func(u *shortLinkUsecase) {
				u.visitorQueue.order = append(u.visitorQueue.order, mockData.shortLink.SlashCode)
				u.visitorQueue.counts[mockData.shortLink.SlashCode] = 1
			},
			expected: mockData.shortLink.Destination,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockDomain.NewMockShortLinkRepository(ctrl)
			usecase := NewShortLinkUsecase(mock)
			if tt.modUcase != nil {
				tt.modUcase(usecase)
			}
			tt.setup(mock)

			dest, err := usecase.Redirect(mockData.shortLink.SlashCode)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Empty(t, dest)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, dest)
			}
		})
	}
}
