import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp20: KPTemplate = {
  id: 'kp20',
  name: 'Certification to File Action (Lupon Secretary)',
  description: 'Certification issued by the Lupon Secretary when a settlement is repudiated',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'array', required: true, placeholder: 'e.g., Juan Dela Cruz' },
    { key: 'respondents', label: 'Respondent/s', type: 'array', required: true, placeholder: 'e.g., Pedro Reyes' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true },
    { key: 'repudiantee', label: 'Name of Repudiantee', type: 'text', required: true, placeholder: 'Person who repudiated the settlement' },
    { key: 'reasonForRepudiation', label: 'Ground for Repudiation', type: 'textarea', required: true },
    { key: 'dateIssued', label: 'Date Issued', type: 'date', required: true },
    { key: 'luponSecretary', label: 'Lupon Secretary', type: 'text', required: true },
    { key: 'luponChairman', label: 'Lupon Chairman (Attested by)', type: 'text', required: true }
  ],
  generatePdf: (data): TDocumentDefinitions => {
    const formatDate = (dateStr: any): string => {
      if (!dateStr) return '';
      const date = new Date(dateStr);
      return date.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
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
      header: function () {
        return {
          image: 'header',
          width: 594,
          height: 93.6,
          alignment: 'center',
          margin: [0, 7.92, 0, 0]
        };
      },
      background: function () {
        return {
          image: 'seal',
          width: 320.4,
          height: 320.4,
          opacity: 0.15,
          alignment: 'center',
          margin: [0, 308, 0, 0]
        };
      },
      content: [
        {
          alignment: 'center',
          stack: [
            'Republic of the Philippines\n',
            'Province of Metro Manila\n',
            'City of Taguig\n',
            'Barangay Bagumbayan\n',
            'OFFICE OF THE LUPONG TAGAPAMAYAPA\n\n',
          ]
        },
        {
          columns: [
            {
              stack: [
                { text: data.complainants || '', decoration: 'underline' },
                'Complainant/s\n\n',
                '----- against -----\n\n',
                { text: data.respondents || '', decoration: 'underline' },
                'Respondent/s\n\n',
              ]
            },
            {
              marginLeft: 70,
              stack: [
                { text: `Barangay Case No. ${data.caseNo || ''}` },
                { text: `For: ${data.for || ''}`, decoration: 'underline' },
              ]
            }
          ]
        },
        {
          alignment: 'center',
          bold: true,
          text: '\nCERTIFICATION TO FILE ACTION\n\n',
          margin: [0, 10, 0, 10]
        },
        {
          text: [
            'This is to certify that:\n',
            '1. There has been a personal confrontation between the parties before the Punong Barangay/Pangkat ng Tagapagkasundo;\n',
            '2. A settlement was reached;\n',
            '3. The settlement has been repudiated in a statement sworn to before the Punong Barangay by ',
            { text: data.repudiantee || '', decoration: 'underline' },
            ' on ground of ',
            { text: data.reasonForRepudiation || '', decoration: 'underline' },
            '; and\n',
            '4. Therefore, the corresponding complaint for the dispute may now be filed in court/government office.\n\n',
            `This ${formatDate(data.dateIssued)}.\n\n\n`,
          ],
          lineHeight: 1.3
        },
        {
          stack: [
            { text: data.luponSecretary || '', decoration: 'underline' },
            'Lupon Secretary\n\n',
            'Attested:\n\n',
            { text: data.luponChairman || '', decoration: 'underline' },
            'Lupon Chairman\n\n\n',
          ]
        },
      ]
    };
  }
};