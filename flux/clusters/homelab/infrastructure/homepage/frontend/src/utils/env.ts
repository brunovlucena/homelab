/**
 * Environment variable helper
 * Centralized place to access environment variables
 * 
 * In Vite: Uses import.meta.env
 * In Jest: This module is mocked (see __mocks__/env.ts)
 */

export const env = {
  API_URL: import.meta.env.VITE_API_URL || '/api/v1',
  APP_ENV: import.meta.env.VITE_APP_ENV || 'production',
}

