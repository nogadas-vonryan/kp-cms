import type { AxiosError } from 'axios';
import type { ApiErrorResponse } from '@/types';

export interface AppError {
  message: string;
  statusCode?: number;
  originalError?: unknown;
}

/**
 * Extracts a user-friendly error message from various error types
 */
export function extractErrorMessage(error: unknown): string {
  if (!error) return 'An unknown error occurred';

  // Axios error with response
  if (isAxiosError(error)) {
    const axiosError = error as AxiosError<ApiErrorResponse>;
    
    if (axiosError.response?.data) {
      // Prefer explicit error field, but fall back to generic message
      const data = axiosError.response.data as Partial<ApiErrorResponse & { message?: string }>;
      if (data.error) return data.error;
      if (data.message) return data.message;
    }
    
    if (axiosError.response?.status) {
      return getStatusMessage(axiosError.response.status);
    }
    
    if (axiosError.message) {
      return axiosError.message;
    }
  }

  // Standard Error object
  if (error instanceof Error) {
    return error.message;
  }

  // String error
  if (typeof error === 'string') {
    return error;
  }

  return 'An unexpected error occurred';
}

/**
 * Creates a standardized AppError object
 */
export function createAppError(error: unknown): AppError {
  const message = extractErrorMessage(error);
  
  let statusCode: number | undefined;
  if (isAxiosError(error)) {
    statusCode = error.response?.status;
  }

  return {
    message,
    statusCode,
    originalError: error,
  };
}

/**
 * Type guard for Axios errors
 */
function isAxiosError(error: unknown): error is AxiosError {
  return typeof error === 'object' && error !== null && 'isAxiosError' in error;
}

/**
 * Maps HTTP status codes to user-friendly messages
 */
function getStatusMessage(status: number): string {
  const messages: Record<number, string> = {
    400: 'Invalid request. Please check your input.',
    401: 'Authentication required. Please log in.',
    403: 'You do not have permission to perform this action.',
    404: 'The requested resource was not found.',
    409: 'A conflict occurred. Please try again.',
    422: 'The provided data is invalid.',
    429: 'Too many requests. Please slow down.',
    500: 'Server error. Please try again later.',
    502: 'Service temporarily unavailable.',
    503: 'Service unavailable. Please try again later.',
  };

  return messages[status] || `Request failed with status ${status}`;
}
