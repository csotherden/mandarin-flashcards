import { Link, NavLink, Outlet } from "react-router-dom";
import { useAuth } from "../auth";
import { Icon, type IconName } from "./Icon";

const tabs: { to: string; label: string; icon: IconName; end?: boolean }[] = [
  { to: "/", label: "Today", icon: "home", end: true },
  { to: "/cards", label: "Cards", icon: "cards" },
  { to: "/sets", label: "Sets", icon: "sets" },
];

export function BrandMark({ className = "h-9 w-9 text-xl" }: { className?: string }) {
  return <span className={`grid place-items-center rounded-[11px] bg-flag font-hanzi font-black text-gold ${className}`}>卡</span>;
}

/** Phones: content above a dark tab bar with a raised add button. Large screens: a dark sidebar. */
export function Shell() {
  const { user } = useAuth();
  return (
    <div className="min-h-screen lg:flex">
      <aside className="sticky top-0 hidden h-screen w-64 shrink-0 flex-col gap-1.5 bg-ink px-4 py-6 text-white lg:flex">
        <Link to="/" className="mb-6 flex items-center gap-3 px-2 text-lg font-bold">
          <BrandMark />
          Mandarin Cards
        </Link>
        {[...tabs, { to: "/me", label: "Me", icon: "user" as IconName }].map((tab) => (
          <NavLink
            key={tab.to}
            to={tab.to}
            end={tab.end}
            className={({ isActive }) =>
              `relative flex items-center gap-3 rounded-2xl px-4 py-3 font-semibold ${isActive ? "bg-white/10 text-gold before:absolute before:-left-4 before:inset-y-2.5 before:w-1 before:rounded before:bg-gold" : "text-stone-300 hover:bg-white/5"}`
            }
          >
            <Icon name={tab.icon} className="text-xl" />
            {tab.label}
          </NavLink>
        ))}
        <Link to="/add" className="btn-primary mt-4 h-12">
          <Icon name="plus" className="text-lg" /> Add cards
        </Link>
        <p className="mt-auto truncate px-2 text-sm text-stone-400">{user?.email}</p>
      </aside>

      <main className="mx-auto w-full max-w-5xl px-5 pb-32 pt-5 lg:px-10 lg:pb-12 lg:pt-8">
        <Outlet />
      </main>

      <nav className="fixed inset-x-0 bottom-0 z-30 bg-ink pb-[env(safe-area-inset-bottom)] lg:hidden" aria-label="Main">
        <div className="mx-auto flex h-[72px] max-w-lg items-center justify-around px-2">
          {tabs.slice(0, 2).map((tab) => (
            <TabLink key={tab.to} {...tab} />
          ))}
          <Link
            to="/add"
            aria-label="Add cards"
            className="-mt-7 grid h-14 w-14 place-items-center rounded-[20px] border-4 border-canvas bg-crimson text-white shadow-[0_8px_20px_rgba(200,16,46,.35)]"
          >
            <Icon name="plus" className="text-2xl" />
          </Link>
          <TabLink {...tabs[2]} />
          <TabLink to="/me" label="Me" icon="user" />
        </div>
      </nav>
    </div>
  );
}

function TabLink({ to, label, icon, end }: { to: string; label: string; icon: IconName; end?: boolean }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) => `grid w-16 justify-items-center gap-1 text-xs font-semibold ${isActive ? "text-gold" : "text-stone-400"}`}
    >
      {({ isActive }) => (
        <>
          <Icon name={icon} className="text-2xl" />
          {label}
          <span className={`h-[3px] w-[18px] rounded ${isActive ? "bg-gold" : "bg-transparent"}`} />
        </>
      )}
    </NavLink>
  );
}
