import { Link } from 'react-router-dom';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout = ({ children }: LayoutProps) => {
  return (
    <div className="min-h-screen bg-gray-100 flex flex-col">
      {/* Header */}
      <header className="bg-blue-900 text-white p-4 shadow-md">
        <div className="container mx-auto px-4 flex justify-between items-center">
          <Link to="/" className="text-2xl font-bold no-underline text-white">Gogen UI</Link>
          <nav>
            <ul className="flex space-x-4">
              <li><Link to="/" className="hover:text-cyan-400 transition-colors">Home</Link></li>
            </ul>
          </nav>
        </div>
      </header>

      {/* Main content */}
      {children}

      {/* Footer */}
      <footer className="bg-gray-800 text-white p-4 mt-auto">
        <div className="container mx-auto px-4">
          <p>&copy; 2023 Gogen UI. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
};

export default Layout; 