import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp18: KPTemplate = {
  id: 'kp18',
  name: 'KP Form #18: Notice of Hearing (Re: Failure to Appear)',
  description: 'Notice to complainant to explain failure to appear for mediation/conciliation',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'text', required: true },
    { key: 'respondents', label: 'Respondent/s', type: 'text', required: true },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true },
    { key: 'hearingDate', label: 'New Hearing Date', type: 'date', required: true },
    { key: 'hearingTime', label: 'New Hearing Time', type: 'text', required: true, placeholder: 'e.g., 10:00 AM' },
    { key: 'oldHearingDate', label: 'Previous (Missed) Date', type: 'date', required: true },
    { key: 'punongBarangay', label: 'Punong Barangay/Chairman', type: 'text', required: true },
    { key: 'dateIssued', label: 'Date Issued', type: 'date', required: true },
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
          text: '\nNOTICE OF HEARING\n(RE: FAILURE TO APPEAR)\n\n',
          margin: [0, 10, 0, 10]
        },
        {
          columns: [
            { width: 30, text: 'TO:' },
            {
              stack: [
                { text: data.complainants || '', decoration: 'underline' },
                'Complainant/s\n\n'
              ]
            },
          ]
        },
        {
          text: [
            'You are hereby required to appear before me/the Pangkat on the ',
            { text: formatDate(data.hearingDate), decoration: 'underline' },
            ' , at ',
            { text: data.hearingTime, decoration: 'underline' },
            ' o’clock in the morning/afternoon to explain why you failed to appear for mediation/conciliation scheduled on ',
            { text: formatDate(data.oldHearingDate), decoration: 'underline' },
            ' and why your complaint should not be dismissed, a certificate to bar the filing of your action on court/government office should not be issued, and contempt proceedings should not be initiated in court for willful failure or refusal to appear before the Punong Barangay/Pangkat ng Tagapagkasundo.\n\n',
            `This ${formatDate(data.dateIssued)}.\n\n\n`,
          ],
          lineHeight: 1.2
        },
        {
          stack: [
            { text: data.punongBarangay || '', decoration: 'underline' },
            'Punong Barangay/Punong Chairman\n',
            '(Cross out whichever one is not applicable.)\n\n\n',
          ]
        },
        {
          text: `Notified this ${formatDate(data.dateIssued)}.\n\n\n`
        },
        {
          alignment: 'center',
          columns: [
            {
              stack: [
                'Complainant/s',
                { text: data.complainants || '', decoration: 'underline' },
              ]
            },
            {
              stack: [
                'Respondent/s',
                { text: data.respondents || '', decoration: 'underline' },
              ]
            }
          ]
        },
      ]
    };
  }
};