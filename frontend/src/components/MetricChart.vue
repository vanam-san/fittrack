<script setup lang="ts">
import {
  BarChart,
  LineChart,
  type BarSeriesOption,
  type LineSeriesOption,
} from 'echarts/charts'
import {
  GridComponent,
  LegendComponent,
  TooltipComponent,
  type GridComponentOption,
  type LegendComponentOption,
  type TooltipComponentOption,
} from 'echarts/components'
import { init, use, type ComposeOption, type EChartsType } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { onBeforeUnmount, onMounted, ref, useId, watch } from 'vue'
import { formatDateRange, formatNumber } from '@/lib/format'

use([
  BarChart,
  LineChart,
  GridComponent,
  LegendComponent,
  TooltipComponent,
  CanvasRenderer,
])

type ChartOption = ComposeOption<
  | BarSeriesOption
  | LineSeriesOption
  | GridComponentOption
  | LegendComponentOption
  | TooltipComponentOption
>

const props = defineProps<{
  title: string
  unit: string
  from: string
  to: string
  type: 'bar' | 'line'
  series: { name: string; values: number[] }[]
  dates: string[]
  periodEnds?: string[]
  empty?: string
  table?: boolean
}>()

const id = useId()
const plot = ref<HTMLDivElement>()
const showData = ref(false)
let chart: EChartsType | undefined
let resize: ResizeObserver | undefined
let theme: MutationObserver | undefined

const timestamp = (date: string) => Date.parse(`${date}T00:00:00Z`)
const dayMs = 86_400_000
const shortDate = new Intl.DateTimeFormat(undefined, {
  month: 'short',
  day: 'numeric',
  timeZone: 'UTC',
})

function rowLabel(index: number) {
  const date = props.dates[index]!
  return formatDateRange(date, props.periodEnds?.[index] ?? date)
}

function draw() {
  if (!plot.value || props.empty || !plot.value.clientWidth) return
  chart ??= init(plot.value, undefined, { renderer: 'canvas' })
  const dark = document.documentElement.classList.contains('dark')
  const text = dark ? '#d4d4d8' : '#52525b'
  const grid = dark ? '#3f3f46' : '#e4e4e7'
  const legend = props.series.length > 1
  const option: ChartOption = {
    animation: false,
    useUTC: true,
    color: dark
      ? ['#60a5fa', '#fbbf24', '#4ade80']
      : ['#2563eb', '#b45309', '#15803d'],
    grid: {
      top: 38,
      left: 12,
      right: 18,
      bottom: legend ? 44 : 14,
      outerBoundsMode: 'same',
    },
    legend: { show: legend, bottom: 0, textStyle: { color: text } },
    tooltip: {
      trigger: 'item',
      renderMode: 'richText',
      confine: true,
      formatter(item) {
        if (Array.isArray(item)) return ''
        const value = Array.isArray(item.value) ? item.value[1] : item.value
        return `${rowLabel(item.dataIndex)}\n${item.seriesName}: ${formatNumber(Number(value))} ${props.unit}`
      },
    },
    xAxis:
      props.type === 'bar'
        ? {
            type: 'category',
            data: props.dates.map((date) => shortDate.format(timestamp(date))),
            axisLabel: { color: text, hideOverlap: true },
          }
        : {
            type: 'time',
            min: timestamp(props.from),
            max: timestamp(props.to) + dayMs - 1,
            minInterval: dayMs,
            axisLabel: {
              color: text,
              hideOverlap: true,
              formatter: (value: number) => shortDate.format(value),
            },
          },
    yAxis: {
      type: 'value',
      name: props.unit,
      nameTextStyle: { color: text, align: 'left' },
      scale: props.type === 'line',
      minInterval:
        props.unit === 'sessions' || props.unit === 'reps' ? 1 : undefined,
      axisLabel: {
        color: text,
        formatter: (value: number) => formatNumber(value),
      },
      splitLine: { lineStyle: { color: grid } },
    },
    series: props.series.map(
      ({ name, values }): BarSeriesOption | LineSeriesOption =>
        props.type === 'bar'
          ? { name, type: 'bar', data: values, barMaxWidth: 32 }
          : {
              name,
              type: 'line',
              data: props.dates.map((date, index) => [
                timestamp(date) + dayMs / 2,
                values[index] ?? null,
              ]),
              showSymbol: values.length < 60,
              symbolSize: 7,
            },
    ),
  }
  chart.setOption(option, { notMerge: true })
}

function disconnect() {
  resize?.disconnect()
  theme?.disconnect()
  chart?.dispose()
  chart = undefined
}

function observePlot() {
  // Empty states can replace the plot element, so dispose its old chart and observers.
  disconnect()
  if (!plot.value) return
  resize = new ResizeObserver(() => {
    if (chart) chart.resize()
    else draw()
  })
  resize.observe(plot.value)
  theme = new MutationObserver(draw)
  theme.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class'],
  })
  draw()
}

onMounted(observePlot)
watch(plot, observePlot, { flush: 'post' })
watch(props, draw, { flush: 'post' })
onBeforeUnmount(disconnect)
</script>

<template>
  <figure class="min-w-0 space-y-3 rounded-lg border border-border p-3 sm:p-4">
    <figcaption :id="id">
      <h3 class="font-medium">
        {{ title }} <span class="text-muted-foreground">({{ unit }})</span>
      </h3>
      <p class="text-xs text-muted-foreground">
        {{ formatDateRange(from, to) }}
      </p>
    </figcaption>
    <p v-if="empty" class="py-5 text-sm text-muted-foreground">{{ empty }}</p>
    <template v-else>
      <div ref="plot" role="img" :aria-labelledby="id" class="h-60 w-full" />
      <details
        v-if="table"
        class="text-sm"
        @toggle="showData = ($event.target as HTMLDetailsElement).open"
      >
        <summary class="cursor-pointer text-muted-foreground">
          View chart data
        </summary>
        <div v-if="showData" class="mt-2 max-h-60 overflow-auto">
          <table class="w-full border-collapse text-left">
            <thead>
              <tr>
                <th class="p-2" scope="col">Date</th>
                <th
                  v-for="item in series"
                  :key="item.name"
                  class="p-2"
                  scope="col"
                >
                  {{ item.name }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(date, index) in dates"
                :key="`${date}-${index}`"
                class="border-t border-border"
              >
                <th class="p-2 font-normal" scope="row">
                  {{ rowLabel(index) }}
                </th>
                <td
                  v-for="item in series"
                  :key="item.name"
                  class="p-2 tabular-nums"
                >
                  {{
                    item.values[index] === undefined
                      ? '—'
                      : formatNumber(item.values[index])
                  }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </details>
    </template>
  </figure>
</template>
