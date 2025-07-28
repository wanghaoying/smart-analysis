import api from './api';
import { 
  Session, 
  Query, 
  QueryRequest, 
  QueryResponse,
  VisualizationRequest,
  VisualizationResponse,
  ReportRequest,
  ReportResponse,
  LLMConfig,
  LLMConfigRequest,
  UsageResponse,
  ApiResponse 
} from './types';

export const analysisService = {
  // 创建会话
  async createSession(name: string, fileId?: number): Promise<Session> {
    const response: ApiResponse<Session> = await api.post('/analysis/session', {
      name,
      file_id: fileId,
    });
    return response.data!;
  },

  // 获取会话详情
  async getSession(id: number): Promise<Session> {
    const response: ApiResponse<Session> = await api.get(`/analysis/session/${id}`);
    return response.data!;
  },

  // 发送查询
  async query(data: QueryRequest): Promise<QueryResponse> {
    const response: ApiResponse<QueryResponse> = await api.post('/analysis/query', data);
    return response.data!;
  },

  // 生成可视化
  async visualize(data: VisualizationRequest): Promise<VisualizationResponse> {
    const response: ApiResponse<VisualizationResponse> = await api.post('/analysis/visualize', data);
    return response.data!;
  },

  // 生成报告
  async generateReport(data: ReportRequest): Promise<ReportResponse> {
    const response: ApiResponse<ReportResponse> = await api.post('/analysis/report', data);
    return response.data!;
  },

  // 获取查询历史
  async getHistory(sessionId?: number): Promise<Query[]> {
    const params = sessionId ? { session_id: sessionId } : {};
    const response: ApiResponse<Query[]> = await api.get('/analysis/history', { params });
    return response.data!;
  },

  // 配置LLM
  async configLLM(data: LLMConfigRequest): Promise<LLMConfig> {
    const response: ApiResponse<LLMConfig> = await api.post('/llm/config', data);
    return response.data!;
  },

  // 获取LLM配置
  async getLLMConfig(): Promise<LLMConfig[]> {
    const response: ApiResponse<LLMConfig[]> = await api.get('/llm/config');
    return response.data!;
  },

  // 获取使用量统计
  async getUsage(): Promise<UsageResponse> {
    const response: ApiResponse<UsageResponse> = await api.get('/llm/usage');
    return response.data!;
  },
};
