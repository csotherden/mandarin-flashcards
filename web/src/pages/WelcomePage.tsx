import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api";
import { Icon } from "../components/Icon";
import { LibraryGrid } from "../components/LibraryGrid";
import { BrandMark } from "../components/Shell";
import type { LibrarySet } from "../types";

/**
 * Where a new account lands: try a lesson first, from a ready-made set, before building
 * anything. Picking a set opens it with a one-tap start; building your own is right below.
 */
export function WelcomePage() {
  const [sets, setSets] = useState<LibrarySet[] | null>(null);

  useEffect(() => {
    api<LibrarySet[]>("/api/library/sets").then(setSets).catch(() => setSets([]));
  }, []);

  const starters = sets ?? [];

  return (
    <div className="mx-auto max-w-3xl">
      <div className="flex items-center gap-3 text-lg font-bold lg:hidden">
        <BrandMark />
        Mandarin Cards
      </div>
      <section className="relative mt-5 overflow-hidden rounded-[28px] bg-hero p-6 text-white lg:mt-0 lg:p-8">
        <span className="absolute -right-6 -top-10 font-hanzi text-[9rem] font-black leading-none text-white/10" aria-hidden>
          学
        </span>
        <p className="text-sm font-semibold uppercase tracking-[0.08em] text-white/85">Welcome</p>
        <h1 className="mt-1.5 max-w-md text-3xl font-extrabold leading-tight lg:text-4xl">Want to try a lesson?</h1>
        <p className="mt-2 max-w-lg text-white/90">
          Pick a set below and start talking. You'll say each word and see exactly which tones landed. The set is added to
          your cards, so you can change it, add to it, or build your own next.
        </p>
      </section>

      <Link to="/tones" className="surface mt-5 flex items-center gap-4 p-4 ring-2 ring-crimson/70">
        <span className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-ink font-hanzi text-2xl font-bold text-gold">声</span>
        <span className="min-w-0 flex-1">
          <b className="block">New to tones? Start here</b>
          <span className="block text-sm text-muted">Short lessons on the four tones, with your speaking graded. A great first step.</span>
        </span>
        <Icon name="right" className="shrink-0 text-xl text-crimson" />
      </Link>

      {sets === null ? (
        <p className="mt-6 text-muted">Loading…</p>
      ) : starters.length > 0 ? (
        <section className="mt-6">
          <h2 className="mb-3 font-bold">Or jump into a set</h2>
          <LibraryGrid sets={starters} compact from="welcome" />
        </section>
      ) : null}

      <section className="mt-7">
        <h2 className="mb-3 font-bold">{starters.length > 0 ? "Or build your own" : "Build your own deck"}</h2>
        <div className="grid gap-3 sm:grid-cols-2">
          <Link to="/add" className="surface flex items-center gap-3 p-4">
            <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-crimson text-xl text-white"><Icon name="plus" /></span>
            <span>
              <b className="block">Add your own words</b>
              <span className="text-sm text-muted">Type hanzi, English or pinyin</span>
            </span>
          </Link>
          <Link to="/sets/generate" className="surface flex items-center gap-3 p-4">
            <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-ink text-xl text-gold"><Icon name="spark" /></span>
            <span>
              <b className="block">Generate a set</b>
              <span className="text-sm text-muted">Describe a theme, get the cards</span>
            </span>
          </Link>
        </div>
        <Link to="/" className="mt-5 block text-center font-semibold text-muted underline underline-offset-4">
          Skip for now
        </Link>
      </section>
    </div>
  );
}

