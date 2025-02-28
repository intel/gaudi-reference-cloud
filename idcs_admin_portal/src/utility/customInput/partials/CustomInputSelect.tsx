// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useMemo, useState } from 'react'
import { getCustomInputId, type CustomInputOption, type CustomInputProps } from '../CustomInput.types'
import Form from 'react-bootstrap/Form'
import Button from 'react-bootstrap/Button'
import './_CustomInputs.scss'
import CustomInputLabel from './CustomInputLabel'

interface CustomInputSelectOption {
  name: string
  value: string | number
}

const CustomInputSelect: React.FC<CustomInputProps> = ({
  label = '',
  labelButton,
  fieldSize,
  options = [],
  isReadOnly,
  hiddenLabel,
  value,
  autocomplete,
  placeholder = '',
  onChanged,
  extraButton
}): JSX.Element => {
  const [optionSelected, setOptionSelected] = useState<CustomInputOption>()
  const shouldDisable = isReadOnly === true || (value && options && options.length <= 1)

  const labelId = getCustomInputId(label)

  const availableOptions: CustomInputSelectOption[] = useMemo(() => {
    const newOptions = options.map((x) => ({
      name: x.name,
      value: x.value as string | number
    }))
    return newOptions
  }, [options])

  useEffect(() => {
    const selectedOption = options.find((item) => item.value === value)
    setOptionSelected(selectedOption)
  }, [value])

  return (
    <Form.Group className="d-flex-customInput">
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <Form.Select
        intc-id={`${labelId}Select`}
        size={fieldSize}
        aria-label={label}
        autoComplete={autocomplete}
        value={value || placeholder}
        onChange={(e) => {
          if (onChanged) {
            onChanged(e)
          }
        }}
        className="form-dropdown-select"
        disabled={shouldDisable}
      >
        <option value={placeholder} disabled intc-id={`${labelId}OptionPlaceholder`}>
          {placeholder}
        </option>
        {availableOptions.map((option, index) => (
          <option
            key={index}
            value={option.value}
            className="dropdownOption"
            intc-id={`${getCustomInputId(option.name)}Option`}
          >
            {option.name}
          </option>
        ))}
      </Form.Select>
      {optionSelected?.feedbackHtml && <Form.Control.Feedback>{optionSelected.feedbackHtml}</Form.Control.Feedback>}
      {extraButton && (
        <Button intc-id={`${labelId}btnExtra`} variant="outline-primary" onClick={extraButton.buttonFunction}>
          {extraButton.label}
        </Button>
      )}
    </Form.Group>
  )
}

export default CustomInputSelect
