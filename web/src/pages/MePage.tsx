import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api, ApiError } from "../api";
import { useAuth } from "../auth";
import { Icon } from "../components/Icon";
import { Sheet } from "../components/Sheet";
import { chineseVoices, speakChinese } from "../speech";
import type { Home, LoginResponse, Preferences } from "../types";

const speeds = [
  { value: 0.7, label: "Slower" },
  { value: 0.85, label: "A bit slower" },
  { value: 1, label: "Normal" },
  { value: 1.15, label: "A bit faster" },
];

/** Me: progress at a glance, ways back into learning, and settings. */
export function MePage() {
  const { user, logout } = useAuth();
  const [home, setHome] = useState<Home | null>(null);
  const [sheet, setSheet] = useState<"" | "password" | "delete">("");
  const [note, setNote] = useState("");

  useEffect(() => {
    api<Home>("/api/home").then(setHome).catch(() => undefined);
  }, []);

  return (
    <div className="mx-auto max-w-xl">
      <h1 className="mb-5 text-2xl font-extrabold lg:text-3xl">Me</h1>
      {home && (
        <div className="grid grid-cols-2 gap-3">
          <Stat value={home.streak_days} label={home.streak_today || home.streak_days === 0 ? "day streak" : "day streak, practice today"} />
          <Stat value={home.streak_longest} label="longest streak" />
          <Stat value={home.xp_total} label="total XP" />
          <Stat value={home.total_cards} label="cards" />
        </div>
      )}

      <h2 className="mb-2 mt-7 font-bold">Keep learning</h2>
      <div className="surface divide-y divide-line">
        <MeLink to="/tones" glyph="声" title="Tone practice" detail={home ? (home.tones_next ? `Intro to tones · ${home.tones_completed} of ${home.tones_total} done` : "Revisit any tones lesson") : ""} />
        <MeLink to="/library" glyph="书" title="Library" detail="Ready-made sets to practice or add" />
        <MeLink to="/cards" glyph="卡" title="Your cards" detail={home ? `${home.total_cards} cards` : ""} />
      </div>

      <h2 className="mb-2 mt-7 font-bold">Speech</h2>
      <SpeechSettings />

      <h2 className="mb-2 mt-7 font-bold">Account</h2>
      {note && (
        <p className="mb-3 flex items-center gap-2 rounded-2xl bg-ink px-4 py-3 text-sm text-white" role="status">
          <Icon name="check" className="shrink-0 text-gold" /> {note}
        </p>
      )}
      <div className="surface divide-y divide-line">
        <div className="px-4 py-3">
          <p className="label">Email</p>
          <p className="mt-0.5 break-all font-semibold">{user?.email}</p>
        </div>
        <div className="px-4 py-3">
          <p className="label">Your days count in</p>
          <p className="mt-0.5 font-semibold">{user?.timezone || "UTC"}</p>
          <p className="text-sm text-muted">Set from this device, for streaks and daily XP.</p>
        </div>
        <button type="button" className="flex min-h-14 w-full items-center gap-3 px-4 text-left font-semibold" onClick={() => setSheet("password")}>
          <span className="flex-1">Change password</span>
          <Icon name="right" className="text-lg text-muted" />
        </button>
        <button type="button" className="flex min-h-14 w-full items-center gap-3 px-4 text-left font-semibold" onClick={logout}>
          <span className="flex-1">Sign out</span>
          <Icon name="logout" className="text-lg text-muted" />
        </button>
      </div>
      <button type="button" className="mx-auto mt-6 flex min-h-11 items-center gap-1.5 px-3 text-sm font-semibold text-crimson" onClick={() => setSheet("delete")}>
        <Icon name="trash" /> Delete account
      </button>

      <PasswordSheet open={sheet === "password"} onClose={() => setSheet("")} onDone={() => setNote("Password changed. Other devices will need to sign in again.")} />
      <DeleteSheet open={sheet === "delete"} onClose={() => setSheet("")} />
    </div>
  );
}

const isLocale = (voice: string) => voice === "" || /^[a-z]{2}(-[A-Z]{2})?$/.test(voice);

function SpeechSettings() {
  const { user, setUser } = useAuth();
  const [voices, setVoices] = useState(chineseVoices());
  const [saving, setSaving] = useState(false);
  const prefs = user?.preferences ?? { default_tts_speed: 1, preferred_voice: "" };

  useEffect(() => {
    const synth = window.speechSynthesis;
    const update = () => setVoices(chineseVoices());
    synth?.addEventListener?.("voiceschanged", update);
    update();
    return () => synth?.removeEventListener?.("voiceschanged", update);
  }, []);

  async function save(next: Preferences) {
    if (!user) return;
    setUser({ ...user, preferences: next }); // applies at once; the save follows
    setSaving(true);
    try {
      const saved = await api<Preferences>("/api/user/preferences", { method: "PUT", body: JSON.stringify(next) });
      setUser({ ...user, preferences: saved });
    } finally {
      setSaving(false);
    }
  }

  const nearest = speeds.reduce((best, s) => (Math.abs(s.value - prefs.default_tts_speed) < Math.abs(best.value - prefs.default_tts_speed) ? s : best), speeds[2]);

  return (
    <div className="surface grid gap-4 p-4">
      <fieldset>
        <legend className="label">Speed</legend>
        <div className="mt-2 grid grid-cols-2 gap-2 sm:grid-cols-4">
          {speeds.map((s) => (
            <button
              key={s.value}
              type="button"
              aria-pressed={nearest.value === s.value}
              onClick={() => save({ ...prefs, default_tts_speed: s.value })}
              className={`min-h-11 rounded-2xl border-2 px-3 text-sm font-semibold ${nearest.value === s.value ? "border-crimson bg-crimson text-white" : "border-line bg-white"}`}
            >
              {s.label}
            </button>
          ))}
        </div>
      </fieldset>
      <div className="grid gap-1.5">
        <label className="label" htmlFor="voice">Voice</label>
        <select
          id="voice"
          aria-describedby="voice-hint"
          className="field font-semibold"
          value={isLocale(prefs.preferred_voice) ? "" : prefs.preferred_voice}
          onChange={(event) => save({ ...prefs, preferred_voice: event.target.value || "zh-CN" })}
        >
          <option value="">Automatic</option>
          {voices.map((v) => (
            <option key={v.name} value={v.name}>
              {v.name}
            </option>
          ))}
        </select>
        <span id="voice-hint" className="text-xs text-muted">{voices.length ? "Voices come from this device and browser." : "This browser didn't offer any Mandarin voices."}</span>
      </div>
      <button type="button" className="btn-soft min-h-11" onClick={() => speakChinese("你好，我们一起学中文吧。", 1)}>
        <Icon name="volume" /> Try it {saving && <span className="text-sm text-muted">(saving…)</span>}
      </button>
    </div>
  );
}

function PasswordSheet({ open, onClose, onDone }: { open: boolean; onClose: () => void; onDone: () => void }) {
  const { replaceSession } = useAuth();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (open) {
      setCurrent("");
      setNext("");
      setConfirm("");
      setError("");
    }
  }, [open]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (next !== confirm) {
      setError("The new passwords don't match");
      return;
    }
    setBusy(true);
    setError("");
    try {
      replaceSession(await api<LoginResponse>("/api/user/password", { method: "PUT", body: JSON.stringify({ current_password: current, new_password: next }) }));
      onClose();
      onDone();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not change the password");
    } finally {
      setBusy(false);
    }
  }

  return (
    <Sheet open={open} onClose={onClose} title="Change password">
      <form className="grid gap-4" onSubmit={submit}>
        <label className="grid gap-1.5">
          <span className="label">Current password</span>
          <input className="field" type="password" autoComplete="current-password" required value={current} onChange={(e) => setCurrent(e.target.value)} />
        </label>
        <div className="grid gap-1.5">
          <label className="label" htmlFor="new-password">New password</label>
          <input id="new-password" className="field" type="password" autoComplete="new-password" required minLength={8} aria-describedby="new-password-hint" value={next} onChange={(e) => setNext(e.target.value)} />
          <span id="new-password-hint" className="text-xs text-muted">At least 8 characters.</span>
        </div>
        <label className="grid gap-1.5">
          <span className="label">New password again</span>
          <input className="field" type="password" autoComplete="new-password" required minLength={8} value={confirm} onChange={(e) => setConfirm(e.target.value)} />
        </label>
        {error && <p className="text-sm font-semibold text-crimson" role="alert">{error}</p>}
        <button type="submit" className="btn-primary min-h-14 text-lg" disabled={busy}>
          {busy ? "Saving…" : "Change password"}
        </button>
      </form>
    </Sheet>
  );
}

function DeleteSheet({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { logout } = useAuth();
  const navigate = useNavigate();
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (open) {
      setPassword("");
      setError("");
    }
  }, [open]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError("");
    try {
      await api("/api/user", { method: "DELETE", body: JSON.stringify({ password }) });
      logout();
      navigate("/register", { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not delete the account");
      setBusy(false);
    }
  }

  return (
    <Sheet open={open} onClose={onClose} title="Delete account">
      <form className="grid gap-4" onSubmit={submit}>
        <p>
          This permanently deletes your account and everything in it: cards, photos you uploaded, sets, progress, streak and XP. It
          can't be undone.
        </p>
        <label className="grid gap-1.5">
          <span className="label">Password</span>
          <input className="field" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} />
        </label>
        {error && <p className="text-sm font-semibold text-crimson" role="alert">{error}</p>}
        <button type="submit" className="min-h-14 rounded-2xl bg-crimson text-lg font-bold text-white disabled:opacity-60" disabled={busy || !password}>
          {busy ? "Deleting…" : "Delete my account"}
        </button>
        <button type="button" className="btn-ghost min-h-12" onClick={onClose}>
          Keep my account
        </button>
      </form>
    </Sheet>
  );
}

function Stat({ value, label }: { value: number; label: string }) {
  return (
    <div className="surface p-4 text-center">
      <b className="block text-3xl font-extrabold">{value}</b>
      <span className="text-sm text-muted">{label}</span>
    </div>
  );
}

function MeLink({ to, glyph, title, detail }: { to: string; glyph: string; title: string; detail: string }) {
  return (
    <Link to={to} className="flex items-center gap-3 px-4 py-3">
      <span className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-ink font-hanzi text-lg font-bold text-gold" aria-hidden>
        {glyph}
      </span>
      <span className="min-w-0 flex-1">
        <b className="block">{title}</b>
        <span className="block truncate text-sm text-muted">{detail}</span>
      </span>
      <Icon name="right" className="shrink-0 text-lg text-muted" />
    </Link>
  );
}
