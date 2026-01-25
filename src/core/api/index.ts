/**
 * Core API module exports
 * Centralized exports for cleaner imports throughout the app
 */

export { default as api } from './client';
export { extractErrorMessage, createAppError, type AppError } from './errorHandler';
