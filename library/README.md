# Global library content

Curated sets that every user can browse and import (see `internal/library`). Each file here is
loader JSON: a list of sets, matched to the database by `slug`.

Load or update a file on a running deployment (safe to rerun; sets not in the file are left alone):

```sh
docker exec -i mandarin-api /app/api-server -library-load - < library/starter.json
```

The loader checks the whole file first (one numbered syllable per character in `tonal_pinyin`, no
duplicate words or slugs) and writes nothing if anything is wrong. It then notifies Unsplash once
for each photo that is new to the library, as their API guidelines ask. To load into a test copy
without notifying Unsplash, blank the key for that run: `docker exec -i -e UNSPLASH_ACCESS_KEY= ...`.

Photos are Unsplash images; keep `unsplash_attribution` (including `download_location`) with each.

## starter.json

Ten beginner sets of ten everyday, picture-friendly words, no word repeated across sets: Family,
Food & Drink, Fruit, Animals, At Home, Clothes, Getting Around, Weather & Nature, The Body, Around
Town. Tonal pinyin uses citation tones (5 = neutral) and was cross-checked against the marked pinyin.
