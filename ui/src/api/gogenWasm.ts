import { Configuration } from './gogenApi';

export interface ExecutionParams {
  intervals: number;
  intervalSeconds: number;
  eventCount: number;
  outputTemplate: 'raw' | 'json' | 'configured';
}

// Define the Go type from wasm_exec.js
declare global {
  interface Window {
    Go: any;
  }
}

interface VirtualFsSetup {
  output: string[];
  outputBuf: string;
  decoder: TextDecoder;
  originalFs: {
    writeSync: any;
    stat: any;
    open: any;
    read: any;
    readSync: any;
  };
  cleanup: () => void;
}

const setupVirtualFileSystem = (
  configuration: Configuration,
  onOutput?: (line: string) => void
): VirtualFsSetup => {
  const output: string[] = [];
  const decoder = new TextDecoder();
  let outputBuf = "";

  // Set up virtual filesystem for config
  const virtualFiles: { [key: string]: string } = {
    '/config.yml': configuration.config || ''
  };

  // Track file positions for virtual files
  const virtualFilePositions: { [key: number]: number } = {};

  // Create a complete stats object that matches Node.js fs.Stats
  const createStats = (isFile: boolean, size: number = 0) => ({
    isFile: () => isFile,
    isDirectory: () => !isFile,
    isBlockDevice: () => false,
    isCharacterDevice: () => false,
    isSymbolicLink: () => false,
    isFIFO: () => false,
    isSocket: () => false,
    dev: 0,
    ino: 0,
    mode: isFile ? 0o666 : 0o777,
    nlink: 1,
    uid: 0,
    gid: 0,
    rdev: 0,
    size,
    blksize: 4096,
    blocks: Math.ceil(size / 4096),
    atimeMs: Date.now(),
    mtimeMs: Date.now(),
    ctimeMs: Date.now(),
    birthtimeMs: Date.now(),
    atime: new Date(),
    mtime: new Date(),
    ctime: new Date(),
    birthtime: new Date()
  });

  const global = globalThis as any;
  if (!global.fs) {
    global.fs = {
      constants: { O_WRONLY: -1, O_RDWR: -1, O_CREAT: -1, O_TRUNC: -1, O_APPEND: -1, O_EXCL: -1 }
    };
  }

  // Store original methods
  const originalFs = {
    writeSync: global.fs.writeSync,
    stat: global.fs.stat,
    open: global.fs.open,
    read: global.fs.read,
    readSync: global.fs.readSync
  };

  // Set up the fs methods
  let stdoutBuf = '';
  let stderrBuf = '';

  global.fs.writeSync = (fd: number, buf: Uint8Array) => {
    // fd 1 is stdout, fd 2 is stderr
    const line = decoder.decode(buf);
    
    if (fd === 1) {
      stdoutBuf = processOutput(stdoutBuf, line, (line) => {
        output.push(line);
        if (onOutput) {
          onOutput(line);
        }
        console.log(line);
      });
    } else if (fd === 2) {
      stderrBuf = processOutput(stderrBuf, line, (line) => {
        console.error(line);
        if (onOutput) {
          onOutput(`ERROR: ${line}`);
        }
      });
    }
    
    return buf.length;
  };

  // Helper function to process output buffers
  const processOutput = (
    buffer: string, 
    newData: string, 
    lineHandler: (line: string) => void
  ): string => {
    buffer += newData;
    const nl = buffer.lastIndexOf("\n");
    if (nl !== -1) {
      const line = buffer.substring(0, nl);
      lineHandler(line);
      return buffer.substring(nl + 1);
    }
    return buffer;
  };

  global.fs.stat = (path: string, callback: Function) => {
    console.log('stat', path);
    if (path in virtualFiles) {
      callback(null, createStats(true, virtualFiles[path].length));
    } else if (path === '/') {
      callback(null, createStats(false));
    } else if (typeof originalFs.stat === 'function') {
      originalFs.stat(path, callback);
    } else {
      callback(new Error(`ENOENT: no such file or directory, stat '${path}'`));
    }
  };

  global.fs.open = (path: string, flags: string, mode: number, callback: Function) => {
    console.log('open', path, flags, mode);
    if (path in virtualFiles) {
      const fd = 3;
      virtualFilePositions[fd] = 0;
      callback(null, fd);
    } else if (typeof originalFs.open === 'function') {
      originalFs.open(path, flags, mode, callback);
    } else {
      callback(new Error(`ENOENT: no such file or directory, open '${path}'`));
    }
  };

  global.fs.read = (fd: number, buffer: Uint8Array, offset: number, length: number, position: number | null, callback: Function) => {
    console.log('read', fd, offset, length, position);
    if (fd === 3 && '/config.yml' in virtualFiles) {
      const content = virtualFiles['/config.yml'];
      let pos = position !== null ? position : (virtualFilePositions[fd] || 0);
      
      if (pos >= content.length) {
        callback(null, 0);
        return;
      }

      const data = new TextEncoder().encode(content).slice(pos, pos + length);
      buffer.set(data, offset);

      if (position === null) {
        virtualFilePositions[fd] = pos + data.length;
      }

      callback(null, data.length);
    } else if (typeof originalFs.read === 'function') {
      originalFs.read(fd, buffer, offset, length, position, callback);
    } else {
      callback(new Error('EBADF: bad file descriptor'));
    }
  };

  const cleanup = () => {
    // Restore original fs methods
    Object.assign(global.fs, originalFs);
  };

  return {
    output,
    outputBuf,
    decoder,
    originalFs,
    cleanup
  };
};

/**
 * Execute a Gogen configuration using WebAssembly
 * 
 * @param configuration The configuration to execute
 * @param params The execution parameters
 * @param onOutput Optional callback for streaming output
 * @returns Array of output lines
 */
export const executeConfiguration = async (
  configuration: Configuration,
  params: ExecutionParams,
  onOutput?: (line: string) => void
): Promise<string[]> => {
  const fsSetup = setupVirtualFileSystem(configuration, onOutput);
  
  try {
    // Initialize a new Go instance for each execution
    const go = new window.Go();
    
    // Fetch and instantiate a new WASM module
    const wasmResponse = await fetch('/gogen.wasm');
    const wasmBytes = await wasmResponse.arrayBuffer();
    const wasmResult = await WebAssembly.instantiate(wasmBytes, go.importObject);
    
    // Build command line arguments
    const args = ['gogen', '-c', '/config.yml'];

    // Add output template (-ot) if not 'configured'
    if (params.outputTemplate !== 'configured') {
      args.push('-ot', params.outputTemplate);
    }

    // Add gen command and remaining arguments
    args.push('gen');

    // Add event interval count (-ei)
    args.push('-ei', params.intervals.toString());

    // Add interval seconds (-i)
    args.push('-i', params.intervalSeconds.toString());

    // Add event count (-c)
    args.push('-c', params.eventCount.toString());

    // Set up arguments and run
    go.argv = args;
    
    try {
      await go.run(wasmResult.instance);
    } catch (error: any) {
      if (error.message === 'Go program has already exited') {
        // This is expected - the Go program exits after completion
        console.log('Go program completed successfully');
      } else {
        throw error;
      }
    }
    
    // Process any remaining output
    if (fsSetup.outputBuf.length > 0) {
      fsSetup.output.push(fsSetup.outputBuf);
      if (onOutput) {
        onOutput(fsSetup.outputBuf);
      }
    }
    
    return fsSetup.output;
  } catch (error: any) {
    console.error('Error executing WASM:', error);
    throw error;
  } finally {
    fsSetup.cleanup();
  }
};

export default {
  executeConfiguration,
}; 