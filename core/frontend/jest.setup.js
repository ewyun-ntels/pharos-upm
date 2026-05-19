require('@testing-library/jest-dom');

// Mock ResizeObserver (not available in jsdom)
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Polyfill for React Router v7 (requires TextEncoder/TextDecoder)
const { TextEncoder, TextDecoder } = require('util');
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder;

// Mock lucide-react icons
jest.mock('lucide-react', () => ({
  CheckIcon: () => null,
  ChevronDownIcon: () => null,
  ChevronRightIcon: () => null,
  AlertTriangle: () => null,
  __esModule: true,
}));
