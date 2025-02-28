// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Doughnut } from 'react-chartjs-2'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'

ChartJS.register(ArcElement, Tooltip, Legend)

interface DoughnutChartProps {
  isDarkMode: boolean
  running: number
  total: number
}

const DoughnutChart: React.FC<DoughnutChartProps> = (props): JSX.Element => {
  const valueColor = props.isDarkMode ? '#30E9CB' : '#040E35'
  const backgroundColor = props.isDarkMode ? '#585A62' : '#C4C7CF'

  const emptyDataSet = [
    {
      label: '',
      data: [1],
      backgroundColor: [backgroundColor],
      borderWidth: 0
    }
  ]

  const dataSet = [
    {
      label: '',
      data: [props.running, props.total - props.running],
      backgroundColor: [valueColor, backgroundColor],
      borderWidth: 0
    }
  ]

  const data = {
    labels: [],
    datasets: props.total > 0 ? dataSet : emptyDataSet
  }

  const options = {
    cutoutPercentage: 150,
    plugins: {
      legend: {
        display: false,
        position: 'right' as const
      },
      tooltip: {
        enabled: false
      }
    }
  }

  return <Doughnut data={data} options={options} />
}

export default DoughnutChart
