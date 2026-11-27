import { useRef, useState, type PointerEvent } from "react";
import { unlockSpeech } from "./speech";

function mimeType() {
  const types = ["audio/webm;codecs=opus", "audio/webm", "audio/mp4"];
  return types.find((type) => typeof MediaRecorder !== "undefined" && MediaRecorder.isTypeSupported(type));
}

/**
 * Hold-to-speak recording: press starts the mic, release stops it and hands back the take.
 * Shared by the card drill and the tones lessons.
 */
export function useRecorder(onTake: (audio: Blob) => void, onError: (message: string) => void) {
  const [recording, setRecording] = useState(false);
  const holding = useRef(false);
  const recorder = useRef<MediaRecorder | null>(null);
  const chunks = useRef<Blob[]>([]);

  async function pointerDown(event: PointerEvent<HTMLButtonElement>) {
    event.preventDefault();
    unlockSpeech();
    onError("");
    holding.current = true;
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      if (!holding.current) {
        stream.getTracks().forEach((track) => track.stop());
        return;
      }
      const mime = mimeType();
      const rec = new MediaRecorder(stream, mime ? { mimeType: mime } : undefined);
      chunks.current = [];
      rec.ondataavailable = (item) => {
        if (item.data.size) chunks.current.push(item.data);
      };
      rec.onstop = () => {
        stream.getTracks().forEach((track) => track.stop());
        onTake(new Blob(chunks.current, { type: rec.mimeType || mime || "audio/webm" }));
      };
      recorder.current = rec;
      rec.start();
      setRecording(true);
    } catch {
      onError("Allow the microphone to grade pronunciation.");
      holding.current = false;
    }
  }

  function pointerUp() {
    holding.current = false;
    if (recorder.current && recorder.current.state === "recording") recorder.current.stop();
    setRecording(false);
  }

  /** Spread onto the mic button. */
  const bind = {
    onPointerDown: pointerDown,
    onPointerUp: pointerUp,
    onPointerCancel: pointerUp,
    onContextMenu: (event: { preventDefault: () => void }) => event.preventDefault(),
  };
  return { recording, bind };
}
