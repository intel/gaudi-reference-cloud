// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import CustomAlerts from '../../customAlerts/CustomAlerts'
import CustomInput from '../../customInput/CustomInput'

interface ExtendReservationModalProps {
  showModal: boolean
  state: any
  onChange: (event: any) => void
  onSubmit: (event: any, action: string) => void
  onHide: () => void
}

interface NavigationBottomItem {
  buttonLabel: string
  buttonVariant: string
  buttonFunction: (event: any, item: any) => void
}

const ExtendReservationModal: React.FC<ExtendReservationModalProps> = (props): JSX.Element => {
  const showModal = props.showModal
  const state = props.state
  const onChange = props.onChange
  const onSubmit = props.onSubmit
  const onHide = props.onHide
  const mainTitle = state.mainTitle
  const warningMessage = state.warningMessage
  const bodyMessage = state.bodyMessage
  const formElement = state.form.days
  const navigationBottom = state.navigationBottom

  const modalBody = (
    <>
      <p>
        {bodyMessage.split('\n').map((line: string, index: number) => (
          <React.Fragment key={index}>
            {line}
            <br />
          </React.Fragment>
        ))}
      </p>
      <CustomInput
        key={formElement.label}
        type={formElement.type}
        maxLength={formElement.maxLength}
        fieldSize={formElement.fieldSize}
        placeholder={formElement.placeholder}
        validationMessage={formElement.validationMessage}
        helperMessage={formElement.helperMessage}
        label={formElement.validationRules.isRequired ? String(formElement.label) + ' *' : formElement.label}
        value={formElement.value}
        isValid={formElement.isValid}
        isTouched={formElement.isTouched}
        customClass=""
        onChanged={(event: any) => {
          onChange(event)
        }}
      />
    </>
  )

  const modalFooter = (
    <ButtonGroup>
      {navigationBottom.map((item: NavigationBottomItem, index: number) => (
        <Button
          aria-label={`Click ${item.buttonLabel}`}
          intc-id={`btn-reservation-extension-${item.buttonLabel}`}
          key={index}
          disabled={item.buttonLabel === 'Request' ? !state.isValidForm : false}
          variant={item.buttonVariant}
          onClick={(event) => {
            onSubmit(event, item.buttonLabel)
          }}
        >
          {item.buttonLabel}
        </Button>
      ))}
    </ButtonGroup>
  )

  return (
    <Modal
      show={showModal}
      onHide={() => {
        onHide()
      }}
      backdrop="static"
      keyboard={false}
      centered
      aria-label="Extend reservation modal"
    >
      <Modal.Header closeButton closeLabel="Close request extention modal">
        <Modal.Title>{mainTitle}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <>
          {warningMessage && (
            <CustomAlerts message={warningMessage} showIcon={true} alertType="warning" showAlert={true}></CustomAlerts>
          )}
          <div className="section">{modalBody}</div>
        </>
      </Modal.Body>
      <Modal.Footer>{modalFooter}</Modal.Footer>
    </Modal>
  )
}

export default ExtendReservationModal
