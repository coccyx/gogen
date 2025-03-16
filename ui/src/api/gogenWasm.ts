import { Configuration } from './gogenApi';
import { init, Wasmer, Directory } from '@wasmer/sdk';

// Define the Go type from wasm_exec.js
declare global {
  interface Window {
    Go: any;
  }
}

/**
 * Execute a Gogen configuration using WebAssembly
 * 
 * @param configuration The configuration to execute
 * @param count The number of events to generate
 * @param onOutput Optional callback for streaming output
 * @returns Array of output lines
 */
export const executeConfiguration = async (
  configuration: Configuration,
  count: number,
  onOutput?: (line: string) => void
): Promise<string[]> => {
  try {
    // Enable Wasmer debug logging if available
    try {
      console.log('Enabling Wasmer debug logging...');
      
      // Check Wasmer version if available
      if ((Wasmer as any).version) {
        console.log('Wasmer SDK version:', (Wasmer as any).version);
      } else {
        console.log('Wasmer SDK version not available');
      }
      
      // Check if debug logging is available in this version of Wasmer SDK
      if (typeof (Wasmer as any).setDebug === 'function') {
        (Wasmer as any).setDebug(true);
        console.log('Wasmer debug logging enabled via setDebug');
      } else if (typeof (Wasmer as any).enableDebugMode === 'function') {
        (Wasmer as any).enableDebugMode();
        console.log('Wasmer debug logging enabled via enableDebugMode');
      } else {
        console.log('Wasmer debug logging not available in this SDK version');
      }
    } catch (logError) {
      console.warn('Failed to enable Wasmer debug logging:', logError);
    }

    // Check if WebAssembly is supported
    if (typeof WebAssembly === 'undefined') {
      throw new Error('WebAssembly is not supported in this browser');
    }
    
    // Check for Cross-Origin Isolation support
    if (typeof crossOriginIsolated === 'undefined' || !crossOriginIsolated) {
      console.warn('Cross-Origin Isolation is not enabled. Wasmer requires this for SharedArrayBuffer support.');
      console.warn('See https://docs.wasmer.io/javascript-sdk/explainers/troubleshooting#sharedarraybuffer-and-cross-origin-isolation');
      
      if (onOutput) {
        onOutput('Warning: Cross-Origin Isolation is not enabled. This may cause issues with WebAssembly execution.');
        onOutput('Please visit the site using a URL that has Cross-Origin Isolation enabled.');
      }
    }

    console.log('Initializing Wasmer...');
    await init();
    console.log('Wasmer initialized successfully');

    // Convert the configuration to a string (YAML or JSON)
    const fs = new Directory();
    const configString = configuration.config || '';
    await fs.writeFile('/config.yml', configString);
    
    // Set up arguments for gogen
    const args = ['gogen', '-vv', '-c', '/config.yml', 'gen', '-ei 1'];
    
    // Add count if specified
    if (count > 0) {
      args.push('-c', count.toString());
    }
    
    // Log the arguments and configuration being used
    console.log('Executing gogen with arguments:', args);
    console.log('Configuration:', configString.substring(0, 200) + (configString.length > 200 ? '...' : ''));
    
    try {
      // Fetch the WASM module
      console.log('Fetching WASM module...');
      // Use an absolute path to ensure we're getting the right file
      const wasmUrl = new URL('/gogen.wasm', window.location.origin).href;
      console.log('WASM URL:', wasmUrl);
      const response = await fetch(wasmUrl);
      
      // Add detailed error handling for the fetch response
      if (!response.ok) {
        console.error(`Failed to fetch WASM module: ${response.status} ${response.statusText}`);
        console.error(`Response URL: ${response.url}`);
        
        // Try to read the response content to see what was returned instead
        const responseText = await response.text();
        console.error(`Response content (first 200 chars): ${responseText.substring(0, 200)}`);
        
        throw new Error(`Failed to fetch WASM module: ${response.status} ${response.statusText}`);
      }
      
      const wasmBuffer = await response.arrayBuffer()
      console.log('WASM module fetched, instantiating...');
      
      // Debug: Check the first few bytes of the WASM file
      const firstBytes = new Uint8Array(wasmBuffer.slice(0, 8));
      console.log('First bytes of WASM file:', 
        Array.from(firstBytes).map(b => b.toString(16).padStart(2, '0')).join(' '));
      
      // Check for the WebAssembly magic number (0x0061736d or \0asm)
      if (firstBytes[0] !== 0x00 || firstBytes[1] !== 0x61 || 
          firstBytes[2] !== 0x73 || firstBytes[3] !== 0x6d) {
        console.error('Invalid WASM file: Missing WebAssembly magic number');
        console.error('First 4 bytes:', firstBytes.slice(0, 4));
        
        // Try to decode as text to see what we got instead
        const decoder = new TextDecoder('utf-8');
        const textContent = decoder.decode(wasmBuffer.slice(0, 100));
        console.error('Content as text:', textContent);
        
        throw new Error('Invalid WASM file: Missing WebAssembly magic number');
      }

      // Instantiate the Wasmer module
      try {
        console.log('Creating Wasmer instance from WASM bytes...');
        
        // Check if the WASM module has the expected format for Wasmer
        // This is a basic check - Wasmer might need specific WASM features
        if (wasmBuffer.byteLength < 8) {
          throw new Error('WASM file is too small to be valid');
        }
        
        // Log the WASM module size
        console.log(`WASM module size: ${wasmBuffer.byteLength} bytes`);
        
        // Try to instantiate the module with Wasmer
        let pkg;
        try {
          pkg = await Wasmer.fromWasm(new Uint8Array(wasmBuffer));
          console.log('Wasmer package created successfully');
          
          // Inspect the package to see what's available
          console.log('Inspecting Wasmer package...');
          try {
            // Log available exports and properties
            console.log('Package properties:', Object.keys(pkg as object));
            
            // Check if we can access exports directly
            if ((pkg as any).exports) {
              console.log('Package exports:', Object.keys((pkg as any).exports));
            }
            
            // Check for available functions
            if ((pkg as any).functions) {
              console.log('Package functions:', Object.keys((pkg as any).functions));
            }
            
            // Check for available namespaces
            if ((pkg as any).namespaces) {
              console.log('Package namespaces:', Object.keys((pkg as any).namespaces));
            }
          } catch (inspectError) {
            console.warn('Error inspecting package:', inspectError);
          }
          
          console.log('Accessing entrypoint...');
          const instance = await pkg.entrypoint;

          if (!instance) {
            throw new Error('Failed to instantiate WASM module: No entrypoint found');
          }

          // Inspect the instance
          console.log('Inspecting instance...');
          try {
            console.log('Instance properties:', Object.keys(instance as object));
            
            // Check for available methods
            const methods = Object.getOwnPropertyNames(Object.getPrototypeOf(instance));
            console.log('Instance methods:', methods);
            
            // Check if run method exists and what parameters it accepts
            if (typeof instance.run === 'function') {
              console.log('Run method exists');
              console.log('Run method toString:', instance.run.toString().substring(0, 200));
            } else {
              console.warn('Run method does not exist on instance');
            }
          } catch (inspectError) {
            console.warn('Error inspecting instance:', inspectError);
          }
          
          console.log('WASM module instantiated successfully');
          
          // Run the Go instance with the WASM module
          console.log('Running Go instance with args:', args);
          
          // Create a callback for streaming output if provided
          const outputCallback = onOutput ? 
            (line: string) => {
              console.log('Output:', line);
              onOutput(line);
            } : undefined;
          
          // Set a timeout to detect if execution is hanging
          const executionTimeout = setTimeout(() => {
            console.warn('Execution seems to be taking a long time (10 seconds). This might indicate a hang.');
            if (onOutput) {
              onOutput('Warning: Execution is taking longer than expected. This might indicate an issue with the WASM module.');
            }
          }, 10000);
          
          // Run the instance with the provided arguments
          console.log('Starting WASM execution...');
          
          // Add environment variables for debugging
          const env: Record<string, string> = {
            // Common Go environment variables
            'GOOS': 'js',
            'GOARCH': 'wasm',
            // Debug flags
            'WASMER_DEBUG': '1',
            'WASMER_BACKTRACE': '1',
            'RUST_BACKTRACE': 'full',
            // Go specific debug flags
            'GODEBUG': 'asyncpreemptoff=1',
          };
          
          console.log('Setting environment variables for debugging:', env);
          
          // Run with debug environment variables
          const results = await instance.run({
            mount: { '/': fs },
            args: args,
            env: env,
          });
          console.log('WASM execution returned, clearing timeout');
          clearTimeout(executionTimeout);
          
          console.log('Go runtime completed');
          console.log('Results object:', Object.keys(results));
          
          // Check if stderr has any content that might indicate an error
          const stderrReader = results.stderr.getReader();
          let stderrContent = '';
          let stderrDone = false;
          
          console.log('Checking stderr for errors...');
          while (!stderrDone) {
            const { value, done } = await stderrReader.read();
            stderrDone = done;
            if (value) {
              const text = new TextDecoder().decode(value);
              stderrContent += text;
              console.error('STDERR:', text);
              if (onOutput) {
                onOutput(`Error output: ${text}`);
              }
            }
          }
          
          if (stderrContent) {
            console.warn('WASM execution produced error output:', stderrContent);
          } else {
            console.log('No stderr output detected');
          }
          
          // Collect all stdout output using the reader approach
          console.log('Starting to read stdout...');
          const reader = results.stdout.getReader();
          const decoder = new TextDecoder('utf-8');
          const chunks: string[] = [];
          let done = false;
          
          // Set a timeout for stdout reading
          const stdoutTimeout = setTimeout(() => {
            console.warn('Reading stdout is taking a long time (10 seconds). This might indicate a hang.');
            if (onOutput) {
              onOutput('Warning: Reading output is taking longer than expected. The process might be hanging.');
            }
          }, 10000);

          // Process output in real-time if callback is provided
          console.log('Reading stdout in a loop...');
          let readCount = 0;
          let maxReadAttempts = 100; // Prevent infinite loops
          
          while (!done && readCount < maxReadAttempts) {
            console.log(`Stdout read attempt ${++readCount}...`);
            
            try {
              // Set a timeout for each read operation
              const readPromise = reader.read();
              
              // Create a timeout promise
              const timeoutPromise = new Promise<{value: undefined, done: true}>((resolve) => {
                setTimeout(() => {
                  console.warn(`Read operation timed out on attempt ${readCount}`);
                  resolve({value: undefined, done: true});
                }, 5000); // 5 second timeout per read
              });
              
              // Race the read operation against the timeout
              const { value, done: isDone } = await Promise.race([
                readPromise,
                timeoutPromise
              ]);
              
              console.log(`Read result - done: ${isDone}, has value: ${!!value}, value length: ${value ? value.length : 0}`);
              done = isDone;
              
              if (value) {
                const text = decoder.decode(value, { stream: !done });
                console.log(`Decoded text (${text.length} chars): ${text.substring(0, 100)}${text.length > 100 ? '...' : ''}`);
                chunks.push(text);
                
                // If we have a callback, send each line as it comes
                if (outputCallback) {
                  const lines = text.split('\n');
                  console.log(`Split into ${lines.length} lines`);
                  for (const line of lines) {
                    if (line.trim()) {
                      outputCallback(line);
                    }
                  }
                }
              }
            } catch (error) {
              console.error(`Error during stdout read attempt ${readCount}:`, error);
              if (onOutput) {
                onOutput(`Error reading output: ${error instanceof Error ? error.message : 'Unknown error'}`);
              }
              // Break the loop on error
              done = true;
            }
          }
          
          if (readCount >= maxReadAttempts) {
            console.warn(`Reached maximum read attempts (${maxReadAttempts}). Breaking out of the loop.`);
            if (onOutput) {
              onOutput(`Warning: Reached maximum read attempts (${maxReadAttempts}). Some output may be missing.`);
            }
          }
          
          clearTimeout(stdoutTimeout);
          console.log('Stdout collection completed');
          
          // Split by newlines to get individual events
          const outputLines = chunks.join('').split('\n').filter(line => line.trim() !== '');
          
          return outputLines;
        } catch (error: any) {
          console.error('Error executing configuration:', error);
          
          const errorMessage = `Error: ${error.message || 'Unknown error'}`;
          
          if (onOutput) {
            onOutput(errorMessage);
          }
          
          // Generate mock data for testing
          const mockResults = [];
          for (let i = 0; i < 5; i++) {
            const mockEvent = JSON.stringify({
              _raw: `Mock event ${i + 1} (after error)`,
              timestamp: new Date().toISOString(),
              source: configuration.gogen,
              index: i
            });
            
            if (onOutput) {
              onOutput(mockEvent);
            }
            
            mockResults.push(mockEvent);
          }
          
          return [errorMessage, ...mockResults];
        }
      } catch (error: any) {
        console.error('Error executing configuration:', error);
        
        const errorMessage = `Error: ${error.message || 'Unknown error'}`;
        
        if (onOutput) {
          onOutput(errorMessage);
        }
        
        // Generate mock data for testing
        const mockResults = [];
        for (let i = 0; i < 5; i++) {
          const mockEvent = JSON.stringify({
            _raw: `Mock event ${i + 1} (after error)`,
            timestamp: new Date().toISOString(),
            source: configuration.gogen,
            index: i
          });
          
          if (onOutput) {
            onOutput(mockEvent);
          }
          
          mockResults.push(mockEvent);
        }
        
        return [errorMessage, ...mockResults];
      }
    } catch (error: any) {
      console.error('Error executing configuration:', error);
      
      const errorMessage = `Error: ${error.message || 'Unknown error'}`;
      
      if (onOutput) {
        onOutput(errorMessage);
      }
      
      // Generate mock data for testing
      const mockResults = [];
      for (let i = 0; i < 5; i++) {
        const mockEvent = JSON.stringify({
          _raw: `Mock event ${i + 1} (after error)`,
          timestamp: new Date().toISOString(),
          source: configuration.gogen,
          index: i
        });
        
        if (onOutput) {
          onOutput(mockEvent);
        }
        
        mockResults.push(mockEvent);
      }
      
      return [errorMessage, ...mockResults];
    }
  } catch (error: any) {
    console.error('Error setting up execution:', error);
    
    const errorMessage = `Setup Error: ${error.message || 'Unknown error'}`;
    
    if (onOutput) {
      onOutput(errorMessage);
    }
    
    return [errorMessage];
  }
};

export default {
  executeConfiguration,
}; 