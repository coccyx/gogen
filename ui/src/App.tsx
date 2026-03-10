import { Suspense, lazy } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';
import LoadingSpinner from './components/LoadingSpinner';

const HomePage = lazy(() => import('./pages/HomePage'));
const ConfigurationDetailPage = lazy(() => import('./pages/ConfigurationDetailPage'));
const LoginPage = lazy(() => import('./pages/LoginPage'));
const AuthCallbackPage = lazy(() => import('./pages/AuthCallbackPage'));
const MyConfigurationsPage = lazy(() => import('./pages/MyConfigurationsPage'));
const EditConfigurationPage = lazy(() => import('./pages/EditConfigurationPage'));
const NotFoundPage = lazy(() => import('./pages/NotFoundPage'));

function App() {
  return (
    <AuthProvider>
      <Router>
        <Layout>
          <Suspense fallback={<div className="container mx-auto px-4 py-6"><LoadingSpinner /></div>}>
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
          </Suspense>
        </Layout>
      </Router>
    </AuthProvider>
  );
}

export default App;
