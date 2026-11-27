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

// CardSet is a curated deck. The table name stays "sets" as in the product schema.
type CardSet struct {
	ent.Schema
}

func (CardSet) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "sets"}}
}

func (CardSet) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			NotEmpty().
			MaxLen(120),
		field.String("description").
			Default("").
			MaxLen(500),
		// The library set this was imported from, if any. A plain id, not an edge: the copy is
		// the user's own and outlives changes to the library.
		field.UUID("library_set_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (CardSet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("card_sets").
			Unique().
			Required(),
		edge.To("memberships", SetFlashcard.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("lessons", Lesson.Type).
			Ref("card_sets"),
	}
}

func (CardSet) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("owner"),
	}
}

func (CardSet) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
