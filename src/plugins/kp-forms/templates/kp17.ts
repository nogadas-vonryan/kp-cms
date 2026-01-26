import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp17: KPTemplate = {
  id: 'kp17',
  name: 'KP Form #17: Repudiation',
  description: 'Official form for repudiating a settlement or an agreement for arbitration on grounds of fraud, violence, or intimidation.',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'array', required: true, placeholder: 'e.g., Juan Dela Cruz' },
    { key: 'respondents', label: 'Respondent/s', type: 'array', required: true, placeholder: 'e.g., Pedro Reyes' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true, placeholder: 'Nature of dispute' },
    { key: 'repudiationGrounds', label: 'Grounds for Repudiation', type: 'textarea', required: true, helpText: 'Describe the fraud, violence, or intimidation used to get your consent' },
    { key: 'arbitratedBy', label: 'Arbitrated/Settled Before', type: 'text', required: true, placeholder: 'Name of Punong Barangay/Chairman/Member' },
    { key: 'attestedBy', label: 'Attested By (Punong Barangay)', type: 'text', required: true },
    { key: 'dateFiled', label: 'Date Filed', type: 'date', required: true }
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

    const formatTime = (dateStr: any): string => {
      if (!dateStr) return '';
      const date = new Date(dateStr);
      return date.toLocaleTimeString('en-US', { 
        hour: '2-digit', 
        minute: '2-digit' 
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
          text: '\nR E P U D I A T I O N\n\n',
          fontSize: 14
        },
        {
          text: [
            'I/WE hereby repudiate the settlement/agreement for arbitration ',
            'entered into in this case on the ground that my/our consent was ',
            'vitiated by:\n\n'
          ]
        },
        {
          text: data.repudiationGrounds || '',
          margin: [20, 0, 20, 20],
          decoration: 'underline',
          lineHeight: 1.2
        },
        {
          text: '(Cross out whichever is not applicable)\n\n',
          fontSize: 8,
          italics: true
        },
        {
          alignment: 'center',
          columns: [
            {
              stack: [
                'Complainant/s\n',
                { text: `${data.complainants || ''}\n`, decoration: 'underline' }
              ]
            },
            {
              stack: [
                'Respondent/s\n',
                { text: `${data.respondents || ''}\n`, decoration: 'underline' }
              ]
            }
          ],
          margin: [0, 0, 0, 20]
        },
        {
          text: `SUBSCRIBED AND SWORN TO before me this ${formatDate(data.dateFiled)} at ${formatTime(data.dateFiled)}.\n\n`
        },
        {
          stack: [
            { text: `${data.arbitratedBy || ''}\n`, decoration: 'underline' },
            'Punong Barangay/Punong Chairman/Member\n\n'
          ]
        },
        {
          stack: [
            `Received and filed this ${formatDate(data.dateFiled)}.\n\n`,
            { text: `${data.attestedBy || ''}\n`, decoration: 'underline' },
            'Punong Barangay\n\n',
            { 
              text: '* Failure to repudiate the settlement or the arbitration agreement within the time limits respectively set (ten [10] days from the date of settlement and five [5] days from the date of arbitration agreement) shall be deemed a waiver of the right to challenge on said grounds.', 
              fontSize: 8, 
              italics: true 
            }
          ]
        }
      ]
    };
  }
};