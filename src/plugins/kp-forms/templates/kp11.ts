import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp11: KPTemplate = {
  id: 'kp11',
  name: 'KP Form #11: Notice to Chosen Pangkat Member',
  description: 'Official notice to a person informing them they have been chosen as a member of the Pangkat ng Tagapagkasundo.',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'text', required: true, placeholder: 'Name/s of Complainant/s' },
    { key: 'respondents', label: 'Respondent/s', type: 'text', required: true, placeholder: 'Name/s of Respondent/s' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true, placeholder: 'Nature of dispute' },
    { key: 'pangkatMember', label: 'Chosen Pangkat Member', type: 'text', required: true },
    { key: 'dateIssued', label: 'Date Issued', type: 'date', required: true },
    { key: 'barangayOfficial', label: 'Punong Barangay/Lupon Secretary', type: 'text', required: true }
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
            'Republic of the Philippines\n' +
            'Province of Metro Manila\n' +
            'City of Taguig\n' +
            'Barangay Bagumbayan\n' +
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
          fontSize: 13,
          text: '\nNOTICE TO CHOSEN PANGKAT MEMBERS \n\n'
        },
        {
          columns: [
            {
              alignment: 'left',
              text: [
                `TO: `,
                { decoration: 'underline', text: `${data.pangkatMember || ''}\n` },
              ]
            },
            {
              alignment: 'right',
              text: [
                `Date: `,
                { decoration: 'underline', text: `${formatDate(data.dateIssued)}\n\n` },
              ]
            }
          ]
        },
        {
          text: [
            'Notice is hereby given that you have been chosen member of the ',
            'Pangkat ng Tagapagkasundo amicably conciliate the dispute between ',
            'the parties in the above-entitled case.\n\n',
          ],
          lineHeight: 1.3
        },
        {
          stack: [
            { decoration: 'underline', text: `${data.barangayOfficial || ''}\n` },
            'Punong Barangay/Lupon Secretary\n\n\n',
          ],
          margin: [0, 20, 0, 30]
        },
        {
          text: `Received this ${formatDate(data.dateIssued)}.\n\n`,
        },
        {
          stack: [
            { decoration: 'underline', text: `${data.pangkatMember || ''}\n` },
            'Pangkat Member'
          ]
        }
      ]
    };
  }
};