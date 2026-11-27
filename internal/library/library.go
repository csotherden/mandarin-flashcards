// Package library is the global card library: curated sets owned by no one that users browse
// and import. Importing hydrates a library set into the user's own account (a CardSet and
// Flashcards they own outright, free to edit, extend or delete); the library itself never
// changes because of what users do with their copies.
package library

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardset"
	"github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/librarycard"
	"github.com/csotherden/mandarin-flashcards/ent/libraryset"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/own"
	"github.com/csotherden/mandarin-flashcards/internal/pinyin"
	"github.com/google/uuid"
)

// File is the JSON the loader reads: any number of sets, each matched to the database by slug.
type File struct {
	Sets []SetSpec `json:"sets"`
}

type SetSpec struct {
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Level       string     `json:"level"`
	SortOrder   int        `json:"sort_order"`
	Published   *bool      `json:"published"` // default true
	Cards       []CardSpec `json:"cards"`
}

type CardSpec struct {
	Hanzi                 string                      `json:"hanzi"`
	Pinyin                string                      `json:"pinyin"`
	TonalPinyin           string                      `json:"tonal_pinyin"`
	English               string                      `json:"english"`
	ExampleSentenceHanzi  string                      `json:"example_sentence_hanzi"`
	ExampleSentencePinyin string                      `json:"example_sentence_pinyin"`
	ImageURL              string                      `json:"image_url"`
	UnsplashAttribution   *models.UnsplashAttribution `json:"unsplash_attribution"`
	Categories            []string                    `json:"categories"`
}

// Validate checks a file before anything is written, returning every problem found.
func (f File) Validate() []string {
	var problems []string
	slugs := map[string]bool{}
	for i, s := range f.Sets {
		where := fmt.Sprintf("set %d (%s)", i+1, s.Slug)
		if strings.TrimSpace(s.Slug) == "" || strings.TrimSpace(s.Title) == "" {
			problems = append(problems, where+": slug and title are required")
		}
		if slugs[s.Slug] {
			problems = append(problems, where+": duplicate slug")
		}
		slugs[s.Slug] = true
		seen := map[string]bool{}
		for j, c := range s.Cards {
			at := fmt.Sprintf("%s card %d (%s)", where, j+1, c.Hanzi)
			switch {
			case strings.TrimSpace(c.Hanzi) == "":
				problems = append(problems, at+": hanzi is required")
			case seen[c.Hanzi]:
				problems = append(problems, at+": duplicate word in this set")
			case c.TonalPinyin != "" && !pinyin.ValidTonal(c.Hanzi, c.TonalPinyin):
				problems = append(problems, at+": tonal_pinyin needs one numbered syllable per character")
			}
			seen[c.Hanzi] = true
		}
	}
	return problems
}

// LoadResult summarizes a load.
type LoadResult struct {
	SetsCreated, SetsUpdated                 int
	CardsCreated, CardsUpdated, CardsRemoved int
	// NewPhotos are the Unsplash download locations of photos this load put into the library
	// (new words, or words whose photo changed). The loader notifies Unsplash of each once;
	// user imports don't, since copying a library card isn't choosing a photo.
	NewPhotos []string
}

func (r *LoadResult) notePhoto(c CardSpec) {
	if c.ImageURL == "" || c.UnsplashAttribution == nil || c.UnsplashAttribution.DownloadLocation == "" {
		return
	}
	for _, seen := range r.NewPhotos {
		if seen == c.UnsplashAttribution.DownloadLocation {
			return
		}
	}
	r.NewPhotos = append(r.NewPhotos, c.UnsplashAttribution.DownloadLocation)
}

// Load reads a File and makes the library match it, set by set: sets are created or updated by
// slug, and each listed set's cards are made to match the file (matched by hanzi, so a card's id
// survives edits). Sets not in the file are left alone. All or nothing.
func Load(ctx context.Context, db *ent.Client, r io.Reader) (LoadResult, error) {
	var file File
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&file); err != nil {
		return LoadResult{}, fmt.Errorf("read library file: %w", err)
	}
	if problems := file.Validate(); len(problems) > 0 {
		return LoadResult{}, fmt.Errorf("library file has problems:\n  %s", strings.Join(problems, "\n  "))
	}
	tx, err := db.Tx(ctx)
	if err != nil {
		return LoadResult{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var res LoadResult
	for _, spec := range file.Sets {
		if err := loadSet(ctx, tx, spec, &res); err != nil {
			return LoadResult{}, fmt.Errorf("set %s: %w", spec.Slug, err)
		}
	}
	return res, tx.Commit()
}

func loadSet(ctx context.Context, tx *ent.Tx, spec SetSpec, res *LoadResult) error {
	published := spec.Published == nil || *spec.Published
	set, err := tx.LibrarySet.Query().Where(libraryset.Slug(spec.Slug)).Only(ctx)
	switch {
	case ent.IsNotFound(err):
		set, err = tx.LibrarySet.Create().
			SetSlug(spec.Slug).SetTitle(spec.Title).SetDescription(spec.Description).
			SetLevel(spec.Level).SetSortOrder(spec.SortOrder).SetPublished(published).
			Save(ctx)
		if err != nil {
			return err
		}
		res.SetsCreated++
	case err != nil:
		return err
	default:
		set, err = set.Update().
			SetTitle(spec.Title).SetDescription(spec.Description).
			SetLevel(spec.Level).SetSortOrder(spec.SortOrder).SetPublished(published).
			Save(ctx)
		if err != nil {
			return err
		}
		res.SetsUpdated++
	}

	existing, err := tx.LibraryCard.Query().Where(librarycard.SetID(set.ID)).All(ctx)
	if err != nil {
		return err
	}
	byHanzi := map[string]*ent.LibraryCard{}
	for _, c := range existing {
		byHanzi[c.Hanzi] = c
	}
	for i, c := range spec.Cards {
		if prior, ok := byHanzi[c.Hanzi]; ok {
			delete(byHanzi, c.Hanzi)
			if prior.ImageURL != c.ImageURL {
				res.notePhoto(c)
			}
			update := prior.Update().
				SetSortOrder(i).SetPinyin(c.Pinyin).SetTonalPinyin(pinyin.NormalizeTonal(c.TonalPinyin)).
				SetEnglish(c.English).SetExampleSentenceHanzi(c.ExampleSentenceHanzi).
				SetExampleSentencePinyin(c.ExampleSentencePinyin).SetImageURL(c.ImageURL).
				SetCategories(c.Categories)
			if c.UnsplashAttribution != nil {
				update.SetUnsplashAttribution(c.UnsplashAttribution)
			} else {
				update.ClearUnsplashAttribution()
			}
			if err := update.Exec(ctx); err != nil {
				return err
			}
			res.CardsUpdated++
			continue
		}
		create := tx.LibraryCard.Create().
			SetSetID(set.ID).SetSortOrder(i).SetHanzi(c.Hanzi).SetPinyin(c.Pinyin).
			SetTonalPinyin(pinyin.NormalizeTonal(c.TonalPinyin)).SetEnglish(c.English).
			SetExampleSentenceHanzi(c.ExampleSentenceHanzi).SetExampleSentencePinyin(c.ExampleSentencePinyin).
			SetImageURL(c.ImageURL).SetCategories(c.Categories)
		if c.UnsplashAttribution != nil {
			create.SetUnsplashAttribution(c.UnsplashAttribution)
		}
		if err := create.Exec(ctx); err != nil {
			return err
		}
		res.notePhoto(c)
		res.CardsCreated++
	}
	for _, gone := range byHanzi {
		if err := tx.LibraryCard.DeleteOne(gone).Exec(ctx); err != nil {
			return err
		}
		res.CardsRemoved++
	}
	return nil
}

// Imported returns the user's copy of a library set, or nil if they don't have one.
func Imported(ctx context.Context, db *ent.Client, userID, librarySetID uuid.UUID) (*ent.CardSet, error) {
	set, err := db.CardSet.Query().
		Where(cardset.HasOwnerWith(user.ID(userID)), cardset.LibrarySetID(librarySetID)).
		Order(ent.Desc(cardset.FieldCreatedAt)).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return set, err
}

// ImportResult is the user's copy of a library set.
type ImportResult struct {
	Set *ent.CardSet
	// Created is false when the user already had a copy (which is returned untouched).
	Created bool
	// NewCards counts cards created; the rest were words the user already had, reused as-is.
	NewCards int
}

// Import hydrates a library set into the user's account. If they already have a copy, that
// copy is returned unchanged (they may have edited it). Otherwise a new set is made: each word
// they already have is reused (their card, their edits), every other word becomes a new card
// of theirs. The user's row is locked so a double tap imports once.
func Import(ctx context.Context, db *ent.Client, userID, librarySetID uuid.UUID) (ImportResult, error) {
	tx, err := db.Tx(ctx)
	if err != nil {
		return ImportResult{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.User.Query().Where(user.ID(userID)).ForUpdate().Only(ctx); err != nil {
		return ImportResult{}, fmt.Errorf("lock user: %w", err)
	}
	if existing, err := Imported(ctx, tx.Client(), userID, librarySetID); err != nil || existing != nil {
		return ImportResult{Set: existing}, err
	}
	lib, err := tx.LibrarySet.Query().
		Where(libraryset.ID(librarySetID), libraryset.Published(true)).
		WithCards(func(q *ent.LibraryCardQuery) { q.Order(ent.Asc(librarycard.FieldSortOrder)) }).
		Only(ctx)
	if err != nil {
		return ImportResult{}, err
	}
	set, err := tx.CardSet.Create().
		SetOwnerID(userID).SetTitle(lib.Title).SetDescription(lib.Description).SetLibrarySetID(lib.ID).
		Save(ctx)
	if err != nil {
		return ImportResult{}, fmt.Errorf("create set: %w", err)
	}
	res := ImportResult{Set: set, Created: true}
	for i, c := range lib.Edges.Cards {
		card, err := tx.Flashcard.Query().
			Where(flashcard.HasOwnerWith(user.ID(userID)), flashcard.Hanzi(c.Hanzi)).
			Order(ent.Asc(flashcard.FieldCreatedAt)).
			First(ctx)
		if ent.IsNotFound(err) {
			card, err = copyCard(ctx, tx, userID, c)
			res.NewCards++
		}
		if err != nil {
			return ImportResult{}, fmt.Errorf("card %s: %w", c.Hanzi, err)
		}
		err = tx.SetFlashcard.Create().SetSetID(set.ID).SetFlashcardID(card.ID).SetSortOrder(i).Exec(ctx)
		if err != nil {
			return ImportResult{}, fmt.Errorf("add %s to set: %w", c.Hanzi, err)
		}
	}
	return res, tx.Commit()
}

func copyCard(ctx context.Context, tx *ent.Tx, userID uuid.UUID, c *ent.LibraryCard) (*ent.Flashcard, error) {
	create := tx.Flashcard.Create().
		SetOwnerID(userID).SetLibraryCardID(c.ID).
		SetHanzi(c.Hanzi).SetPinyin(c.Pinyin).SetTonalPinyin(c.TonalPinyin).SetEnglish(c.English).
		SetExampleSentenceHanzi(c.ExampleSentenceHanzi).SetExampleSentencePinyin(c.ExampleSentencePinyin).
		SetImageURL(c.ImageURL)
	if c.UnsplashAttribution != nil && c.ImageURL != "" {
		create.SetUnsplashAttribution(c.UnsplashAttribution)
	}
	if len(c.Categories) > 0 {
		ids, err := own.CategoryNames(ctx, tx.Client(), userID, c.Categories)
		if err != nil {
			return nil, err
		}
		create.AddCategoryIDs(ids...)
	}
	return create.Save(ctx)
}
