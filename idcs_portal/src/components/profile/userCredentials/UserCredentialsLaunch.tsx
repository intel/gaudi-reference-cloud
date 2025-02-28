// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import CustomInput from '../../../utils/customInput/CustomInput'
import { ButtonGroup } from 'react-bootstrap'
import Button from 'react-bootstrap/Button'
import UserCredentialsModal from './UserCredentialsModal'

const UserCredentialsLaunch = (props: any): JSX.Element => {
  // ****
  // Varaibles
  // ****
  const title = props.title
  const form = props.form
  const onChangeInput = props.onChangeInput
  const navigationBottom = props.navigationBottom
  const onSubmit = props.onSubmit
  const accessTokenModal = props.accessTokenModal
  const onCloseModal = props.onCloseModal
  const formElements = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    formElements.push({
      id: key,
      configInput: formItem
    })
  }

  const buildCustomInput = (element: any, key: number): JSX.Element => {
    return (
      <CustomInput
        key={key}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired
            ? String(element.configInput.label) + ' *'
            : element.configInput.label
        }
        value={element.configInput.value}
        onChanged={(event) => onChangeInput(event, element.id)}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        refreshButton={element.configInput.refreshButton}
        extraButton={element.configInput.extraButton}
        selectAllButton={element.configInput.selectAllButton}
        labelButton={element.configInput.labelButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
        hidden={element.configInput.hidden}
      />
    )
  }

  return (
    <>
      <UserCredentialsModal
        show={accessTokenModal.show}
        onClickShow={onCloseModal}
        isTokenReady={accessTokenModal.isTokenReady}
        token={accessTokenModal.token}
        command={accessTokenModal.command}
      />
      <div className="section">
        <h2 intc-id="userCredentialsTitle">{title}</h2>
      </div>
      <div className="section">
        {formElements.map((element, index) => buildCustomInput(element, index))}
        <ButtonGroup>
          {navigationBottom.map((item: any, index: string) => (
            <Button
              intc-id={`btn-user-credential-navigationBottom ${item.buttonLabel}`}
              data-wap_ref={`btn-user-credential-navigationBottom ${item.buttonLabel}`}
              aria-label={item.buttonLabel}
              key={index}
              variant={item.buttonVariant}
              onClick={item.buttonAction === 'Submit' ? onSubmit : item.buttonFunction}
            >
              {item.buttonLabel}
            </Button>
          ))}
        </ButtonGroup>
      </div>
    </>
  )
}

export default UserCredentialsLaunch
