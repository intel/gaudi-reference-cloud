// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import moment from 'moment'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'
import { UpdateFormHelper, setFormValue, setSelectOptions } from '../../utils/updateFormHelper/UpdateFormHelper'

import CloudAccountService from '../../services/CloudAccountService'
import ClustersGraphs from '../../components/metrics/ClustersGraphs'

interface MetricsResponse {
  data: any[]
  unit: string
  xDataKey: string
  yDataKey: string
  label: string
  item: string
  dateFormat: string
  selectedView: string
  startTime: string
  endTime: string
  isError: boolean
  errorMessage: string
  min: number | null
  max: number | null
}

interface ChartData {
  time: string
  value: number
}

type MetricData = Record<string, MetricsResponse[]>

const ClusterGraphsContainer = (props: any): JSX.Element => {
  const throwError = useErrorBoundary()

  const clusters = props?.clusters
  const showDropdown = props?.showClusters || false

  const metricsTypes = [
    'apiserver_cpu',
    'apiserver_memory',
    'apiserver_requestsbyverb',
    'apiserver_requestsbycode',
    'apiserver_latencybyhostname',
    'apiserver_latencybyverb',
    'apiserver_errorsbyhostname',
    'apiserver_errorsbyverb',
    'apiserver_httprequestsbyhostname',
    'etcd_cpu',
    'etcd_memory',
    'etcd_clienttrafficin',
    'etcd_clienttrafficout',
    'etcd_peertrafficin',
    'etcd_peertrafficout',
    'etcd_dbsizeinuse',
    'etcd_heartbeatsendfailurestotal'
  ]

  const metricConfiguration: any = {
    apiserver_cpu: {
      label: 'CPU',
      min: 0,
      max: null
    },
    apiserver_memory: {
      label: 'Memory',
      min: null,
      max: null
    },
    apiserver_requestsbyverb: {
      label: '',
      min: null,
      max: null
    },
    apiserver_requestsbycode: {
      label: '',
      min: null,
      max: null
    },
    apiserver_latencybyhostname: {
      label: '',
      min: null,
      max: null
    },
    apiserver_latencybyverb: {
      label: '',
      min: null,
      max: null
    },
    apiserver_errorsbyhostname: {
      label: '',
      min: 0,
      max: null
    },
    apiserver_errorsbyverb: {
      label: '',
      min: 0,
      max: null
    },
    apiserver_httprequestsbyhostname: {
      label: '',
      min: null,
      max: null
    },
    etcd_cpu: {
      label: 'CPU',
      min: 0,
      max: null
    },
    etcd_memory: {
      label: 'Memory',
      min: null,
      max: null
    },
    etcd_clienttrafficin: {
      label: '',
      min: null,
      max: null
    },
    etcd_clienttrafficout: {
      label: '',
      min: null,
      max: null
    },
    etcd_peertrafficin: {
      label: '',
      min: null,
      max: null
    },
    etcd_peertrafficout: {
      label: '',
      min: null,
      max: null
    },
    etcd_dbsizeinuse: {
      label: '',
      min: 0,
      max: null
    },
    etcd_heartbeatsendfailurestotal: {
      label: '',
      min: 0,
      max: null
    }
  }

  const initialState = {
    form: {
      clusters: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Clusters:',
        maxWidth: '25rem',
        placeholder: 'Please select',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        options: [],
        validationRules: {
          isRequired: false
        }
      },
      viewType: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'View:',
        maxWidth: '15rem',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        options: [],
        validationRules: {
          isRequired: false
        }
      }
    }
  }

  const initialViewsTypes = [
    {
      name: 'Last hour',
      value: '01hours',
      step: 10,
      dateFormat: 'hh:mm a'
    },
    {
      name: 'Last 6 hours',
      value: '06hours',
      step: 30,
      dateFormat: 'hh:mm a'
    },
    {
      name: 'Last 12 hours',
      value: '12hours',
      step: 60,
      dateFormat: 'hh:mm a'
    },
    {
      name: 'Last 24 hours',
      value: '24hours',
      step: 60,
      dateFormat: 'hh:mm a'
    },
    {
      name: 'Last 7 days',
      value: '07days',
      step: 600,
      dateFormat: 'DD/MM hh:mm a'
    },
    {
      name: 'Last 30 days',
      value: '30days',
      step: 3600,
      dateFormat: 'DD/MM hh:mm a'
    }
  ]

  // Local State

  const [selectedResource, setSelectedResource] = useState('')
  const [selectedView, setSelectedView] = useState('01hours')
  const [state, setState] = useState(initialState)
  const [metricData, setMetricData] = useState<MetricData>({})

  // Hooks

  useEffect(() => {
    setForm()
  }, [])

  useEffect(() => {
    getGraphsData()
  }, [selectedResource, selectedView])

  // functions

  const setForm = (): void => {
    const stateUpdated = { ...state }

    if (clusters.length > 0) {
      stateUpdated.form = setSelectOptions('clusters', clusters, stateUpdated.form)
      const selectedValue = clusters[0]?.value

      stateUpdated.form = setFormValue('clusters', selectedValue, stateUpdated.form)
      setSelectedResource(selectedValue)
    }

    stateUpdated.form = setSelectOptions('viewType', initialViewsTypes, stateUpdated.form)
    stateUpdated.form = setFormValue('viewType', initialViewsTypes[0].value, stateUpdated.form)

    setState(stateUpdated)
  }

  const onChange = (event: any, inputName: string): void => {
    const updatedState = { ...state }
    const value: string = event.target.value

    const updatedForm = UpdateFormHelper(value, inputName, updatedState.form)

    if (inputName === 'clusters') {
      setSelectedResource(value)
    } else {
      setSelectedView(value)
    }

    updatedState.form = updatedForm
    setState(updatedState)
  }

  const getGraphsData = (): void => {
    if (selectedResource) {
      setMetricData({})
      const { startTime, endTime } = getUnixTime()
      setMetricData({})
      Promise.all(
        metricsTypes.map((metric) =>
          callQueryAPI(metric, startTime, endTime)
            .then((data: any) => {
              setGraphData(metric, data, startTime, endTime)
            })
            .catch((error: any) => {
              setGraphError(metric, error)
            })
        )
      ).catch((error) => {
        throwError(error)
      })
    }
  }

  const callQueryAPI = (metric: string, startTime: string, endTime: string): any => {
    const selectedViewStep = initialViewsTypes.find((x) => x.value === selectedView)

    const payload = {
      start: String(startTime),
      end: String(endTime),
      step: String(selectedViewStep?.step),
      category: 'metrics',
      metric,
      resourceType: 'IKS'
    }

    return CloudAccountService.getMetricsQueryData(selectedResource, payload)
  }

  const getUnixTime = (): any => {
    const value = selectedView.substring(0, 2)
    const unit = selectedView.substring(2, selectedView.length)

    const startTime = moment()
      .subtract(value, unit as moment.unitOfTime.DurationConstructor)
      .utc()
      .unix()
    const endTime = moment().utc().unix()
    return { startTime, endTime }
  }

  const setGraphData = (metric: string, response: any, startTime: any, endTime: any): void => {
    setMetricData((oldMetricData) => {
      const apiResponse: any[] = response.data.response

      const metricArray: any[] = []
      for (const res of apiResponse) {
        const unit = res.unit
        const item = res.item
        const queryValue = res.queryvalue

        const chartData = getChartData(queryValue, unit)
        const selectedViewDateFormat = initialViewsTypes.find((x) => x.value === selectedView)

        const chartObj: MetricsResponse = {
          data: chartData,
          unit,
          xDataKey: 'time',
          yDataKey: 'value',
          label: metricConfiguration[metric].label,
          item,
          dateFormat: String(selectedViewDateFormat?.dateFormat),
          selectedView,
          startTime: moment.unix(startTime).toISOString(),
          endTime: moment.unix(endTime).toISOString(),
          isError: false,
          errorMessage: '',
          min: metricConfiguration[metric].min,
          max: metricConfiguration[metric].max
        }

        metricArray.push(chartObj)
      }

      const metricResponse: MetricData = {
        [metric]: metricArray
      }
      const newMetricData = { ...oldMetricData, ...metricResponse }

      return newMetricData
    })
  }

  const setGraphError = (metric: string, error: any): void => {
    setMetricData((oldMetricData) => {
      const chartObj: MetricsResponse = {
        data: [],
        unit: '',
        xDataKey: '',
        yDataKey: '',
        label: '',
        item: '',
        dateFormat: '',
        selectedView: '',
        startTime: '',
        endTime: '',
        isError: true,
        errorMessage: 'Service currently unavailable. Please check back later.',
        min: null,
        max: null
      }

      const metricResponse: MetricData = {
        [metric]: [chartObj]
      }
      const newMetricData = { ...oldMetricData, ...metricResponse }

      return newMetricData
    })
  }

  const getChartData = (data: any, unit: string): ChartData[] => {
    let newData: ChartData[] = [
      {
        time: '',
        value: 0
      }
    ]

    const getValue = (value: number): number => {
      if (unit === 'percentage') {
        value = value * 100
      }

      if (unit === 'b/s' || unit === 'bytes') {
        value = value / 1000000
      }

      if (unit === 'B/s') {
        value = value / 1000
      }

      return formatNumber(value, 2)
    }

    newData = data.map((item: any) => {
      return {
        value: getValue(item.value),
        time: moment.unix(item.epochtime).toISOString()
      }
    })

    return newData
  }

  return (
    <ClustersGraphs
      onChange={onChange}
      state={state}
      metricData={metricData}
      getGraphsData={getGraphsData}
      showDropdown={showDropdown}
    />
  )
}

export default ClusterGraphsContainer
