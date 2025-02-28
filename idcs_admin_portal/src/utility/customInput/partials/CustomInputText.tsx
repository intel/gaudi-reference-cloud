// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Form from 'react-bootstrap/Form'
import InputGroup from 'react-bootstrap/InputGroup'
import Button from 'react-bootstrap/Button'
import { getCustomInputId, type CustomInputProps } from '../CustomInput.types'
import CustomInputLabel from './CustomInputLabel'
import './_CustomInputs.scss'

const CustomInputText: React.FC<CustomInputProps> = ({
  intcId,
  label = '',
  type = 'text',
  labelButton,
  maxWidth,
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
  prepend,
  maxLength,
  minLength,
  validationMessage,
  isValid,
  isTouched,
  customClass
}): JSX.Element => {
  const labelId = getCustomInputId(label)
  return (
    <Form.Group className="d-flex-customInput" style={maxWidth ? { maxWidth } : undefined}>
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <InputGroup size={fieldSize} style={maxWidth ? { maxWidth } : undefined}>
        {prepend && <InputGroup.Text>{prepend}</InputGroup.Text>}
        <Form.Control
          intc-id={intcId ?? `${labelId}Input`}
          aria-label={label}
          as="input"
          type={type}
          size={fieldSize}
          value={value}
          onChange={onChanged}
          onBlur={onBlur}
          disabled={isReadOnly}
          placeholder={placeholder}
          maxLength={maxLength}
          minLength={minLength}
          autoComplete={autocomplete}
          isValid={isValid && isTouched}
          isInvalid={!isValid && isTouched}
          className={`${customClass ?? ''}`}
        />
      </InputGroup>
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

export default CustomInputText
