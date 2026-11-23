import { formatNumber } from './format'

const kgPerPound = 0.45359237

export function displayWeight(kg: number, unit: string) {
  return unit === 'lb' ? kg / kgPerPound : kg
}

export function formatWeight(kg: number, unit: string) {
  return `${formatNumber(displayWeight(kg, unit))} ${unit}`
}
