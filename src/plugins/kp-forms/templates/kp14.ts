import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp14: KPTemplate = {
  id: 'kp14',
  name: 'KP Form #14: Agreement for Arbitration',
  description: 'Official agreement where parties agree to abide by the arbitration award of the Lupon Chairman or Pangkat.',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'text', required: true, placeholder: 'Name/s of Complainant/s' },
    { key: 'respondents', label: 'Respondent/s', type: 'text', required: true, placeholder: 'Name/s of Respondent/s' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true, placeholder: 'Nature of dispute' },
    { key: 'dateEntered', label: 'Date Entered', type: 'date', required: true },
    { key: 'punongBarangayOrPangkatChairman', label: 'Punong Barangay/Pangkat Chairman', type: 'text', required: true }
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
          text: '\nAGREEMENT FOR ARBITRATION\n\n',
          fontSize: 14
        },
        {
          text: [
            'We hereby agree to submit our dispute for arbitration to the ',
            'Lupon Chairman/Pangkat ng Tagapagkasundo (cross out whichever ',
            'is not applicable) and bind ourselves to abide by the condition ',
            'of the arbitration award to be rendered herein after a hearing ',
            'and that we know of the same to be final and executory upon the ',
            'expiration of fifteen (15) days from the date of the award, ',
            'unless there is a repudiation of the agreement or a petition ',
            'to nullify the award has been filed before the proper court.\n\n',
            'We further state that we have entered into this agreement freely ',
            'and voluntarily, after we have been explained to the nature and ',
            'consequences.\n\n',
            `Entered into this ${formatDate(data.dateEntered)}.\n\n\n`
          ],
          lineHeight: 1.3
        },
        {
          alignment: 'center',
          columns: [
            {
              stack: [
                { text: `${data.complainants || ''}\n`, decoration: 'underline' },
                'Complainant/s'
              ]
            },
            {
              stack: [
                { text: `${data.respondents || ''}\n`, decoration: 'underline' },
                'Respondent/s\n\n\n'
              ]
            }
          ]
        },
        {
          bold: true,
          text: 'ATTESTATION\n\n',
        },
        {
          text: [
            'I hereby certify that the foregoing Agreement for Arbitration was ',
            'entered into by the parties freely and voluntarily, after I had explained ',
            'to them the nature and the consequences of such agreement.\n\n\n'
          ]
        },
        {
          stack: [
            { text: `${data.punongBarangayOrPangkatChairman || ''}\n`, decoration: 'underline' },
            'Punong Barangay/Pangkat Chairman\n',
            { text: '(Cross out whichever one is not applicable.)', italics: true, fontSize: 8 }
          ]
        }
      ]
    };
  }
};