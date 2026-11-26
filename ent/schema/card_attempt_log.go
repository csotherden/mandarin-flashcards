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

// CardAttemptLog is one graded interaction with a card inside a session.
type CardAttemptLog struct {
	ent.Schema
}

func (CardAttemptLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("mode").
			Values("image_speak", "hanzi_speak", "picture_hanzi_match"),
		field.Float("overall_score").Default(0),
		field.Float("tone_accuracy_score").Optional().Nillable(),
		field.JSON("raw_api_feedback", &models.StoredFeedback{}).Optional(),
		field.Text("llm_coaching_tip").Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("lesson_attempt_id", uuid.UUID{}),
		field.UUID("flashcard_id", uuid.UUID{}),
	}
}

func (CardAttemptLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("card_attempt_logs").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("lesson_attempt", LessonAttempt.Type).
			Ref("logs").
			Unique().
			Required().
			Field("lesson_attempt_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("flashcard", Flashcard.Type).
			Ref("attempt_logs").
			Unique().
			Required().
			Field("flashcard_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (CardAttemptLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("owner", "flashcard"),
		index.Edges("lesson_attempt"),
	}
}

func (CardAttemptLog) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
