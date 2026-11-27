import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api";
import { Icon } from "../components/Icon";
import { LibraryGrid } from "../components/LibraryGrid";
import type { LibrarySet } from "../types";

/** The global library: ready-made sets to practice or add to your own. */
export function LibraryPage() {
  const [sets, setSets] = useState<LibrarySet[] | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    api<LibrarySet[]>("/api/library/sets").then(setSets).catch((err: Error) => setError(err.message));
  }, []);

  return (
    <div>
      <Link to="/sets" className="mb-4 flex w-fit items-center gap-1 font-semibold text-muted">
        <Icon name="back" /> Sets
      </Link>
      <h1 className="text-2xl font-extrabold lg:text-3xl">Library</h1>
      <p className="mb-5 mt-1 text-muted">Ready-made sets with photos and tones. Practice one and it becomes yours to change.</p>
      {error && <p className="text-sm font-semibold text-crimson" role="alert">{error}</p>}
      {sets && sets.length === 0 && <p className="surface p-6 text-center text-muted">The library is empty for now.</p>}
      {sets && <LibraryGrid sets={sets} />}
    </div>
  );
}
