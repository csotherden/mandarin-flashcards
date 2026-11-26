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

// Flashcard is one vocabulary item owned by a user.
type Flashcard struct {
	ent.Schema
}

func (Flashcard) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("hanzi").
			NotEmpty().
			MaxLen(64),
		field.String("pinyin").
			Default("").
			MaxLen(160),
		field.String("english").
			Default("").
			MaxLen(300),
		field.String("example_sentence_hanzi").
			Default("").
			MaxLen(500),
		field.String("example_sentence_pinyin").
			Default("").
			MaxLen(500),
		field.String("image_url").
			Default("").
			MaxLen(2048),
		field.String("image_object_key").
			Optional().
			Nillable().
			MaxLen(512),
		field.JSON("unsplash_attribution", &models.UnsplashAttribution{}).
			Optional(),
	}
}

func (Flashcard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("flashcards").
			Unique().
			Required(),
		edge.To("categories", Category.Type).
			StorageKey(edge.Table("flashcard_categories"), edge.Columns("flashcard_id", "category_id")),
		edge.To("set_memberships", SetFlashcard.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("disabled_in", Lesson.Type).
			Ref("disabled_cards"),
		edge.To("attempt_logs", CardAttemptLog.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Flashcard) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("owner"),
		index.Fields("hanzi").Edges("owner"),
	}
}

func (Flashcard) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
