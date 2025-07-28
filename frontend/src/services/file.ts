import api from './api';
import { 
  FileInfo, 
  CSVData, 
  ApiResponse 
} from './types';

export const fileService = {
  // 上传文件
  async upload(file: File): Promise<FileInfo> {
    const formData = new FormData();
    formData.append('file', file);
    
    const response: ApiResponse<{file: FileInfo}> = await api.post('/file/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data!.file;
  },

  // 获取文件列表
  async getList(): Promise<FileInfo[]> {
    const response: ApiResponse<FileInfo[]> = await api.get('/file/list');
    return response.data!;
  },

  // 删除文件
  async delete(id: number): Promise<void> {
    await api.delete(`/file/${id}`);
  },

  // 预览文件数据
  async preview(id: number, limit?: number): Promise<CSVData> {
    const params = limit ? { limit } : {};
    const response: ApiResponse<CSVData> = await api.get(`/file/${id}/preview`, { params });
    return response.data!;
  },
};
