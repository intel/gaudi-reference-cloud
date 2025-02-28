// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { getCustomInputId, type CustomInputProps } from '../CustomInput.types'
import Form from 'react-bootstrap/Form'
import Button from 'react-bootstrap/Button'
import './_CustomInputs.scss'
import CustomInputLabel from './CustomInputLabel'

const CustomInputInteger: React.FC<CustomInputProps> = ({
  label = '',
  labelButton,
  fieldSize,
  isReadOnly,
  hiddenLabel,
  autocomplete,
  value = '',
  placeholder = '',
  onChanged,
  onBlur,
  extraButton,
  helperMessage,
  maxLength,
  checkMaxValue,
  validationMessage,
  isValid,
  isTouched,
  customClass
}): JSX.Element => {
  const labelId = getCustomInputId(label)
  const maxWidth = checkMaxValue ? checkMaxValue.toString().length : maxLength
  return (
    <Form.Group className="d-flex-customInput">
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <Form.Control
        intc-id={`${labelId}Input`}
        aria-label={label}
        as="input"
        type="text"
        size={fieldSize}
        value={value}
        onChange={onChanged}
        onBlur={onBlur}
        disabled={isReadOnly}
        placeholder={placeholder}
        maxLength={maxWidth}
        style={maxWidth ? { maxWidth: `${maxWidth + 2}rem` } : {}}
        autoComplete={autocomplete}
        isValid={isValid && isTouched}
        isInvalid={!isValid && isTouched}
        onKeyPress={(event) => {
          if (!/[0-9]/.test(event.key)) {
            event.preventDefault()
          }
        }}
        className={`integerInput ${customClass ?? ''}`}
      />
      {!isValid && isTouched && (
        <Form.Control.Feedback intc-id={`${labelId}InvalidMessage`} type="invalid">
          {validationMessage}
        </Form.Control.Feedback>
      )}
      {(helperMessage !== undefined || maxLength !== undefined) && (
        <Form.Control.Feedback type="valid">
          <div className="d-flex flex-row w-100 gap-s8 justify-content-between">
            <span className="d-flex">{helperMessage}</span>
          </div>
        </Form.Control.Feedback>
      )}
      {extraButton && (
        <Button intc-id={`${labelId}btnExtra`} variant="outline-primary" onClick={extraButton.buttonFunction}>
          {extraButton.label}
        </Button>
      )}
    </Form.Group>
  )
}

export default CustomInputInteger
