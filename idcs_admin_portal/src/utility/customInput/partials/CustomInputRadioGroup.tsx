// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { v4 as uuidv4 } from 'uuid'
import Form from 'react-bootstrap/Form'
import InputGroup from 'react-bootstrap/InputGroup'
import Button from 'react-bootstrap/Button'
import { getCustomInputId, type CustomInputProps } from '../CustomInput.types'
import CustomInputLabel from './CustomInputLabel'
import './_CustomInputs.scss'

const CustomInputRadioGroup: React.FC<CustomInputProps> = ({
  label = '',
  labelButton,
  customClass,
  options = [],
  fieldSize,
  isReadOnly,
  hiddenLabel,
  value = '',
  onChanged,
  onBlur,
  extraButton,
  helperMessage,
  validationMessage,
  radioGroupHorizontal,
  isValid,
  isTouched
}): JSX.Element => {
  if (options.length === 0) {
    return <></>
  }

  const labelId = getCustomInputId(label)

  // const guidForOptions =

  return (
    <Form.Group className="d-flex-customInput">
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <InputGroup
        size={fieldSize}
        className={`customInputMultiSelect px-s3 gap-s4 ${radioGroupHorizontal ? 'flex-row gap-s6 text-nowrap' : ''}`}
      >
        {options.map((option, index) => (
          <Form.Check
            key={index}
            id={`${labelId}-Radio-option-${getCustomInputId(option.name)}-${uuidv4()}`}
            intc-id={`${labelId}-Radio-option-${getCustomInputId(option.name)}`}
            type="radio"
            value={option.value as string | readonly string[] | number | undefined}
            label={option.name}
            aria-label={option.name}
            onChange={option.onChanged ?? onChanged}
            onBlur={onBlur}
            disabled={isReadOnly ?? option.disabled}
            checked={value === option.value}
            className={customClass ?? ''}
          />
        ))}
      </InputGroup>
      {!isValid && isTouched && (
        <Form.Control.Feedback intc-id={`${labelId}InvalidMessage`} type="invalid">
          {validationMessage}
        </Form.Control.Feedback>
      )}
      {helperMessage !== undefined && (
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

export default CustomInputRadioGroup
