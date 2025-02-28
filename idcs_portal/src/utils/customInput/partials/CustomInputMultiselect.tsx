// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { v4 as uuidv4 } from 'uuid'
import Form from 'react-bootstrap/Form'
import Button from 'react-bootstrap/Button'
import InputGroup from 'react-bootstrap/InputGroup'
import CustomInputLabel from './CustomInputLabel'
import { type CustomInputOnChangeEvent, getCustomInputId, type CustomInputProps } from '../CustomInput.types'
import './_CustomInputs.scss'
import { ButtonGroup } from 'react-bootstrap'

const CustomInputMultiselect: React.FC<CustomInputProps> = ({
  label = '',
  options = [],
  labelButton,
  emptyOptionsMessage,
  fieldSize,
  isReadOnly,
  hiddenLabel,
  value = [],
  onChangeDropdownMultiple,
  onBlur,
  extraButton,
  refreshButton,
  selectAllButton,
  helperMessage,
  validationMessage,
  isValid,
  isTouched,
  borderlessDropdownMultiple,
  customClass
}): JSX.Element => {
  const onChange = (
    event: React.ChangeEvent<HTMLInputElement>,
    optionOnChange?: (event: CustomInputOnChangeEvent) => void
  ): void => {
    const values = value
    const val = event.target.value
    if (event.target.checked) {
      values.push(val)
    } else {
      const index = values.indexOf(val, 0)
      if (index > -1) {
        values.splice(index, 1)
      }
    }
    if (onChangeDropdownMultiple) {
      onChangeDropdownMultiple(values)
    }

    if (optionOnChange) {
      optionOnChange(event)
    }
  }

  const labelId = getCustomInputId(label)

  return (
    <Form.Group className="d-flex-customInput">
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      {options.length > 0 && (
        <InputGroup
          size={fieldSize}
          className={`customInputMultiSelect gap-s3 ${borderlessDropdownMultiple ? 'px-s3' : 'customBorder p-s6'} ${
            isTouched && !isValid ? (borderlessDropdownMultiple ? '' : 'is-invalid') : ''
          }`}
        >
          {selectAllButton && borderlessDropdownMultiple && (
            <Form.Check
              id={`${labelId}-Input-option-selectAll-${uuidv4()}`}
              intc-id={`${labelId}-Input-option-selectAll`}
              label={selectAllButton.label}
              aria-label={selectAllButton.label}
              onChange={selectAllButton.buttonFunction}
              onBlur={onBlur}
              checked={options.length > 0 && value.length === options.length}
              className={customClass ?? ''}
            />
          )}
          {options.map((option, index) => (
            <Form.Check
              key={index}
              id={`${labelId}-Input-option-${getCustomInputId(option.name)}-${uuidv4()}`}
              intc-id={`${labelId}-Input-option-${getCustomInputId(option.name)}`}
              label={option.name}
              aria-label={value.indexOf(index) !== -1 ? `Uncheck ${option.name}` : `Check ${option.name}`}
              onChange={(e) => {
                onChange(e, option.onChanged)
              }}
              onBlur={onBlur}
              disabled={isReadOnly ?? option.disabled}
              value={option.value as string | number | undefined}
              checked={value.indexOf(option.value) !== -1}
              className={customClass ?? ''}
            />
          ))}
        </InputGroup>
      )}
      {options.length === 0 && (
        <Form.Control
          as="input"
          type="text"
          intc-id={`${labelId}inputEmptyData`}
          placeholder={emptyOptionsMessage}
          readOnly
          disabled
        />
      )}
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
      <ButtonGroup>
        {extraButton && (
          <Button intc-id={`${labelId}btnExtra`} variant="outline-primary" onClick={extraButton.buttonFunction}>
            {extraButton.label}
          </Button>
        )}
        {refreshButton && (
          <Button intc-id={`${labelId}btnRefresh`} variant="outline-primary" onClick={refreshButton.buttonFunction}>
            {refreshButton.label}
          </Button>
        )}
        {selectAllButton && !borderlessDropdownMultiple && (
          <Button intc-id={`${labelId}btnSelectAll`} variant="outline-primary" onClick={selectAllButton.buttonFunction}>
            {selectAllButton.label}
          </Button>
        )}
      </ButtonGroup>
    </Form.Group>
  )
}

export default CustomInputMultiselect
