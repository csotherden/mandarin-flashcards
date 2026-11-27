package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// XPAward is the XP a user has earned for one word on one of their calendar days. Keyed by the
// word (hanzi), not the card, so duplicate or re-created cards don't earn twice. Retries raise
// xp to the best tier reached that day; the difference is what gets added to the user's total.
type XPAward struct {
	ent.Schema
}

func (XPAward) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.String("word").
			NotEmpty().
			MaxLen(64),
		// The user's local date, YYYY-MM-DD, in their saved timezone.
		field.String("day").
			MaxLen(10),
		field.Int("xp").
			Default(0),
		field.String("tier").
			Default(""),
		field.Float("best_score").
			Default(0),
	}
}

func (XPAward) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("xp_awards").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (XPAward) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "word", "day").Unique(),
		index.Fields("user_id", "day"),
	}
}

func (XPAward) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
