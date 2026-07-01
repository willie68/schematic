// Compute SHA-256 hash from base64-encoded file data
export async function getHash(base64Data) {
  // Decode base64 zu binary
  const binary = atob(base64Data)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }

  // Compute SHA-256
  const hashBuffer = await crypto.subtle.digest('SHA-256', bytes)
  
  // Convert zu hex string
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
}
