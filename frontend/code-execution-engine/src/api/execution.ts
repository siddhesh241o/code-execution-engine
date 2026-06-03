const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:5005';
const FRONTEND_SHARED_SECRET = import.meta.env.VITE_FRONTEND_SHARED_SECRET || '';

const buildHeaders = (extraHeaders: Record<string, string> = {}) => ({
  'X-Frontend-Secret': FRONTEND_SHARED_SECRET,
  'ngrok-skip-browser-warning': 'true',
  ...extraHeaders,
});

export interface ExecutionRequest {
  code: string;
  language: string;
  input?: string;
}

export interface ExecutionResponse {
  id: string;
  stdout: string;
  stderr: string;
  time_ms: number;
  memory_kb: number;
  status: string;
}

export interface JobStatus {
  job_id: string;
  status: string;
}

export const executeCode = async (req: ExecutionRequest): Promise<JobStatus> => {
  const response = await fetch(`${API_BASE_URL}/api/execute`, {
    method: 'POST',
    headers: buildHeaders({
      'Content-Type': 'application/json',
    }),
    body: JSON.stringify(req),
  });

  if (!response.ok) {
    const errorBody = await response.text();
    throw new Error(`Execution failed: ${errorBody || response.statusText}`);
  }

  return response.json();
};

export const getResult = async (jobId: string): Promise<ExecutionResponse | JobStatus> => {
  const response = await fetch(`${API_BASE_URL}/api/result/${jobId}`, {
    headers: buildHeaders(),
  });

  if (!response.ok) {
    const errorBody = await response.text();
    throw new Error(`Failed to fetch result: ${errorBody || response.statusText}`);
  }

  return response.json();
};

export const pollResult = async (jobId: string, interval = 2000, maxAttempts = 30): Promise<ExecutionResponse> => {
  let attempts = 0;
  
  while (attempts < maxAttempts) {
    const data = await getResult(jobId);
    
    // Check if it's a full response or just a status
    if ('status' in data && (data.status === 'Queued' || data.status === 'Processing')) {
      await new Promise(resolve => setTimeout(resolve, interval));
      attempts++;
      continue;
    }
    
    // If it has stdout/stderr or a final status, it's done
    return data as ExecutionResponse;
  }
  
  throw new Error('Polling timed out');
};
