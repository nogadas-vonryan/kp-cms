import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp15: KPTemplate = {
  id: 'kp15',
  name: 'KP Form #15: Arbitration Award',
  description: 'Official award rendered by the Lupon Chairman or Pangkat after arbitration proceedings.',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'text', required: true, placeholder: 'Name/s of Complainant/s' },
    { key: 'respondents', label: 'Respondent/s', type: 'text', required: true, placeholder: 'Name/s of Respondent/s' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true, placeholder: 'Nature of dispute' },
    { key: 'awardDetails', label: 'Arbitration Award Details', type: 'textarea', required: true, helpText: 'Describe the terms and resolution of the award' },
    { key: 'madeBy', label: 'Award Made By', type: 'text', required: true, placeholder: 'Name of Punong Barangay/Pangkat Chairman' },
    { key: 'attestedBy', label: 'Attested By', type: 'text', required: true, placeholder: 'Name of Punong Barangay/Lupon Secretary' },
    { key: 'dateIssued', label: 'Date Issued', type: 'date', required: true },
    { key: 'tagapagkasundoMembers', label: 'Tagapagkasundo Members', type: 'textarea', placeholder: 'Enter member names (one per line)' }
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
          text: '\nARBITRATION AWARD\n\n',
          fontSize: 14
        },
        {
          text: [
            'After hearing the testimonies given and careful examination of the evidence ',
            'presented in this case, award is hereby made as follows:\n\n'
          ]
        },
        {
          text: data.awardDetails || '',
          margin: [20, 0, 20, 20],
          decoration: 'underline',
          lineHeight: 1.2
        },
        {
          text: `Made this ${formatDate(data.dateIssued)} at Barangay Bagumbayan, Taguig City.\n\n\n`
        },
        {
          columns: [
            {
              width: '50%',
              stack: [
                { text: `${data.madeBy || ''}\n`, decoration: 'underline' },
                'Punong Barangay/Pangkat Chairman *\n',
                { text: '* To be signed by either, whoever made the arbitration award.', fontSize: 8, italics: true }
              ]
            },
            {
              width: '50%',
              stack: [
                (data.tagapagkasundoMembers || '').split('\n').map((m: string) => ({ 
                  text: `${m.trim()}\n`, 
                  decoration: 'underline',
                  alignment: 'center'
                })),
                { text: 'Member/s', alignment: 'center', fontSize: 10 }
              ]
            }
          ],
          margin: [0, 0, 0, 20]
        },
        {
          stack: [
            'ATTESTED:\n',
            { text: `${data.attestedBy || ''}\n`, decoration: 'underline' },
            'Punong Barangay/Lupon Secretary **\n',
            { 
              text: '** To be signed by the Punong Barangay if the award is made by the Pangkat Chairman, and by the Lupon Secretary if the award is made by the Punong Barangay.', 
              fontSize: 8, 
              italics: true 
            }
          ]
        }
      ]
    };
  }
};