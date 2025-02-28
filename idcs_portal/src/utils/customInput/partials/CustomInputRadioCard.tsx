// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import './_CustomInputs.scss'
import { Card, Form, Button } from 'react-bootstrap'
import { type CustomInputOption, type CustomInputProps, getCustomInputId } from '../CustomInput.types'
import CustomInputLabel from './CustomInputLabel'

interface CustomInputRadioCardBodyProps {
  index: any
  value: any
  option: CustomInputOption
  handleSelect: (option: CustomInputOption) => void
  subtitleClassName?: string
}

const CustomInputRadioCardBody: React.FC<CustomInputRadioCardBodyProps> = ({
  index,
  value,
  option,
  handleSelect,
  subtitleClassName
}): JSX.Element => {
  return (
    <Card.Body className="radio-card-body">
      <div className="d-flex gap-s6 w-100">
        <div className="d-flex flex-column justify-content-center">
          <Form.Check
            key={index}
            type="radio"
            disabled={option.disabled}
            checked={value === option.value}
            onChange={() => {
              handleSelect(option)
            }}
            id={`${option.value}-radio-select`}
            intc-id={`${option.value}-radio-select`}
            aria-label={`Select ${option.value} products`}
          />
        </div>
        <div className={`d-flex flex-column ${subtitleClassName ?? ''}`}>{option.subTitleHtml}</div>
      </div>
    </Card.Body>
  )
}

const CustomInputRadioCard: React.FC<CustomInputProps> = ({
  label = '',
  labelButton,
  options = [],
  hiddenLabel,
  value,
  onChanged,
  validationMessage,
  helperMessage,
  extraButton,
  isValid,
  isTouched
}): JSX.Element => {
  const [selectedOption, setSelectedOption] = useState<CustomInputOption | undefined>()

  const handleSelect = (option: CustomInputOption): void => {
    const e = { target: { value: option.value } }
    if (onChanged && !option.disabled) {
      onChanged(e)
    }
  }

  const labelId = getCustomInputId(label)

  useEffect(() => {
    const newOption = options.find((item) => item.value === value)
    setSelectedOption(newOption)
  }, [value])

  return (
    <Form.Group className="d-flex flex-column gap-s4 w-100">
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <div className="d-flex d-md-none flex-wrap gap-s4">
        {options.map((option, index) => (
          <Card
            key={index}
            id={`${option.value}-card-select`}
            intc-id={`${option.value}-card-select`}
            aria-label={`${option.value}-card-select`}
            className={`radio-card standard-height w-100 ${value === option.value ? 'selected' : ''} ${option.disabled ? 'disabled' : ''}`}
            onClick={() => {
              handleSelect(option)
            }}
          >
            <CustomInputRadioCardBody index={index} value={value} option={option} handleSelect={handleSelect} />
          </Card>
        ))}
      </div>
      <div
        className={`d-none d-md-flex d-lg-none flex-wrap ${options.length > 1 ? 'justify-content-center' : 'justify-content-start'} gap-s4`}
      >
        {options.map((option, index) => (
          <Card
            key={index}
            id={`${option.value}-radio-card-form-select`}
            intc-id={`${option.value}-radio-card-form-select`}
            aria-label={`${option.value}-radio-card-form-select`}
            className={`radio-card standard-height half-width ${value === option.value ? 'selected' : ''} ${option.disabled ? 'disabled' : ''}`}
            onClick={() => {
              handleSelect(option)
            }}
          >
            <CustomInputRadioCardBody index={index} value={value} option={option} handleSelect={handleSelect} />
          </Card>
        ))}
      </div>
      <div
        className={`d-none d-lg-flex flex-wrap ${options.length > 1 ? 'justify-content-center' : 'justify-content-start'} gap-s4`}
      >
        {options.map((option, index) => (
          <Card
            key={index}
            id={`${option.value}-radio-card-form-select`}
            intc-id={`${option.value}-radio-card-form-select`}
            aria-label={`${option.value}-radio-card-form-select`}
            className={`radio-card standard-height ${options.length % 2 === 0 ? 'half-width' : 'third-width'} ${value === option.value ? 'selected' : ''} ${option.disabled ? 'disabled' : ''}`}
            onClick={() => {
              handleSelect(option)
            }}
          >
            <CustomInputRadioCardBody
              index={index}
              value={value}
              option={option}
              handleSelect={handleSelect}
              subtitleClassName="w-100"
            />
          </Card>
        ))}
      </div>
      {selectedOption?.feedbackHtml && <Form.Control.Feedback>{selectedOption.feedbackHtml}</Form.Control.Feedback>}
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
        <Button
          intc-id={`${labelId}btnExtra`}
          data-wap_ref={`${labelId}btnExtra`}
          variant="outline-primary"
          onClick={extraButton.buttonFunction}
        >
          {extraButton.label}
        </Button>
      )}
    </Form.Group>
  )
}

export default CustomInputRadioCard
