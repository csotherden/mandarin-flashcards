// Package tones is the "Intro to tones" tutorial: a fixed sequence of lessons, each a short
// explanation followed by words to say and have graded. The catalog lives in code (it changes
// with releases, not per user); progress and grades are stored per user by lesson and item key.
package tones

// Contour is a tone's pitch path on the five-level scale (5 high, 1 low), as Chao tone letters:
// tone 1 is 55, tone 2 35, tone 3 214, tone 4 51. The neutral tone is short and has no contour
// of its own; it is drawn as a dot at mid height.
type Contour []int

// Item is one thing to say: a syllable or word, read in citation tones.
type Item struct {
	Key     string `json:"key"`
	Hanzi   string `json:"hanzi"`
	Pinyin  string `json:"pinyin"`
	Tonal   string `json:"tonal_pinyin"`
	English string `json:"english"`
}

// Example is a word to listen to while learning, not graded.
type Example struct {
	Hanzi  string `json:"hanzi"`
	Pinyin string `json:"pinyin"`
	Tonal  string `json:"tonal_pinyin"`
	Note   string `json:"note"`
}

// Lesson is one step of the tutorial.
type Lesson struct {
	Key      string    `json:"key"`
	Title    string    `json:"title"`
	Subtitle string    `json:"subtitle"`
	Tone     int       `json:"tone"` // 1-4, 5 for neutral, 0 for mixed lessons
	Contour  Contour   `json:"contour,omitempty"`
	Explain  []string  `json:"explain"` // paragraphs
	Tip      string    `json:"tip"`
	Examples []Example `json:"examples"`
	Items    []Item    `json:"items"`
}

// Lessons is the tutorial, in order.
var Lessons = []Lesson{
	{
		Key: "tone-1", Title: "Tone 1: high and flat", Subtitle: "mā", Tone: 1, Contour: Contour{5, 5},
		Explain: []string{
			"Mandarin uses pitch to tell words apart. The same sound, ma, means mom, hemp, horse or scold depending on its tone.",
			"The first tone is the simplest: start high and stay there, level, like holding a note when you sing. In pinyin it's marked with a flat line: ā.",
		},
		Tip: "Start a little higher than your normal speaking voice and keep it steady to the end. Don't let it drop.",
		Examples: []Example{
			{Hanzi: "妈", Pinyin: "mā", Tonal: "ma1", Note: "mom"},
			{Hanzi: "书", Pinyin: "shū", Tonal: "shu1", Note: "book"},
		},
		Items: []Item{
			{Key: "ma", Hanzi: "妈", Pinyin: "mā", Tonal: "ma1", English: "mom"},
			{Key: "ba", Hanzi: "八", Pinyin: "bā", Tonal: "ba1", English: "eight"},
			{Key: "shu", Hanzi: "书", Pinyin: "shū", Tonal: "shu1", English: "book"},
			{Key: "tian", Hanzi: "天", Pinyin: "tiān", Tonal: "tian1", English: "sky, day"},
			{Key: "he", Hanzi: "喝", Pinyin: "hē", Tonal: "he1", English: "to drink"},
			{Key: "feiji", Hanzi: "飞机", Pinyin: "fēijī", Tonal: "fei1 ji1", English: "airplane"},
		},
	},
	{
		Key: "tone-2", Title: "Tone 2: rising", Subtitle: "má", Tone: 2, Contour: Contour{3, 5},
		Explain: []string{
			"The second tone rises from the middle of your voice to the top, like asking \"what?\" in surprise.",
			"In pinyin it's marked with a mark rising to the right: á.",
		},
		Tip: "Start in your normal speaking voice and push up steadily. Keep it moving the whole way rather than jumping at the end.",
		Examples: []Example{
			{Hanzi: "麻", Pinyin: "má", Tonal: "ma2", Note: "numb, tingly"},
			{Hanzi: "茶", Pinyin: "chá", Tonal: "cha2", Note: "tea"},
		},
		Items: []Item{
			{Key: "ma", Hanzi: "麻", Pinyin: "má", Tonal: "ma2", English: "numb, tingly"},
			{Key: "cha", Hanzi: "茶", Pinyin: "chá", Tonal: "cha2", English: "tea"},
			{Key: "ren", Hanzi: "人", Pinyin: "rén", Tonal: "ren2", English: "person"},
			{Key: "lai", Hanzi: "来", Pinyin: "lái", Tonal: "lai2", English: "to come"},
			{Key: "xue", Hanzi: "学", Pinyin: "xué", Tonal: "xue2", English: "to study"},
			{Key: "yinhang", Hanzi: "银行", Pinyin: "yínháng", Tonal: "yin2 hang2", English: "bank"},
		},
	},
	{
		Key: "tone-3", Title: "Tone 3: low dip", Subtitle: "mǎ", Tone: 3, Contour: Contour{2, 1, 4},
		Explain: []string{
			"The third tone goes low. Said on its own it dips down to the bottom of your voice and comes back up a little. Marked with a v-shaped mark: ǎ.",
			"In real speech the low part matters most. Before another syllable it usually just stays low, and two third tones in a row change: the first becomes a rising tone, so 你好 nǐ hǎo sounds like ní hǎo.",
		},
		Tip: "Think low and relaxed, almost a creak at the bottom of your voice. The dip down is what people listen for.",
		Examples: []Example{
			{Hanzi: "马", Pinyin: "mǎ", Tonal: "ma3", Note: "horse"},
			{Hanzi: "你好", Pinyin: "nǐ hǎo", Tonal: "ni3 hao3", Note: "hello: the first syllable rises"},
		},
		Items: []Item{
			{Key: "ma", Hanzi: "马", Pinyin: "mǎ", Tonal: "ma3", English: "horse"},
			{Key: "wo", Hanzi: "我", Pinyin: "wǒ", Tonal: "wo3", English: "I, me"},
			{Key: "hao", Hanzi: "好", Pinyin: "hǎo", Tonal: "hao3", English: "good"},
			{Key: "shui", Hanzi: "水", Pinyin: "shuǐ", Tonal: "shui3", English: "water"},
			{Key: "mai", Hanzi: "买", Pinyin: "mǎi", Tonal: "mai3", English: "to buy"},
			{Key: "beijing", Hanzi: "北京", Pinyin: "Běijīng", Tonal: "bei3 jing1", English: "Beijing"},
		},
	},
	{
		Key: "tone-4", Title: "Tone 4: sharp fall", Subtitle: "mà", Tone: 4, Contour: Contour{5, 1},
		Explain: []string{
			"The fourth tone starts high and drops fast, firm and short, like saying \"No!\" or \"Stop!\".",
			"Marked with a mark falling to the right: à.",
		},
		Tip: "Commit to it: start at the top and fall all the way. A gentle fall can sound like a different tone.",
		Examples: []Example{
			{Hanzi: "骂", Pinyin: "mà", Tonal: "ma4", Note: "to scold"},
			{Hanzi: "是", Pinyin: "shì", Tonal: "shi4", Note: "to be, yes"},
		},
		Items: []Item{
			{Key: "ma", Hanzi: "骂", Pinyin: "mà", Tonal: "ma4", English: "to scold"},
			{Key: "shi", Hanzi: "是", Pinyin: "shì", Tonal: "shi4", English: "to be, yes"},
			{Key: "qu", Hanzi: "去", Pinyin: "qù", Tonal: "qu4", English: "to go"},
			{Key: "da", Hanzi: "大", Pinyin: "dà", Tonal: "da4", English: "big"},
			{Key: "ai", Hanzi: "爱", Pinyin: "ài", Tonal: "ai4", English: "to love"},
			{Key: "zaijian", Hanzi: "再见", Pinyin: "zàijiàn", Tonal: "zai4 jian4", English: "goodbye"},
		},
	},
	{
		Key: "neutral", Title: "The neutral tone", Subtitle: "ma", Tone: 5,
		Explain: []string{
			"Some syllables are said lightly and quickly, with no tone of their own. Pinyin leaves them unmarked: the second ma in māma.",
			"They mostly come at the end of words, and their pitch just follows the syllable before. Saying them too strongly is a common giveaway of a learner.",
		},
		Tip: "Give the first syllable its full tone, then let the second one fall away, short and soft.",
		Examples: []Example{
			{Hanzi: "妈妈", Pinyin: "māma", Tonal: "ma1 ma5", Note: "mom"},
			{Hanzi: "谢谢", Pinyin: "xièxie", Tonal: "xie4 xie5", Note: "thank you"},
		},
		Items: []Item{
			{Key: "mama", Hanzi: "妈妈", Pinyin: "māma", Tonal: "ma1 ma5", English: "mom"},
			{Key: "baba", Hanzi: "爸爸", Pinyin: "bàba", Tonal: "ba4 ba5", English: "dad"},
			{Key: "xiexie", Hanzi: "谢谢", Pinyin: "xièxie", Tonal: "xie4 xie5", English: "thank you"},
			{Key: "pengyou", Hanzi: "朋友", Pinyin: "péngyou", Tonal: "peng2 you5", English: "friend"},
			{Key: "zhuozi", Hanzi: "桌子", Pinyin: "zhuōzi", Tonal: "zhuo1 zi5", English: "table"},
			{Key: "nide", Hanzi: "你的", Pinyin: "nǐ de", Tonal: "ni3 de5", English: "yours"},
		},
	},
	{
		Key: "tell-apart", Title: "Telling them apart", Subtitle: "mā má mǎ mà", Tone: 0,
		Explain: []string{
			"Now mix them. Each pair here sounds the same except for the tone, and the tone is the whole difference in meaning.",
			"Listen to each one first if you like, then say it. Watch which tone the grader hears.",
		},
		Tip: "Exaggerate at first. Big, clear contours are easier to fix than timid ones.",
		Examples: []Example{
			{Hanzi: "买", Pinyin: "mǎi", Tonal: "mai3", Note: "to buy"},
			{Hanzi: "卖", Pinyin: "mài", Tonal: "mai4", Note: "to sell"},
		},
		Items: []Item{
			{Key: "ma3", Hanzi: "马", Pinyin: "mǎ", Tonal: "ma3", English: "horse"},
			{Key: "ma1", Hanzi: "妈", Pinyin: "mā", Tonal: "ma1", English: "mom"},
			{Key: "ma4", Hanzi: "骂", Pinyin: "mà", Tonal: "ma4", English: "to scold"},
			{Key: "ma2", Hanzi: "麻", Pinyin: "má", Tonal: "ma2", English: "numb, tingly"},
			{Key: "mai4", Hanzi: "卖", Pinyin: "mài", Tonal: "mai4", English: "to sell"},
			{Key: "mai3", Hanzi: "买", Pinyin: "mǎi", Tonal: "mai3", English: "to buy"},
			{Key: "tang2", Hanzi: "糖", Pinyin: "táng", Tonal: "tang2", English: "sugar, candy"},
			{Key: "tang1", Hanzi: "汤", Pinyin: "tāng", Tonal: "tang1", English: "soup"},
		},
	},
	{
		Key: "pairs", Title: "Tones in pairs", Subtitle: "Zhōngguó", Tone: 0,
		Explain: []string{
			"Most everyday words have two syllables, so the real skill is moving from one tone to the next without losing either.",
			"Each word here combines two different tones. Say it as one word, not two separate syllables.",
		},
		Tip: "Say it slowly first, then at normal speed. Keep both contours clear at normal speed.",
		Examples: []Example{
			{Hanzi: "中国", Pinyin: "Zhōngguó", Tonal: "zhong1 guo2", Note: "China: high, then rising"},
			{Hanzi: "电视", Pinyin: "diànshì", Tonal: "dian4 shi4", Note: "TV: two falls"},
		},
		Items: []Item{
			{Key: "zhongguo", Hanzi: "中国", Pinyin: "Zhōngguó", Tonal: "zhong1 guo2", English: "China"},
			{Key: "xiongmao", Hanzi: "熊猫", Pinyin: "xióngmāo", Tonal: "xiong2 mao1", English: "panda"},
			{Key: "tianqi", Hanzi: "天气", Pinyin: "tiānqì", Tonal: "tian1 qi4", English: "weather"},
			{Key: "xuexiao", Hanzi: "学校", Pinyin: "xuéxiào", Tonal: "xue2 xiao4", English: "school"},
			{Key: "mifan", Hanzi: "米饭", Pinyin: "mǐfàn", Tonal: "mi3 fan4", English: "cooked rice"},
			{Key: "laoshi", Hanzi: "老师", Pinyin: "lǎoshī", Tonal: "lao3 shi1", English: "teacher"},
			{Key: "shanghai", Hanzi: "上海", Pinyin: "Shànghǎi", Tonal: "shang4 hai3", English: "Shanghai"},
			{Key: "dianshi", Hanzi: "电视", Pinyin: "diànshì", Tonal: "dian4 shi4", English: "TV"},
		},
	},
}

// Find returns a lesson and, if itemKey is non-empty, one of its items.
func Find(lessonKey, itemKey string) (*Lesson, *Item) {
	for i := range Lessons {
		l := &Lessons[i]
		if l.Key != lessonKey {
			continue
		}
		if itemKey == "" {
			return l, nil
		}
		for j := range l.Items {
			if l.Items[j].Key == itemKey {
				return l, &l.Items[j]
			}
		}
		return l, nil
	}
	return nil, nil
}
