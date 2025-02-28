// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  registerables
} from 'chart.js'
import { Line } from 'react-chartjs-2'
import { formatNumber } from '../numberFormatHelper/NumberFormatHelper'
import moment from 'moment'
import 'chartjs-adapter-moment'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, ...registerables)

const LineGraph = (props) => {
  // props
  const data = props.data[0]

  const getUnitLabel = (unitType) => {
    let unitLabel = ''

    switch (unitType) {
      case 'percentage':
        unitLabel = '%'
        break
      case 'ms':
        unitLabel = ' ms'
        break
      case 'io/s':
        unitLabel = ' io/s'
        break
      case 'B/s':
        unitLabel = ' kB/s'
        break
      case 'b/s':
        unitLabel = ' Mb/s'
        break
      case 'bytes':
        unitLabel = ' MiB'
        break
      default:
        break
    }

    return unitLabel
  }

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      intersect: false
    },
    plugins: {
      legend: {
        display: props.data.length > 1,
        position: 'top',
        labels: {
          usePointStyle: true,
          generateLabels: (chart) => {
            const original = ChartJS.defaults.plugins.legend.labels.generateLabels
            const labelsOriginal = original.call(this, chart)
            labelsOriginal.forEach((label) => {
              label.pointStyle = 'line'
              label.lineWidth = 4
            })
            return labelsOriginal
          }
        }
      },
      title: {
        display: false
      },
      tooltip: {
        callbacks: {
          title: function (context) {
            return moment(context[0].raw.time).format('YYYY-MM-DD hh:mm a')
          },
          label: function (context) {
            let label = context.dataset.label || ''
            const parsedY = context.parsed.y
            if (parsedY !== null) {
              label += ' ' + parsedY + getUnitLabel(data.unit)
            }
            return label
          }
        }
      }
    },
    parsing: {
      xAxisKey: data.xDataKey,
      yAxisKey: data.yDataKey
    },
    elements: {
      point: {
        radius: 0
      }
    },
    scales: {
      y: {
        grid: {
          display: true
        },
        ticks: {
          callback: (value) => String(formatNumber(value, 2)) + getUnitLabel(data.unit)
        },
        min: data.min,
        max: data.max
      },
      x: {
        type: 'time',
        grid: {
          display: false
        },
        title: {
          display: false
        },
        min: data.startTime,
        max: data.endTime
      }
    }
  }

  const datasets = props.data.map((x, index) => {
    const label = String(x.label)
    const item = String(x.item)
    return {
      label: label ? label + (item ? ` (${item})` : '') : item,
      data: x.data,
      borderWidth: 1
    }
  })

  const newData = { datasets }

  return <Line options={options} data={newData} height={'300vh'} />
}

export default LineGraph
