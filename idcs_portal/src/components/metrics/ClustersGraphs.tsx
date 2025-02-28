// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import { BsArrowClockwise } from 'react-icons/bs'

import './MetricsGraphs.scss'
import LoaderItem from './partials/LoaderItem'
import ChartItem from './partials/ChartItem'
import BuildMetricsCustomInput from './partials/BuildMetricsCustomInput'
import { Accordion } from 'react-bootstrap'
import { useMediaQuery } from 'react-responsive'

const ClustersGraphs = (props: any): JSX.Element => {
  // props
  const onChange = props.onChange
  const state = props.state
  const metricData = props.metricData
  const getGraphsData = props.getGraphsData
  const showDropdown = props.showDropdown

  const isMDScreen = useMediaQuery({
    query: '(max-width: 991px)'
  })

  const customInput = showDropdown ? (
    <BuildMetricsCustomInput name="clusters" input={state.form.clusters} onChange={onChange} />
  ) : (
    ''
  )
  const viewTypeInput = <BuildMetricsCustomInput name="viewType" input={state.form.viewType} onChange={onChange} />

  const getMetric = (): JSX.Element => {
    return (
      <Accordion intc-id={'ClusterMetrics-accordion-'} className="border-0 w-100" defaultActiveKey="apiserver">
        <Accordion.Item key="apiserver" eventKey="apiserver" className="w-100 border-top">
          <Accordion.Header as="div" intc-id={'ClusterMetricsTitle-apiserver'}>
            <h3>Kubernetes API Server cloud monitor</h3>
          </Accordion.Header>
          <Accordion.Body className="d-flex flex-column gap-s6 py-s5 px-s0 px-lg-s5">
            <div className={isMDScreen ? 'd-flex flex-column gap-s8' : 'row g-s8'}>
              {metricData?.apiserver_cpu ? (
                <ChartItem metricData={metricData.apiserver_cpu} title="CPU" />
              ) : (
                <LoaderItem title="CPU" />
              )}

              {metricData?.apiserver_memory ? (
                <ChartItem metricData={metricData.apiserver_memory} title="Memory" />
              ) : (
                <LoaderItem title="Memory" />
              )}

              {metricData?.apiserver_requestsbycode ? (
                <ChartItem metricData={metricData.apiserver_requestsbycode} title="Requests by code" />
              ) : (
                <LoaderItem title="Requests by code" />
              )}

              {metricData?.apiserver_requestsbyverb ? (
                <ChartItem metricData={metricData.apiserver_requestsbyverb} title="Requests by verb" />
              ) : (
                <LoaderItem title="Requests by verb" />
              )}

              {metricData?.apiserver_latencybyhostname ? (
                <ChartItem metricData={metricData.apiserver_latencybyhostname} title="Latency by node name" />
              ) : (
                <LoaderItem title="Latency by node name" />
              )}

              {metricData?.apiserver_latencybyverb ? (
                <ChartItem metricData={metricData.apiserver_latencybyverb} title="Latency by verb" />
              ) : (
                <LoaderItem title="Latency by verb" />
              )}

              {metricData?.apiserver_errorsbyhostname ? (
                <ChartItem metricData={metricData.apiserver_errorsbyhostname} title="Errors by node name" />
              ) : (
                <LoaderItem title="Errors by node name" />
              )}

              {metricData?.apiserver_errorsbyverb ? (
                <ChartItem metricData={metricData.apiserver_errorsbyverb} title="Errors by verb" />
              ) : (
                <LoaderItem title="Errors by verb" />
              )}

              {metricData?.apiserver_httprequestsbyhostname ? (
                <ChartItem
                  metricData={metricData.apiserver_httprequestsbyhostname}
                  title="HTTP Requests by node name"
                />
              ) : (
                <LoaderItem title="HTTP Requests by node name" />
              )}
            </div>
          </Accordion.Body>
        </Accordion.Item>
        <Accordion.Item key="etcd" eventKey="etcd" className="w-100 border-top">
          <Accordion.Header as="div" intc-id={'ClusterMetricsTitle-etcd'}>
            <h3>Kubernetes etcd cloud monitor</h3>
          </Accordion.Header>
          <Accordion.Body className="d-flex flex-column gap-s6 py-s5 px-s0 px-lg-s5">
            <div className={isMDScreen ? 'd-flex flex-column gap-s8' : 'row g-s8'}>
              {metricData?.etcd_cpu ? (
                <ChartItem metricData={metricData.etcd_cpu} title="CPU" />
              ) : (
                <LoaderItem title="CPU" />
              )}

              {metricData?.etcd_memory ? (
                <ChartItem metricData={metricData.etcd_memory} title="Memory" />
              ) : (
                <LoaderItem title="Memory" />
              )}

              {metricData?.etcd_clienttrafficin ? (
                <ChartItem metricData={metricData.etcd_clienttrafficin} title="Client traffic in" />
              ) : (
                <LoaderItem title="Client traffic in" />
              )}

              {metricData?.etcd_clienttrafficout ? (
                <ChartItem metricData={metricData.etcd_clienttrafficout} title="Client traffic out" />
              ) : (
                <LoaderItem title="Client traffic out" />
              )}

              {metricData?.etcd_peertrafficin ? (
                <ChartItem metricData={metricData.etcd_peertrafficin} title="Peer traffic in" />
              ) : (
                <LoaderItem title="Peer traffic in" />
              )}

              {metricData?.etcd_peertrafficout ? (
                <ChartItem metricData={metricData.etcd_peertrafficout} title="Peer traffic out" />
              ) : (
                <LoaderItem title="Peer traffic out" />
              )}

              {metricData?.etcd_dbsizeinuse ? (
                <ChartItem metricData={metricData.etcd_dbsizeinuse} title="DB Size in use" />
              ) : (
                <LoaderItem title="DB Size in use" />
              )}

              {metricData?.etcd_heartbeatsendfailurestotal ? (
                <ChartItem
                  metricData={metricData.etcd_heartbeatsendfailurestotal}
                  title="Total heartbeats send failures"
                />
              ) : (
                <LoaderItem title="Total heartbeats send failures" />
              )}
            </div>
          </Accordion.Body>
        </Accordion.Item>
      </Accordion>
    )
  }

  return (
    <>
      <div className="section">
        <div className="d-flex w-100 flex-xs-column align-items-xs-start align-items-md-end gap-s6 chart-flex">
          {customInput}
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

export default ClustersGraphs
