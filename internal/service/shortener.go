package service

import (
	"net/url"
	"strings"

	"github.com/artem/url-shortener/internal/model"
	"github.com/artem/url-shortener/pkg/base62"
)

type ShortenerService struct {
	linkRepo  model.LinkRepository
	cacheRepo model.CacheRepository
}

func NewShortenerService(linkRepo model.LinkRepository, cacheRepo model.CacheRepository) *ShortenerService {
	return &ShortenerService{linkRepo: linkRepo, cacheRepo: cacheRepo}
}

func (s *ShortenerService) Create(originalURL string, ownerID *string, baseURL string) (*model.CreateLinkResponse, error) {
	originalURL = strings.TrimSpace(originalURL)
	if !isValidURL(originalURL) {
		return nil, ErrInvalidURL
	}

	link, err := s.linkRepo.Insert(originalURL, ownerID)
	if err != nil {
		return nil, err
	}

	shortCode := base62.Encode(uint64(link.ID))
	if err := s.linkRepo.UpdateShortCode(link.ID, shortCode); err != nil {
		return nil, err
	}

	shortURL := strings.TrimRight(baseURL, "/") + "/" + shortCode

	return &model.CreateLinkResponse{
		ShortURL:    shortURL,
		ShortCode:   shortCode,
		OriginalURL: originalURL,
	}, nil
}

func (s *ShortenerService) Resolve(code string) (string, error) {
	if cached, ok := s.cacheRepo.Get(code); ok {
		return cached, nil
	}

	link, err := s.linkRepo.FindByShortCode(code)
	if err != nil {
		return "", err
	}
	if link == nil {
		return "", ErrLinkNotFound
	}

	s.cacheRepo.Set(code, link.OriginalURL, 0)

	return link.OriginalURL, nil
}

func (s *ShortenerService) GetStats(linkID int64) (*model.LinkStats, error) {
	return s.linkRepo.GetStats(linkID)
}

func isValidURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
