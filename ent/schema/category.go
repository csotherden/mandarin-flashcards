package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Category is a user-owned tag such as "Food & Dining".
type Category struct {
	ent.Schema
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(80),
		field.String("color").
			Default("#9f2d20").
			MaxLen(7),
	}
}

func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("categories").
			Unique().
			Required(),
		edge.From("flashcards", Flashcard.Type).
			Ref("categories"),
	}
}

func (Category) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Edges("owner").Unique(),
	}
}

func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
