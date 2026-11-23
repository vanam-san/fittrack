const number = new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 })

export const formatNumber = (value: number) => number.format(value)

export function localDate(date = new Date()) {
  const year = String(date.getFullYear()).padStart(4, '0')
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return [year, month, day].join('-')
}

export function formatDate(date: string) {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(
    new Date(`${date}T12:00:00`),
  )
}

export function formatDateRange(from: string, to: string) {
  return from === to
    ? formatDate(from)
    : `${formatDate(from)} – ${formatDate(to)}`
}

export function recentDateRange(days: number) {
  const now = new Date()
  const to = localDate(now)
  now.setDate(now.getDate() - days + 1)
  return { from: localDate(now), to }
}

export function errorMessage(error: unknown) {
  return error instanceof Error && error.message
    ? error.message
    : 'FitTrack could not complete the request. Try again.'
}
