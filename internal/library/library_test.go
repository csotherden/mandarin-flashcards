package library

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/enttest"
	"github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/libraryset"
	"github.com/csotherden/mandarin-flashcards/ent/setflashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func TestValidate(t *testing.T) {
	f := File{Sets: []SetSpec{
		{Slug: "a", Title: "A", Cards: []CardSpec{{Hanzi: "老公", TonalPinyin: "lao3"}, {Hanzi: "老公"}, {Hanzi: ""}}},
		{Slug: "a", Title: ""},
	}}
	got := strings.Join(f.Validate(), "\n")
	for _, want := range []string{"one numbered syllable", "duplicate word", "hanzi is required", "duplicate slug", "slug and title are required"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	ok := File{Sets: []SetSpec{{Slug: "a", Title: "A", Cards: []CardSpec{{Hanzi: "老公", TonalPinyin: "lao3 gong1"}}}}}
	if p := ok.Validate(); len(p) != 0 {
		t.Errorf("valid file reported %v", p)
	}
}

// The rest runs against Postgres when TEST_DATABASE_URL is set (a database of its own: other
// packages migrate theirs in parallel).
func testDB(t *testing.T) *ent.Client {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	return enttest.Open(t, "postgres", url)
}

const family = `{"sets": [{
  "slug": "%s", "title": "Family", "description": "The people closest to you", "level": "Beginner",
  "cards": [
    {"hanzi": "妈妈", "pinyin": "māma", "tonal_pinyin": "ma1 ma5", "english": "mom", "image_url": "https://example.com/mom.jpg",
     "unsplash_attribution": {"photographer_name": "A", "photographer_url": "https://u/a", "photo_url": "https://u/p"},
     "categories": ["Family"]},
    {"hanzi": "爸爸", "pinyin": "bàba", "tonal_pinyin": "ba4 ba5", "english": "dad", "categories": ["Family"]},
    {"hanzi": "老公", "pinyin": "lǎogōng", "tonal_pinyin": "lao3 gong1", "english": "husband", "categories": ["Family", "Marriage"]}
  ]}]}`

func load(t *testing.T, db *ent.Client, body string) LoadResult {
	t.Helper()
	res, err := Load(context.Background(), db, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestLoadIsRepeatable(t *testing.T) {
	db := testDB(t)
	slug := "family-" + uuid.NewString()[:8]
	if res := load(t, db, sprintf(family, slug)); res.SetsCreated != 1 || res.CardsCreated != 3 {
		t.Fatalf("first load: %+v", res)
	}
	if res := load(t, db, sprintf(family, slug)); res.SetsUpdated != 1 || res.CardsUpdated != 3 || res.CardsCreated != 0 {
		t.Fatalf("reload: %+v", res)
	}
	// Drop 老公, add 女儿, retitle.
	edited := strings.Replace(sprintf(family, slug), `"title": "Family"`, `"title": "My family"`, 1)
	edited = strings.Replace(edited, `{"hanzi": "老公", "pinyin": "lǎogōng", "tonal_pinyin": "lao3 gong1", "english": "husband", "categories": ["Family", "Marriage"]}`,
		`{"hanzi": "女儿", "pinyin": "nǚ'ér", "tonal_pinyin": "nü3 er2", "english": "daughter"}`, 1)
	if res := load(t, db, edited); res.CardsCreated != 1 || res.CardsUpdated != 2 || res.CardsRemoved != 1 {
		t.Fatalf("edit: %+v", res)
	}
	set := db.LibrarySet.Query().Where(libraryset.Slug(slug)).WithCards().OnlyX(context.Background())
	if set.Title != "My family" || len(set.Edges.Cards) != 3 {
		t.Fatalf("after edit: %q with %d cards", set.Title, len(set.Edges.Cards))
	}
	// A bad file changes nothing.
	if _, err := Load(context.Background(), db, strings.NewReader(`{"sets":[{"slug":"`+slug+`","title":"x","cards":[{"hanzi":"好","tonal_pinyin":"hao3 hao3"}]}]}`)); err == nil {
		t.Fatal("bad tonal pinyin loaded")
	}
}

func TestImport(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	slug := "family-" + uuid.NewString()[:8]
	load(t, db, sprintf(family, slug))
	lib := db.LibrarySet.Query().Where(libraryset.Slug(slug)).OnlyX(ctx)
	u := db.User.Create().SetEmail(uuid.NewString() + "@example.com").SetPasswordHash("x").SetPasswordSalt("x").SaveX(ctx)
	// The user already has 老公, edited their way.
	mine := db.Flashcard.Create().SetOwnerID(u.ID).SetHanzi("老公").SetEnglish("hubby").SaveX(ctx)

	res, err := Import(ctx, db, u.ID, lib.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created || res.NewCards != 2 {
		t.Fatalf("import: created=%v new=%d", res.Created, res.NewCards)
	}
	members := db.SetFlashcard.Query().Where(setflashcard.SetID(res.Set.ID)).Order(ent.Asc(setflashcard.FieldSortOrder)).AllX(ctx)
	if len(members) != 3 || members[2].FlashcardID != mine.ID {
		t.Fatalf("set should hold 3 cards in library order, reusing the user's 老公: %+v", members)
	}
	if got := db.Flashcard.GetX(ctx, mine.ID); got.English != "hubby" {
		t.Fatalf("import overwrote the user's card: %q", got.English)
	}
	mom := db.Flashcard.Query().Where(flashcard.HasOwnerWith(user.ID(u.ID)), flashcard.Hanzi("妈妈")).WithCategories().OnlyX(ctx)
	if mom.TonalPinyin != "ma1 ma5" || mom.ImageURL == "" || mom.UnsplashAttribution == nil || len(mom.Edges.Categories) != 1 || mom.LibraryCardID == nil {
		t.Fatalf("copied card: %+v", mom)
	}

	// Importing again returns the same copy, even after the user changed it.
	db.SetFlashcard.DeleteOneID(members[0].ID).ExecX(ctx)
	again, err := Import(ctx, db, u.ID, lib.ID)
	if err != nil || again.Created || again.Set.ID != res.Set.ID {
		t.Fatalf("re-import: %+v %v", again, err)
	}
	if n := db.SetFlashcard.Query().Where(setflashcard.SetID(res.Set.ID)).CountX(ctx); n != 2 {
		t.Fatalf("re-import changed the user's set: %d cards", n)
	}
	// After deleting their copy, importing makes a fresh one and reuses their cards.
	db.CardSet.DeleteOneID(res.Set.ID).ExecX(ctx)
	fresh, err := Import(ctx, db, u.ID, lib.ID)
	if err != nil || !fresh.Created || fresh.NewCards != 0 {
		t.Fatalf("import after delete: %+v %v", fresh, err)
	}
	if n := db.Flashcard.Query().Where(flashcard.HasOwnerWith(user.ID(u.ID))).CountX(ctx); n != 3 {
		t.Fatalf("user has %d cards, want 3 (no duplicates)", n)
	}
	// Another user's import is entirely their own.
	other := db.User.Create().SetEmail(uuid.NewString() + "@example.com").SetPasswordHash("x").SetPasswordSalt("x").SaveX(ctx)
	theirs, err := Import(ctx, db, other.ID, lib.ID)
	if err != nil || theirs.NewCards != 3 || theirs.Set.ID == fresh.Set.ID {
		t.Fatalf("second user: %+v %v", theirs, err)
	}
}

func sprintf(format, slug string) string { return strings.Replace(format, "%s", slug, 1) }
