import { ApiFailure } from './api.service';

export interface ViewError {
  title: string;
  message: string;
  code: string;
  requestId: string;
  status: number;
}

export function toViewError(error: unknown, resource = 'data'): ViewError {
  if (error instanceof ApiFailure) {
    if (error.status === 404) {
      return {
        title: `${capitalize(resource)} not found`,
        message: `The requested ${resource} does not exist or is no longer available.`,
        code: error.code,
        requestId: error.requestId,
        status: error.status,
      };
    }
    if (error.status === 0 || error.code === 'NETWORK') {
      return {
        title: 'Release intelligence is temporarily unavailable',
        message: 'The API could not be reached. Check the local backend or try the request again.',
        code: error.code,
        requestId: error.requestId,
        status: error.status,
      };
    }
    return {
      title: `Unable to load ${resource}`,
      message: error.message || `The ${resource} request failed.`,
      code: error.code,
      requestId: error.requestId,
      status: error.status,
    };
  }
  return {
    title: `Unable to load ${resource}`,
    message: `An unexpected error occurred while loading ${resource}.`,
    code: 'UNKNOWN',
    requestId: '',
    status: 0,
  };
}

function capitalize(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1);
}
