import { render, screen, fireEvent } from '../utils/test-utils';
import JsonBrowser from './JsonBrowser';

describe('JsonBrowser', () => {
  const sampleData = [
    { host: 'web01', status: 200, path: '/api/health' },
    { host: 'web02', status: 500, path: '/api/login' },
    { _raw: 'ts=2024-01-01 message=Hello World' },
  ];

  test('shows empty state when data is empty', () => {
    render(<JsonBrowser data={[]} />);
    expect(screen.getByText(/no output yet/i)).toBeInTheDocument();
  });

  test('renders all events collapsed by default', () => {
    render(<JsonBrowser data={sampleData} />);

    expect(screen.getByText('[0]')).toBeInTheDocument();
    expect(screen.getByText('[1]')).toBeInTheDocument();
    expect(screen.getByText('[2]')).toBeInTheDocument();

    // Collapsed summaries should be visible
    expect(screen.getByText(/host: "web01"/)).toBeInTheDocument();
    expect(screen.getByText(/host: "web02"/)).toBeInTheDocument();

    // Full JSON should not be visible
    expect(screen.queryByText(/"path": "\/api\/health"/)).not.toBeInTheDocument();
  });

  test('shows _raw value as summary for raw events', () => {
    render(<JsonBrowser data={sampleData} />);
    expect(screen.getByText('ts=2024-01-01 message=Hello World')).toBeInTheDocument();
  });

  test('truncates summary to 3 keys with ellipsis for objects with more', () => {
    const data = [{ a: 1, b: 2, c: 3, d: 4 }];
    render(<JsonBrowser data={data} />);
    expect(screen.getByText(/\.\.\./)).toBeInTheDocument();
  });

  test('does not show ellipsis for objects with 3 or fewer keys', () => {
    const data = [{ a: 1, b: 2 }];
    render(<JsonBrowser data={data} />);
    expect(screen.queryByText(/\.\.\./)).not.toBeInTheDocument();
  });

  test('expands a single event on click', () => {
    render(<JsonBrowser data={sampleData} />);

    fireEvent.click(screen.getByText('[0]'));

    // Expanded JSON should now be visible
    expect(screen.getByText(/"path": "\/api\/health"/)).toBeInTheDocument();

    // Other events should still be collapsed
    expect(screen.queryByText(/"path": "\/api\/login"/)).not.toBeInTheDocument();
  });

  test('collapses an expanded event on second click', () => {
    render(<JsonBrowser data={sampleData} />);

    fireEvent.click(screen.getByText('[0]'));
    expect(screen.getByText(/"path": "\/api\/health"/)).toBeInTheDocument();

    fireEvent.click(screen.getByText('[0]'));
    expect(screen.queryByText(/"path": "\/api\/health"/)).not.toBeInTheDocument();
  });

  test('expand all button expands every event', () => {
    render(<JsonBrowser data={sampleData} />);

    const expandAllButton = screen.getByTitle('Expand all');
    fireEvent.click(expandAllButton);

    expect(screen.getByText(/"path": "\/api\/health"/)).toBeInTheDocument();
    expect(screen.getByText(/"path": "\/api\/login"/)).toBeInTheDocument();
  });

  test('collapse all button collapses every event', () => {
    render(<JsonBrowser data={sampleData} />);

    // Expand all first
    fireEvent.click(screen.getByTitle('Expand all'));
    expect(screen.getByText(/"path": "\/api\/health"/)).toBeInTheDocument();

    // Now collapse all
    fireEvent.click(screen.getByTitle('Collapse all'));
    expect(screen.queryByText(/"path": "\/api\/health"/)).not.toBeInTheDocument();
    expect(screen.queryByText(/"path": "\/api\/login"/)).not.toBeInTheDocument();
  });

  test('expand all button toggles title based on state', () => {
    render(<JsonBrowser data={sampleData} />);

    expect(screen.getByTitle('Expand all')).toBeInTheDocument();

    fireEvent.click(screen.getByTitle('Expand all'));
    expect(screen.getByTitle('Collapse all')).toBeInTheDocument();

    fireEvent.click(screen.getByTitle('Collapse all'));
    expect(screen.getByTitle('Expand all')).toBeInTheDocument();
  });
});
