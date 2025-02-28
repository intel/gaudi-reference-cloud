// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import { Link } from 'react-router-dom'

interface EmptyViewAction {
  label: string
  rightIcon?: React.ReactElement | null
  href?: any
  type: 'redirect' | 'function'
  btnType?: 'link' | 'primary' | 'outline-primary'
}

export interface EmptyViewProps {
  title: string
  subTitle: string
  action?: EmptyViewAction
}

const EmptyView: React.FC<EmptyViewProps> = ({ title, subTitle, action }): JSX.Element => {
  const getClass = (type: string | undefined): string => {
    switch (type) {
      case 'link':
        return 'btn btn-link'
      default:
        return 'btn btn-primary'
    }
  }

  return (
    <div className="section align-items-center">
      <div className="text-center" intc-id="data-view-empty">
        <span className="h4">{title}</span>
        <p className="add-break-line mt-3 lead">{subTitle}</p>
        {action ? (
          action.type === 'redirect' ? (
            <Link
              intc-id={action.label.replaceAll(' ', '') + 'EmptyViewButton'}
              data-wap_ref={action.label.replaceAll(' ', '') + 'EmptyViewButton'}
              to={action.href}
              className={getClass(action.btnType)}
            >
              {action.rightIcon ? action.rightIcon : null}
              {action.label}
            </Link>
          ) : (
            <Button
              onClick={() => action.href()}
              variant={action.btnType ?? 'outline-primary'}
              intc-id={action.label.replaceAll(' ', '') + 'EmptyViewButton'}
              data-wap_ref={action.label.replaceAll(' ', '') + 'EmptyViewButton'}
            >
              {action.rightIcon ? action.rightIcon : null}
              {action.label}
            </Button>
          )
        ) : null}
      </div>
      <br />
    </div>
  )
}

export default EmptyView
