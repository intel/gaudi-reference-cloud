// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button } from 'react-bootstrap'

import './Stepper.scss'

interface StepperProps {
  state: any
  onRedirectTo: (value: string) => void
}

const Stepper: React.FC<StepperProps> = (props): JSX.Element => {
  const onRedirectTo = props.onRedirectTo
  const state = props.state
  const steps = state.steps
  const stepElements = []

  for (const key in steps) {
    const step = { ...steps[key] }
    stepElements.push(step)
  }

  const getLabel = (step: any): string => {
    if (step?.isDone) {
      return 'Done'
    }
    if (step?.hasError) {
      return 'Unavailable'
    }
    return step?.labelButton?.label
  }

  return (
    <div className="d-flex flex-sm-row flex-xs-column w-100">
      {stepElements.map((step: any, index: number) => {
        return (
          <div
            className="d-flex flex-sm-column flex-xs-row align-items-sm-center align-items-xs-start flex-fill position-relative w-100 gap-s4"
            key={index}
          >
            <div className="d-flex flex-sm-row flex-xs-column align-items-center justify-content-center stepContainer">
              <div
                className={`d-flex rounded-pill align-items-center justify-content-center stepProgressIndicator fw-semibold ${
                  step?.isDone || index === state.activeStep ? 'activeStatus' : 'inactiveStatus'
                }`}
              >
                {index + 1}
              </div>
              {index + 1 < stepElements.length && (
                <div className={`stepConnector ${index < state.activeStep ? 'activeStatus' : 'inactiveStatus'}`} />
              )}
            </div>
            <div className="d-flex flex-sm-column w-100 h-100 flex-xs-row justify-content-xs-between align-items-sm-center align-items-xs-start gap-s4">
              <div className="text-sm-center text-xs-start">{step?.label}</div>
              <Button
                intc-id={`btn-stepper-${step?.label}`}
                data-wap_ref={`btn-stepper-${step?.label}`}
                variant={step?.isDone || index === state.activeStep ? 'primary' : 'outline-primary'}
                aria-label={`${step?.label}`}
                className="mt-sm-auto mt-xs-0"
                size="sm"
                onClick={() => {
                  onRedirectTo(step?.labelButton?.redirectTo)
                }}
                disabled={step?.isDone || step?.hasError}
              >
                {getLabel(step)}
              </Button>
            </div>
          </div>
        )
      })}
    </div>
  )
}

export default Stepper
