// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useRef, useState, useEffect } from 'react'
import { v4 as uuidv4 } from 'uuid'
import Form from 'react-bootstrap/Form'
import Button from 'react-bootstrap/Button'
import { getCustomInputId, type CustomInputProps } from '../CustomInput.types'
import CustomInputLabel from './CustomInputLabel'
import './_CustomInputs.scss'
import { OverlayTrigger, Tooltip } from 'react-bootstrap'

const CustomInputRange: React.FC<CustomInputProps> = ({
  label = '',
  labelButton,
  fieldSize,
  isReadOnly,
  hiddenLabel,
  minRange = 0,
  maxRange = 100,
  value = '',
  step = 1,
  onChanged,
  onBlur,
  extraButton,
  helperMessage,
  validationMessage,
  isValid,
  isTouched
}): JSX.Element => {
  const labelId = getCustomInputId(label)
  const [offset, setOffset] = useState(0)
  const sliderRef = useRef<HTMLInputElement>(null)

  const updateTooltipPosition = (): void => {
    if (sliderRef.current) {
      const { min, max, value } = sliderRef.current
      const range = parseFloat(max) - parseFloat(min)
      const percent = ((parseFloat(value) - parseFloat(min)) / range) * 100
      const sliderWidth = sliderRef.current.offsetWidth
      const newOffset = sliderWidth * ((percent - 50) / 100)
      setOffset(newOffset)
    }
  }

  const renderTooltip = (props: any): JSX.Element => <Tooltip {...props}>{value}</Tooltip>

  useEffect(() => {
    updateTooltipPosition()
  }, [value])

  return (
    <Form.Group className="d-flex-customInput">
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <OverlayTrigger
        overlay={renderTooltip}
        popperConfig={{
          modifiers: [
            {
              name: 'offset',
              options: {
                offset: [offset, 0],
                placement: 'top'
              }
            }
          ]
        }}
      >
        <Form.Control
          id={`${labelId}Range-${uuidv4()}`}
          intc-id={`${labelId}Range`}
          ref={sliderRef}
          type="range"
          className={'form-range'}
          value={value}
          onChange={onChanged}
          onBlur={onBlur}
          min={minRange}
          max={maxRange}
          step={step}
          disabled={isReadOnly}
          size={fieldSize}
        />
      </OverlayTrigger>
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

export default CustomInputRange
