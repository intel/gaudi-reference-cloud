// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'

import './MetricsGraphs.scss'
import BuildMetricsCustomInput from './partials/BuildMetricsCustomInput'
import idcConfig from '../../config/configurator'
import moment from 'moment'
import CustomAlerts from '../../utils/customAlerts/CustomAlerts'
import { Spinner } from 'react-bootstrap'
import MetricsGraphsContainer from '../../containers/metrics/MetricsGraphsContainer'

const InstanceGroupBuild = (props: any): JSX.Element => {
  // props
  const onChange = props.onChange
  const state = props.state
  const selectedResourceCategory = props.selectedResourceCategory
  const selectedResource = props.selectedResource
  const instancesFilteredValues = props.instancesFilteredValues

  const instanceInput = (
    <BuildMetricsCustomInput name="instanceGroups" input={state.form.instanceGroups} onChange={onChange} />
  )

  return (
    <>
      <div className="section">
        {selectedResourceCategory === 'BareMetalHost' && (
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
        </div>
      </div>
      {selectedResource && instancesFilteredValues.length === 0 ? (
        <Spinner />
      ) : selectedResource && instancesFilteredValues.length > 0 ? (
        <MetricsGraphsContainer instances={instancesFilteredValues} showInstances={true} hideNotice={true} />
      ) : null}
    </>
  )
}

export default InstanceGroupBuild
