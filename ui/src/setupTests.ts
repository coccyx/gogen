import '@testing-library/jest-dom';
import { TextEncoder, TextDecoder } from 'util';

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe = jest.fn();
  unobserve = jest.fn();
  disconnect = jest.fn();
};

// Mock TextEncoder/TextDecoder
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder as typeof global.TextDecoder;

const originalConsoleWarn = console.warn;
const originalConsoleError = console.error;

beforeAll(() => {
  jest.spyOn(console, 'warn').mockImplementation((...args: unknown[]) => {
    const message = String(args[0] ?? '');
    if (
      message.includes('React Router Future Flag Warning')
    ) {
      return;
    }

    originalConsoleWarn(...args);
  });

  jest.spyOn(console, 'error').mockImplementation((...args: unknown[]) => {
    const message = String(args[0] ?? '');
    if (
      message.includes('not wrapped in act') ||
      message.startsWith('Error fetching configuration') ||
      message.startsWith('Error fetching configurations:') ||
      message.startsWith('Error searching configurations:') ||
      message.startsWith('Error executing WASM:')
    ) {
      return;
    }

    originalConsoleError(...args);
  });
});

afterAll(() => {
  jest.restoreAllMocks();
});
