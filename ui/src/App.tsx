import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import HomePage from './pages/HomePage';
import ConfigurationDetailPage from './pages/ConfigurationDetailPage';

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/configurations/:owner/:configName" element={<ConfigurationDetailPage />} />
        </Routes>
      </Layout>
    </Router>
  );
}

export default App; 