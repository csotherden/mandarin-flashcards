export type Preferences = {
  default_tts_speed: number;
  preferred_voice: string;
};

export type User = {
  id: string;
  email: string;
  preferences: Preferences;
  timezone: string;
  created_at: string;
};

export type LoginResponse = {
  access_token: string;
  refresh_token: string;
  user: User;
};

export type Category = {
  id: string;
  name: string;
  color: string;
  created_at: string;
};

export type Attribution = {
  photographer_name: string;
  photographer_url: string;
  photo_url: string;
  download_location?: string;
};

export type Flashcard = {
  id: string;
  hanzi: string;
  pinyin: string;
  tonal_pinyin: string;
  english: string;
  example_sentence_hanzi: string;
  example_sentence_pinyin: string;
  image_url: string;
  has_local_image: boolean;
  unsplash_attribution?: Attribution;
  category_ids: string[];
  set_ids: string[];
  created_at: string;
};

export type Autofill = {
  hanzi: string;
  pinyin: string;
  tonal_pinyin: string;
  english: string;
  example_sentence_hanzi: string;
  example_sentence_pinyin: string;
  image_search: string;
  categories: string[];
  existing_card_id?: string;
};

export type ImageCandidate = {
  id: string;
  thumb_url: string;
  full_url: string;
  photographer_name: string;
  photographer_url: string;
  photo_url: string;
  download_location: string;
};

export type CardSet = {
  id: string;
  title: string;
  description: string;
  card_count: number;
  created_at: string;
};

export type SetDetail = CardSet & { cards: Flashcard[] };

export type Lesson = {
  id: string;
  title: string;
  set_ids: string[];
  card_count: number;
  enabled_count: number;
  created_at: string;
};

export type LessonCard = { card: Flashcard; enabled: boolean };

export type LessonDetail = Lesson & { sets: CardSet[]; cards: LessonCard[] };

export type StudyMode = "image_speak" | "hanzi_speak" | "picture_hanzi_match";

export type Syllable = {
  character: string;
  pinyin: string;
  score: number;
  expected_tone: string;
  heard_tone: string;
  status: "ok" | "warn" | "mismatch" | string;
  detail?: string;
};

export type CardResult = {
  flashcard_id: string;
  overall_score: number;
  tone_accuracy_score?: number;
  coaching_tip?: string;
  syllables?: Syllable[];
  /** XP this grade added (0 when the word already reached this tier today) and the tier reached. */
  xp_earned: number;
  xp_tier: XPTier;
  created_at: string;
};

export type Attempt = {
  id: string;
  lesson_id?: string;
  mode: StudyMode;
  card_count_requested: number;
  card_ids: string[];
  overall_score?: number;
  started_at: string;
  completed_at?: string;
};

export type Session = {
  attempt: Attempt;
  cards: Flashcard[];
  results: CardResult[];
  overall_score?: number;
  /** Every take's XP in this session, retries included. */
  xp_earned: number;
};

export type XPTier = "" | "keep" | "close" | "nice" | "perfect";

export type AssessResponse = {
  result: CardResult;
  assessable: boolean;
  reason?: string;
  xp_total: number;
};

export type HomeSet = { set: CardSet; cover?: Flashcard };

export type Home = {
  /** The live streak; streak_today says whether today already counts. */
  streak_days: number;
  streak_today: boolean;
  streak_longest: number;
  xp_total: number;
  xp_today: number;
  tones_completed: number;
  tones_total: number;
  /** Next unfinished tones lesson; empty once the tutorial is done. */
  tones_next: string;
  tones_next_title: string;
  practiced_today: number;
  total_cards: number;
  suggested_count: number;
  last_mode?: StudyMode;
  sets: HomeSet[];
  slipped: { card: Flashcard; average: number }[];
  saved_sessions: { id: string; title: string; card_count: number }[];
  /** The latest unfinished session from the last week: done of total cards graded. */
  open_session?: { id: string; mode: StudyMode; done: number; total: number; started_at: string };
};

export type GeneratedSet = { set: CardSet; cards: Flashcard[]; skipped: string[] };

/** Where a session's cards come from. */
export type SessionSource =
  | { kind: "all" }
  | { kind: "sets"; setIds: string[] }
  | { kind: "lesson"; lessonId: string }
  | { kind: "missed"; attemptId: string };


/** A set in the global library. imported_set_id is the viewer's own copy, if they have one. */
export type LibrarySet = {
  id: string;
  slug: string;
  title: string;
  description: string;
  level: string;
  card_count: number;
  cover_url: string;
  preview: string[];
  imported_set_id?: string;
};

export type LibraryCard = {
  id: string;
  hanzi: string;
  pinyin: string;
  tonal_pinyin: string;
  english: string;
  example_sentence_hanzi: string;
  example_sentence_pinyin: string;
  image_url: string;
  unsplash_attribution?: Attribution;
  /** The viewer already has this word; importing reuses their card. */
  owned: boolean;
};

export type LibrarySetDetail = LibrarySet & { cards: LibraryCard[] };

export type LibraryImport = { set: CardSet; created: boolean; new_cards: number };

export type ToneItem = { key: string; hanzi: string; pinyin: string; tonal_pinyin: string; english: string };

export type ToneLesson = {
  key: string;
  title: string;
  subtitle: string;
  /** 1-4, 5 for neutral, 0 for mixed lessons. */
  tone: number;
  /** Pitch path on the 1-5 scale (Chao tone letters), e.g. [2, 1, 4] for tone 3. */
  contour?: number[];
  explain: string[];
  tip: string;
  examples: { hanzi: string; pinyin: string; tonal_pinyin: string; note: string }[];
  items: ToneItem[];
  completed: boolean;
  completed_at?: string;
  best_score: number;
};

export type TonesOverview = { lessons: ToneLesson[]; completed: number; next: string };

export type ToneAssess = {
  overall_score: number;
  assessable: boolean;
  reason?: string;
  syllables: Syllable[];
  xp_earned: number;
  xp_tier: XPTier;
  /** True until the lesson is first finished; after that takes are practice (no XP). */
  first_pass: boolean;
  xp_total: number;
};

export type ToneComplete = { first_time: boolean; best_score: number; xp_earned: number; next: string };
