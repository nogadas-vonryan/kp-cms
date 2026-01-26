import api from '@/core/api/client';
import type {
  Document,
  CreateDocumentRequest,
  UpdateDocumentRequest,
  SyncIssue,
  ReloadResponse,
} from '@/types';

export const DocumentService = {
  getAll: (offset = 0, limit = 15) => 
    api.get<Document[]>('/documents', { params: { offset, limit } }),

  search: (params: {
    uuid?: string;
    code?: string;
    folder_name?: string;
    field_key?: string;
    date_from?: string;
    date_to?: string;
    [key: string]: any;
    sort_by?: string;
    sort_desc?: boolean;
    offset?: number;
    limit?: number;
  }) => {
    const cleanParams: any = {};
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== '' && value !== null) {
        cleanParams[key] = value;
      }
    });
    return api.get<Document[]>('/documents/search', { params: cleanParams });
  },

  getOne: (uuid: string) => 
    api.get<Document>(`/documents/${uuid}`),

  getById: (uuid: string) =>
    api.get<Document>(`/documents/${uuid}`),

  getByCode: (code: string) =>
    api.get<Document>(`/documents/code/${code}`),

  // Only exposes admin methods if generic type allows, but API enforces security
  create: (data: CreateDocumentRequest) => 
    api.post<Document>('/documents', data),

  update: (uuid: string, data: UpdateDocumentRequest) =>
    api.put<Document>(`/documents/${uuid}`, data),

  remove: (uuid: string) =>
    api.delete<void>(`/documents/${uuid}`),
    
  uploadFile: (uuid: string, file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return api.post(`/documents/${uuid}/files`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
  },

  /**
   * Downloads a file with proper authentication.
   * Returns the file as a blob so it can be downloaded with auth headers.
   */
  downloadFile: async (uuid: string, fileName: string) => {
    const response = await api.get(`/documents/${uuid}/files/${fileName}`, {
      responseType: 'blob'
    });
    return response.data;
  },

  /**
   * Updates the description and notes for a specific file entry in files.json.
   * Matches: PUT /documents/{uuid}/files/{fileName}
   */
  updateFileMetadata: (uuid: string, fileName: string, data: { description: string; note: string }) =>
    api.put(`/documents/${uuid}/files/${fileName}`, data),

  /**
   * Deletes both the physical file and its entry in files.json.
   * Matches: DELETE /documents/{uuid}/files/{fileName}
   */
  deleteFile: (uuid: string, fileName: string) =>
    api.delete(`/documents/${uuid}/files/${fileName}`),

  listConflicts: () =>
    api.get<SyncIssue[]>('/documents/conflicts'),

  reload: () =>
    api.post<ReloadResponse>('/documents/reload'),
};
