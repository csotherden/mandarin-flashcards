import { useState } from "react";
import { Link, Navigate } from "react-router-dom";
import { ApiError } from "../api";
import { useAuth } from "../auth";

export function AuthPage({ mode }: { mode: "login" | "register" }) {
  const { user, login, register } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  if (user) return <Navigate to="/" replace />;

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
    <main className="mx-auto flex min-h-screen max-w-md flex-col justify-center px-5">
      <p className="text-xs uppercase tracking-[0.18em] text-cinnabar">Mandarin cards</p>
      <h1 className="mt-2 font-hanzi text-4xl font-semibold">{mode === "login" ? "Welcome back" : "Create an account"}</h1>
      <p className="mt-2 text-stone-600">Drill the words that stick, then hear exactly which tone slipped.</p>
      <form className="mt-8 grid gap-3" onSubmit={submit}>
        <label className="grid gap-1 text-sm">
          Email
          <input className="rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" type="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
        </label>
        <label className="grid gap-1 text-sm">
          Password
          <input className="rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" type="password" autoComplete={mode === "login" ? "current-password" : "new-password"} value={password} onChange={(event) => setPassword(event.target.value)} required minLength={8} />
        </label>
        {error && <p className="text-sm text-cinnabar">{error}</p>}
        <button className="rounded-2xl bg-ink py-3 font-semibold text-paper" disabled={busy} type="submit">
          {busy ? "Please wait…" : mode === "login" ? "Sign in" : "Create account"}
        </button>
      </form>
      <p className="mt-4 text-sm text-stone-600">
        {mode === "login" ? (
          <Link className="underline" to="/register">Need an account?</Link>
        ) : (
          <Link className="underline" to="/login">Already have one?</Link>
        )}
      </p>
    </main>
  );
}
