import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import gogenApi, { ConfigurationSummary } from '../api/gogenApi';
import Hero from '../components/Hero';
import ConfigurationList from '../components/ConfigurationList';

const HomePage = () => {
  const [configurations, setConfigurations] = useState<ConfigurationSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchConfigurations = async () => {
      try {
        setLoading(true);
        const data = await gogenApi.listConfigurations();
        setConfigurations(data);
      } catch (err) {
        setError('Failed to fetch configurations');
        console.error('Error fetching configurations:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchConfigurations();
  }, []);

  return (
    <>
      <Hero />
      <main className="container mx-auto px-4 py-8">
        <section className="mb-12">
          <h2 className="text-3xl font-bold text-gray-800 mb-4">Configurations</h2>
          <ConfigurationList 
            configurations={configurations}
            loading={loading}
            error={error}
          />
        </section>
      </main>
    </>
  );
};

export default HomePage; 