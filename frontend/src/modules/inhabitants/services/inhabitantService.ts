import api from '@/core/api/client';

export interface Inhabitant {
  id?: number;
  first_name: string;
  last_name: string;
  middle_name: string;
  suffix: string;
  birthday: string;
  contact_no: string;
  address: string;
}

export const InhabitantService = {
  getAll(offset = 0, limit = 100) {
    return api.get<Inhabitant[]>('/api/inhabitants', { params: { offset, limit } });
  },
  getById(id: number) {
    return api.get<Inhabitant>(`/api/inhabitants/${id}`);
  },
  getInhabitantDocuments(id: number) {
    return api.get<any[]>(`/api/inhabitants/${id}/documents`);
  },
  create(data: Inhabitant) {
    return api.post<Inhabitant>('/api/inhabitants', data);
  },
  update(id: number, data: Inhabitant) {
    return api.put<Inhabitant>(`/api/inhabitants/${id}`, data);
  },
  delete(id: number) {
    return api.delete(`/api/inhabitants/${id}`);
  }
};