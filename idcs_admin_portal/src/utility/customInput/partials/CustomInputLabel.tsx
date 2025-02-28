// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import Form from 'react-bootstrap/Form'
import { getCustomInputId, type CustomInputExtraButton } from '../CustomInput.types'

interface CustomInputLabelProps {
  intcId?: string
  label?: string
  hiddenLabel?: boolean
  labelButton?: CustomInputExtraButton
}

const CustomInputLabel: React.FC<CustomInputLabelProps> = ({
  label = '',
  hiddenLabel = false,
  labelButton,
  intcId
}) => {
  if ((!label || hiddenLabel) && !labelButton) {
    return null
  }

  const labelId = getCustomInputId(label)

  return (
    <div className="customInputLabel gap-s6">
      {!hiddenLabel && (
        <Form.Label intc-id={intcId ?? `${labelId}InputLabel`} className="mb-0">
          {label}
        </Form.Label>
      )}
      {labelButton !== undefined && (
        <Button
          variant="link"
          className="float-end text-decoration-underline"
          onClick={labelButton.buttonFunction}
          intc-id={label.replaceAll(' ', '').replaceAll('*', '') + 'btnLabel'}
          aria-label={labelButton.label}
        >
          {labelButton.label}
        </Button>
      )}
    </div>
  )
}

export default CustomInputLabel
