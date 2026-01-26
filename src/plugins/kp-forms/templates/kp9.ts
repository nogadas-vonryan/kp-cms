import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp9: KPTemplate = {
  id: 'kp9',
  name: 'KP Form #9: Patawag (Summons)',
  description: 'Official summons for the respondent to appear for mediation/conciliation',
  fields: [
    { key: 'respondents', label: 'Respondent/s', type: 'text', required: true, placeholder: 'Name/s of Respondent/s' },
    { key: 'complainants', label: 'Complainant/s', type: 'text', required: true, placeholder: 'Name/s of Complainant/s' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true, placeholder: 'e.g., Unpaid Debt' },
    { key: 'hearingDate', label: 'Hearing Date', type: 'date', required: true },
    { key: 'hearingTime', label: 'Hearing Time', type: 'text', required: true, placeholder: 'e.g., 9:00 AM' },
    { key: 'punongBarangay', label: 'Punong Barangay/Lupon Chairman', type: 'text', required: true }
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
          text: 'S U M M O N S\n\n',
          fontSize: 14
        },
        {
          columns: [
            { width: 30, text: 'TO:' },
            {
              stack: [
                { text: `${data.respondents || ''}\n`, decoration: 'underline' },
                'Respondents\n\n\n'
              ]
            }
          ]
        },
        {
          text: [
            'You are hereby summoned to appear before me in person, together ',
            'with your witnesses, on the ',
            { text: formatDate(data.hearingDate), bold: true },
            ' at ',
            { text: `${data.hearingTime}`, bold: true },
            ' o\'clock in the morning/afternoon, then and there to ',
            'answer to a complaint made before me, copy of which is attached ',
            'hereto, for mediation/conciliation of your dispute with complainant/s.\n\n',
            'You are hereby warned that if you refuse or willfully fail to appear in ',
            'obedience to this summons, you may be barred from filing any ',
            'counterclaim arising from said complaint.\n\n',
            'FAIL NOT or else face punishment as for contempt of court.\n\n\n'
          ]
        },
        {
          stack: [
            { text: `${data.punongBarangay || ''}\n`, decoration: 'underline' },
            'Punong Barangay/Lupon Chairman'
          ],
          margin: [0, 20, 0, 0]
        }
      ]
    };
  }
};