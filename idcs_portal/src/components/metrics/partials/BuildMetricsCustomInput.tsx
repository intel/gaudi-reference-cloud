// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { type CustomInputOnChangeEvent } from '../../../utils/customInput/CustomInput.types'
import CustomInput from '../../../utils/customInput/CustomInput'

interface BuildCustomInputProps {
  name: string
  input: any
  onChange: (event: CustomInputOnChangeEvent, name: string) => void
}

const BuildMetricsCustomInput: React.FC<BuildCustomInputProps> = ({ name, input, onChange }): JSX.Element => {
  return (
    <CustomInput
      key={name}
      type={input.type}
      fieldSize={input.fieldSize}
      placeholder={input.placeholder}
      label={input.label}
      value={input.value}
      isValid={input.isValid}
      isTouched={input.isTouched}
      options={input.options}
      maxWidth={input.maxWidth}
      onChanged={(event: any) => {
        onChange(event, name)
      }}
    />
  )
}

export default BuildMetricsCustomInput
