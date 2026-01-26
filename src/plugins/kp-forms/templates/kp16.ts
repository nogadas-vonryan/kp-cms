import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp16: KPTemplate = {
  id: 'kp16',
  name: 'KP Form #16: Amicable Settlement',
  description: 'Official agreement outlining the terms of settlement reached between the complainant and respondent.',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'array', required: true, placeholder: 'e.g., Juan Dela Cruz' },
    { key: 'respondents', label: 'Respondent/s', type: 'array', required: true, placeholder: 'e.g., Pedro Reyes' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true, placeholder: 'Nature of dispute' },
    { key: 'settlementPlan', label: 'Settlement Plan/Terms', type: 'textarea', required: true, helpText: 'Describe the specific terms agreed upon by both parties' },
    { key: 'punongBarangayOrPangkatChairman', label: 'Punong Barangay/Pangkat Chairman', type: 'text', required: true },
    { key: 'dateEntered', label: 'Date Entered', type: 'date', required: true }
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
          text: '\nAMICABLE SETTLEMENT\n\n',
          fontSize: 14
        },
        {
          text: [
            'We, complainant/s and respondent/s in the above-captioned case, ',
            'do hereby agree to settle our dispute as follows:\n\n'
          ]
        },
        {
          text: data.settlementPlan || '',
          margin: [20, 0, 20, 20],
          decoration: 'underline',
          lineHeight: 1.2
        },
        {
          text: [
            'and bind ourselves to comply with good faith with the above ',
            'terms of settlement.\n\n',
            `Entered into this ${formatDate(data.dateEntered)}.\n\n\n`
          ]
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
          margin: [0, 0, 0, 30]
        },
        {
          bold: true,
          text: 'ATTESTATION\n\n',
        },
        {
          text: [
            'I hereby certify that the foregoing amicable settlement was entered into ',
            'by the parties freely and voluntarily, after I had explained to them the ',
            'nature and consequence of such settlement.\n\n\n'
          ]
        },
        {
          stack: [
            { text: `${data.punongBarangayOrPangkatChairman || ''}\n`, decoration: 'underline' },
            'Punong Barangay/Pangkat Chairman'
          ]
        }
      ]
    };
  }
};