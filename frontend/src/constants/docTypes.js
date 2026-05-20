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
