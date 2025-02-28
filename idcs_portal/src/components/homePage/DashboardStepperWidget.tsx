// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import Card from 'react-bootstrap/Card'
import Stepper from '../../utils/stepper/Stepper'
import Spinner from '../../utils/spinner/Spinner'

interface DashboardStepperWidgetProps {
  state: any
  loading: boolean
  onRedirectTo: (value: string) => void
}

const DashboardStepperWidget: React.FC<DashboardStepperWidgetProps> = (props): JSX.Element => {
  const state = props.state
  const loading = props.loading
  const onRedirectTo = props.onRedirectTo

  const spinner = <Spinner />

  return (
    <Card>
      <Card.Body>
        <Card.Title as="h2" className="h6">
          {state.title}
        </Card.Title>
        <div className="p-sm-s8 px-xs-s6 py-xs-s8 w-100">
          {loading ? spinner : <Stepper state={state} onRedirectTo={onRedirectTo} />}
        </div>
      </Card.Body>
      <Button
        variant="close"
        aria-label="Close"
        className="position-absolute top-0 end-0 m-s6"
        onClick={() => {}}
      ></Button>
    </Card>
  )
}

export default DashboardStepperWidget
