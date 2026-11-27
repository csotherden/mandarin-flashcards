// Tone-colored pinyin: drawn from the numbered form (lao3 gong1) so every syllable knows its tone.
const marks: Record<string, string[]> = {
  a: ["ā", "á", "ǎ", "à"],
  e: ["ē", "é", "ě", "è"],
  i: ["ī", "í", "ǐ", "ì"],
  o: ["ō", "ó", "ǒ", "ò"],
  u: ["ū", "ú", "ǔ", "ù"],
  ü: ["ǖ", "ǘ", "ǚ", "ǜ"],
};

const toneClass = ["text-tone1", "text-tone2", "text-tone3", "text-tone4", "text-tone5"];

/** Turns one numbered syllable ("lao3", "nü3", "nv3", "ma5") into a tone-marked one and its tone (1-5). */
export function markSyllable(numbered: string): { text: string; tone: number } {
  const match = /^([a-zü:v]+)([1-5])$/i.exec(numbered.trim());
  if (!match) return { text: numbered, tone: 5 };
  let base = match[1].toLowerCase().replace(/u:|v/g, "ü");
  const tone = Number(match[2]);
  if (tone === 5) return { text: base, tone };
  // Standard placement: a or e takes the mark; in "ou" the o does; otherwise the last vowel.
  let index = base.search(/[ae]/);
  if (index === -1) index = base.indexOf("ou");
  if (index === -1) {
    for (let i = base.length - 1; i >= 0; i -= 1) {
      if ("iouü".includes(base[i])) {
        index = i;
        break;
      }
    }
  }
  if (index === -1) return { text: base, tone };
  base = base.slice(0, index) + marks[base[index]][tone - 1] + base.slice(index + 1);
  return { text: base, tone };
}

export function Pinyin({
  tonal,
  plain,
  className = "",
  spaced = false,
  colored = true,
}: {
  tonal?: string;
  plain?: string;
  className?: string;
  spaced?: boolean;
  colored?: boolean;
}) {
  const syllables = (tonal ?? "").trim().split(/\s+/).filter(Boolean);
  if (syllables.length === 0) return <span className={className}>{plain}</span>;
  return (
    <span className={className}>
      {syllables.map((syllable, index) => {
        const { text, tone } = markSyllable(syllable);
        return (
          <span key={`${syllable}-${index}`} className={colored ? toneClass[tone - 1] : undefined}>
            {spaced && index > 0 ? " " : ""}
            {text}
          </span>
        );
      })}
    </span>
  );
}

export function toneLabel(tone: string) {
  return { "1": "1st", "2": "2nd", "3": "3rd", "4": "4th", "5": "neutral" }[tone] ?? tone;
}

export function toneColor(tone: string) {
  const n = Number(tone);
  return n >= 1 && n <= 5 ? toneClass[n - 1] : "text-ink";
}
