import { useEffect, useSyncExternalStore } from "react";

/** Where microphone permission stands: unknown until checked, then prompt, granted or denied. */
export type MicState = "unknown" | "prompt" | "granted" | "denied";

const okKey = "mandarin.micOK";
let state: MicState = "unknown";
const listeners = new Set<() => void>();

function set(next: MicState) {
  if (next === state) return;
  state = next;
  if (next === "granted") localStorage.setItem(okKey, "1");
  if (next === "denied") localStorage.removeItem(okKey);
  listeners.forEach((fn) => fn());
}

/** Reads the browser's permission without prompting (where the browser lets us). */
async function check() {
  try {
    const status = await navigator.permissions?.query({ name: "microphone" as PermissionName });
    if (status) {
      set(status.state === "granted" ? "granted" : status.state === "denied" ? "denied" : localStorage.getItem(okKey) ? "granted" : "prompt");
      status.onchange = () => set(status.state === "granted" ? "granted" : status.state === "denied" ? "denied" : "prompt");
      return;
    }
  } catch {
    // Some browsers (older Safari) can't query the microphone; fall back below.
  }
  set(localStorage.getItem(okKey) ? "granted" : "prompt");
}

/**
 * Asks for the microphone (the browser shows its prompt), then releases it. Returns a message to
 * show when it failed for a reason other than being blocked (no microphone, one in use).
 */
export async function requestMic(): Promise<string> {
  if (!navigator.mediaDevices?.getUserMedia) return "This browser can't record audio. Try Safari or Chrome.";
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    stream.getTracks().forEach((track) => track.stop());
    set("granted");
    return "";
  } catch (err) {
    const name = err instanceof DOMException ? err.name : "";
    if (name === "NotAllowedError" || name === "SecurityError") {
      set("denied");
      return "";
    }
    if (name === "NotFoundError") return "No microphone was found. Connect one (or use your phone) and try again.";
    if (name === "NotReadableError") return "The microphone is busy in another app. Close it there and try again.";
    return "Couldn't start the microphone. Try again.";
  }
}

/** Called by the recorder when the browser refuses the microphone mid-practice. */
export function markMicDenied() {
  set("denied");
}

/** The current microphone state, checked once on first use. */
export function useMic(): MicState {
  const current = useSyncExternalStore(
    (fn) => {
      listeners.add(fn);
      return () => listeners.delete(fn);
    },
    () => state,
  );
  useEffect(() => {
    if (state === "unknown") void check();
  }, []);
  return current;
}
