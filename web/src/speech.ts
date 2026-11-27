let unlocked = false;

export function unlockSpeech() {
  if (unlocked || typeof window === "undefined" || !window.speechSynthesis) return;
  const utterance = new SpeechSynthesisUtterance(" ");
  utterance.volume = 0;
  utterance.lang = "zh-CN";
  window.speechSynthesis.speak(utterance);
  unlocked = true;
}

/** The user's speech settings: a speed multiplier and a voice name ("" or a locale = automatic). */
let preferences = { speed: 1, voice: "" };

export function setSpeechPreferences(speed: number, voice: string) {
  preferences = { speed: speed > 0 ? speed : 1, voice: /^[a-z]{2}(-[A-Z]{2})?$/.test(voice) ? "" : voice };
}

/** Mandarin voices this device offers (they differ by device and browser). */
export function chineseVoices(): SpeechSynthesisVoice[] {
  if (typeof window === "undefined" || !window.speechSynthesis) return [];
  return window.speechSynthesis
    .getVoices()
    .filter((voice) => /^(zh|cmn)/i.test(voice.lang) && !/HK|yue/i.test(voice.lang));
}

// Held so the engine can't garbage-collect an utterance mid-sentence (it then never plays).
let current: SpeechSynthesisUtterance | null = null;

/**
 * Speaks Chinese at `rate` (1 normal, below 1 slower), scaled by the user's speed setting.
 *
 * iOS Safari quirks handled here: speaking immediately after cancel() is often silently dropped,
 * so we only cancel when something is actually playing and give it a moment first; and the engine
 * can be left paused, so we resume it before speaking.
 */
export function speakChinese(text: string, rate: number) {
  const synth = typeof window !== "undefined" ? window.speechSynthesis : undefined;
  if (!text || !synth) return;
  const utterance = new SpeechSynthesisUtterance(text);
  utterance.lang = "zh-CN";
  utterance.rate = Math.min(2, Math.max(0.3, rate * preferences.speed));
  const voices = chineseVoices();
  const voice =
    (preferences.voice && voices.find((item) => item.name === preferences.voice)) ||
    voices.find((item) => /tingting/i.test(item.name)) ||
    voices.find((item) => item.lang.toLowerCase().startsWith("zh"));
  if (voice) utterance.voice = voice;
  utterance.onend = utterance.onerror = () => {
    if (current === utterance) current = null;
  };
  const go = () => {
    current = utterance;
    if (synth.paused) synth.resume();
    synth.speak(utterance);
  };
  if (synth.speaking || synth.pending) {
    synth.cancel();
    window.setTimeout(go, 120);
  } else {
    go();
  }
}

if (typeof window !== "undefined" && window.speechSynthesis) {
  window.speechSynthesis.onvoiceschanged = () => {
    window.speechSynthesis.getVoices();
  };
}
