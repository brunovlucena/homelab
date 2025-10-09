/**
 * Mock for env.ts in Jest tests
 * This avoids the import.meta.env issue in Jest
 */

export const env = {
  API_URL: '/api/v1',
  APP_ENV: 'test',
}

