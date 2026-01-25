/**
 * Centralized configuration for environment variables
 * Makes it easier to track what env vars are used and provide defaults
 */

interface AppConfig {
  apiUrl: string;
  apiTimeout: number;
  isDevelopment: boolean;
  isProduction: boolean;
}

function getEnvVar(key: string, defaultValue?: string): string {
  const value = import.meta.env[key];
  if (value !== undefined) return String(value);
  if (defaultValue !== undefined) return defaultValue;
  throw new Error(`Missing required environment variable: ${key}`);
}

export const config: AppConfig = {
  apiUrl: getEnvVar('VITE_API_URL', 'http://localhost:8080'),
  apiTimeout: parseInt(getEnvVar('VITE_API_TIMEOUT', '30000'), 10),
  isDevelopment: import.meta.env.DEV,
  isProduction: import.meta.env.PROD,
};
