import type { KPTemplate } from '../types';
import type { TDocumentDefinitions } from 'pdfmake/interfaces';
import { HEADER_BASE64, SEAL_BASE64 } from '../old_templates/image-headers-base64';

export const kp9Return: KPTemplate = {
  id: 'kp9-return',
  name: "KP Form #9: Officer's Return",
  description: "Official return form indicating how and when the summons was served to the respondent/s",
  fields: [
    { key: 'respondent1', label: 'Primary Respondent', type: 'text', required: true },
    { key: 'dateServed1', label: 'Date Served (Primary)', type: 'date', required: true },
    { key: 'respondent2', label: 'Secondary Respondent (Optional)', type: 'text' },
    { key: 'dateServed2', label: 'Date Served (Secondary)', type: 'date' },
    { 
      key: 'serviceMethod', 
      label: 'Service Method', 
      type: 'select', 
      options: ['Handed in person', 'Refused to receive', 'Left at dwelling', 'Left at office'],
      required: true 
    },
    { key: 'receivedByThirdPartyName', label: 'Third Party Name (if applicable)', type: 'text', helpText: 'Used for methods 3 and 4', placeholder: 'Name of person who received summons' },
    { key: 'officerName', label: 'Serving Officer Name', type: 'text', required: true, placeholder: 'Name of the officer serving summons' }
  ],
  generatePdf: (data): TDocumentDefinitions => {
    const formatDate = (dateStr: any): string => {
      if (!dateStr) return '___________________';
      const date = new Date(dateStr);
      return date.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
    };

    const today = formatDate(new Date());

    return {
      pageSize: { width: 612, height: 936 },
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
          bold: true,
          text: '\nOFFICER\'S RETURN \n\n',
          fontSize: 14
        },
        {
          text: [
            `I serve this summons upon respondent `,
            { decoration: 'underline', text: `${data.respondent1 || '___________________'}` },
            ` on `,
            { text: formatDate(data.dateServed1) },
            ` and upon respondent `,
            { decoration: 'underline', text: `${data.respondent2 || '___________________'}` },
            ` on `,
            { text: formatDate(data.dateServed2) },
            ` by:\n\n`,
            { text: '(Write name/s of respondent/s before mode by which he/they was/were served.)\n\n', italics: true, fontSize: 10 },
            'Respondent/s\n',
          ]
        },
        {
          columns: [
            {
              width: '40%',
              stack: [
                { text: data.serviceMethod === 'handed_in_person' ? data.respondent1 : '__________________________', decoration: 'underline', alignment: 'center', margin: [0, 0, 0, 15] },
                { text: data.serviceMethod === 'refused_to_receive' ? data.respondent1 : '__________________________', decoration: 'underline', alignment: 'center', margin: [0, 0, 0, 15] },
                { text: data.serviceMethod === 'left_at_dwelling' ? data.respondent1 : '__________________________', decoration: 'underline', alignment: 'center', margin: [0, 0, 0, 25] },
                { text: data.serviceMethod === 'left_at_office' ? data.respondent1 : '__________________________', decoration: 'underline', alignment: 'center', margin: [0, 0, 0, 15] },
              ]
            },
            {
              width: '60%',
              fontSize: 10,
              stack: [
                { text: '1. Handing to him/them said summons in person, or\n\n', margin: [0, 0, 0, 5] },
                { text: '2. Handing to him/them said summons and he/they refused to receive, or\n\n', margin: [0, 0, 0, 5] },
                { text: `3. Leaving said summons at his/their dwelling with ${data.receivedByThirdPartyName || '________'} (name) a person and of suitable age and discretion residing therein, or\n\n`, margin: [0, 0, 0, 5] },
                { text: `4. Leaving said summons at his/their office/place of business with ${data.receivedByThirdPartyName || '________'}, (name) a competent person in charge thereof.\n\n`, margin: [0, 0, 0, 5] }
              ]
            }
          ],
          margin: [0, 10, 0, 20]
        },
        {
          stack: [
            { decoration: 'underline', text: `${data.officerName || ''}\n` },
            'Officer\n\n\n',
            { text: 'Received by Respondent/s representative/s:\n\n', bold: true }
          ]
        },
        {
          columns: [
            {
              alignment: 'center',
              stack: [
                '__________________________\n',
                { text: 'Signature\n\n', fontSize: 9 },
                '__________________________\n',
                { text: 'Signature', fontSize: 9 },
              ]
            },
            {
              alignment: 'center',
              stack: [
                { decoration: 'underline', text: `${today}\n` },
                { text: 'Date\n\n', fontSize: 9 },
                { decoration: 'underline', text: `${today}\n` },
                { text: 'Date', fontSize: 9 },
              ]
            }
          ]
        }
      ]
    };
  }
};