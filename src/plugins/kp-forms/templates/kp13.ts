import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp13: KPTemplate = {
  id: 'kp13',
  name: 'KP Form #13: Subpoena (Saksi)',
  description: 'Official subpoena for witnesses to appear and testify in a hearing.',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'text', required: true, placeholder: 'Name/s of Complainant/s' },
    { key: 'respondents', label: 'Respondent/s', type: 'text', required: true, placeholder: 'Name/s of Respondent/s' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true, placeholder: 'Nature of dispute' },
    { key: 'witnesses', label: 'Witnesses', type: 'textarea', required: true, placeholder: 'Enter names of witnesses (one per line)' },
    { key: 'appearanceDate', label: 'Appearance Date', type: 'date', required: true },
    { key: 'appearanceTime', label: 'Appearance Time', type: 'text', required: true, placeholder: 'e.g., 9:00 AM' },
    { key: 'dateIssued', label: 'Date Issued', type: 'date', required: true },
    { key: 'punongBarangay', label: 'Punong Barangay/Lupon Chairman', type: 'text', required: true }
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
            'OFFICE OF THE LUPONG TAGAPAMAYAPA\n\n',
          ]
        },
        {
          columns: [
            {
              marginLeft: 30,
              stack: [
                { text: `${data.complainants || ''}\n`, decoration: 'underline' },
                'Complainant/s\n\n',
                '----- against -----\n\n',
                { text: `${data.respondents || ''}\n`, decoration: 'underline' },
                'Respondent/s\n\n\n'
              ]
            },
            {
              marginLeft: 70,
              stack: [
                `Barangay Case No. ${data.caseNo || ''}\n`,
                {
                  text: [
                    'For: ',
                    { text: `${data.for || ''}`, decoration: 'underline' }
                  ]
                }
              ]
            }
          ]
        },
        {
          alignment: 'center',
          bold: true,
          text: '\nS U B P O E N A\n\n',
          fontSize: 14
        },
        {
          alignment: 'center',
          columns: [
            { width: 30, text: 'TO:' },
            {
              decoration: 'underline',
              stack: (data.witnesses || '').split('\n').map((w: string) => ({ text: w.trim() }))
            }
          ]
        },
        {
          marginLeft: 30,
          alignment: 'center',
          text: 'Witnesses\n\n',
          fontSize: 10
        },
        {
          text: [
            'You are hereby commanded to appear before me on ',
            { text: formatDate(data.appearanceDate), bold: true },
            ', at ',
            { text: `${data.appearanceTime}`, bold: true },
            ' o\'clock, then and there to testify in the hearing of the above-captioned case.\n\n',
            `This ${formatDate(data.dateIssued)}.\n\n\n`
          ],
          lineHeight: 1.5
        },
        {
          stack: [
            { text: `${data.punongBarangay || ''}\n`, decoration: 'underline' },
            'Punong Barangay/Lupon Chairman\n',
            { text: '(Cross out whichever one is not applicable.)', fontSize: 8, italics: true }
          ]
        }
      ]
    };
  }
};