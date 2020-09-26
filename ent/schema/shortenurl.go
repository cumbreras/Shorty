package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/field"
	"github.com/google/uuid"
)

// ShortenURL holds the schema definition for the ShortenURL entity.
type ShortenURL struct {
	ent.Schema
}

// Fields of the ShortenURL.
func (ShortenURL) Fields() []ent.Field {
	return []ent.Field{field.String("URL"), field.UUID("Code", uuid.UUID{}).Default(uuid.New)}
}

// Edges of the ShortenURL.
func (ShortenURL) Edges() []ent.Edge {
	return nil
}
