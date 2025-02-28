// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import { BsArrowClockwise } from 'react-icons/bs'

import './MetricsGraphs.scss'
import LoaderItem from './partials/LoaderItem'
import ChartItem from './partials/ChartItem'
import BuildMetricsCustomInput from './partials/BuildMetricsCustomInput'
import idcConfig from '../../config/configurator'
import moment from 'moment'
import CustomAlerts from '../../utils/customAlerts/CustomAlerts'

const MetricsGraphs = (props: any): JSX.Element => {
  // props
  const onChange = props.onChange
  const state = props.state
  const metricData = props.metricData
  const getGraphsData = props.getGraphsData
  const showInstances = props.showInstances
  const selectedResourceCategory = props.selectedResourceCategory
  const hideNotice = props.hideNotice

  const instanceInput = showInstances ? (
    <BuildMetricsCustomInput name="instances" input={state.form.instances} onChange={onChange} />
  ) : (
    ''
  )
  const viewTypeInput = <BuildMetricsCustomInput name="viewType" input={state.form.viewType} onChange={onChange} />

  const getMetric = (): JSX.Element => {
    return (
      <div className="row g-s8">
        {metricData?.cpu ? (
          <ChartItem metricData={metricData.cpu} title="CPU usage" />
        ) : (
          <LoaderItem title="CPU usage" />
        )}

        {metricData?.memory ? (
          <ChartItem metricData={metricData.memory} title="Memory usage" />
        ) : (
          <LoaderItem title="Memory usage" />
        )}

        {selectedResourceCategory === 'BareMetalHost' ? (
          <>
            {metricData?.network_receive_bytes ? (
              <ChartItem metricData={metricData.network_receive_bytes} title="Network usage received" />
            ) : (
              <LoaderItem title="Network usage received" />
            )}

            {metricData?.network_transmit_bytes ? (
              <ChartItem metricData={metricData.network_transmit_bytes} title="Network usage transmit" />
            ) : (
              <LoaderItem title="Network usage transmit" />
            )}

            {metricData?.disk ? <ChartItem metricData={metricData.disk} title="Disk" /> : <LoaderItem title="Disk" />}

            {metricData?.io_traffic_read && metricData?.io_traffic_write ? (
              <ChartItem
                metricData={metricData.io_traffic_read.concat(metricData.io_traffic_write)}
                title="IO Traffic"
              />
            ) : (
              <LoaderItem title="IO Traffic" />
            )}
          </>
        ) : (
          <>
            {metricData?.network_receive_bytes && metricData?.network_transmit_bytes ? (
              <ChartItem
                metricData={metricData.network_receive_bytes.concat(metricData.network_transmit_bytes)}
                title="Network usage"
              />
            ) : (
              <LoaderItem title="Network usage" />
            )}

            {metricData?.storage_read_traffic_bytes && metricData?.storage_write_traffic_bytes ? (
              <ChartItem
                metricData={metricData.storage_read_traffic_bytes.concat(metricData.storage_write_traffic_bytes)}
                title="IO Traffic"
              />
            ) : (
              <LoaderItem title="IO Traffic" />
            )}

            {metricData?.storage_iops_read_total && metricData?.storage_iops_write_total ? (
              <ChartItem
                metricData={metricData.storage_iops_read_total.concat(metricData.storage_iops_write_total)}
                title="IOPS usage"
              />
            ) : (
              <LoaderItem title="IOPS usage" />
            )}

            {metricData?.storage_read_times_ms_total && metricData?.storage_write_times_ms_total ? (
              <ChartItem
                metricData={metricData.storage_read_times_ms_total.concat(metricData.storage_write_times_ms_total)}
                title="IO Times"
              />
            ) : (
              <LoaderItem title="IO Times" />
            )}
          </>
        )}
      </div>
    )
  }

  return (
    <>
      <div className="section">
        {!hideNotice && selectedResourceCategory === 'BareMetalHost' && (
          <CustomAlerts
            alertType="secondary"
            showAlert
            showIcon
            message={`Cloud Monitor is enabled for BMs created after ${moment('2025-02-19').format('MM/DD/YYYY')}. For earlier BMs or if you encounter any issues, please `}
            link={{ href: idcConfig.REACT_APP_SUPPORT_PAGE, label: 'contact support', openInNewTab: true }}
            className="w-100"
          />
        )}
        <div className="d-flex w-100 flex-xs-column align-items-xs-start align-items-md-end gap-s6 chart-flex">
          {instanceInput}
          {viewTypeInput}
          <Button variant="outline-primary" onClick={() => getGraphsData()} intc-id="metricsRefreshButton">
            <BsArrowClockwise intc-id="metricsRefreshButtonIcon" /> Refresh
          </Button>
        </div>
      </div>
      <div className="section">
        {state.mainTitle && (
          <h2 intc-id="metricsTitle" className="h5">
            {state.mainTitle}
          </h2>
        )}
        {getMetric()}
      </div>
    </>
  )
}

export default MetricsGraphs
