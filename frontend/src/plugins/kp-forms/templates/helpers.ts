import type { Content, TDocumentDefinitions, StyleDictionary } from 'pdfmake/interfaces';
import type { KPFormField, KPTemplate } from '../types';

const sharedStyles: StyleDictionary = {
  overline: {
    fontSize: 10,
    color: '#4b5563',
    margin: [0, 0, 0, 2] as [number, number, number, number]
  },
  title: {
    fontSize: 18,
    bold: true,
    margin: [0, 4, 0, 4] as [number, number, number, number]
  },
  docTitle: {
    fontSize: 14,
    bold: true,
    margin: [0, 0, 0, 2] as [number, number, number, number]
  },
  description: {
    fontSize: 10,
    italics: true,
    color: '#374151',
    margin: [0, 0, 0, 12] as [number, number, number, number]
  },
  tableHeader: {
    bold: true,
    fillColor: '#e5e7eb'
  },
  note: {
    fontSize: 10,
    italics: true,
    color: '#4b5563'
  },
  footer: {
    fontSize: 9,
    color: '#6b7280'
  }
};

function normalizeValue(raw: unknown): string {
  if (raw === undefined || raw === null) return '';
  if (Array.isArray(raw)) return raw.join(', ');
  if (raw instanceof Date) return raw.toLocaleDateString();
  return String(raw);
}

export function createStandardTemplate(config: {
  id: string;
  name: string;
  description: string;
  fields: KPFormField[];
  footerNote?: string;
}): KPTemplate {
  return {
    id: config.id,
    name: config.name,
    description: config.description,
    fields: config.fields,
    generatePdf: (data): TDocumentDefinitions => {
      const body = [
        [
          { text: 'Field', style: 'tableHeader' },
          { text: 'Details', style: 'tableHeader' }
        ],
        ...config.fields.map((field) => [
          { text: field.label, bold: true },
          { text: normalizeValue((data as Record<string, unknown>)[field.key]) }
        ])
      ];

      const today = new Date();

      const content: Content[] = [
        { text: 'Republic of the Philippines', style: 'overline', alignment: 'center' },
        { text: 'Katarungang Pambarangay', style: 'title', alignment: 'center' },
        { text: config.name, style: 'docTitle', alignment: 'center' },
        { text: config.description, style: 'description', alignment: 'center' },
        {
          table: {
            widths: ['32%', '*'],
            body
          },
          layout: 'lightHorizontalLines'
        },
        config.footerNote
          ? { text: config.footerNote, style: 'note', margin: [0, 12, 0, 0] }
          : undefined,
        { text: `Generated on ${today.toLocaleDateString()}`, style: 'footer', margin: [0, 16, 0, 0] }
      ].filter(Boolean) as Content[];

      return {
        content,
        styles: sharedStyles
      };
    }
  };
}
