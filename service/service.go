package service

import (
	"context"
	"net/url"

	"github.com/cumbreras/shortener/ent"
	"github.com/cumbreras/shortener/model"
	"github.com/cumbreras/shortener/repository"
	"github.com/hashicorp/go-hclog"
)


// RepositoryService interface
type RepositoryService interface {
	Find(context.Context, string) (*ent.ShortenURL, error)
	Create(context.Context, *model.ShortenURL) (*ent.ShortenURL, error)
	Destroy(context.Context, string) error
	NewUS() *model.ShortenURL
}

// repositoryService is the structure for the service
type repositoryService struct {
	repository repository.Repository
	logger     hclog.Logger
	RepositoryService
}

// New creates a New repositoryService
func New(repository repository.Repository, logger hclog.Logger) RepositoryService {
	return &repositoryService{repository: repository, logger: logger}
}

// New produces a new ShortenURL
func (rs *repositoryService) NewUS() *model.ShortenURL {
	return rs.repository.NewUS()
}

// Find finds a ShortenURL by Code
func (rs *repositoryService) Find(ctx context.Context, code string) (*ent.ShortenURL, error) {
	rs.logger.Info(code)
	s, err := rs.repository.Find(ctx, code)
	if err != nil {
		rs.logger.Error("RepositoryService: Could not Find", err)
		return nil, err
	}

	return s, nil
}

// Create creates a ShortenURL
func (rs *repositoryService) Create(ctx context.Context, su *model.ShortenURL) (*ent.ShortenURL, error) {
	_, err := url.ParseRequestURI(su.URL)
	if err != nil {
		rs.logger.Error("RepositoryService: Error validating: %#v; got err", su, err.Error())
		return nil, err
	}

	s, err := rs.repository.Save(ctx, su)
	if err != nil {
		rs.logger.Error("RepositoryService: Could not create")
		return nil, err
	}
	rs.logger.Info("Creating %#v", s)
	return s, nil
}

// Destroy deletes a ShortenURL
func (rs *repositoryService) Destroy(ctx context.Context, code string) error {
	err := rs.repository.Delete(ctx, code)
	if err != nil {
		rs.logger.Error("RepositoryService: Could not delete")
		return err
	}
	return nil
}
