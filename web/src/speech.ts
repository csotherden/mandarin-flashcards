let unlocked = false;

export function unlockSpeech() {
  if (unlocked || typeof window === "undefined" || !window.speechSynthesis) return;
  const utterance = new SpeechSynthesisUtterance(" ");
  utterance.volume = 0;
  utterance.lang = "zh-CN";
  window.speechSynthesis.speak(utterance);
  unlocked = true;
}

export function speakChinese(text: string, rate: number) {
  if (!text || !window.speechSynthesis) return;
  window.speechSynthesis.cancel();
  const utterance = new SpeechSynthesisUtterance(text);
  utterance.lang = "zh-CN";
  utterance.rate = rate;
  const voices = window.speechSynthesis.getVoices();
  const voice =
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
