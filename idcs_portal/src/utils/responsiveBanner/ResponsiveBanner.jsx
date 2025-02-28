// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import idcConfig from '../../config/configurator'

const ResponsiveBanner = () => {
  return (
    <div className="d-flex flex-row">
      <div className="d-flex flex-column bd-highlight p-3 ms-4 me-4 section-component">
        <div className="d-flex justify-content-center my-3">
          <span className="h3">Sorry,</span>
        </div>
        <div className="d-flex justify-content-center">
          <p className="small">
            {`${idcConfig.REACT_APP_CONSOLE_LONG_NAME} is not supported on mobile devices. Please access it from a desktop or laptop computer for the best experience.`}
          </p>
        </div>
      </div>
    </div>
  )
}

export default ResponsiveBanner
