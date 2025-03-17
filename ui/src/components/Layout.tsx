import Header from './Header';
import Footer from './Footer';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout = ({ children }: LayoutProps) => {
  return (
    <div className="min-h-screen bg-gray-100 flex flex-col" data-testid="layout-container">
      <Header />
      {/* Main content */}
      {children}
      <Footer />
    </div>
  );
};

export default Layout; 