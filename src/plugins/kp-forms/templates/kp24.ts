import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp24: KPTemplate = {
  id: 'kp24',
  name: 'KP Form #24: Certification to Bar Counterclaim',
  description: 'Certification barring respondent from filing counterclaims due to willful failure to appear',
  fields: [
    { key: 'complainants', label: 'Complainant/s', type: 'array', required: true, placeholder: 'e.g., Juan Dela Cruz' },
    { key: 'respondents', label: 'Respondent/s', type: 'array', required: true, placeholder: 'e.g., Pedro Reyes' },
    { key: 'caseNo', label: 'Barangay Case No.', type: 'text', required: true },
    { key: 'for', label: 'For', type: 'text', required: true, placeholder: 'Nature of dispute' },
    { key: 'respondent1', label: 'First Respondent Name', type: 'text', required: true, placeholder: 'Primary respondent from above' },
    { key: 'respondent2', label: 'Second Respondent Name (Optional)', type: 'text', required: false },
    { key: 'dateIssued', label: 'Date Issued', type: 'date', required: true },
    { key: 'secretaryName', label: 'Secretary Name', type: 'text', required: true },
    { key: 'chairmanName', label: 'Chairman Name (Attested by)', type: 'text', required: true },
    { 
      key: 'officeType', 
      label: 'Office Type', 
      type: 'select', 
      options: ['Lupon', 'Pangkat'], 
      defaultValue: 'Lupon',
      required: true 
    }
  ],
  generatePdf: (data): TDocumentDefinitions => {
    const formatDate = (dateStr: any): string => {
      if (!dateStr) return '';
      const date = new Date(dateStr);
      return date.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
    };

    const secretaryLabel = data.officeType === 'Pangkat' ? 'Pangkat Secretary' : 'Lupon Secretary';
    const chairmanLabel = data.officeType === 'Pangkat' ? 'Pangkat Chairman' : 'Lupon Chairman';

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
          text: '\nCERTIFICATION TO BAR COUNTERCLAIM\n\n',
          margin: [0, 10, 0, 10]
        },
        {
          text: [
            'This is to certify that respondent/s ',
            { text: data.respondent1 || '_____________', decoration: 'underline' },
            ' (name) and ',
            { text: data.respondent2 || '_____________', decoration: 'underline' },
            ' (name) have been found to have willfully failed or refused to appear without justifiable reason before the Punong Barangay/Pangkat ng Tagapagkasundo and therefore respondent/s is/are barred from filing his/their counterclaim (if any) arising from the complaint in court/government office.\n\n',
            `This ${formatDate(data.dateIssued)}.\n\n\n`,
          ],
          lineHeight: 1.3
        },
        {
          stack: [
            { text: data.secretaryName || '', decoration: 'underline' },
            `${secretaryLabel}\n\n`,
            'Attested:\n\n',
            { text: data.chairmanName || '', decoration: 'underline' },
            `${chairmanLabel}\n\n\n`,
          ] 
        },
        {
          fontSize: 8,
          italics: true,
          text: 'IMPORTANT: If Lupon Secretary makes the certification, the Lupon Chairman attests. If the Pangkat Secretary makes the certification, the Pangkat Chairman attests.'
        }
      ]
    };
  }
};