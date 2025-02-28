// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import Button from 'react-bootstrap/Button'

const CannotGetInformation = (props) => {
  const [currentPath] = useState(window.location.pathname)

  const tryAgain = () => {
    if (props.clearError) {
      props.clearError()
    }
  }

  useEffect(() => {
    if (currentPath !== window.location.pathname) {
      // Clear the error flag when the Component is unmounted, ex: Changing routes
      props.clearError()
    }
  }, [window.location.pathname])

  return (
    <div className="section text-center align-items-center">
      <h1>Something went wrong</h1>
      <p>
        There was an error trying to access this information.
        <br />
        Please try again in a few minutes
      </p>
      <Button variant="link" aria-label="Try again" className="btn btn-link" onClick={tryAgain}>
        Try Again
      </Button>
    </div>
  )
}

export default CannotGetInformation
