import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp8: KPTemplate = {
  id: 'kp8',
  name: 'KP Form #8: Notice of Hearing (Mediation)',
  description: 'Notice of hearing for mediation proceedings before the Lupon',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'array', required: true, placeholder: 'e.g., Juan Dela Cruz' },
    { key: 'hearingDate', label: 'Hearing Date', type: 'date', required: true },
    { key: 'hearingTime', label: 'Hearing Time', type: 'text', required: true, placeholder: 'e.g., 2:00 PM' },
    { key: 'timeOfDay', label: 'Time of Day', type: 'select', options: ['morning', 'afternoon'], required: true, defaultValue: 'afternoon' },
    { key: 'punongBarangay', label: 'Punong Barangay', type: 'select', required: true, defaultValue: 'Hon. [Name]' },
    { key: 'noticeDate', label: 'Notice Date', type: 'date', required: true }
  ],
  generatePdf: (data): TDocumentDefinitions => {
    const formatDate = (dateStr: any): string => {
      if (!dateStr) return '';
      const date = new Date(dateStr);
      return date.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
    };

    const complainantsList = data.complainants ? String(data.complainants).split('\n').filter(Boolean) : [];

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
          text: [
            'Republic of the Philippines\n' +
            'Province of Metro Manila\n' +
            'City of Taguig\n' +
            'Barangay Bagumbayan\n' +
            'OFFICE OF THE LUPONG TAGAPAMAYAPA\n\n'
          ]
        },
        {
          alignment: 'center',
          bold: true,
          text: 'NOTICE OF HEARING\n' +
            '(MEDIATION PROCEEDINGS)\n\n'
        },
        {
          alignment: 'left',
          columns: [
            {
              width: 30,
              text: 'TO: '
            },
            {
              stack: [
                { text: complainantsList.join('\n') + '\n', decoration: 'underline' },
                'Complainant/s \n\n'
              ]
            }
          ]
        },
        {
          text: `You are hereby required to appear before me on the ${formatDate(data.hearingDate)} at ${data.hearingTime || ''} o'clock in the ${data.timeOfDay || 'afternoon'} for the hearing of your complaint.\n\n` +
            `This ${formatDate(data.noticeDate)}.\n\n\n`
        },
        {
          text: [
            { text: `${data.punongBarangay || ''}\n`, decoration: 'underline' },
            'Punong Barangay/Lupon Chairman\n\n' +
            `Notified this ${formatDate(data.noticeDate)}.\n\n`
          ]
        },
        {
          stack: [
            'Complainant/s \n',
            { text: complainantsList.join('\n') + '\n', decoration: 'underline' }
          ]
        }
      ]
    };
  }
};
