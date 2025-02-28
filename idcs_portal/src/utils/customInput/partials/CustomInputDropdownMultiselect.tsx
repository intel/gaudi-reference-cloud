// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useRef, useState } from 'react'
import { v4 as uuidv4 } from 'uuid'
import {
  type CustomInputProps,
  type CustomInputOption,
  type CustomInputOnChangeEvent,
  getCustomInputId
} from '../CustomInput.types'
import Form from 'react-bootstrap/Form'
import Button from 'react-bootstrap/Button'
import Dropdown from 'react-bootstrap/Dropdown'
import './_CustomInputs.scss'
import CustomInputLabel from './CustomInputLabel'

const CustomInputDropdownMultiselect: React.FC<CustomInputProps> = ({
  label = '',
  labelButton,
  maxWidth,
  helperMessage,
  options = [],
  isReadOnly,
  hiddenLabel,
  value = [],
  placeholder = '',
  onChangeDropdownMultiple,
  extraButton,
  validationMessage,
  isValid,
  isTouched,
  customClass,
  emptyOptionsMessage,
  refreshButton,
  selectAllButton,
  isFilter = true
}): JSX.Element => {
  const [filterText, setFilterText] = useState('')
  const dropdownRef = useRef<HTMLDivElement>(null)
  const [showDropdown, setShowDropdown] = useState(false)

  const onChange = (
    event: React.ChangeEvent<HTMLInputElement>,
    optionOnChange?: (event: CustomInputOnChangeEvent) => void
  ): void => {
    event.stopPropagation()

    const values = [...value]
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

  const onOptionSelect = (event: any, option: CustomInputOption): void => {
    event.stopPropagation()

    const values = [...value]
    const val = option.value

    const index = values.indexOf(val, 0)
    if (index !== -1) values.splice(index, 1)
    else values.push(val)

    if (onChangeDropdownMultiple) {
      onChangeDropdownMultiple(values)
    }

    if (option?.onChanged) {
      option?.onChanged(event)
    }
  }

  const onClickOutside = (event: MouseEvent): void => {
    if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
      setShowDropdown(false)
    }
  }

  useEffect(() => {
    document.addEventListener('mousedown', onClickOutside)
    return () => {
      document.removeEventListener('mousedown', onClickOutside)
    }
  }, [])

  const onFilterChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
    setFilterText(event.target.value)
  }

  const labelId = getCustomInputId(label)
  const filteredOptions = options.filter((option) => option.name.toLowerCase().includes(filterText.toLowerCase()))

  const displayOptions = options.filter((option) => value.includes(option.value)).map((x) => x.name)

  const displayText =
    displayOptions.length > 2
      ? `${displayOptions.length} options selected`
      : displayOptions.length > 0
        ? displayOptions.join(', ')
        : placeholder
  const isAllSelected = filteredOptions.length > 0 && filteredOptions.every((option) => value.includes(option.value))

  return (
    <Form.Group className="d-flex-customInput" style={maxWidth ? { maxWidth } : undefined}>
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      {options.length > 0 && (
        <div ref={dropdownRef}>
          <Dropdown
            style={maxWidth ? { maxWidth } : undefined}
            show={showDropdown}
            onToggle={() => {
              setShowDropdown(!showDropdown)
            }}
          >
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
              disabled={isReadOnly}
              onClick={() => {
                setShowDropdown(!showDropdown)
              }}
            >
              <div className={`gap-s4 dropdownInput ${value.length === 0 ? 'empty' : ''} `}>{displayText}</div>
            </Dropdown.Toggle>
            <Dropdown.Menu
              id={`${labelId}-menu-form-select`}
              renderOnMount
              flip
              aria-labelledby={`${labelId}-toggle-form-select`}
              className="form-dropdown-select-menu customInputMultiSelectDropdown"
            >
              {isFilter && (
                <Dropdown.Item
                  intc-id={`${labelId}-form-select-option-Filter-Option`}
                  aria-label={'Select option Filter-Option'}
                >
                  <Form.Control
                    autoFocus
                    placeholder="Filter options..."
                    value={filterText}
                    onChange={onFilterChange}
                    onClick={(e) => {
                      e.stopPropagation()
                    }}
                  />
                </Dropdown.Item>
              )}
              {selectAllButton && filteredOptions.length > 0 && (
                <Dropdown.Item
                  intc-id={`${labelId}-form-select-option-Filter-Option`}
                  aria-label={'Select option Filter-Option'}
                >
                  <Form.Check
                    type="checkbox"
                    id="select-all"
                    label={selectAllButton.label}
                    checked={isAllSelected}
                    onChange={selectAllButton.buttonFunction}
                    onClick={(e) => {
                      e.stopPropagation()
                    }}
                    disabled={Boolean(filterText)}
                  />
                </Dropdown.Item>
              )}

              {filteredOptions.map((option, index) => (
                <Dropdown.Item
                  key={index}
                  intc-id={`${labelId}-form-select-option-${getCustomInputId(option.name)}`}
                  aria-label={`Select option ${option.name}`}
                  onClick={(e) => {
                    onOptionSelect(e, option)
                  }}
                >
                  <Form.Check
                    type="checkbox"
                    id={`${labelId}-Input-option-${getCustomInputId(option.name)}-${uuidv4()}`}
                    intc-id={`${labelId}-Input-option-${getCustomInputId(option.name)}`}
                    label={option.name}
                    aria-label={
                      value.indexOf(index) !== -1 ? 'Uncheck ' + String(option.name) : 'Check ' + String(option.name)
                    }
                    onChange={(e) => {
                      onChange(e, option.onChanged)
                    }}
                    disabled={option.disabled ?? isReadOnly}
                    value={option.value as string | number | undefined}
                    checked={value.includes(option.value)}
                    className={customClass ?? ''}
                    onClick={(e) => {
                      e.stopPropagation()
                    }}
                  />
                </Dropdown.Item>
              ))}
            </Dropdown.Menu>
          </Dropdown>
        </div>
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
    </Form.Group>
  )
}

export default CustomInputDropdownMultiselect
