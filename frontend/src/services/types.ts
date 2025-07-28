export interface User {
  id: number;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
}

export interface FileInfo {
  id: number;
  user_id: number;
  name: string;
  orig_name: string;
  path: string;
  size: number;
  type: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface CSVData {
  headers: string[];
  rows: string[][];
  summary: {
    total_rows: number;
    total_cols: number;
  };
}

export interface Session {
  id: number;
  user_id: number;
  name: string;
  file_id?: number;
  created_at: string;
  updated_at: string;
}

export interface Query {
  id: number;
  session_id: number;
  user_id: number;
  question: string;
  answer: string;
  query_type: string;
  status: string;
  created_at: string;
}

export interface QueryRequest {
  session_id: number;
  question: string;
  file_id?: number;
}

export interface QueryResponse {
  answer: string;
  data?: any;
  query_type: string;
  status: string;
}

export interface VisualizationRequest {
  session_id: number;
  query: string;
  file_id: number;
  chart_type: string;
}

export interface VisualizationResponse {
  chart_data: any;
  chart_type: string;
  title: string;
}

export interface ReportRequest {
  session_id: number;
  file_id: number;
  dimensions: string[];
  description: string;
}

export interface ReportResponse {
  content: string;
  charts: any[];
  summary: string;
  export_url?: string;
}

export interface LLMConfig {
  id: number;
  user_id: number;
  provider: string;
  api_key: string;
  model: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface LLMConfigRequest {
  provider: string;
  api_key: string;
  model: string;
  is_default: boolean;
}

export interface Usage {
  id: number;
  user_id: number;
  provider: string;
  model: string;
  tokens: number;
  cost: number;
  query_id: number;
  created_at: string;
}

export interface UsageResponse {
  total_tokens: number;
  total_cost: number;
  usage: Usage[];
}
