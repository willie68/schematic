import api from '../services/api'
import { getHash } from './fileHash'

export async function checkFilePresence(file) {
  try {
    const hash = await getHash(file.data)
    const { data } = await api.post('/api/v1/files/presence', {
      hash,
    })
    return Boolean(data?.presence)
  } catch (_err) {
    return false
  }
}
