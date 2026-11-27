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
  practiced_today: number;
  total_cards: number;
  suggested_count: number;
  last_mode?: StudyMode;
  sets: HomeSet[];
  slipped: { card: Flashcard; average: number }[];
  saved_sessions: { id: string; title: string; card_count: number }[];
};

export type GeneratedSet = { set: CardSet; cards: Flashcard[]; skipped: string[] };

/** Where a session's cards come from. */
export type SessionSource =
  | { kind: "all" }
  | { kind: "sets"; setIds: string[] }
  | { kind: "lesson"; lessonId: string }
  | { kind: "missed"; attemptId: string };

