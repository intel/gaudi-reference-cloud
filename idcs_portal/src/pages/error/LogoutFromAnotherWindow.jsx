// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Button from 'react-bootstrap/Button'

const LogoutFromAnotherWindow = () => {
  const refreshPage = () => {
    window.location.href = window.location
  }

  return (
    <div className="section text-center align-items-center">
      <h1>Please sign in again</h1>
      <p>
        You were signed out of your account.
        <br />
        Please press reload to sign into the Console again.
      </p>
      <Button variant="link" aria-label="Reload page" onClick={refreshPage}>
        Reload
      </Button>
    </div>
  )
}

export default LogoutFromAnotherWindow
