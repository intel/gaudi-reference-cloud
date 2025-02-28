// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useRef, useState } from 'react'
import Button from 'react-bootstrap/Button'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import idcConfig from '../../config/configurator'

const ConsoleAccessDenied = () => {
  const [timerCounter, setTimerCounter] = useState(30)
  const interval = useRef(null)

  useEffect(() => {
    interval.current = setInterval(() => {
      setTimerCounter((timerCounter) => timerCounter - 1)
    }, 1000)

    return () => clearInterval(interval.current)
  }, [])

  const redirectAfterTimeOut = async () => {
    window.location.href = idcConfig.REACT_APP_GUI_DOMAIN
  }

  useEffect(() => {
    if (timerCounter === 0) {
      clearInterval(interval.current)
      redirectAfterTimeOut()
    }
  }, [timerCounter])

  return (
    <div className="section text-center align-items-center">
      <h1>Access Restricted</h1>
      <p>
        We appreciate your interest. If you believe you require access to this site, please contact our support team.
        <br />
        You will be redirected to {idcConfig.REACT_APP_GUI_DOMAIN} in {timerCounter}
      </p>

      <ButtonGroup>
        <Button aria-label="Go to IDC Console" variant="link" onClick={redirectAfterTimeOut}>
          Go to {idcConfig.REACT_APP_GUI_DOMAIN}
        </Button>
      </ButtonGroup>
    </div>
  )
}

export default ConsoleAccessDenied
