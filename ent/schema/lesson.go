package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Lesson is a saved study configuration: one or more sets, with cards that can be toggled off.
type Lesson struct {
	ent.Schema
}

func (Lesson) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			NotEmpty().
			MaxLen(120),
	}
}

func (Lesson) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("lessons").
			Unique().
			Required(),
		edge.To("card_sets", CardSet.Type).
			StorageKey(edge.Table("lesson_sets"), edge.Columns("lesson_id", "set_id")),
		edge.To("disabled_cards", Flashcard.Type).
			StorageKey(edge.Table("lesson_disabled_cards"), edge.Columns("lesson_id", "flashcard_id")),
		edge.To("attempts", LessonAttempt.Type),
	}
}

func (Lesson) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("owner"),
	}
}

func (Lesson) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
