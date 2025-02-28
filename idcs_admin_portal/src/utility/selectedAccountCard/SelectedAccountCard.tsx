// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Card, Button } from 'react-bootstrap'
import { BsCopy } from 'react-icons/bs'
import { useCopy } from '../../hooks/useCopy'

interface SelectedAccountCardProps {
  selectedCloudAccount: any
}

const SelectedAccountCard: React.FC<SelectedAccountCardProps> = (props): JSX.Element => {
  const { copyToClipboard } = useCopy()
  const selectedCloudAccount = props.selectedCloudAccount

  if (!selectedCloudAccount) return <></>

  return (
    <Card>
      <Card.Header style={{ display: 'inline-flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>Selected Cloud Account</div>
      </Card.Header>
      <Card.Body>
        <Card.Title intc-id="selectedCloudAccountName" className="d-flex flex-row w-100 align-items-center">
          <span className="col-10">{selectedCloudAccount.name}</span>{' '}
          <Button
            variant="icon-simple"
            size="sm"
            intc-id="btn-copy-account-email"
            data-wap_ref="btn-copy-account-email"
            onClick={(event) => {
              copyToClipboard(selectedCloudAccount.name)
            }}
          >
            <BsCopy className="text-primary" />
          </Button>
        </Card.Title>
        <Card.Subtitle intc-id="selectedCloudAccountId" className="d-flex flex-row w-100 align-items-center">
          <span className="col-10">{selectedCloudAccount.id}</span>
          <Button
            variant="icon-simple"
            size="sm"
            intc-id="btn-copy-account-id"
            data-wap_ref="btn-copy-account-id"
            onClick={(event) => {
              copyToClipboard(selectedCloudAccount.id)
            }}
          >
            <BsCopy className="text-primary" />
          </Button>
        </Card.Subtitle>
        <Card.Text intc-id="selectedCloudAccountType">{selectedCloudAccount.type}</Card.Text>
      </Card.Body>
    </Card>
  )
}

export default SelectedAccountCard
