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

/** Speaks Chinese at `rate` (1 normal, below 1 slower), scaled by the user's speed setting. */
export function speakChinese(text: string, rate: number) {
  if (!text || !window.speechSynthesis) return;
  window.speechSynthesis.cancel();
  const utterance = new SpeechSynthesisUtterance(text);
  utterance.lang = "zh-CN";
  utterance.rate = Math.min(2, Math.max(0.3, rate * preferences.speed));
  const voices = chineseVoices();
  const voice =
    (preferences.voice && voices.find((item) => item.name === preferences.voice)) ||
    voices.find((item) => /tingting/i.test(item.name)) ||
    voices.find((item) => item.lang.toLowerCase().startsWith("zh"));
  if (voice) utterance.voice = voice;
  window.speechSynthesis.speak(utterance);
}

if (typeof window !== "undefined" && window.speechSynthesis) {
  window.speechSynthesis.onvoiceschanged = () => {
    window.speechSynthesis.getVoices();
  };
}
