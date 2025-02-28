// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import HowToConnectInstance from './partials/HowToConnectInstance'
import HowToConnectStorage from './partials/HowToConnectStorage'
import HowToConnectVastStorage from './partials/HowToConnectVastStorage'
import ToastContainer from '../../toast/ToastContainer'
import { appFeatureFlags, isFeatureFlagEnable } from '../../../config/configurator'

const HowToConnect = (props) => {
  const flowType = props?.flowType ? props?.flowType : 'other'
  const data = props.data || null
  if (!data) {
    return null
  }

  const getTitle = () => {
    switch (flowType) {
      case 'compute':
        return 'How to connect to your instance'
      case 'mount':
        return 'How to mount storage volume'
      case 'unmount':
        return 'How to unmount storage volume'
      default:
        return ''
    }
  }

  return (
    <Modal
      show={props.showHowToConnectModal}
      onHide={() => props.onClickHowToConnect(false)}
      backdrop="static"
      size="xl"
      aria-labelledby="contained-modal-title-vcenter"
      centered
      aria-label="How to connect modal"
    >
      <ToastContainer />
      <Modal.Header closeButton>
        <Modal.Title>
          <h1 className="h4">{getTitle()}</h1>
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {flowType === 'compute' ? <HowToConnectInstance data={data} /> : null}
        {flowType === 'mount' ? (
          isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_STORAGE_VAST) ? (
            <HowToConnectVastStorage data={data} mount closeModal={() => props.onClickHowToConnect(false)} />
          ) : (
            <HowToConnectStorage data={data} mount closeModal={() => props.onClickHowToConnect(false)} />
          )
        ) : null}
        {flowType === 'unmount' ? (
          <HowToConnectStorage data={data} mount={false} closeModal={() => props.onClickHowToConnect(false)} />
        ) : null}
      </Modal.Body>
      <Modal.Footer>
        <Button variant="outline-primary" onClick={() => props.onClickHowToConnect(false)} intc-id="HowToConnectClose">
          Close
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default HowToConnect
