import { executeConfiguration, ExecutionParams } from './gogenWasm';
import { Configuration } from './gogenApi';

// Get the global object
const global = globalThis as any;

// Define Go instance interface
interface MockGoInstance {
  argv: string[];
  env: Record<string, string>;
  importObject: {
    env: Record<string, unknown>;
    go: Record<string, unknown>;
  };
  run: jest.Mock;
  exit: jest.Mock;
}

// Mock Go class with static instance tracking
class MockGo implements MockGoInstance {
  static lastInstance: MockGo | null = null;

  // Instance properties
  argv: string[] = [];
  env: Record<string, string> = {};
  importObject: {
    env: Record<string, unknown>;
    go: Record<string, unknown>;
  } = {
    env: {
      'runtime.ticks': jest.fn(() => 0),
      'runtime.sleepTicks': jest.fn(),
      'syscall/js.valueGet': jest.fn(),
      'syscall/js.valueSet': jest.fn(),
      'syscall/js.valueIndex': jest.fn(),
      'syscall/js.valueCall': jest.fn(),
      'syscall/js.valueNew': jest.fn(),
      'syscall/js.valueLength': jest.fn(),
      'syscall/js.valuePrepareString': jest.fn(),
      'syscall/js.valueLoadString': jest.fn(),
      'syscall/js.finalizeRef': jest.fn(),
    },
    go: {
      'runtime.wasmExit': jest.fn(),
      'runtime.wasmWrite': jest.fn((sp: number) => {
        // Mock writing to stdout
        const instance = MockGo.getInstance();
        if (instance && global.fs) {
          const fd = Number(instance.mem.getBigInt64(sp + 8));
          const p = Number(instance.mem.getBigInt64(sp + 16));
          const n = instance.mem.getInt32(sp + 24);
          const buf = new Uint8Array(instance.mem.buffer, p, n);
          global.fs.writeSync(fd, buf);
        }
      }),
      'runtime.resetMemoryDataView': jest.fn(),
      'runtime.nanotime1': jest.fn(() => 0),
      'runtime.walltime1': jest.fn(() => 0),
      'runtime.scheduleTimeoutEvent': jest.fn(),
      'runtime.clearTimeoutEvent': jest.fn(),
      'runtime.getRandomData': jest.fn(),
    },
  };
  run: jest.Mock = jest.fn().mockResolvedValue(undefined);
  exit: jest.Mock = jest.fn();
  mem: DataView;

  constructor() {
    // Set up prototype chain
    Object.setPrototypeOf(this, MockGo.prototype);
    
    // Initialize instance properties
    this.argv = [];
    this.env = {};
    this.mem = new DataView(new ArrayBuffer(1024 * 1024)); // 1MB buffer for testing

    // Track instance
    MockGo.lastInstance = this;
  }

  static clearInstance() {
    MockGo.lastInstance = null;
  }

  static getInstance(): MockGo {
    if (!MockGo.lastInstance) {
      throw new Error('No MockGo instance available');
    }
    return MockGo.lastInstance;
  }
}

// Set up constructor properties
Object.defineProperty(MockGo, 'prototype', {
  writable: false,
  enumerable: false,
  configurable: false,
});

// Mock WebAssembly
const mockWebAssembly = {
  instantiate: jest.fn().mockResolvedValue({
    instance: { exports: {} },
    module: {},
  })
};

// Mock TextDecoder and TextEncoder
const mockTextDecoder = {
  decode: jest.fn((buf) => Buffer.from(buf).toString())
};

const mockTextEncoder = {
  encode: jest.fn((str) => Buffer.from(str))
};

describe('gogenWasm', () => {
  let originalWindow: any;
  let originalWebAssembly: any;
  let originalTextDecoder: any;
  let originalTextEncoder: any;
  let mockFetch: jest.Mock;

  beforeEach(() => {
    // Store original globals
    originalWindow = global.window;
    originalWebAssembly = global.WebAssembly;
    originalTextDecoder = global.TextDecoder;
    originalTextEncoder = global.TextEncoder;

    // Create window if it doesn't exist
    if (!global.window) {
      global.window = {};
    }

    // Directly assign Go constructor
    global.window.Go = MockGo;

    // Debug logging
    console.log('Mock setup:', {
      'window.Go exists': !!global.window.Go,
      'window.Go is constructor': typeof global.window.Go === 'function',
      'window.Go prototype': Object.getPrototypeOf(global.window.Go),
      'MockGo is constructor': typeof MockGo === 'function',
      'MockGo prototype': Object.getPrototypeOf(MockGo),
    });

    // Mock WebAssembly
    global.WebAssembly = mockWebAssembly as any;

    // Mock TextDecoder and TextEncoder
    global.TextDecoder = jest.fn(() => mockTextDecoder) as any;
    global.TextEncoder = jest.fn(() => mockTextEncoder) as any;

    // Mock fetch
    mockFetch = jest.fn(() => 
      Promise.resolve({
        arrayBuffer: () => Promise.resolve(new ArrayBuffer(0))
      })
    );
    global.fetch = mockFetch;

    // Reset all mocks and static instance
    jest.clearAllMocks();
    MockGo.clearInstance();
  });

  afterEach(() => {
    // Restore original globals
    global.window = originalWindow;
    global.WebAssembly = originalWebAssembly;
    global.TextDecoder = originalTextDecoder;
    global.TextEncoder = originalTextEncoder;

    // Clean up virtual filesystem and static instance
    if (global.fs) {
      delete global.fs;
    }
    MockGo.clearInstance();
  });

  describe('executeConfiguration', () => {
    let mockConfig: Configuration;
    let mockParams: ExecutionParams;
    let outputLines: string[] = [];
    let onOutput: jest.Mock;
    let mockGoInstance: MockGo;
    let outputBuffer: { stdout: string; stderr: string };

    beforeEach(() => {
      outputLines = [];
      onOutput = jest.fn();
      outputBuffer = {
        stdout: '',
        stderr: ''
      };

      mockConfig = {
        gogen: 'test',
        description: 'Test configuration',
        config: 'test: true'
      };

      mockParams = {
        eventCount: 10,
        intervals: 5,
        intervalSeconds: 1,
        outputTemplate: 'json' as const
      };

      // Set up virtual filesystem with proper file descriptors
      global.fs = {
        constants: {
          O_WRONLY: 1,
          O_RDWR: 2,
          O_CREAT: 64,
          O_TRUNC: 512,
          O_APPEND: 1024,
          O_EXCL: 128,
        },
        writeSync: jest.fn((fd: number, buf: Uint8Array) => {
          const text = new TextDecoder().decode(buf);
          
          // Accumulate output in buffer
          if (fd === 1) {
            outputBuffer.stdout += text;
            const lines = outputBuffer.stdout.split('\n');
            if (lines.length > 1) {
              // Process all complete lines
              for (let i = 0; i < lines.length - 1; i++) {
                const line = lines[i];
                if (line.length > 0) {
                  outputLines.push(line);
                  onOutput(line);
                }
              }
              // Keep the partial line
              outputBuffer.stdout = lines[lines.length - 1];
            }
          } else if (fd === 2) {
            outputBuffer.stderr += text;
            const lines = outputBuffer.stderr.split('\n');
            if (lines.length > 1) {
              // Process all complete lines
              for (let i = 0; i < lines.length - 1; i++) {
                const line = lines[i];
                if (line.length > 0) {
                  onOutput(`ERROR: ${line}`);
                }
              }
              // Keep the partial line
              outputBuffer.stderr = lines[lines.length - 1];
            }
          }
          return buf.length;
        }),
        readSync: jest.fn(),
        read: jest.fn(),
        open: jest.fn((_path, _flags, _mode, callback) => callback(null, 3)), // Start at fd 3 for files
        stat: jest.fn((_path, callback) => callback(null, { isFile: () => true, size: 100 })),
      };

      // Ensure window.Go is set and create instance
      if (!global.window) {
        global.window = {};
      }
      mockGoInstance = new MockGo();
      global.window.Go = jest.fn(() => mockGoInstance);
    });

    // Test output handling separately from WASM execution
    describe('output handling', () => {
      it('should handle stdout output correctly', () => {
        // Write test output
        const testOutput = 'test output\n';
        const outputBuffer = Buffer.from(testOutput);
        global.fs.writeSync(1, outputBuffer);

        expect(outputLines).toContain('test output');
        expect(onOutput).toHaveBeenCalledWith('test output');
      });

      it('should handle stderr output correctly', () => {
        // Write error output
        const errorOutput = 'error message\n';
        const errorBuffer = Buffer.from(errorOutput);
        global.fs.writeSync(2, errorBuffer);

        expect(onOutput).toHaveBeenCalledWith('ERROR: error message');
      });
    });

    // Original tests for executeConfiguration
    it('should verify Go constructor is available', () => {
      expect(global.window.Go).toBeDefined();
      expect(typeof global.window.Go).toBe('function');
      expect(() => new global.window.Go()).not.toThrow();
      expect(new global.window.Go()).toBe(mockGoInstance);
    });

    it('should handle Go execution errors', async () => {
      mockGoInstance.run.mockRejectedValue(new Error('Execution failed'));
      await expect(executeConfiguration(mockConfig, mockParams, onOutput)).rejects.toThrow('Execution failed');
    });

    it('should clean up virtual filesystem after execution', async () => {
      const originalFs = { ...global.fs };
      mockGoInstance.run.mockRejectedValue(new Error('Execution failed'));
      
      try {
        await executeConfiguration(mockConfig, mockParams, onOutput);
      } catch (error) {
        // Expected error
      }

      expect(global.fs).toEqual(originalFs);
    });

    it('should handle partial output lines correctly', async () => {
      // Write partial output
      const partialOutput = 'partial output';
      const outputBuffer = Buffer.from(partialOutput); // No newline
      global.fs.writeSync(1, outputBuffer);

      expect(outputLines).not.toContain(partialOutput); // Should not process without newline
      expect(onOutput).not.toHaveBeenCalled();
    });

    it('should pass correct arguments to Go instance', async () => {
      try {
        await executeConfiguration(mockConfig, mockParams, onOutput);
      } catch (error) {
        // Expected error
      }

      expect(mockGoInstance.argv).toEqual([
        'gogen',
        '-c',
        '/config.yml',
        '-ot',
        'json',
        'gen',
        '-ei',
        '5',
        '-i',
        '1',
        '-c',
        '10'
      ]);
    });

    it('should handle different output templates', async () => {
      const configuredParams = { 
        ...mockParams, 
        outputTemplate: 'configured' as const 
      };
      try {
        await executeConfiguration(mockConfig, configuredParams, onOutput);
      } catch (error) {
        // Expected error
      }

      expect(mockGoInstance.argv).not.toContain('-ot');
    });
  });
}); 