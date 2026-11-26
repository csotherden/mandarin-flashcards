package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/google/uuid"
)

// User is an account. Passwords are argon2id with a unique salt stored on the row.
// Sessions are local JWTs; this build does not delegate identity to Cognito.
type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("email").
			Unique().
			MaxLen(255).
			NotEmpty(),
		field.String("password_hash").
			Sensitive().
			NotEmpty(),
		field.String("password_salt").
			Sensitive().
			NotEmpty(),
		field.String("refresh_token_id").
			Optional().
			Sensitive().
			Nillable(),
		field.JSON("preferences", models.Preferences{}).
			Default(models.Preferences{
				DefaultTTSSpeed: 1,
				PreferredVoice:  "zh-CN",
			}),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("categories", Category.Type).
			StorageKey(edge.Column("user_id")).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("flashcards", Flashcard.Type).
			StorageKey(edge.Column("user_id")).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("card_sets", CardSet.Type).
			StorageKey(edge.Column("user_id")).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("lessons", Lesson.Type).
			StorageKey(edge.Column("user_id")).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("lesson_attempts", LessonAttempt.Type).
			StorageKey(edge.Column("user_id")).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("card_attempt_logs", CardAttemptLog.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}
