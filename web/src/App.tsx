import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "./auth";
import { Shell } from "./components/Shell";
import { AuthPage } from "./pages/AuthPage";
import { CardPage } from "./pages/CardPage";
import { DrillPage } from "./pages/DrillPage";
import { LessonPage } from "./pages/LessonPage";
import { LessonsPage } from "./pages/LessonsPage";
import { LibraryPage } from "./pages/LibraryPage";
import { SetPage } from "./pages/SetPage";
import { SetsPage } from "./pages/SetsPage";
import { SummaryPage } from "./pages/SummaryPage";

function Guard({ children }: { children: React.ReactNode }) {
  const { user, ready } = useAuth();
  if (!ready) return <p className="p-6 text-stone-500">Loading…</p>;
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
        <Route path="/" element={<LibraryPage />} />
        <Route path="/cards/new" element={<CardPage />} />
        <Route path="/cards/:id" element={<CardPage />} />
        <Route path="/sets" element={<SetsPage />} />
        <Route path="/sets/:id" element={<SetPage />} />
        <Route path="/lessons" element={<LessonsPage />} />
        <Route path="/lessons/:id" element={<LessonPage />} />
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
