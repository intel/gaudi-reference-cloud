// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { BsSearch } from 'react-icons/bs'
import Button from 'react-bootstrap/Button'
import InputGroup from 'react-bootstrap/InputGroup'
import Form from 'react-bootstrap/Form'
import './searchBox.scss'

type FormControlElement = HTMLInputElement | HTMLTextAreaElement

interface SearchBoxProps {
  'intc-id'?: string
  value: string
  className?: string
  classCustom?: string
  placeholder: string
  'aria-label'?: string
  onChange: React.ChangeEventHandler<FormControlElement>
  onClickSearchButton?: React.MouseEventHandler<HTMLButtonElement>
}

const SearchBox: React.FC<SearchBoxProps> = (props): JSX.Element => {
  const { value, placeholder, onChange, classCustom, className } = props
  const onClickSearchButton: any = props.onClickSearchButton
  const onEnterKeyPress = (event: React.KeyboardEvent<FormControlElement>): void => {
    if (event.key === 'Enter') {
      // Cancel the default action, if needed
      event.preventDefault()
      if (onClickSearchButton) onClickSearchButton()
    }
  }

  return (
    <InputGroup className={`searchBox ${className ?? ''}`}>
      <Form.Control
        type="text"
        role="searchbox"
        className={classCustom}
        intc-id={props['intc-id']}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        aria-label={props['aria-label']}
        onKeyUp={onEnterKeyPress}
      />
      <Button variant="icon-simple" id="search-icon" onClick={onClickSearchButton} aria-label="search-icon">
        <BsSearch />
      </Button>
    </InputGroup>
  )
}

export default SearchBox
