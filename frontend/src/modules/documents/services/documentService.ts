import api from '@/core/api/client';
import type {
  Document,
  CreateDocumentRequest,
  UpdateDocumentRequest,
  BackupFile,
  SyncIssue,
  ReloadResponse,
} from '@/types';

export const DocumentService = {
  getAll: (offset = 0, limit = 15, sort_by?: string, sort_desc?: boolean) => 
    api.get<Document[]>('/api/documents', { params: { offset, limit, sort_by, sort_desc } }),

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
      // Allow field_* parameters even if empty, but filter out other empty values
      if (value !== undefined && value !== null) {
        if (key.startsWith('field_') || value !== '') {
          cleanParams[key] = value;
        }
      }
    });
    return api.get<Document[]>('/api/documents/search', { params: cleanParams });
  },

  getOne: (uuid: string) => 
    api.get<Document>(`/api/documents/${uuid}`),

  getById: (uuid: string) =>
    api.get<Document>(`/api/documents/${uuid}`),

  getByCode: (code: string) =>
    api.get<Document>(`/api/documents/code/${code}`),

  // Only exposes admin methods if generic type allows, but API enforces security
  create: (data: CreateDocumentRequest) => 
    api.post<Document>('/api/documents', data),

  update: (uuid: string, data: UpdateDocumentRequest) =>
    api.put<Document>(`/api/documents/${uuid}`, data),

  remove: (uuid: string) =>
    api.delete<void>(`/api/documents/${uuid}`),
    
  uploadFile: (uuid: string, file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return api.post(`/api/documents/${uuid}/files`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
  },

  /**
   * Downloads a file with proper authentication.
   * Returns the file as a blob so it can be downloaded with auth headers.
   */
  downloadFile: async (uuid: string, fileName: string) => {
    const response = await api.get(`/api/documents/${uuid}/files/${fileName}`, {
      responseType: 'blob'
    });
    return response.data;
  },

  /**
   * Updates the description and notes for a specific file entry in files.json.
   * Matches: PUT /documents/{uuid}/files/{fileName}
   */
  updateFileMetadata: (uuid: string, fileName: string, data: { description: string; note: string }) =>
    api.put(`/api/documents/${uuid}/files/${fileName}`, data),

  /**
   * Deletes both the physical file and its entry in files.json.
   * Matches: DELETE /documents/{uuid}/files/{fileName}
   */
  deleteFile: (uuid: string, fileName: string) =>
    api.delete(`/api/documents/${uuid}/files/${fileName}`),

  /**
   * Updates the file contents directly.
   * Matches: PATCH /documents/{uuid}/files/{fileName}/contents
   */
  updateFileContents: (uuid: string, fileName: string, data: { contents: string }) =>
    api.patch(`/api/documents/${uuid}/files/${fileName}/contents`, data),

  /**
   * Renames a file.
   * Matches: PATCH /documents/{uuid}/files/{fileName}/rename
   */
  renameFile: (uuid: string, fileName: string, data: { new_name: string }) =>
    api.patch(`/api/documents/${uuid}/files/${fileName}/rename`, data),

  /**
   * Lists all sync conflicts (admin only).
   * Matches: GET /documents/conflicts
   */
  listConflicts: () =>
    api.get<SyncIssue[]>('/api/documents/conflicts'),

  /**
   * Reloads all documents (admin only).
   * Matches: POST /documents/reload
   */
  reload: () =>
    api.post<ReloadResponse>('/api/documents/reload'),

  /**
   * Reloads a specific folder/document (admin only).
   * Matches: POST /documents/reload/{folderName}
   */
  reloadDocument: (folderName: string) =>
    api.post<ReloadResponse>(`/api/documents/reload/${folderName}`),

  /**
   * Lists available backup files.  
   * Matches: GET /api/backup
   */
  listBackups: () =>
    api.get<{ backups: BackupFile[] }>('/api/documents/backup'),

  /**
   * Downloads a specific backup file by name.
   * Matches: GET /api/backup/download/{fileName}
   */
  downloadBackup: async (fileName: string) => {
    const response = await api.get(`/api/documents/backup/download/${fileName}`, {
      responseType: 'blob'
    });
    return response.data;
  },

  /**
   * Triggers the creation of backup.
   * Returns a job_id for progress tracking.
   */
  createBackup: async () => {
    const response = await api.post<{ job_id: string; message: string }>('/api/documents/backup');
    return response.data;
  },

  /**
   * Restores a backup file that is already on the server.
   * Returns a JobID for progress tracking.
   */
  restoreBackup: (fileName: string, mode: 'merge' | 'overwrite' = 'merge') => {
    return api.post<{ 
      job_id: string; 
      message: string; 
    }>(`/api/documents/backup/restore?file_name=${fileName}&mode=${mode}`);
  },

  /**
   * Polls the status of any background job (backup or restore).
   * Matches: GET /api/jobs/{jobID}
   */
  getJobStatus: (jobId: string) => {
    return api.get<{
      type: 'backup' | 'restore';
      progress: number;
      status: 'processing' | 'completed' | 'failed';
      error?: string;
    }>(`/api/jobs/${jobId}`); 
  },
};
