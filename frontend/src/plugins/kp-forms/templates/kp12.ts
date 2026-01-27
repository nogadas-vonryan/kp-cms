import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp12: KPTemplate = {
  id: 'kp12',
  name: 'KP Form #12: Notice of Hearing (Conciliation Proceedings)',
  description: 'Official notice for parties to appear for conciliation proceedings before the Pangkat Chairman.',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'array', required: true, placeholder: 'e.g., Juan Dela Cruz' },
    { key: 'respondents', label: 'Respondent/s', type: 'array', required: true, placeholder: 'e.g., Pedro Reyes' },
    { key: 'hearingDate', label: 'Hearing Date', type: 'date', required: true },
    { key: 'hearingTime', label: 'Hearing Time', type: 'text', required: true, placeholder: 'e.g., 9:00 AM' },
    { key: 'pangkatChairman', label: 'Pangkat Chairman', type: 'text', required: true },
    { key: 'dateIssued', label: 'Date Issued', type: 'date', required: true }
  ],
  generatePdf: (data): TDocumentDefinitions => {
    const formatDate = (dateStr: any): string => {
      if (!dateStr) return '';
      const date = new Date(dateStr);
      return date.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'long', 
        day: 'numeric' 
      });
    };

    return {
      pageSize: {
        width: 612,
        height: 936
      },
      pageMargins: [72, 108, 72, 72],
      images: {
        header: HEADER_BASE64,
        seal: SEAL_BASE64
      },
      header: () => ({
        image: 'header',
        width: 594,
        height: 93.6,
        alignment: 'center',
        margin: [0, 7.92, 0, 0]
      }),
      background: () => ({
        image: 'seal',
        width: 320.4,
        height: 320.4,
        opacity: 0.15,
        alignment: 'center',
        margin: [0, 308, 0, 0]
      }),
      content: [
        {
          alignment: 'center',
          stack: [
            'Republic of the Philippines\n',
            'Province of Metro Manila\n',
            'City of Taguig\n',
            'Barangay Bagumbayan\n',
            'OFFICE OF THE PANGKAT TAGAPAGKASUNDO\n\n\n',
          ]
        },
        {
          alignment: 'center',
          bold: true,
          text: 'NOTICE OF HEARING\n(Conciliation Proceedings)\n\n\n',
          fontSize: 14
        },
        {
          columns: [
            { width: 30, text: 'TO:' },
            {
              stack: [
                { text: `${data.complainants || ''}\n`, decoration: 'underline' },
                'Complainants'
              ]
            },
            {
              stack: [
                { text: `${data.respondents || ''}\n`, decoration: 'underline' },
                'Respondents\n\n\n'
              ]
            }
          ],
          margin: [0, 0, 0, 20]
        },
        {
          text: [
            'You are hereby required to appear before me on the ',
            { text: formatDate(data.hearingDate), bold: true },
            ' at ',
            { text: `${data.hearingTime}`, bold: true },
            ' o\'clock in the morning/afternoon for a hearing of the above-entitled complaint.\n\n',
            `This ${formatDate(data.dateIssued)}.\n\n\n`
          ],
          lineHeight: 1.5
        },
        {
          stack: [
            { text: `${data.pangkatChairman || ''}\n`, decoration: 'underline' },
            'Pangkat Chairman\n\n\n'
          ]
        },
        {
          text: `Notified this ${formatDate(data.dateIssued)}.\n\n\n`
        },
        {
          alignment: 'center',
          columns: [
            { width: 30, text: 'TO:' },
            {
              stack: [
                { text: `${data.complainants || ''}\n`, decoration: 'underline' },
                'Complainants'
              ]
            },
            {
              stack: [
                { text: `${data.respondents || ''}\n`, decoration: 'underline' },
                'Respondents'
              ]
            }
          ]
        }
      ]
    };
  }
};