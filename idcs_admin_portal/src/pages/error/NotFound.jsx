// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Link } from 'react-router-dom'

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
          Intel Tiber Developer Admin Console
        </button>
      </Link>
    </div>
  )
}

export default NotFound
