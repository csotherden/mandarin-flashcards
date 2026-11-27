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

// ToneProgress is a user's standing in one lesson of the tones tutorial. The lessons themselves
// are a fixed catalog in code (internal/tones); rows refer to them by key.
type ToneProgress struct {
	ent.Schema
}

func (ToneProgress) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.String("lesson_key").NotEmpty().MaxLen(40),
		// Set the first time the lesson is finished. Before then its items earn XP; after, they
		// are practice only (streak credit, no XP).
		field.Time("completed_at").Optional().Nillable(),
		// Average of each item's best score, as of the latest finish.
		field.Float("best_score").Default(0),
	}
}

func (ToneProgress) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("tone_progress").Unique().Required().Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (ToneProgress) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "lesson_key").Unique()}
}

func (ToneProgress) Mixin() []ent.Mixin { return []ent.Mixin{TimeMixin{}} }

// ToneAttempt is one graded take of a tones tutorial item.
type ToneAttempt struct {
	ent.Schema
}

func (ToneAttempt) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.String("lesson_key").NotEmpty().MaxLen(40),
		field.String("item_key").NotEmpty().MaxLen(40),
		field.Float("overall_score").Default(0),
		field.String("xp_tier").Default(""),
		field.Int("xp_earned").Default(0),
		field.JSON("raw_api_feedback", &models.StoredFeedback{}).Optional(),
	}
}

func (ToneAttempt) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("tone_attempts").Unique().Required().Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (ToneAttempt) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "lesson_key")}
}

func (ToneAttempt) Mixin() []ent.Mixin { return []ent.Mixin{TimeMixin{}} }
