import { useState, useEffect, useRef } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { Terminal } from 'xterm';
import gogenApi, { Configuration } from '../api/gogenApi';
import gogenWasm from '../api/gogenWasm';
import LoadingSpinner from '../components/LoadingSpinner';

// Import xterm.css directly
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

const ExecutionPage = () => {
  const { owner, configName } = useParams<{ owner: string; configName: string }>();
  const navigate = useNavigate();
  const [configuration, setConfiguration] = useState<Configuration | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);
  const [outputMode, setOutputMode] = useState<'terminal' | 'structured'>('terminal');
  const [structuredOutput, setStructuredOutput] = useState<any[]>([]);
  const [eventCount, setEventCount] = useState<number>(100);
  
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminalInstance = useRef<Terminal | null>(null);

  // Fetch configuration details
  useEffect(() => {
    const fetchConfiguration = async () => {
      if (!owner || !configName) {
        setError('Configuration details are incomplete');
        setIsLoading(false);
        return;
      }

      try {
        setIsLoading(true);
        const fullConfigName = `${owner}/${configName}`;
        const data = await gogenApi.getConfiguration(fullConfigName);
        setConfiguration(data);
        setError(null);
      } catch (err) {
        setError('Failed to load configuration. Please try again later.');
        console.error(`Error fetching configuration ${owner}/${configName}:`, err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchConfiguration();
  }, [owner, configName]);

  // Initialize terminal
  useEffect(() => {
    console.log('Terminal initialization effect running', {
      terminalRefExists: !!terminalRef.current,
      terminalInstanceExists: !!terminalInstance.current,
      outputMode,
      isCssLoaded: isCssLoaded('xterm.css'),
      isTerminalClassDefined: typeof Terminal !== 'undefined'
    });
    
    // Only initialize terminal after loading is complete and we're in terminal mode
    if (!isLoading && terminalRef.current && !terminalInstance.current && outputMode === 'terminal') {
      console.log('Creating new terminal instance');
      
      const initializeTerminal = async () => {
        try {
          // Ensure the CSS is loaded
          if (!isCssLoaded('xterm.css')) {
            console.warn('xterm.css is not loaded, dynamically adding it');
            const link = document.createElement('link');
            link.rel = 'stylesheet';
            link.type = 'text/css';
            link.href = 'https://unpkg.com/xterm@5.3.0/css/xterm.css';
            document.head.appendChild(link);
          }
          
          // Check if Terminal class is available
          if (typeof Terminal === 'undefined') {
            console.warn('Terminal class is not defined, attempting to load xterm.js dynamically');
            try {
              await loadScript('https://unpkg.com/xterm@5.3.0/lib/xterm.js');
              console.log('xterm.js loaded dynamically');
            } catch (error) {
              console.error('Failed to load xterm.js dynamically:', error);
              setError('Terminal component failed to load. Please refresh the page or try again later.');
              return;
            }
          }
          
          // Small delay to ensure DOM is ready
          setTimeout(() => {
            try {
              // Re-check if Terminal is defined after dynamic loading
              if (typeof Terminal === 'undefined') {
                throw new Error('Terminal class is still not defined after dynamic loading');
              }
              
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
              
              term.open(terminalRef.current!);
              term.writeln('Gogen Execution Terminal');
              term.writeln('------------------------');
              term.writeln('Click "Execute" to run the configuration.');
              term.writeln('');
              
              terminalInstance.current = term;
              console.log('Terminal instance created successfully');
            } catch (error) {
              console.error('Error creating terminal:', error);
              setError(`Failed to initialize terminal: ${error instanceof Error ? error.message : 'Unknown error'}`);
            }
          }, 100);
        } catch (error) {
          console.error('Error in terminal initialization:', error);
          setError(`Terminal initialization error: ${error instanceof Error ? error.message : 'Unknown error'}`);
        }
      };
      
      initializeTerminal();
    }

    return () => {
      if (terminalInstance.current) {
        console.log('Disposing terminal instance');
        terminalInstance.current.dispose();
        terminalInstance.current = null;
      }
    };
  }, [outputMode, isLoading]);

  // Execute configuration
  const executeConfiguration = async () => {
    if (!configuration) return;
    
    setIsExecuting(true);
    
    try {
      if (outputMode === 'terminal' && terminalInstance.current) {
        terminalInstance.current.clear();
        terminalInstance.current.writeln(`Executing configuration: ${configuration.gogen}`);
        terminalInstance.current.writeln('------------------------');
        terminalInstance.current.writeln('');
        terminalInstance.current.writeln('Configuration will be passed to the WASM module via a virtual file system.');
        terminalInstance.current.writeln('');
        
        try {
          // Execute the configuration using the WASM module with streaming output
          await gogenWasm.executeConfiguration(
            configuration,
            eventCount,
            (line) => {
              if (terminalInstance.current) {
                terminalInstance.current.writeln(line);
              }
            }
          );
          
          terminalInstance.current.writeln('Execution completed successfully.');
        } catch (execError: any) {
          console.error('Error during execution:', execError);
          terminalInstance.current.writeln(`\x1b[31mExecution error: ${execError.message || 'Unknown error'}\x1b[0m`);
          
          // Generate some mock output for demonstration purposes
          terminalInstance.current.writeln('\x1b[33mGenerating mock output for demonstration purposes...\x1b[0m');
          for (let i = 0; i < 5; i++) {
            const mockEvent = JSON.stringify({
              _raw: `Mock event ${i + 1}`,
              timestamp: new Date().toISOString(),
              source: configuration.gogen,
              index: i
            }, null, 2);
            terminalInstance.current.writeln(mockEvent);
          }
        }
      } else if (outputMode === 'structured') {
        try {
          // For structured output, get all results at once
          const results = await gogenWasm.executeConfiguration(configuration, eventCount);
          
          // Parse the results as JSON objects if needed
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
          console.error('Error during structured execution:', execError);
          
          // Generate mock results for demonstration purposes
          const mockResults = [];
          for (let i = 0; i < 5; i++) {
            mockResults.push({
              _raw: `Mock event ${i + 1}`,
              timestamp: new Date().toISOString(),
              source: configuration.gogen,
              index: i
            });
          }
          setStructuredOutput(mockResults);
        }
      }
    } catch (error: any) {
      console.error('Error executing configuration:', error);
      if (terminalInstance.current) {
        terminalInstance.current.writeln(`\x1b[31mError: ${error.message || 'Unknown error'}\x1b[0m`);
        terminalInstance.current.writeln('\x1b[31mCheck the browser console for more details.\x1b[0m');
      }
    } finally {
      setIsExecuting(false);
    }
  };

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (error || !configuration) {
    return (
      <div className="text-center py-8">
        <p className="text-red-600 mb-4">{error || 'Configuration not found'}</p>
        <button
          onClick={() => navigate('/configurations')}
          className="btn-primary"
        >
          Back to Configurations
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-cribl-primary">
          Execute: {configuration.gogen}
        </h1>
        <div className="flex space-x-4">
          <Link to={`/configurations/${owner}/${configName}`} className="btn-primary">
            Back to Details
          </Link>
        </div>
      </div>

      <div className="mb-6">
        <h2 className="text-xl font-semibold mb-2">Execution Options</h2>
        <div className="bg-white rounded-lg shadow-md p-4">
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Output Mode
            </label>
            <div className="flex space-x-4">
              <button
                onClick={() => setOutputMode('terminal')}
                className={`px-4 py-2 rounded-md ${
                  outputMode === 'terminal'
                    ? 'bg-cribl-purple text-white'
                    : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                }`}
              >
                Terminal Output
              </button>
              <button
                onClick={() => setOutputMode('structured')}
                className={`px-4 py-2 rounded-md ${
                  outputMode === 'structured'
                    ? 'bg-cribl-purple text-white'
                    : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                }`}
              >
                Structured Output
              </button>
            </div>
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Event Count
            </label>
            <input
              type="number"
              min="1"
              max="10000"
              value={eventCount}
              onChange={(e) => setEventCount(Math.max(1, parseInt(e.target.value) || 100))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-cribl-purple focus:border-cribl-purple"
            />
          </div>

          <button
            onClick={executeConfiguration}
            disabled={isExecuting}
            className={`btn-secondary w-full ${
              isExecuting ? 'opacity-70 cursor-not-allowed' : ''
            }`}
          >
            {isExecuting ? 'Executing...' : 'Execute Configuration'}
          </button>
        </div>
      </div>

      <div className="mb-6">
        <h2 className="text-xl font-semibold mb-2">Output</h2>
        <div className="bg-white rounded-lg shadow-md p-4">
          {outputMode === 'terminal' ? (
            <div 
              ref={terminalRef} 
              className="h-96 rounded-md overflow-hidden border border-gray-300"
              style={{ 
                backgroundColor: '#f8f9fa', 
                position: 'relative',
                minHeight: '384px' // Ensure minimum height even if empty
              }}
            >
              {!terminalInstance.current && (
                <div className="absolute inset-0 flex items-center justify-center text-gray-500">
                  Initializing terminal...
                </div>
              )}
            </div>
          ) : (
            <div className="h-96 overflow-auto">
              {structuredOutput.length === 0 ? (
                <div className="flex items-center justify-center h-full text-gray-500">
                  Execute the configuration to see structured output
                </div>
              ) : (
                <div className="space-y-2">
                  {structuredOutput.map((event, index) => (
                    <div key={index} className="border border-gray-200 rounded-md p-3 hover:bg-gray-50">
                      <div className="flex justify-between">
                        <span className="font-medium">Event {index + 1}</span>
                        <span className="text-sm text-gray-500">{event.timestamp}</span>
                      </div>
                      <pre className="mt-2 text-sm overflow-x-auto">
                        {JSON.stringify(event, null, 2)}
                      </pre>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ExecutionPage;