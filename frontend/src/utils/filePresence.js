import api from '../services/api'

export async function checkFilePresence(file) {
  try {
    const { data } = await api.post('/api/v1/files/presence', {
      name: file.name,
      mimetype: file.mimetype,
      data: file.data,
    })
    return Boolean(data?.presence)
  } catch (_err) {
    return false
  }
}
