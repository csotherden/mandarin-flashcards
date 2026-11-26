package own

import (
	"context"
	"errors"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardset"
	"github.com/csotherden/mandarin-flashcards/ent/category"
	"github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/google/uuid"
)

// ErrNotFound means one of the requested rows is missing or belongs to someone else.
var ErrNotFound = errors.New("not found")

func Unique(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return []uuid.UUID{}
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func Categories(ctx context.Context, client *ent.Client, userID uuid.UUID, ids []uuid.UUID) error {
	return sameOwnerCount(len(ids), func() (int, error) {
		if len(ids) == 0 {
			return 0, nil
		}
		return client.Category.Query().
			Where(category.IDIn(ids...), category.HasOwnerWith(user.ID(userID))).
			Count(ctx)
	})
}

func CardSets(ctx context.Context, client *ent.Client, userID uuid.UUID, ids []uuid.UUID) error {
	return sameOwnerCount(len(ids), func() (int, error) {
		if len(ids) == 0 {
			return 0, nil
		}
		return client.CardSet.Query().
			Where(cardset.IDIn(ids...), cardset.HasOwnerWith(user.ID(userID))).
			Count(ctx)
	})
}

func Flashcards(ctx context.Context, client *ent.Client, userID uuid.UUID, ids []uuid.UUID) error {
	return sameOwnerCount(len(ids), func() (int, error) {
		if len(ids) == 0 {
			return 0, nil
		}
		return client.Flashcard.Query().
			Where(flashcard.IDIn(ids...), flashcard.HasOwnerWith(user.ID(userID))).
			Count(ctx)
	})
}

func sameOwnerCount(want int, count func() (int, error)) error {
	if want == 0 {
		return nil
	}
	got, err := count()
	if err != nil {
		return err
	}
	if got != want {
		return ErrNotFound
	}
	return nil
}
