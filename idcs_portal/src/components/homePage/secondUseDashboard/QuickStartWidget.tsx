// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Card, ListGroup } from 'react-bootstrap'
import { type InitialState } from '../../../containers/homePage/secondUseDashboard/QuickStartWidgetContainer'

interface QuickStartWidgetProps {
  state: InitialState
  navigateTo: (url: string) => void
}

const QuickStartWidget: React.FC<QuickStartWidgetProps> = (props): JSX.Element => {
  const state = props.state
  const actions = state.actions
  const navigateTo = props.navigateTo

  return (
    <Card>
      <Card.Body>
        <Card.Title>
          <h2 className="h6">Quick Start</h2>
        </Card.Title>
        <small>Most common actions</small>
        <ListGroup variant="flush" className="w-100">
          {actions.map((action, index) =>
            !action.disabled ? (
              <ListGroup.Item
                key={index}
                action
                intc-id={`link-dashboard-quick-start-${action.label}`}
                data-wap_ref={`link-dashboard-quick-start-${action.label}`}
                aria-label={`${action.label} link`}
                onClick={() => {
                  navigateTo(action.redirectTo)
                }}
              >
                {action.icon}
                {action.label}
              </ListGroup.Item>
            ) : null
          )}
        </ListGroup>
      </Card.Body>
    </Card>
  )
}

export default QuickStartWidget
