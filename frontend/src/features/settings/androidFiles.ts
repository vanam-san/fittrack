type AndroidFilePicker = {
  pick(id: number, operation: 'open' | 'save', path: string): void
  discard(path: string): void
}

const host = window as Window & {
  fitTrackFiles?: AndroidFilePicker
  onFitTrackDocument?: (id: number, path: string, error: string) => void
}

export const androidFiles = host.fitTrackFiles
let nextRequest = 0

export function pickAndroidDocument(operation: 'open' | 'save', path = '') {
  return new Promise<string>((resolve, reject) => {
    if (!androidFiles)
      throw new Error('The Android file picker is unavailable.')
    const id = ++nextRequest
    host.onFitTrackDocument = (request, result, error) => {
      if (request !== id) return
      delete host.onFitTrackDocument
      if (error) reject(new Error(error))
      else resolve(result)
    }
    androidFiles.pick(id, operation, path)
  })
}
