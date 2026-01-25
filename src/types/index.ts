export type UserRole = 'RoleAdmin' | 'RoleUser';

export type { PermissionAction } from '@/core/auth/usePermission';
export type { ArchivistPlugin, PluginLocations, PluginContext } from '@/core/plugins/pluginRegistry';

export interface AuthUser {
  id: string;
  role: UserRole;
}

export type DocumentFields = {
  status?: string;
  complainants?: string[];
  respondents?: string[];
  [key: string]: any;
};

export interface DocumentFile {
  file_name: string;
  type: string;
  size: number;
  created_at: string;
}

export interface Document {
  uuid: string;
  code: string;
  folder_name: string;
  title: string;
  fields: DocumentFields;
  files: DocumentFile[];
  created_at: string;
  updated_at: string;
}

export interface CreateDocumentRequest {
  code?: string | null;
  folder_name?: string | null;
  title: string;
  fields: DocumentFields;
}

export interface UpdateDocumentRequest {
  code?: string;
  title?: string;
  fields?: DocumentFields;
}

export interface ApiErrorResponse {
  error: string;
}

export interface SyncIssue {
  type: string;
  path: string;
  message: string;
}

export interface ReloadResponse {
  status: string;
  conflicts: SyncIssue[];
}

export interface AuthResponse {
  user: AuthUser;
  csrf_token: string;
}

// Server returns capitalized keys; we map these to AuthResponse before storing
export interface ServerAuthUser {
  ID: string;
  Role: string;
}

export interface ServerAuthResponse {
  user: ServerAuthUser;
  csrf_token: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}
