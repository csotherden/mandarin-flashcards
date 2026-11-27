import { useState } from "react";
import { Link, Navigate } from "react-router-dom";
import { ApiError } from "../api";
import { useAuth } from "../auth";
import { BrandMark } from "../components/Shell";

export function AuthPage({ mode }: { mode: "login" | "register" }) {
  const { user, login, register } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  // A new account starts with a lesson offer rather than an empty Today.
  if (user) return <Navigate to={mode === "register" ? "/welcome" : "/"} replace />;

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError("");
    try {
      if (mode === "login") await login(email, password);
      else await register(email, password);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not continue");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="min-h-dvh lg:grid lg:grid-cols-2">
      <section className="relative hidden overflow-hidden bg-hero p-12 text-white lg:flex lg:flex-col lg:justify-end">
        <span className="absolute -right-10 -top-16 font-hanzi text-[22rem] font-black leading-none text-white/10" aria-hidden>卡</span>
        <h2 className="text-5xl font-extrabold leading-tight">Say it.<br />Hear which tone slipped.</h2>
        <p className="mt-4 max-w-md text-lg text-white/90">Picture cards for the words that keep getting away, graded syllable by syllable.</p>
      </section>
      <div className="mx-auto flex min-h-dvh max-w-md flex-col px-6 pb-10 lg:min-h-0 lg:justify-center lg:pb-0">
        <div className="relative -mx-6 overflow-hidden bg-hero px-6 pb-7 pt-10 text-white lg:hidden">
          <span className="absolute -right-6 -top-8 font-hanzi text-[9rem] font-black leading-none text-white/10" aria-hidden>
            卡
          </span>
          <div className="flex items-center gap-3 text-lg font-bold">
            <BrandMark />
            Mandarin Cards
          </div>
          <p className="mt-4 text-2xl font-extrabold leading-tight">Say it. Hear which tone slipped.</p>
        </div>
        <div className="hidden items-center gap-3 text-lg font-bold lg:flex">
          <BrandMark />
          Mandarin Cards
        </div>
        <h1 className="mt-8 text-3xl font-extrabold">{mode === "login" ? "Welcome back" : "Create an account"}</h1>
        <p className="mt-1 text-muted">{mode === "login" ? "Pick up where you left off." : "Add your first words in seconds."}</p>
        <form className="mt-7 grid gap-4" onSubmit={submit}>
          <label className="grid gap-1.5">
            <span className="label">Email</span>
            <input className="field" type="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
          </label>
          <label className="grid gap-1.5">
            <span className="label">Password</span>
            <input className="field" type="password" autoComplete={mode === "login" ? "current-password" : "new-password"} value={password} onChange={(event) => setPassword(event.target.value)} required minLength={8} />
          </label>
          {error && <p className="text-sm font-semibold text-crimson" role="alert">{error}</p>}
          <button className="btn-primary h-13 py-3.5 text-lg" disabled={busy} type="submit">
            {busy ? "Please wait…" : mode === "login" ? "Sign in" : "Create account"}
          </button>
        </form>
        <p className="mt-3 text-muted">
          {mode === "login" ? (
            <>New here? <Link className="link tap" to="/register">Create an account</Link></>
          ) : (
            <>Have an account? <Link className="link tap" to="/login">Sign in</Link></>
          )}
        </p>
      </div>
    </main>
  );
}
