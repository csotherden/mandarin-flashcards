package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/google/uuid"
)

// LibrarySet is a curated set in the global library, owned by no one. Users browse it and
// import it: importing copies it into their own CardSet and Flashcards, which they then own
// outright. Loaded from JSON by the api-server -library-load command, keyed by slug.
type LibrarySet struct {
	ent.Schema
}

func (LibrarySet) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("slug").
			Unique().
			NotEmpty().
			MaxLen(80),
		field.String("title").
			NotEmpty().
			MaxLen(120),
		field.String("description").
			Default("").
			MaxLen(500),
		// A short label like "Beginner", shown on the browse card.
		field.String("level").
			Default("").
			MaxLen(40),
		// Browse order, lowest first.
		field.Int("sort_order").
			Default(0),
		// Unpublished sets stay in the database (imports keep pointing at them) but aren't listed.
		field.Bool("published").
			Default(true),
	}
}

func (LibrarySet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("cards", LibraryCard.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (LibrarySet) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// LibraryCard is one word in a LibrarySet, with everything a Flashcard copy needs.
type LibraryCard struct {
	ent.Schema
}

func (LibraryCard) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("set_id", uuid.UUID{}),
		field.Int("sort_order").Default(0),
		field.String("hanzi").NotEmpty().MaxLen(64),
		field.String("pinyin").Default("").MaxLen(200),
		field.String("tonal_pinyin").Default("").MaxLen(200),
		field.String("english").Default("").MaxLen(500),
		field.String("example_sentence_hanzi").Default("").MaxLen(500),
		field.String("example_sentence_pinyin").Default("").MaxLen(500),
		field.String("image_url").Default("").MaxLen(2048),
		field.JSON("unsplash_attribution", &models.UnsplashAttribution{}).Optional(),
		// Category names given to the user's copy (created in their account if missing).
		field.Strings("categories").Optional(),
	}
}

func (LibraryCard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("set", LibrarySet.Type).
			Ref("cards").
			Unique().
			Required().
			Field("set_id"),
	}
}

func (LibraryCard) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("set_id", "hanzi").Unique(),
	}
}

func (LibraryCard) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
