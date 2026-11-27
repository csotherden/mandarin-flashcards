import { useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { requestMic, useMic } from "../mic";
import { Icon } from "./Icon";

function device() {
  const ua = navigator.userAgent;
  if (/iPhone|iPad|iPod/.test(ua)) return "ios";
  if (/Android/.test(ua)) return "android";
  return "desktop";
}

/**
 * Shows its children (a speaking activity) once the microphone is allowed. Before the browser's
 * first prompt it explains why we need the mic; if the mic is blocked it explains how to unblock
 * it, with a way out that doesn't need speaking.
 */
export function MicGate({ children, exitTo = "/", exitLabel = "Back to Today" }: { children: ReactNode; exitTo?: string; exitLabel?: string }) {
  const mic = useMic();
  const [asking, setAsking] = useState(false);
  const [problem, setProblem] = useState("");

  if (mic === "granted") return <>{children}</>;
  if (mic === "unknown") return <p className="py-10 text-center text-muted">Checking the microphone…</p>;

  async function ask() {
    setAsking(true);
    setProblem("");
    setProblem(await requestMic());
    setAsking(false);
  }

  if (mic === "denied") {
    const where = device();
    return (
      <section className="rise surface p-6" aria-labelledby="mic-blocked">
        <span className="grid h-14 w-14 place-items-center rounded-2xl bg-crimson/10 text-3xl text-crimson" aria-hidden>
          <Icon name="mic" />
        </span>
        <h2 id="mic-blocked" className="mt-4 text-2xl font-extrabold">The microphone is blocked</h2>
        <p className="mt-2 text-muted">Speaking practice needs it to hear your tones. To turn it back on:</p>
        <ol className="mt-3 list-decimal space-y-1.5 pl-5">
          {where === "ios" ? (
            <>
              <li>Tap <b>aA</b> (or the page menu) in Safari's address bar.</li>
              <li>Choose <b>Website Settings</b>, then set <b>Microphone</b> to <b>Allow</b>.</li>
              <li>Or open the Settings app → <b>Safari</b> → <b>Microphone</b> → <b>Allow</b>.</li>
            </>
          ) : where === "android" ? (
            <>
              <li>Tap the icon to the left of the address in Chrome.</li>
              <li>Choose <b>Permissions</b> (or Site settings) → <b>Microphone</b> → <b>Allow</b>.</li>
            </>
          ) : (
            <>
              <li>Click the icon to the left of the address bar.</li>
              <li>Set <b>Microphone</b> to <b>Allow</b>, then reload if asked.</li>
            </>
          )}
        </ol>
        <div className="mt-5 grid gap-3 sm:grid-cols-2">
          <button type="button" className="btn-primary min-h-12" disabled={asking} onClick={ask}>
            {asking ? "Checking…" : "Try again"}
          </button>
          {problem && (
            <p className="text-sm font-semibold text-crimson sm:col-span-2" role="alert">
              {problem}
            </p>
          )}
          <Link to={exitTo} className="btn-ghost min-h-12">
            {exitLabel}
          </Link>
        </div>
        <p className="mt-4 text-sm text-muted">
          No mic right now? The <b>Picture ↔ hanzi</b> match drill doesn't need one.
        </p>
      </section>
    );
  }

  return (
    <section className="rise surface p-6 text-center" aria-labelledby="mic-primer">
      <span className="mx-auto grid h-16 w-16 place-items-center rounded-full bg-crimson text-3xl text-white shadow-[0_12px_28px_rgba(200,16,46,.35)]" aria-hidden>
        <Icon name="mic" />
      </span>
      <h2 id="mic-primer" className="mt-4 text-2xl font-extrabold">Let's hear your tones</h2>
      <p className="mx-auto mt-2 max-w-sm text-muted">
        You'll hold the mic button and say each word, and we'll show which tones landed. Your browser will ask to use the
        microphone; recordings are only used to grade you, and the app doesn't keep them.
      </p>
      {problem && (
        <p className="mt-4 rounded-2xl bg-crimson/10 px-4 py-3 text-sm font-semibold text-crimson" role="alert">
          {problem}
        </p>
      )}
      <button type="button" className="btn-primary mt-5 min-h-14 w-full text-lg sm:w-auto sm:px-10" disabled={asking} onClick={ask}>
        <Icon name="mic" /> {asking ? "Waiting for your browser…" : "Turn on the microphone"}
      </button>
      <Link to={exitTo} className="mt-2 flex min-h-11 items-center justify-center font-semibold text-muted underline underline-offset-4">
        Not now
      </Link>
    </section>
  );
}
