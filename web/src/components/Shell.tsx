import { NavLink, Outlet } from "react-router-dom";
import { useAuth } from "../auth";

const link = "rounded-full px-3 py-2 text-sm font-medium text-stone-600";
const active = "bg-ink text-paper";

export function Shell() {
  const { user, logout } = useAuth();
  return (
    <div className="mx-auto flex min-h-screen w-full max-w-3xl flex-col px-4 pb-24 pt-4">
      <header className="mb-5 flex items-center justify-between gap-3">
        <div>
          <p className="text-xs uppercase tracking-[0.18em] text-cinnabar">Mandarin</p>
          <h1 className="font-hanzi text-2xl font-semibold">Cards</h1>
        </div>
        <button className="text-sm text-stone-500" onClick={logout} type="button">
          {user?.email.split("@")[0]} · Sign out
        </button>
      </header>
      <main className="flex-1">
        <Outlet />
      </main>
      <nav className="fixed inset-x-0 bottom-0 border-t border-stone-300/80 bg-paper/95 px-4 py-3 backdrop-blur">
        <div className="mx-auto flex max-w-3xl justify-around">
          <NavLink to="/" end className={({ isActive }) => `${link} ${isActive ? active : ""}`}>
            Library
          </NavLink>
          <NavLink to="/sets" className={({ isActive }) => `${link} ${isActive ? active : ""}`}>
            Sets
          </NavLink>
          <NavLink to="/lessons" className={({ isActive }) => `${link} ${isActive ? active : ""}`}>
            Lessons
          </NavLink>
        </div>
      </nav>
    </div>
  );
}
