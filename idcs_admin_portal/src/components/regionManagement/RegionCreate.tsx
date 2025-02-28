// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import CustomInput from '../../utility/customInput/CustomInput'
import Button from 'react-bootstrap/Button'
import { ButtonGroup } from 'react-bootstrap'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'
import EmptyView from '../../utility/emptyView/EmptyView'

const RegionCreate = (props: any): JSX.Element => {
  const state = props.state
  const desciption = state.desciption
  const form = state.form
  const navigationTop = state.navigationTop
  const navigationBottom = state.navigationBottom
  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  const showModal = props.showModal
  const emptyView = props.emptyView

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

  let content = (
    <>
      {formElements.map((element, index) => (
        <CustomInput
          key={index}
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
          emptyOptionsMessage={element.configInput.emptyOptionsMessage}
          hiddenLabel={element.configInput.hiddenLabel}
        />
      ))}

      <ButtonGroup>
        {navigationBottom.map((item: any, index: number) => (
          <Button
            key={index}
            intc-id={`navigationBottom${item.label}`}
            variant={item.buttonVariant}
            onClick={item.buttonAction === 'Submit' ? onSubmit : item.function}
          >
            {item.label}
          </Button>
        ))}
      </ButtonGroup>
    </>
  )

  if (emptyView?.show) {
    content = <EmptyView title={emptyView.title} action={emptyView.action} subTitle={emptyView.label}/>
  }

  return (
    <>
      <div className="section">
        {navigationTop.map((item: any, index: number) => (
          <div className="m-lg-0" key={index}>
            {
              <Button
                intc-id={`navigationTop${item.label}`}
                variant={item.buttonVariant}
                className="p-s0"
                onClick={item.function}
              >
                {item.label}
              </Button>
            }
          </div>
        ))}
      </div>
      <OnSubmitModal showModal={showModal} message="Working on your request" />

      <div className="section">
        <div className="d-flex flex-row flex-wrap justify-content-between align-items-center w-100">
          <h2 className="h4">{desciption}</h2>
        </div>
      </div>
      <div className="section">{content}</div>
    </>
  )
}

export default RegionCreate
