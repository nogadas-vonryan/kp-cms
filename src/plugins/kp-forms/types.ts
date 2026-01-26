import type { TDocumentDefinitions } from 'pdfmake/interfaces';

export interface KPFormField {
  key: string;
  label: string;
  type: 'text' | 'textarea' | 'date' | 'select' | 'array';
  options?: string[]; // For select inputs
  defaultValue?: string;
  required?: boolean;
  placeholder?: string;
  helpText?: string;
}

export interface KPTemplate {
  id: string;
  name: string; // e.g., "KP Form #7: Complaint"
  description: string;
  fields: KPFormField[];
  // Function to convert user data into PDFMake JSON format
  generatePdf: (data: Record<string, any>) => TDocumentDefinitions;
}