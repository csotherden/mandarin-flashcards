package study

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardattemptlog"
	"github.com/csotherden/mandarin-flashcards/ent/cardset"
	"github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/lesson"
	"github.com/csotherden/mandarin-flashcards/ent/setflashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/google/uuid"
)

// LessonDeck is every unique card in a lesson, in set order, plus the disabled set.
type LessonDeck struct {
	Cards    []*ent.Flashcard
	Disabled map[uuid.UUID]struct{}
}

func LoadLesson(ctx context.Context, client *ent.Client, userID, lessonID uuid.UUID) (*LessonDeck, error) {
	row, err := client.Lesson.Query().
		Where(lesson.ID(lessonID), lesson.HasOwnerWith(user.ID(userID))).
		WithDisabledCards().
		WithCardSets(func(q *ent.CardSetQuery) {
			q.Order(cardset.ByCreatedAt()).
				WithMemberships(func(mq *ent.SetFlashcardQuery) {
					mq.Order(setflashcard.BySortOrder()).
						WithFlashcard(func(fq *ent.FlashcardQuery) {
							fq.WithCategories().WithSetMemberships()
						})
				})
		}).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	disabled := map[uuid.UUID]struct{}{}
	for _, card := range row.Edges.DisabledCards {
		disabled[card.ID] = struct{}{}
	}
	seen := map[uuid.UUID]struct{}{}
	var cards []*ent.Flashcard
	for _, set := range row.Edges.CardSets {
		for _, membership := range set.Edges.Memberships {
			card := membership.Edges.Flashcard
			if card == nil {
				continue
			}
			if _, ok := seen[card.ID]; ok {
				continue
			}
			seen[card.ID] = struct{}{}
			cards = append(cards, card)
		}
	}
	return &LessonDeck{Cards: cards, Disabled: disabled}, nil
}

func LoadSet(ctx context.Context, client *ent.Client, userID, setID uuid.UUID) ([]*ent.Flashcard, error) {
	memberships, err := client.SetFlashcard.Query().
		Where(
			setflashcard.SetID(setID),
			setflashcard.HasCardSetWith(cardset.HasOwnerWith(user.ID(userID))),
		).
		Order(setflashcard.BySortOrder()).
		WithFlashcard(func(q *ent.FlashcardQuery) {
			q.WithCategories().WithSetMemberships()
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}
	cards := make([]*ent.Flashcard, 0, len(memberships))
	for _, membership := range memberships {
		if membership.Edges.Flashcard != nil {
			cards = append(cards, membership.Edges.Flashcard)
		}
	}
	return cards, nil
}

func LoadOwned(ctx context.Context, client *ent.Client, userID uuid.UUID, ids []uuid.UUID) ([]*ent.Flashcard, error) {
	if len(ids) == 0 {
		return []*ent.Flashcard{}, nil
	}
	rows, err := client.Flashcard.Query().
		Where(flashcard.IDIn(ids...), flashcard.HasOwnerWith(user.ID(userID))).
		WithCategories().
		WithSetMemberships().
		All(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]*ent.Flashcard, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	ordered := make([]*ent.Flashcard, 0, len(ids))
	for _, id := range ids {
		if row, ok := byID[id]; ok {
			ordered = append(ordered, row)
		}
	}
	return ordered, nil
}

// History returns a stat for every id, including cards that have never been practiced.
func History(ctx context.Context, client *ent.Client, userID uuid.UUID, ids []uuid.UUID) ([]PracticeStat, error) {
	stats := make([]PracticeStat, len(ids))
	index := make(map[uuid.UUID]int, len(ids))
	for i, id := range ids {
		stats[i] = PracticeStat{ID: id}
		index[id] = i
	}
	if len(ids) == 0 {
		return stats, nil
	}
	logs, err := client.CardAttemptLog.Query().
		Where(cardattemptlog.UserID(userID), cardattemptlog.FlashcardIDIn(ids...)).
		Order(cardattemptlog.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, err
	}
	type acc struct {
		sum   float64
		count int
		last  time.Time
	}
	totals := map[uuid.UUID]*acc{}
	for _, log := range logs {
		item := totals[log.FlashcardID]
		if item == nil {
			item = &acc{}
			totals[log.FlashcardID] = item
		}
		item.sum += log.OverallScore
		item.count++
		item.last = log.CreatedAt
	}
	for id, item := range totals {
		i, ok := index[id]
		if !ok || item.count == 0 {
			continue
		}
		avg := item.sum / float64(item.count)
		last := item.last
		stats[i].Average = &avg
		stats[i].LastPracticed = &last
	}
	return stats, nil
}
