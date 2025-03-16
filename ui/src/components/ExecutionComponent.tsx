import { useState, useEffect, useRef } from 'react';
import { Terminal } from 'xterm';
import gogenWasm, { ExecutionParams } from '../api/gogenWasm';
import { Configuration } from '../api/gogenApi';
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
  const [outputMode, setOutputMode] = useState<'terminal' | 'structured'>('terminal');
  const [structuredOutput, setStructuredOutput] = useState<any[]>([]);
  const [eventCount, setEventCount] = useState<number>(100);
  const [intervals, setIntervals] = useState<number>(1);
  const [intervalSeconds, setIntervalSeconds] = useState<number>(1);
  const [outputTemplate, setOutputTemplate] = useState<'raw' | 'json' | 'configured'>('raw');
  const [error, setError] = useState<string | null>(null);
  
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminalInstance = useRef<Terminal | null>(null);

  // Initialize terminal
  useEffect(() => {
    // Clean up function to properly dispose of terminal
    const cleanupTerminal = () => {
      if (terminalInstance.current) {
        terminalInstance.current.dispose();
        terminalInstance.current = null;
      }
      if (terminalRef.current) {
        while (terminalRef.current.firstChild) {
          terminalRef.current.removeChild(terminalRef.current.firstChild);
        }
      }
    };

    // Only initialize if we're in terminal mode and don't have an instance
    if (outputMode === 'terminal') {
      // Clean up any existing terminal first
      cleanupTerminal();

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
                background: '#f8f9fa',
                foreground: '#212529',
              },
            });
            
            term.open(terminalRef.current);
            
            terminalInstance.current = term;
          }
        } catch (error) {
          setError(`Terminal initialization error: ${error instanceof Error ? error.message : 'Unknown error'}`);
        }
      };
      
      initializeTerminal();
    } else {
      // Clean up terminal when switching to structured mode
      cleanupTerminal();
    }

    // Cleanup on unmount or mode change
    return cleanupTerminal;
  }, [outputMode]);

  // Execute configuration
  const executeConfiguration = async () => {
    setIsExecuting(true);
    
    const executionParams: ExecutionParams = {
      eventCount,
      intervals,
      intervalSeconds,
      outputTemplate
    };
    
    try {
      if (outputMode === 'terminal' && terminalInstance.current) {
        terminalInstance.current.clear();
        
        try {
          await gogenWasm.executeConfiguration(
            configuration,
            executionParams,
            (line) => {
              if (terminalInstance.current) {
                terminalInstance.current.writeln(line);
              }
            }
          );
          
        } catch (execError: any) {
          terminalInstance.current.writeln(`\x1b[31mExecution error: ${execError.message || 'Unknown error'}\x1b[0m`);
        }
      } else if (outputMode === 'structured') {
        try {
          const results = await gogenWasm.executeConfiguration(configuration, executionParams);
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
          setError(`Error during structured execution: ${execError.message || 'Unknown error'}`);
        }
      }
    } catch (error: any) {
      setError(`Error: ${error.message || 'Unknown error'}`);
      if (terminalInstance.current) {
        terminalInstance.current.writeln(`\x1b[31mError: ${error.message || 'Unknown error'}\x1b[0m`);
        terminalInstance.current.writeln('\x1b[31mCheck the browser console for more details.\x1b[0m');
      }
    } finally {
      setIsExecuting(false);
    }
  };

  return (
    <div className="mt-8 border-t pt-8">
      <h2 className="text-xl font-semibold mb-4">Execute Configuration</h2>
      
      <div className="mb-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 items-end">
        <div>
          <label htmlFor="intervals" className="block text-sm font-medium text-gray-700">
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
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm bg-white text-gray-900"
          />
        </div>

        <div>
          <label htmlFor="eventCount" className="block text-sm font-medium text-gray-700">
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
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm bg-white text-gray-900"
          />
        </div>

        <div>
          <label htmlFor="intervalSeconds" className="block text-sm font-medium text-gray-700">
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
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm bg-white text-gray-900"
          />
        </div>
        
        <div>
          <label htmlFor="outputTemplate" className="block text-sm font-medium text-gray-700">
            Output Template
          </label>
          <select
            id="outputTemplate"
            value={outputTemplate}
            onChange={(e) => setOutputTemplate(e.target.value as 'raw' | 'json' | 'configured')}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm bg-white text-gray-900"
          >
            <option value="raw">Raw</option>
            <option value="json">JSON</option>
            <option value="configured">As configured</option>
          </select>
        </div>
        
        <div>
          <label htmlFor="outputMode" className="block text-sm font-medium text-gray-700">
            Output Mode
          </label>
          <select
            id="outputMode"
            value={outputMode}
            onChange={(e) => setOutputMode(e.target.value as 'terminal' | 'structured')}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm bg-white text-gray-900"
          >
            <option value="terminal">Terminal</option>
            <option value="structured">Structured</option>
          </select>
        </div>

        <div className="lg:col-span-5">
          <button
            onClick={executeConfiguration}
            disabled={isExecuting}
            className={`w-full px-4 py-2 rounded-md ${
              isExecuting
                ? 'bg-gray-400 cursor-not-allowed'
                : 'bg-blue-800 hover:bg-blue-700'
            } text-white font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500`}
          >
            {isExecuting ? 'Executing...' : 'Execute'}
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-4 p-4 bg-red-100 text-red-700 rounded-md">
          {error}
        </div>
      )}

      {outputMode === 'terminal' ? (
        <div ref={terminalRef} className="border rounded-md p-4 bg-white shadow-md" />
      ) : (
        <div className="border rounded-md p-4 bg-white shadow-md">
          <pre className="overflow-x-auto">
            <code>{JSON.stringify(structuredOutput, null, 2)}</code>
          </pre>
        </div>
      )}
    </div>
  );
};

export default ExecutionComponent; 