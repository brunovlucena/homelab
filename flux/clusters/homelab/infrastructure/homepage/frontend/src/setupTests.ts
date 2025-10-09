// Jest setup file
// This file is run before each test file

// Mock the env module to avoid import.meta.env issues
jest.mock('./utils/env')

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  // Uncomment to ignore a specific log level
  // log: jest.fn(),
  // debug: jest.fn(),
  // info: jest.fn(),
  // warn: jest.fn(),
  // error: jest.fn(),
};

// Global test timeout
jest.setTimeout(10000);
