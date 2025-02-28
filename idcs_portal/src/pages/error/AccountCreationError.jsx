// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useState, useEffect, useRef } from 'react'
import Button from 'react-bootstrap/Button'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import useLogout from '../../hooks/useLogout'
import idcConfig from '../../config/configurator'

const AccountCreationError = () => {
  const [timerCounter, setTimerCounter] = useState(20)
  const interval = useRef(null)

  const { logoutHandler } = useLogout()

  useEffect(() => {
    interval.current = setInterval(() => {
      setTimerCounter((timerCounter) => timerCounter - 1)
    }, 1000)

    return () => clearInterval(interval.current)
  }, [])

  const logOutAfterTimeout = async () => {
    await logoutHandler()
  }

  useEffect(() => {
    if (timerCounter === 0) {
      clearInterval(interval.current)
      logOutAfterTimeout()
    }
  }, [timerCounter])

  return (
    <div className="section text-center align-items-center">
      <h1>Account creation error</h1>
      <p>
        There was an error creating your account. You will be redirected to the sign in page in {timerCounter} seconds.
        <br />
        Please sign in again to retry or contact our support team for assistance. Thank you for understanding.
      </p>
      <ButtonGroup>
        <Button variant="link" onClick={logoutHandler} aria-label="Sign in again">
          Sign in Again
        </Button>
        <a
          href={idcConfig.REACT_APP_SUPPORT_PAGE}
          target="_blank"
          rel="noreferrer"
          className="btn btn-link"
          aria-label="Contact Support"
        >
          Contact Support
        </a>
      </ButtonGroup>
    </div>
  )
}

export default AccountCreationError
