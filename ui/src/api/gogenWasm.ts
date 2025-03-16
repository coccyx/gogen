import { Configuration } from './gogenApi';

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

    global.fs.writeSync = (fd: number, buf: Uint8Array) => {
      outputBuf += decoder.decode(buf);
      const nl = outputBuf.lastIndexOf("\n");
      if (nl !== -1) {
        const line = outputBuf.substring(0, nl);
        output.push(line);
        if (onOutput) {
          onOutput(line);
        }
        console.log(line);
        outputBuf = outputBuf.substring(nl + 1);
      }
      return buf.length;
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
        // Return a fake file descriptor > 2 (0, 1, 2 are reserved for stdin/stdout/stderr)
        const fd = 3;
        virtualFilePositions[fd] = 0; // Initialize file position
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
        // Use the provided position or the current file position
        let pos = position !== null ? position : (virtualFilePositions[fd] || 0);
        
        // Return 0 bytes if we're reading past the end of the file
        if (pos >= content.length) {
          callback(null, 0);
          return;
        }

        const data = new TextEncoder().encode(content).slice(pos, pos + length);
        buffer.set(data, offset);

        // Update the file position if we're using the internal pointer
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

    global.fs.readSync = (fd: number, buffer: Uint8Array, offset: number, length: number, position: number | null) => {
      console.log('readSync', fd, offset, length, position);
      if (fd === 3 && '/config.yml' in virtualFiles) {
        const content = virtualFiles['/config.yml'];
        // Use the provided position or the current file position
        let pos = position !== null ? position : (virtualFilePositions[fd] || 0);

        // Return 0 bytes if we're reading past the end of the file
        if (pos >= content.length) {
          return 0;
        }

        const data = new TextEncoder().encode(content).slice(pos, pos + length);
        buffer.set(data, offset);

        // Update the file position if we're using the internal pointer
        if (position === null) {
          virtualFilePositions[fd] = pos + data.length;
        }

        return data.length;
      } else if (typeof originalFs.readSync === 'function') {
        return originalFs.readSync(fd, buffer, offset, length, position);
      }
      throw new Error('EBADF: bad file descriptor');
    };
    
    // Initialize Go WASM
    const go = new window.Go();
    
    // Fetch and instantiate the WASM module
    const wasmResponse = await fetch('/gogen.wasm');
    const wasmBytes = await wasmResponse.arrayBuffer();
    const wasmResult = await WebAssembly.instantiate(wasmBytes, go.importObject);
    
    // Run the WASM instance with minimal args
    go.argv = ['gogen', '-c', '/config.yml'];
    await go.run(wasmResult.instance);
    
    // Process any remaining output
    if (outputBuf.length > 0) {
      output.push(outputBuf);
      if (onOutput) {
        onOutput(outputBuf);
      }
    }
    
    return output;
  } catch (error: any) {
    console.error('Error executing WASM:', error);
    return [error.message || 'Unknown error'];
  }
};

export default {
  executeConfiguration,
}; 