package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"

	"github.com/cumbreras/shortener/ent"
	"github.com/cumbreras/shortener/ent/shortenurl"
	"github.com/cumbreras/shortener/model"
)

// Repository is the interface
type Repository interface {
	Save(context.Context, *model.ShortenURL) (*ent.ShortenURL, error)
	Find(context.Context, string) (*ent.ShortenURL, error)
	Delete(context.Context, string) error
	NewUS() *model.ShortenURL
}

// Repository is the struct
type repository struct {
	dbClient *ent.Client
	logger   hclog.Logger
}

// New creates a new repository
func New(dbClient *ent.Client, logger hclog.Logger) Repository {
	return &repository{dbClient: dbClient, logger: logger}
}

func (r *repository) NewUS() *model.ShortenURL {
	return model.New()
}

// Find finds a ShortenURL based on Code
func (r *repository) Find(ctx context.Context, code string) (*ent.ShortenURL, error) {
	uid, err := uuid.Parse(code)

	if err != nil {
		return nil, err
	}

	s, err := r.dbClient.ShortenURL.
		Query().
		Where(shortenurl.CodeEQ(uid)).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	r.logger.Info("Repository: ShortenURL found: %#v", s)

	return s, nil
}

// Save stores new shortenURL to the storage
func (r *repository) Save(ctx context.Context, su *model.ShortenURL) (*ent.ShortenURL, error) {
	s, err := r.dbClient.ShortenURL.
		Create().
		SetURL(su.URL).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	r.logger.Info("Repository: ShortenURL created: %#v", s)
	return s, nil
}

// Delete eliminates a shortenURL from the storage
func (r *repository) Delete(ctx context.Context, code string) error {
	s, err := r.Find(ctx, code)

	if err != nil {
		return err
	}

	err = r.dbClient.ShortenURL.
		DeleteOne(s).
		Exec(ctx)

	if err != nil {
		return err
	}

	r.logger.Info("Repository: ShortenURL found: %#v", s)
	return nil
}
