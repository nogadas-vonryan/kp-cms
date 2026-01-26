import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp25: KPTemplate = {
  id: 'kp25',
  name: 'KP Form #25: Motion for Execution',
  description: 'Request for a writ of execution after a settlement or award becomes final',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'text', required: true },
    { key: 'respondents', label: 'Respondent/s', type: 'text', required: true },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true },
    { key: 'settlementDate', label: 'Date of Settlement/Award', type: 'date', required: true },
    { key: 'dateIssued', label: 'Date of Motion', type: 'date', required: true },
    { 
      key: 'movingParty', 
      label: 'Moving Party Type', 
      type: 'select', 
      options: ['Complainant/s', 'Respondent/s'], 
      defaultValue: 'Complainant/s',
      required: true 
    },
    { key: 'movingPartyName', label: 'Name of Moving Party', type: 'text', required: true }
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
          text: '\nMOTION FOR EXECUTION\n\n',
          margin: [0, 10, 0, 10]
        },
        {
          text: [
            `${data.movingParty || 'Complainant/s/Respondent/s'} state as follows:\n\n`,
            '1. On ',
            { text: formatDate(data.settlementDate), decoration: 'underline' },
            ' the parties in this case signed an amicable settlement/received the arbitration award rendered by the Lupon Chairman/Pangkat ng Tagapagkasundo;\n\n',
            '2. The period of ten (10) days from the above-stated date has expired without any of the parties filing a sworn statement of repudiation of the settlement before the Lupon Chairman or a petition for nullification of the arbitration award in court; and\n\n',
            '3. The amicable settlement/arbitration award is now final and executory.\n\n',
            'WHEREFORE, ',
            { text: data.movingParty || 'Complainant/s/Respondent/s' },
            ' request/s that the corresponding writ of execution be issued by the Lupon Chairman in this case.\n\n\n',
          ],
          lineHeight: 1.2
        },
        {
          stack: [
            { text: formatDate(data.dateIssued), decoration: 'underline' },
            '(Date)\n\n\n',
            '_______________________\n',
            { text: data.movingPartyName || 'Moving Party Signature' },
            `\n${data.movingParty || 'Complainant/s/Respondent/s'}`
          ]
        }
      ]
    };
  }
};