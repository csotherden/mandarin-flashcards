package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// SetFlashcard is the ordered membership of a card in a set.
type SetFlashcard struct {
	ent.Schema
}

func (SetFlashcard) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "set_flashcards"}}
}

func (SetFlashcard) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("set_id", uuid.UUID{}),
		field.UUID("flashcard_id", uuid.UUID{}),
		field.Int("sort_order").Default(0),
	}
}

func (SetFlashcard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("card_set", CardSet.Type).
			Ref("memberships").
			Unique().
			Required().
			Field("set_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("flashcard", Flashcard.Type).
			Ref("set_memberships").
			Unique().
			Required().
			Field("flashcard_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (SetFlashcard) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("set_id", "flashcard_id").Unique(),
		index.Fields("set_id", "sort_order"),
	}
}

func (SetFlashcard) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
