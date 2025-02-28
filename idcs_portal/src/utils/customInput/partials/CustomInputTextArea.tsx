// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { getCustomInputId, type CustomInputProps } from '../CustomInput.types'
import Form from 'react-bootstrap/Form'
import Button from 'react-bootstrap/Button'
import './_CustomInputs.scss'
import CustomInputLabel from './CustomInputLabel'

const CustomInputTextArea: React.FC<CustomInputProps> = ({
  intcId,
  label = '',
  labelButton,
  fieldSize,
  isReadOnly,
  hiddenLabel,
  value = '',
  placeholder = '',
  onChanged,
  onBlur,
  extraButton,
  helperMessage,
  maxLength,
  textAreaRows,
  validationMessage,
  isValid,
  isTouched,
  customClass
}): JSX.Element => {
  const labelId = getCustomInputId(label)
  return (
    <Form.Group className="d-flex-customInput">
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <Form.Control
        intc-id={intcId ?? `${labelId}TextArea`}
        aria-label={label}
        as="textarea"
        size={fieldSize}
        rows={textAreaRows ?? 3}
        value={value}
        onChange={onChanged}
        onBlur={onBlur}
        readOnly={isReadOnly}
        placeholder={placeholder}
        maxLength={maxLength}
        isValid={isValid && isTouched}
        isInvalid={!isValid && isTouched}
        className={customClass ?? ''}
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
            {maxLength && (
              <>
                {`${value.length}`}/{maxLength}
              </>
            )}
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

export default CustomInputTextArea
