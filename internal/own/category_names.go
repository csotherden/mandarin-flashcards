package own

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/category"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/google/uuid"
)

// DefaultCategoryColor is the color a category made from a name gets.
const DefaultCategoryColor = "#C8102E"

// CategoryNames returns the user's categories with these names (matched case-insensitively),
// creating the missing ones, in the order given and without repeats. Blank names are skipped;
// names are trimmed and cut to the 80-character limit.
func CategoryNames(ctx context.Context, client *ent.Client, userID uuid.UUID, names []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(names))
	seen := map[string]bool{}
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if utf8.RuneCountInString(name) > 80 {
			name = string([]rune(name)[:80])
		}
		key := strings.ToLower(name)
		if name == "" || seen[key] {
			continue
		}
		seen[key] = true
		existing, err := client.Category.Query().
			Where(category.NameEqualFold(name), category.HasOwnerWith(user.ID(userID))).
			First(ctx)
		if err == nil {
			out = append(out, existing.ID)
			continue
		}
		if !ent.IsNotFound(err) {
			return nil, err
		}
		created, err := client.Category.Create().
			SetName(name).
			SetColor(DefaultCategoryColor).
			SetOwnerID(userID).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		out = append(out, created.ID)
	}
	return out, nil
}
