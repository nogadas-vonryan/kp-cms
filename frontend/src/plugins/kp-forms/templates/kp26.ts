import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp26: KPTemplate = {
  id: 'kp26',
  name: 'KP Form #26: Notice of Hearing (Re: Motion for Execution)',
  description: 'Notice to parties regarding the hearing for a filed motion for execution',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'array', required: true, placeholder: 'e.g., Juan Dela Cruz' },
    { key: 'respondents', label: 'Respondent/s', type: 'array', required: true, placeholder: 'e.g., Pedro Reyes' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true },
    { key: 'hearingDate', label: 'Hearing Date', type: 'date', required: true },
    { key: 'hearingTime', label: 'Hearing Time', type: 'text', required: true, placeholder: 'e.g., 9:00 AM' },
    { key: 'dateIssued', label: 'Date Issued', type: 'date', required: true },
    { key: 'receivedBy', label: 'Punong Barangay/Lupon Chairman', type: 'text', required: true }
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
          text: '\nNOTICE OF HEARING\n(Re: MOTION FOR EXECUTION)\n\n',
          margin: [0, 10, 0, 10]
        },
        {
          text: [
            'You are hereby notified to appear before me on the ',
            { text: formatDate(data.hearingDate), decoration: 'underline' },
            ' at ',
            { text: data.hearingTime || '_______', decoration: 'underline' },
            ' for the hearing of the motion for execution copy of which is attached hereto, filed by (complainant/s/respondent/s).\n\n\n',
          ],
          lineHeight: 1.5
        },
        {
          text: [
            { text: formatDate(data.dateIssued), decoration: 'underline' },
            ' (Date)\n\n',
          ]
        },
        {
          text: [
            { text: data.receivedBy || '', decoration: 'underline' },
            '\nPunong Barangay/Lupon Chairman\n\n\n',
          ]
        },
        {
          text: `Notified this ${formatDate(data.dateIssued)}.\n\n\n`
        },
        {
          stack: [
            '__________________________               ________________________\n',
            {
              marginLeft: 40,
              text: '(Signature)                                         (Signature)\n'
            },
            {
              marginLeft: 30,
              text: 'Complainant/s                                    Respondent/s\n'
            },
          ]
        },
      ]
    };
  }
};