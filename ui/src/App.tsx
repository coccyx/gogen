import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';
import HomePage from './pages/HomePage';
import ConfigurationDetailPage from './pages/ConfigurationDetailPage';
import LoginPage from './pages/LoginPage';
import AuthCallbackPage from './pages/AuthCallbackPage';
import MyConfigurationsPage from './pages/MyConfigurationsPage';
import EditConfigurationPage from './pages/EditConfigurationPage';
import NotFoundPage from './pages/NotFoundPage';

function App() {
  return (
    <AuthProvider>
      <Router>
        <Layout>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/auth/callback" element={<AuthCallbackPage />} />
            <Route path="/configurations/:owner/:configName" element={<ConfigurationDetailPage />} />
            <Route
              path="/my-configurations"
              element={
                <ProtectedRoute>
                  <MyConfigurationsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/new"
              element={
                <ProtectedRoute>
                  <EditConfigurationPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/edit/:owner/:configName"
              element={
                <ProtectedRoute>
                  <EditConfigurationPage />
                </ProtectedRoute>
              }
            />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </Layout>
      </Router>
    </AuthProvider>
  );
}

export default App;
