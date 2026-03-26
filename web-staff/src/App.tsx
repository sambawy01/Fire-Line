import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { isAuthenticated } from './stores/auth';
import LoginPage from './pages/LoginPage';
import Layout from './components/Layout';
import HomePage from './pages/HomePage';
import TasksPage from './pages/TasksPage';
import SchedulePage from './pages/SchedulePage';
import PointsPage from './pages/PointsPage';
import ChatPage from './pages/ChatPage';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  return isAuthenticated() ? <>{children}</> : <Navigate to="/login" replace />;
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<ProtectedRoute><Layout /></ProtectedRoute>}>
          <Route index element={<HomePage />} />
          <Route path="tasks" element={<TasksPage />} />
          <Route path="schedule" element={<SchedulePage />} />
          <Route path="chat" element={<ChatPage />} />
          <Route path="points" element={<PointsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
