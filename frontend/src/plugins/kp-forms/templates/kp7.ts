import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp7: KPTemplate = {
  id: 'kp7',
  name: 'KP Form #7: Pagsumbong (Complaint)',
  description: 'Official complaint form for Barangay Bagumbayan with header and seal',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'array', required: true, placeholder: 'e.g., Juan Dela Cruz' },
    { key: 'respondents', label: 'Respondent/s', type: 'array', required: true, placeholder: 'e.g., Pedro Reyes' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true, placeholder: 'e.g., 2026-001' },
    { key: 'for', label: 'For (Violation)', type: 'text', required: true, placeholder: 'e.g., Noise Complaint, Property Dispute' },
    { key: 'complaint', label: 'Complaint Details', type: 'textarea', required: true, helpText: 'Describe how your rights and interests were violated' },
    { key: 'reliefSought', label: 'Relief Sought', type: 'textarea', required: true, helpText: 'State what relief/remedy you are seeking' },
    { key: 'date', label: 'Date Filed', type: 'date', required: true },
    { key: 'punongBarangay', label: 'Punong Barangay', type: 'select', defaultValue: 'Hon. [Name]', required: true }
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
          marginBottom: 20,
          stack: [
            { text: 'Republic of the Philippines' },
            { text: 'Province of National Capital Region' },
            { text: 'City of Taguig' },
            { text: 'Barangay Bagumbayan', marginBottom: 20 },
            { text: 'OFFICE OF THE LUPONG TAGAPAMAYAPA', bold: true }
          ]
        },
        {
          columns: [
            {
              stack: [
                { text: `${data.complainants || ''}\n`, decoration: 'underline' },
                'Complainant/s \n\n — against —\n\n',
                { text: `${data.respondents || ''}\n`, decoration: 'underline' },
                'Respondent/s'
              ]
            },
            {
              margin: [90, 0, 0, 0],
              stack: [
                { text: `Barangay Case No. ${data.caseNo || ''}\n` },
                { text: 'For: ' },
                { text: `${data.for || ''}`, decoration: 'underline' }
              ]
            }
          ],
          margin: [0, 10, 0, 10]
        },
        { text: 'C O M P L A I N T', bold: true, alignment: 'center', margin: [0, 15, 0, 10] },
        { 
          text: 'I/WE hereby complain against above named respondent/s for violating my/our rights and interests in the following manner:', 
          margin: [0, 0, 0, 10] 
        },
        { text: data.complaint || '', margin: [0, 0, 0, 10], decoration: 'underline' },
        { 
          text: '\nTHEREFORE, I/WE pray that the following relief/s be granted to me/us in accordance with law and/or equity:', 
          margin: [0, 10] 
        },
        { text: data.reliefSought || '', margin: [0, 0, 0, 10], decoration: 'underline' },
        {
          text: `\nMade this ${formatDate(data.date)}.\n\n`,
          margin: [0, 10]
        },
        { text: `${data.complainants || ''}\n`, decoration: 'underline' },
        { text: 'Complainant/s', marginBottom: 10 },
        {
          text: `Received and filed this ${formatDate(data.date)}.\n\n`,
          margin: [0, 10]
        },
        { text: data.punongBarangay || '', decoration: 'underline' },
        'Punong Barangay'
      ],
      styles: {
        header: {
          fontSize: 16,
          bold: true
        }
      }
    };
  }
};