package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"url-shortener/domain"
	mockDomain "url-shortener/domain/mocks"
	"url-shortener/helpers"
	"url-shortener/models"
	"url-shortener/usecases"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestNewShortLinkHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mockDomain.NewMockShortLinkUsecase(ctrl)
	handler := NewShortLinkHandler(mock)

	assert.NotNil(t, handler.shortLinkUcase)
}

func TestShortLinkCreateShortLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortLink := &models.ShortLink{
		SlashCode:   "foo",
		Origin:      "http://example.com/foo",
		Destination: "https://www.google.com",
	}

	tests := []struct {
		name         string
		setup        func(mu *mockDomain.MockShortLinkUsecase)
		requestBody  *domain.CreateShortLinkRequest
		expectedCode int
		expectedBody *models.ShortLink
	}{
		{
			name: "success without schema",
			setup: func(mu *mockDomain.MockShortLinkUsecase) {
				mu.EXPECT().CreateShortLink(gomock.Any()).DoAndReturn(func(req *domain.CreateShortLinkRequest) (*models.ShortLink, error) {
					return &models.ShortLink{
						SlashCode:   mockShortLink.SlashCode,
						Destination: req.Destination,
					}, nil
				})
			},
			requestBody: &domain.CreateShortLinkRequest{
				Destination: "www.google.com",
			},
			expectedCode: fiber.StatusCreated,
			expectedBody: &models.ShortLink{
				SlashCode:   mockShortLink.SlashCode,
				Origin:      mockShortLink.Origin,
				Destination: mockShortLink.Destination,
			},
		}, {
			name: "success with schema",
			setup: func(mu *mockDomain.MockShortLinkUsecase) {
				mu.EXPECT().CreateShortLink(gomock.Any()).DoAndReturn(func(req *domain.CreateShortLinkRequest) (*models.ShortLink, error) {
					return &models.ShortLink{
						SlashCode:   mockShortLink.SlashCode,
						Destination: req.Destination,
					}, nil
				})
			},
			requestBody: &domain.CreateShortLinkRequest{
				Destination: mockShortLink.Destination,
			},
			expectedCode: fiber.StatusCreated,
			expectedBody: &models.ShortLink{
				SlashCode:   mockShortLink.SlashCode,
				Origin:      mockShortLink.Origin,
				Destination: mockShortLink.Destination,
			},
		}, {
			name: "success with custom slash code",
			setup: func(mu *mockDomain.MockShortLinkUsecase) {
				mu.EXPECT().CreateShortLink(gomock.Any()).DoAndReturn(func(req *domain.CreateShortLinkRequest) (*models.ShortLink, error) {
					return &models.ShortLink{
						SlashCode:   req.SlashCode,
						Destination: req.Destination,
					}, nil
				})
			},
			requestBody: &domain.CreateShortLinkRequest{
				SlashCode:   "bar",
				Destination: mockShortLink.Destination,
			},
			expectedCode: fiber.StatusCreated,
			expectedBody: &models.ShortLink{
				SlashCode:   "bar",
				Origin:      "http://example.com/bar",
				Destination: mockShortLink.Destination,
			},
		}, {
			name:         "error invalid request",
			expectedCode: fiber.StatusUnprocessableEntity,
		}, {
			name:         "error empty request",
			requestBody:  &domain.CreateShortLinkRequest{},
			expectedCode: fiber.StatusBadRequest,
		}, {
			name: "error invalid url",
			requestBody: &domain.CreateShortLinkRequest{
				Destination: "invalid-url",
			},
			expectedCode: fiber.StatusBadRequest,
		}, {
			name: "error too long url",
			requestBody: &domain.CreateShortLinkRequest{
				Destination: mockShortLink.Destination + "/" + helpers.StrRandom(512),
			},
			expectedCode: fiber.StatusBadRequest,
		}, {
			name: "error create short link",
			setup: func(mu *mockDomain.MockShortLinkUsecase) {
				mu.EXPECT().CreateShortLink(gomock.Any()).Return(nil, errors.New("error"))
			},
			requestBody: &domain.CreateShortLinkRequest{
				Destination: mockShortLink.Destination,
			},
			expectedCode: fiber.StatusInternalServerError,
		}, {
			name: "error slash code is exists",
			setup: func(mu *mockDomain.MockShortLinkUsecase) {
				mu.EXPECT().CreateShortLink(gomock.Any()).Return(nil, usecases.ErrSlashCodeExists)
			},
			requestBody: &domain.CreateShortLinkRequest{
				SlashCode:   mockShortLink.SlashCode,
				Destination: mockShortLink.Destination,
			},
			expectedCode: fiber.StatusConflict,
		},
	}

	for _, tt := range tests {
		mock := mockDomain.NewMockShortLinkUsecase(ctrl)
		handler := NewShortLinkHandler(mock)
		if tt.setup != nil {
			tt.setup(mock)
		}

		app := fiber.New()
		app.Post("/links", handler.CreateShortLink)

		var buf bytes.Buffer
		if tt.requestBody != nil {
			err := json.NewEncoder(&buf).Encode(tt.requestBody)
			if err != nil {
				t.Errorf("failed to encode request body: %v", err)
			}
		}
		req := httptest.NewRequest("POST", "/links", &buf)
		req.Header.Set("Content-Type", "application/json")
		res, _ := app.Test(req)
		defer res.Body.Close()

		assert.Equal(t, tt.expectedCode, res.StatusCode)
		if tt.expectedBody != nil {
			body := &models.ShortLink{}
			err := json.NewDecoder(res.Body).Decode(body)
			if err != nil {
				t.Errorf("failed to decode response body: %v", err)
			}
			assert.Equal(t, body, tt.expectedBody)
		}
	}
}

func TestShortLinkRedirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	destination := "http://www.example.com"
	err := errors.New("internal error")

	tests := []struct {
		name         string
		setup        func(mu *mockDomain.MockShortLinkUsecase)
		expected     string
		expectedCode int
	}{
		{
			name: "redirect",
			setup: func(mu *mockDomain.MockShortLinkUsecase) {
				mu.EXPECT().Redirect(gomock.Any()).Return(destination, nil)
			},
			expected:     destination,
			expectedCode: fiber.StatusMovedPermanently,
		}, {
			name: "not found",
			setup: func(mu *mockDomain.MockShortLinkUsecase) {
				mu.EXPECT().Redirect(gomock.Any()).Return("", gorm.ErrRecordNotFound)
			},
			expectedCode: fiber.StatusNotFound,
		}, {
			name: "internal error",
			setup: func(mu *mockDomain.MockShortLinkUsecase) {
				mu.EXPECT().Redirect(gomock.Any()).Return("", err)
			},
			expectedCode: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		mock := mockDomain.NewMockShortLinkUsecase(ctrl)
		handler := NewShortLinkHandler(mock)
		if tt.setup != nil {
			tt.setup(mock)
		}

		app := fiber.New()
		app.Get("/:slash", handler.Redirect)
		req := httptest.NewRequest("GET", "/valid-slash", nil)
		res, _ := app.Test(req)
		defer res.Body.Close()

		assert.Equal(t, tt.expectedCode, res.StatusCode)
		if tt.expected != "" {
			assert.Equal(t, tt.expected, res.Header.Get("Location"))
		}
	}
}
