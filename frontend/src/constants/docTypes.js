/**
 * Document types used throughout the application.
 * Centralized definition to ensure consistency across components.
 */
export const DOC_TYPES = [
  { label: 'Schaltplan', value: 'schematic' },
  { label: 'Bedienungsanleitung', value: 'manual' },
  { label: 'Datenblatt', value: 'datasheet' },
  { label: 'Bild', value: 'image' },
  { label: 'Service-Dokumentation', value: 'service' },
  { label: 'Zertifikat', value: 'certificate' },
  { label: 'Dokumentation', value: 'documentation' },
  { label: 'sonstiges', value: 'other' },
]

/**
 * Get a doc type by its value.
 * @param {string} value - The value of the doc type
 * @returns {Object|undefined} The doc type object or undefined if not found
 */
export function getDocTypeByValue(value) {
  return DOC_TYPES.find((dt) => dt.value === value)
}

/**
 * Get a doc type label by its value.
 * @param {string} value - The value of the doc type
 * @returns {string|undefined} The label of the doc type or undefined if not found
 */
export function getDocTypeLabel(value) {
  return getDocTypeByValue(value)?.label
}

/**
 * Get a PrimeIcon class for a doc type value.
 * @param {string} value - The value of the doc type
 * @returns {string} PrimeIcon class name
 */
export function getDocTypeIcon(value) {
  switch (value) {
    case 'schematic':
      return 'pi pi-sitemap'
    case 'manual':
      return 'pi pi-book'
    case 'datasheet':
      return 'pi pi-table'
    case 'image':
      return 'pi pi-image'
    case 'service':
      return 'pi pi-cog'
    case 'certificate':
      return 'pi pi-shield'
    case 'documentation':
      return 'pi pi-book'
    default:
      return 'pi pi-file'
  }
}
