import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp27: KPTemplate = {
  id: 'kp27',
  name: 'KP Form #27: Notice of Execution',
  description: 'Official notice from the Punong Barangay regarding the execution of a settlement or award',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'text', required: true },
    { key: 'respondents', label: 'Respondent/s', type: 'text', required: true },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true },
    { key: 'dateIssued', label: 'Date Signed', type: 'date', required: true },
    { key: 'receivedBy', label: 'Punong Barangay', type: 'text', required: true }
  ],
  generatePdf: (data): TDocumentDefinitions => {
    const formatDate = (dateStr: any): string => {
      if (!dateStr) return '';
      const date = new Date(dateStr);
      return date.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
    };

    return {
      pageSize: { width: 612, height: 936 },
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
          text: '\nNOTICE OF EXECUTION\n\n',
          margin: [0, 10, 0, 10]
        },
        {
          text: [
            'WHEREAS, on (date), an amicable settlement was signed by the parties/arbitration award was rendered by the Punong Barangay/Pangkat ng Tagapagkasundo;\n\n',
            'WHEREAS, the terms and conditions of the settlement/award are as follows:\n',
            '__________________________________________________________________________\n',
            '__________________________________________________________________________\n\n',
            'WHEREAS, the said settlement/award is now final and executory;\n\n',
            'WHEREAS, the (complainant/respondent) filed a motion for execution of the same;\n\n',
            'NOW THEREFORE, in behalf of the Lupong Tagapamayapa and by virtue of the powers vested in me and the Lupon by Rule VII of the Katarungang Pambarangay Law and Rules, I shall cause to be realized from the goods and personal property of (name of party) the sum of (amount of settlement or value of award) and to deliver the same to (complainant/respondent) until the full amount of the settlement or award shall have been made upon receipt hereof.\n\n',
            `Signed this ${formatDate(data.dateIssued)}.\n\n\n`
          ],
          lineHeight: 1.2
        },
        {
          text: [
            { text: data.receivedBy || '', decoration: 'underline' },
            '\nPunong Barangay\n\n\n',
          ]
        },
        {
          text: 'Copy furnished:\n\n'
        },
        {
          alignment: 'center',
          columns: [
            {
              stack: [
                { text: data.complainants || '', decoration: 'underline' },
                'Complainant/s',
              ]
            },
            {
              stack: [
                { text: data.respondents || '', decoration: 'underline' },
                'Respondent/s',
              ]
            }
          ]
        }
      ]
    };
  }
};