// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Link } from 'react-router-dom'
import idcConfig from '../../config/configurator'

const NotFound = () => {
  return (
    <div className="section text-center align-items-center">
      <h1>Page not found</h1>
      <p>
        The page you are trying to access does not exist.
        <br />
        You can go to any of the following links:
      </p>
      <Link to="/">
        <button type="button" className="btn btn-link text-decoration-underline">
          {`${idcConfig.REACT_APP_CONSOLE_SHORT_NAME} Console`}
        </button>
      </Link>
    </div>
  )
}

export default NotFound
