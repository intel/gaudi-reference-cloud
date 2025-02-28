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

const CustomInputCheckbox: React.FC<CustomInputProps> = ({
  label = '',
  labelButton,
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
  isValid,
  isTouched
}): JSX.Element => {
  const labelId = getCustomInputId(label)
  if (options.length === 0) {
    return <></>
  }

  return (
    <Form.Group className="d-flex-customInput">
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <InputGroup
        size={fieldSize}
        className={`customInputMultiSelect px-s3 ${isTouched && !isValid ? 'is-invalid' : ''}`}
      >
        <Form.Check
          id={`${labelId}Checkbox-${uuidv4()}`}
          intc-id={`${labelId}Checkbox`}
          label={options[0].name}
          aria-label={label}
          onChange={onChanged}
          onBlur={onBlur}
          disabled={isReadOnly ?? options[0].disabled}
          value={options[0].value as string | number | undefined}
          checked={value}
        />
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

export default CustomInputCheckbox
