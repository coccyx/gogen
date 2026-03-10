import { useState, useEffect, useRef } from 'react';
import { Terminal } from 'xterm';
import gogenWasm, { ExecutionParams } from '../api/gogenWasm';
import { Configuration } from '../api/gogenApi';
import JsonBrowser from './JsonBrowser';
import 'xterm/css/xterm.css';

// Helper function to check if a CSS file is loaded
const isCssLoaded = (href: string): boolean => {
  const links = document.getElementsByTagName('link');
  for (let i = 0; i < links.length; i++) {
    if (links[i].href.includes(href)) {
      return true;
    }
  }
  return false;
};

// Helper function to dynamically load a script
const loadScript = (src: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    const script = document.createElement('script');
    script.src = src;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error(`Failed to load script: ${src}`));
    document.head.appendChild(script);
  });
};

interface ExecutionComponentProps {
  configuration: Configuration;
}

const ExecutionComponent: React.FC<ExecutionComponentProps> = ({ configuration }) => {
  const [isExecuting, setIsExecuting] = useState(false);
  const [activeTab, setActiveTab] = useState<'terminal' | 'structured'>('terminal');
  const [structuredOutput, setStructuredOutput] = useState<any[]>([]);
  const [eventCount, setEventCount] = useState<number>(1);
  const [intervals, setIntervals] = useState<number>(5);
  const [intervalSeconds, setIntervalSeconds] = useState<number>(1);
  const [outputTemplate, setOutputTemplate] = useState<'raw' | 'json' | 'configured'>('raw');
  const [error, setError] = useState<string | null>(null);

  const terminalRef = useRef<HTMLDivElement>(null);
  const terminalInstance = useRef<Terminal | null>(null);

  // Initialize terminal once on mount
  useEffect(() => {
    const initializeTerminal = async () => {
      try {
        // Ensure the CSS is loaded
        if (!isCssLoaded('xterm.css')) {
          const link = document.createElement('link');
          link.rel = 'stylesheet';
          link.type = 'text/css';
          link.href = 'https://unpkg.com/xterm@5.3.0/css/xterm.css';
          document.head.appendChild(link);
        }

        // Check if Terminal class is available
        if (typeof Terminal === 'undefined') {
          try {
            await loadScript('https://unpkg.com/xterm@5.3.0/lib/xterm.js');
          } catch (error) {
            setError('Terminal component failed to load. Please refresh the page or try again later.');
            return;
          }
        }

        // Create new terminal instance
        if (terminalRef.current && !terminalInstance.current) {
          const term = new Terminal({
            cursorBlink: false,
            disableStdin: true,
            rows: 20,
            cols: 100,
            theme: {
              background: '#0d1117',
              foreground: '#e6edf3',
              cursor: '#e6edf3',
              cursorAccent: '#0d1117',
              selectionBackground: '#30363d',
              black: '#0d1117',
              red: '#f85149',
              green: '#3fb950',
              yellow: '#d29922',
              blue: '#58a6ff',
              magenta: '#bc8cff',
              cyan: '#39c5cf',
              white: '#e6edf3',
              brightBlack: '#8b949e',
              brightRed: '#f85149',
              brightGreen: '#3fb950',
              brightYellow: '#d29922',
              brightBlue: '#58a6ff',
              brightMagenta: '#bc8cff',
              brightCyan: '#39c5cf',
              brightWhite: '#ffffff',
            },
            fontFamily: 'JetBrains Mono, ui-monospace, SFMono-Regular, monospace',
            fontSize: 13,
          });

          term.open(terminalRef.current);
          terminalInstance.current = term;
        }
      } catch (error) {
        setError(`Terminal initialization error: ${error instanceof Error ? error.message : 'Unknown error'}`);
      }
    };

    initializeTerminal();

    return () => {
      if (terminalInstance.current) {
        terminalInstance.current.dispose();
        terminalInstance.current = null;
      }
    };
  }, []);

  // Execute configuration - populates both terminal and structured output
  const executeConfiguration = async () => {
    setIsExecuting(true);
    setError(null);

    const executionParams: ExecutionParams = {
      eventCount,
      intervals,
      intervalSeconds,
      outputTemplate
    };

    try {
      if (terminalInstance.current) {
        terminalInstance.current.clear();
      }

      const results = await gogenWasm.executeConfiguration(
        configuration,
        executionParams,
        (line) => {
          if (terminalInstance.current) {
            terminalInstance.current.writeln(line);
          }
        }
      );

      const parsedResults = results.map((line) => {
        if (typeof line === 'object') return line;
        try {
          return JSON.parse(line);
        } catch (e) {
          return { _raw: line };
        }
      });
      setStructuredOutput(parsedResults);
    } catch (execError: any) {
      const msg = execError.message || 'Unknown error';
      setError(`Execution error: ${msg}`);
      if (terminalInstance.current) {
        terminalInstance.current.writeln(`\x1b[31mExecution error: ${msg}\x1b[0m`);
      }
    } finally {
      setIsExecuting(false);
    }
  };

  return (
    <div className="mt-6 border-t border-term-border pt-6">
      <h2 className="text-lg font-semibold text-term-text mb-4">Execute Configuration</h2>

      <div className="mb-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 items-end">
        <div>
          <label htmlFor="intervals" className="block text-sm font-medium text-term-text-muted mb-1">
            Intervals
          </label>
          <input
            type="text"
            id="intervals"
            value={intervals}
            onChange={(e) => {
              const value = e.target.value;
              if (value === '' || /^\d+$/.test(value)) {
                setIntervals(value === '' ? 1 : parseInt(value, 10));
              }
            }}
            className="input"
          />
        </div>

        <div>
          <label htmlFor="eventCount" className="block text-sm font-medium text-term-text-muted mb-1">
            Events Per Interval
          </label>
          <input
            type="text"
            id="eventCount"
            value={eventCount}
            onChange={(e) => {
              const value = e.target.value;
              if (value === '' || /^\d+$/.test(value)) {
                setEventCount(value === '' ? 1 : parseInt(value, 10));
              }
            }}
            className="input"
          />
        </div>

        <div>
          <label htmlFor="intervalSeconds" className="block text-sm font-medium text-term-text-muted mb-1">
            Interval (in Seconds)
          </label>
          <input
            type="text"
            id="intervalSeconds"
            value={intervalSeconds}
            onChange={(e) => {
              const value = e.target.value;
              if (value === '' || /^\d+$/.test(value)) {
                setIntervalSeconds(value === '' ? 1 : parseInt(value, 10));
              }
            }}
            className="input"
          />
        </div>

        <div>
          <label htmlFor="outputTemplate" className="block text-sm font-medium text-term-text-muted mb-1">
            Output Template
          </label>
          <select
            id="outputTemplate"
            value={outputTemplate}
            onChange={(e) => setOutputTemplate(e.target.value as 'raw' | 'json' | 'configured')}
            className="input"
          >
            <option value="raw">Raw</option>
            <option value="json">JSON</option>
            <option value="configured">As configured</option>
          </select>
        </div>

        <div className="lg:col-span-4">
          <button
            onClick={executeConfiguration}
            disabled={isExecuting}
            className={`w-full px-4 py-1.5 rounded font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-term-green ${
              isExecuting
                ? 'bg-term-bg-muted text-term-text-muted cursor-not-allowed'
                : 'bg-term-green hover:bg-opacity-90 text-term-bg'
            }`}
          >
            {isExecuting ? 'Executing...' : 'Execute'}
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-4 error-box">
          {error}
        </div>
      )}

      <div className="flex border-b border-term-border mb-0">
        <button
          onClick={() => setActiveTab('terminal')}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
            activeTab === 'terminal'
              ? 'border-term-cyan text-term-cyan'
              : 'border-transparent text-term-text-muted hover:text-term-text'
          }`}
        >
          Terminal
        </button>
        <button
          onClick={() => setActiveTab('structured')}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
            activeTab === 'structured'
              ? 'border-term-cyan text-term-cyan'
              : 'border-transparent text-term-text-muted hover:text-term-text'
          }`}
        >
          Structured
        </button>
      </div>

      <div className={activeTab === 'terminal' ? '' : 'hidden'}>
        <div className="border border-t-0 border-term-border rounded-b bg-term-bg p-2" data-testid="terminal-container">
          <div className="terminal" ref={terminalRef} />
        </div>
      </div>
      <div className={activeTab === 'structured' ? '' : 'hidden'}>
        <div className="border border-t-0 border-term-border rounded-b bg-term-bg-elevated p-3">
          <JsonBrowser data={structuredOutput} />
        </div>
      </div>
    </div>
  );
};

export default ExecutionComponent;
