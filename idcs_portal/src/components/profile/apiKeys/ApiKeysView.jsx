// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import Button from 'react-bootstrap/Button'
import CodeLine from '../../../utils/CodeLine'

const ApiKeysView = ({ token, expirationDate, refreshKey, copyKey }) => {
  return (
    <>
      <div className="section">
        <h2>Client Secrets</h2>
      </div>
      <div className="filter align-items-center">
        <Button variant="primary" intc-id="api-key-refresh-key" aria-label="Refresh Key" onClick={refreshKey}>
          Refresh Secret
        </Button>
        <div className="d-flex flex-row gap-s4">
          <span className="h4">Secret expiration date:</span>
          <span intc-id="api-key-expiration-date" className="h4 fw-normal">
            {expirationDate}
          </span>
        </div>
      </div>
      <div className="section">
        <CodeLine codeline={token} />
      </div>
    </>
  )
}

export default ApiKeysView
