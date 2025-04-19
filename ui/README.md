# Gogen UI

A modern React-based user interface for interacting with the Gogen API. This UI allows users to view, search, and execute Gogen configurations directly in the browser.

## Features

- View a list of all available Gogen configurations
- Search and filter configurations
- View detailed information about each configuration
- Execute configurations in the browser using WebAssembly
- Display execution results in terminal or structured format

## Technology Stack

- **Frontend Framework**: React with TypeScript
- **Styling**: Tailwind CSS
- **State Management**: React Context API and hooks
- **Routing**: React Router
- **API Client**: Axios
- **Terminal Emulation**: xterm.js
- **Build Tool**: Vite

## Getting Started

### Prerequisites

- Node.js (v14 or later)
- npm (v6 or later)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/coccyx/gogen.git
   cd gogen/ui
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

4. Open your browser and navigate to `http://localhost:3000`

## Building for Production

To build the application for production:

```bash
npm run build
```

The built files will be in the `dist` directory.

## Project Structure

- `src/` - Source code
  - `api/` - API client and interfaces
  - `components/` - Reusable UI components
  - `pages/` - Page components
  - `hooks/` - Custom React hooks
  - `utils/` - Utility functions
  - `types/` - TypeScript type definitions
  - `context/` - React context providers
  - `assets/` - Static assets

## Styling with Tailwind CSS

This project uses Tailwind CSS for styling. We've extended the default Tailwind configuration with custom colors and component classes based on the Cribl design system.

### Custom Colors

The following custom colors are defined in `tailwind.config.js`:

- `cribl-primary`: #261926 (Dark purple)
- `cribl-cyan`: #00F9BB (Bright teal/cyan)
- `cribl-purple`: #7f66ff (Bright purple)
- `cribl-pink`: #ff3399 (Bright pink)
- `cribl-red`: #f25e65 (Red)
- `cribl-white`: #ffffff (White)
- `cribl-gray`: #f1f1f1 (Light gray)
- `cribl-dark`: #1A1A1A (Dark background)
- `cribl-blue`: #3A85F7 (Blue)
- `cribl-orange`: #FF6B35 (Orange)
- `cribl-light-gray`: #F5F5F5 (Light gray background)

### Custom Component Classes

We've defined several reusable component classes in `src/index.css` using Tailwind's `@apply` directive:

- `.btn-primary`: Primary button style using cribl-cyan
- `.btn-secondary`: Secondary button style using cribl-purple
- `.btn-outline`: Outline button style with cribl-cyan border
- `.card`: Card component with white background and shadow
- `.container-custom`: Container with responsive padding
- `.nav-link`: Navigation link with hover effect
- `.section-heading`: Section heading with consistent styling

### Usage Example

```jsx
<button className="btn-primary">Click Me</button>
<div className="card">
  <h3 className="text-cribl-primary">Card Title</h3>
  <p>Card content</p>
</div>
<a href="#" className="nav-link">Navigation Link</a>
```

## License

This project is licensed under the same license as the Gogen project. 