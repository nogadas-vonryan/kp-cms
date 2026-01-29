// Configuration for specific field options
const FIELD_OPTIONS = {
  nature: [
    { value: 'civil', label: 'Civil' },
    { value: 'criminal', label: 'Criminal' }
  ],
  status: [
    { value: 'case_filed', label: 'Case Filed' },
    { value: 'arbitration', label: 'Arbitration' },
    { value: 'mediation', label: 'Mediation' },
    { value: 'conciliation', label: 'Conciliation' },
    { value: 'repudiation', label: 'Repudiation' },
    { value: 'resolved', label: 'Resolved' }
  ]
};

const FIELD_DISPLAY_ORDER = [
  'nature',
  'status',
  'complaint',
  'complainants',
  'respondents'
];

export function useDocumentFields() {
  
  function getFieldOptions(key: string) {
    const opts = FIELD_OPTIONS[key as keyof typeof FIELD_OPTIONS] as { value: string; label: string }[] | undefined;
    return opts ? [...opts] : undefined;
  }

  function formatLabel(key: string): string {
    return key
      .split('_')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ');
  }

  function toSnakeCase(str: string): string {
    return str
      .trim()
      .toLowerCase()
      .replace(/\s+/g, '_')
      .replace(/-+/g, '_')
      .replace(/[^\w]/g, '')
      .replace(/__+/g, '_');
  }

  function getTextareaRows(value: any, key?: string): number {
    // Special case: complaint field gets 5 rows by default
    if (key === 'complaint') return 5;
    
    if (!value) return 1;
    const text = value.toString();
    // Use 2 rows if text contains newlines or is longer than 60 characters
    if (text.includes('\n') || text.length > 60) {
      return 2;
    }
    return 1;
  }

  // Returns a sorted array of [key, value] pairs based on priority config
  function getSortedFields(fields: Record<string, any>) {
    const fieldsArray = Object.entries(fields || {});
    
    return fieldsArray.sort(([keyA], [keyB]) => {
      const indexA = FIELD_DISPLAY_ORDER.indexOf(keyA);
      const indexB = FIELD_DISPLAY_ORDER.indexOf(keyB);
      
      const priorityA = indexA === -1 ? FIELD_DISPLAY_ORDER.length : indexA;
      const priorityB = indexB === -1 ? FIELD_DISPLAY_ORDER.length : indexB;
      
      if (priorityA !== priorityB) {
        return priorityA - priorityB;
      }
      
      return keyA.localeCompare(keyB);
    });
  }

  return {
    getFieldOptions,
    formatLabel,
    toSnakeCase,
    getTextareaRows,
    getSortedFields
  };
}