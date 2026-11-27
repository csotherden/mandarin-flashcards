import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "./auth";
import { Shell } from "./components/Shell";
import { AddPage } from "./pages/AddPage";
import { AuthPage } from "./pages/AuthPage";
import { CardPage } from "./pages/CardPage";
import { CardsPage } from "./pages/CardsPage";
import { DrillPage } from "./pages/DrillPage";
import { GeneratePage } from "./pages/GeneratePage";
import { LibraryPage } from "./pages/LibraryPage";
import { LibrarySetPage } from "./pages/LibrarySetPage";
import { MePage } from "./pages/MePage";
import { SetPage } from "./pages/SetPage";
import { SetsPage } from "./pages/SetsPage";
import { SummaryPage } from "./pages/SummaryPage";
import { TodayPage } from "./pages/TodayPage";
import { WelcomePage } from "./pages/WelcomePage";

function Guard({ children }: { children: React.ReactNode }) {
  const { user, ready } = useAuth();
  if (!ready) return <p className="p-6 text-muted">Loading…</p>;
  if (!user) return <Navigate to="/login" replace />;
  return children;
}

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<AuthPage mode="login" />} />
      <Route path="/register" element={<AuthPage mode="register" />} />
      <Route
        element={
          <Guard>
            <Shell />
          </Guard>
        }
      >
        <Route path="/" element={<TodayPage />} />
        <Route path="/add" element={<AddPage />} />
        <Route path="/cards" element={<CardsPage />} />
        <Route path="/cards/new" element={<CardPage />} />
        <Route path="/cards/:id" element={<CardPage />} />
        <Route path="/sets" element={<SetsPage />} />
        <Route path="/sets/generate" element={<GeneratePage />} />
        <Route path="/sets/:id" element={<SetPage />} />
        <Route path="/library" element={<LibraryPage />} />
        <Route path="/library/:id" element={<LibrarySetPage />} />
        <Route path="/welcome" element={<WelcomePage />} />
        <Route path="/me" element={<MePage />} />
        {/* Lessons became quick starts on Today. */}
        <Route path="/lessons/*" element={<Navigate to="/" replace />} />
      </Route>
      <Route
        path="/drill/:id"
        element={
          <Guard>
            <DrillPage />
          </Guard>
        }
      />
      <Route
        path="/summary/:id"
        element={
          <Guard>
            <SummaryPage />
          </Guard>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
