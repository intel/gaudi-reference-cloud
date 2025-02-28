// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Form from 'react-bootstrap/Form'
import Button from 'react-bootstrap/Button'
import { BsNodePlus } from 'react-icons/bs'
import CustomInputLabel from './CustomInputLabel'
import { getCustomInputId, type CustomInputProps } from '../CustomInput.types'
import { InputGroup } from 'react-bootstrap'
import CustomInputText from './CustomInputText'
import LineDivider from '../../lineDivider/LineDivider'

const CustomInputDictionary: React.FC<CustomInputProps> = ({
  dictionaryOptions = [],
  onChangeTagValue,
  label = '',
  labelButton,
  hiddenLabel,
  onClickActionTag,
  helperMessage,
  maxLength,
  maxWidth,
  fieldSize,
  validationMessage,
  isValid,
  isTouched
}) => {
  // props functions

  const labelId = getCustomInputId(label)
  return (
    <Form.Group className="d-flex-customInput" style={maxWidth ? { maxWidth } : undefined}>
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <InputGroup size={fieldSize} style={maxWidth ? { maxWidth } : undefined}>
        {dictionaryOptions.map((option, index) => (
          <React.Fragment key={index}>
            <div className="d-flex flex-row gap-s6 w-100 align-items-start">
              <CustomInputText
                intcId={`${labelId}-input-dictionary-${getCustomInputId(String(option.key.label))}-${index}`}
                label={option.key.label}
                value={option.key.value}
                maxLength={option.key.maxLength}
                onChanged={(e) => {
                  if (onChangeTagValue) {
                    onChangeTagValue(e, 'key', index)
                  }
                }}
                placeholder={option.key.placeholder}
                isValid={option.key.isValid}
                validationMessage={option.key.validationMessage}
                isTouched={option.key.isTouched}
              />
              <CustomInputText
                intcId={`${labelId}-input-dictionary-${getCustomInputId(String(option.value.label))}-${index}`}
                label={option.value.label}
                value={option.value.value}
                maxLength={option.value.maxLength}
                onChanged={(e) => {
                  if (onChangeTagValue) {
                    onChangeTagValue(e, 'value', index)
                  }
                }}
                placeholder={option.value.placeholder}
                isValid={option.value.isValid}
                validationMessage={option.value.validationMessage}
                isTouched={option.value.isTouched}
              />
              <div className="d-flex">
                <Button
                  variant="close"
                  onClick={() => {
                    if (onClickActionTag) {
                      onClickActionTag(index, 'Delete')
                    }
                  }}
                  aria-label={`Remove Tag on index ${index}`}
                ></Button>
              </div>
            </div>
            {index !== dictionaryOptions.length - 1 && <LineDivider horizontal className="w-100" />}
          </React.Fragment>
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
      <div className="d-flex flex-row gap-s6 align-items-center">
        <Button
          disabled={maxLength !== undefined && dictionaryOptions.length >= maxLength}
          variant="outline-primary"
          onClick={() => {
            if (onClickActionTag) {
              onClickActionTag(null, 'Add')
            }
          }}
        >
          <BsNodePlus />
          Add Tag
        </Button>
        <Form.Control.Feedback type="valid">
          <div className="d-flex flex-row w-100 gap-s8 justify-content-between">
            {maxLength && <>{`Up to ${maxLength} tags max. (${maxLength - dictionaryOptions.length}) remaining`}</>}
          </div>
        </Form.Control.Feedback>
      </div>
    </Form.Group>
  )
}

export default CustomInputDictionary
