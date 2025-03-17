import { Link } from 'react-router-dom';

const Header = () => {
  return (
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
  );
};

export default Header; 