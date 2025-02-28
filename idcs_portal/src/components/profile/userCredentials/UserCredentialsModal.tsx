// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import CustomAlerts from '../../../utils/customAlerts/CustomAlerts'
import CodeLine from '../../../utils/CodeLine'
import ToastContainer from '../../../utils/toast/ToastContainer'
import Spinner from '../../../utils/spinner/Spinner'

const UserCredentialsModal = (props: any): JSX.Element => {
  const show = props.show
  const onClickShow = props.onClickShow
  const command = props.command
  const token = props.token
  const isTokenReady = props.isTokenReady

  return (
    <Modal
      show={show}
      onHide={() => onClickShow(false)}
      backdrop="static"
      size="lg"
      aria-labelledby="contained-modal-title-vcenter"
      centered
      aria-label="User credential modal"
    >
      <ToastContainer />
      <Modal.Header closeButton={isTokenReady}>
        {isTokenReady ? (
          <Modal.Title as={'h1'} className="h4">
            Client Secret
          </Modal.Title>
        ) : null}
      </Modal.Header>
      <Modal.Body>
        <div className="section">
          {isTokenReady ? (
            <>
              {' '}
              <CustomAlerts
                showAlert={true}
                showIcon={true}
                alertType="warning"
                message="Client Secret generated successfully, it will be visible only once. Make sure to store it properly."
              />
              {command}
              <CodeLine codeline={token} />
            </>
          ) : (
            <div className="section text-center align-self-center">
              <Spinner />
              <h2 className="h4">Working on your request</h2>
              <span className="align-self-center"> Submitting request</span>
            </div>
          )}
        </div>
      </Modal.Body>
      <Modal.Footer>
        {isTokenReady ? (
          <Button
            disabled={!isTokenReady}
            variant="primary"
            onClick={() => props.onClickShow(false)}
            intc-id="TokenClose"
          >
            Continue
          </Button>
        ) : null}
      </Modal.Footer>
    </Modal>
  )
}

export default UserCredentialsModal
