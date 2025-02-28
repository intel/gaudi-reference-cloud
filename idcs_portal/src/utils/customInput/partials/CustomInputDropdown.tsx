// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { type CustomInputProps, type CustomInputOption, getCustomInputId } from '../CustomInput.types'
import Form from 'react-bootstrap/Form'
import Button from 'react-bootstrap/Button'
import Dropdown from 'react-bootstrap/Dropdown'
import './_CustomInputs.scss'
import CustomInputLabel from './CustomInputLabel'

const CustomInputDropdown: React.FC<CustomInputProps> = ({
  label = '',
  labelButton,
  maxWidth,
  maxInputWidth,
  helperMessage,
  options = [],
  isReadOnly,
  hiddenLabel,
  value,
  placeholder = '',
  onChanged,
  extraButton,
  validationMessage,
  isValid,
  isTouched
}): JSX.Element => {
  const [optionSelected, setOptionSelected] = useState<CustomInputOption>()

  const shouldDisable = isReadOnly === true || (value && options && options.length <= 1)

  const onOptionSelect = (option: CustomInputOption): void => {
    const e = { target: { value: option.value } }
    if (onChanged) {
      onChanged(e)
    }
  }

  useEffect(() => {
    const selectedOption = options.find((item) => item.value === value)
    setOptionSelected(selectedOption)
  }, [value])

  const labelId = getCustomInputId(label)

  return (
    <Form.Group className="d-flex-customInput" style={maxWidth ? { maxWidth } : undefined}>
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <Dropdown style={maxInputWidth ? { maxWidth: maxInputWidth } : maxWidth ? { maxWidth } : undefined}>
        <Dropdown.Toggle
          id={`${labelId}-toggle-form-select`}
          variant="simple"
          intc-id={`${labelId}-form-select`}
          role="combobox"
          aria-label={label}
          data-bs-toggle="dropdown"
          aria-expanded="false"
          aria-controls={`${labelId}-menu-form-select`}
          className={`form-dropdown-select ${isTouched && !isValid ? 'is-invalid' : ''}`}
          disabled={shouldDisable}
        >
          <CustomInputDropdownLabel optionSelected={optionSelected} placeholder={placeholder} />
        </Dropdown.Toggle>
        <Dropdown.Menu
          id={`${labelId}-menu-form-select`}
          renderOnMount
          flip
          aria-labelledby={`${labelId}-toggle-form-select`}
          className="form-dropdown-select-menu"
        >
          {options.map((option, index) => (
            <CustonInputDropdownOption
              onOptionSelect={onOptionSelect}
              labelId={labelId}
              option={option}
              optionSelected={optionSelected}
              key={index}
            />
          ))}
        </Dropdown.Menu>
      </Dropdown>
      {optionSelected?.feedbackHtml && (
        <Form.Control.Feedback style={maxWidth ? { maxWidth } : undefined}>
          {optionSelected.feedbackHtml}
        </Form.Control.Feedback>
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
      {extraButton && (
        <Button intc-id={`${labelId}btnExtra`} variant="outline-primary" onClick={extraButton.buttonFunction}>
          {extraButton.label}
        </Button>
      )}
    </Form.Group>
  )
}

interface CustomInputDropdownLabelProps {
  optionSelected?: CustomInputOption
  placeholder: string
}

const CustomInputDropdownLabel: React.FC<CustomInputDropdownLabelProps> = ({ optionSelected, placeholder }) => {
  if (!optionSelected) {
    return <div className="gap-s4 dropdownInput empty">{placeholder}</div>
  }

  return (
    <div intc-id={`selected-option ${optionSelected.name}`} className="gap-s3 dropdownInput">
      <div className="d-flex flex-xs-column flex-md-row w-100 gap-md-s8 gap-xs-s3 justify-content-between">
        <span>{optionSelected.name}</span>
        <span className="fw-lighter me-s10 text-nowrap">{optionSelected.dropSelect}</span>
      </div>
      {optionSelected.subTitleHtml}
    </div>
  )
}

interface CustomInputDropdownOptionProps {
  labelId: string
  optionSelected?: CustomInputOption
  option: CustomInputOption
  onOptionSelect: (option: CustomInputOption) => void
}

const CustonInputDropdownOption: React.FC<CustomInputDropdownOptionProps> = ({
  labelId,
  optionSelected,
  option,
  onOptionSelect
}) => {
  const isSelected = option.value === optionSelected?.value

  return (
    <Dropdown.Item
      intc-id={`${labelId}-form-select-option-${getCustomInputId(option.name)}`}
      onClick={() => {
        onOptionSelect(option)
      }}
      aria-label={`Select option ${option.name}`}
    >
      <div
        className={`dropdownOption ${option.subTitleHtml ? 'py-s4' : 'py-s3'} gap-s3 w-100 ${
          isSelected ? 'active' : ''
        }`}
      >
        <div className="d-flex flex-xs-column flex-md-row w-100 gap-md-s8 gap-xs-s3 justify-content-between">
          <span>{option.name}</span>
          <span className="fw-lighter me-s10 text-nowrap">{option.dropSelect}</span>
        </div>
        {option.subTitleHtml}
      </div>
    </Dropdown.Item>
  )
}

export default CustomInputDropdown
