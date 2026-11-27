import { useRef, useState, type KeyboardEvent, type PointerEvent } from "react";
import { markMicDenied } from "./mic";
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

  async function start() {
    if (holding.current) return;
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
    } catch (err) {
      holding.current = false;
      if (err instanceof DOMException && (err.name === "NotAllowedError" || err.name === "SecurityError")) markMicDenied();
      else onError("Couldn't start the microphone. Check that one is connected and try again.");
    }
  }

  function stop() {
    holding.current = false;
    if (recorder.current && recorder.current.state === "recording") recorder.current.stop();
    setRecording(false);
  }

  const isHoldKey = (event: KeyboardEvent) => event.key === " " || event.key === "Enter";

  /**
   * Spread onto the mic button. Press and hold with a pointer, or hold Space or Enter while the
   * button has focus (keyboard users record the same way).
   */
  const bind = {
    onPointerDown: (event: PointerEvent<HTMLButtonElement>) => {
      event.preventDefault();
      void start();
    },
    onPointerUp: stop,
    onPointerCancel: stop,
    onKeyDown: (event: KeyboardEvent<HTMLButtonElement>) => {
      if (!isHoldKey(event)) return;
      event.preventDefault();
      if (!event.repeat) void start();
    },
    onKeyUp: (event: KeyboardEvent<HTMLButtonElement>) => {
      if (!isHoldKey(event)) return;
      event.preventDefault();
      stop();
    },
    onBlur: stop,
    onContextMenu: (event: { preventDefault: () => void }) => event.preventDefault(),
  };
  return { recording, bind };
}
