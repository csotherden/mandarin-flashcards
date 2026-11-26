package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// LessonAttempt is one study session. lesson is null for a quick start from a set
// or a "drill missed" session. card_ids freezes the deck order for the session.
type LessonAttempt struct {
	ent.Schema
}

func (LessonAttempt) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("mode").
			Values("image_speak", "hanzi_speak", "picture_hanzi_match"),
		field.Int("card_count_requested").Default(0),
		field.Float("overall_score").Optional().Nillable(),
		field.JSON("card_ids", []uuid.UUID{}).Optional(),
		field.UUID("lesson_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("started_at").Default(time.Now).Immutable(),
		field.Time("completed_at").Optional().Nillable(),
	}
}

func (LessonAttempt) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("lesson_attempts").
			Unique().
			Required(),
		edge.From("lesson", Lesson.Type).
			Ref("attempts").
			Unique().
			Field("lesson_id").
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("logs", CardAttemptLog.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (LessonAttempt) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("owner"),
		index.Edges("lesson"),
	}
}

func (LessonAttempt) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
