// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Spinner from '../../../utils/spinner/Spinner'
import './ChartItem.scss'

const LoaderItem = (props: any): JSX.Element => {
  const title = props.title
  const testId = 'metrics' + String(title.replaceAll(' ', '')) + 'Title'
  const minWidth = '20rem'

  return (
    <div className="col-xxl-4 col-lg-6 col-xs-12">
      <div className="d-flex flex-column rounded p-s6 gap-s6 h-100 chartItem" style={{ minWidth }}>
        <h3 intc-id={testId} className="h6">
          {title}
        </h3>
        <Spinner />
      </div>
    </div>
  )
}

export default LoaderItem
