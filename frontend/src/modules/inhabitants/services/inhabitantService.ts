import api from '@/core/api/client';

export interface Inhabitant {
  uuid?: string;
  first_name: string;
  last_name: string;
  middle_name: string;
  suffix: string;
  birthdate: string;
  contact_no: string;
  address: string;
  civil_status?: string;
  citizenship?: string;
  inhabitant_type?: string;
  sex?: string;
  birth_place?: string;
  occupation?: string;
  email_address?: string;
  highest_educational_attainment?: string;
  mother_first_name?: string;
  mother_middle_name?: string;
  mother_last_name?: string;
}

export const CIVIL_STATUS_OPTIONS = [
  'single', 'married', 'widowed', 'divorced', 'separated', 'common_law_live_in', 'unknown', 'annulled'
] as const;

export const CITIZENSHIP_OPTIONS = [
  'filipino', 'foreigner'
] as const;

export const INHABITANT_TYPE_OPTIONS = [
  'non_migrant', 'migrant', 'transient'
] as const;

export const InhabitantService = {
  getAll(offset = 0, limit = 100) {
    return api.get<Inhabitant[]>('/api/inhabitants', { params: { offset, limit } });
  },
  search(query: string, limit = 20) {
    return api.get<Inhabitant[]>('/api/inhabitants', { params: { q: query, limit } });
  },
  getById(id: string) {
    return api.get<Inhabitant>(`/api/inhabitants/${id}`);
  },
  getInhabitantDocuments(id: string) {
    return api.get<any[]>(`/api/inhabitants/${id}/documents`);
  },
  create(data: Inhabitant) {
    return api.post<Inhabitant>('/api/inhabitants', data);
  },
  update(id: string, data: Inhabitant) {
    return api.put<Inhabitant>(`/api/inhabitants/${id}`, data);
  },
  delete(id: string) {
    return api.delete(`/api/inhabitants/${id}`);
  }
};