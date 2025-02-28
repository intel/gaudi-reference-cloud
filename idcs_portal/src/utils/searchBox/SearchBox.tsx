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
  placeholder: string
  'aria-label'?: string
  onChange: React.ChangeEventHandler<FormControlElement>
  onClickSearchButton?: React.MouseEventHandler<HTMLButtonElement>
}

const SearchBox: React.FC<SearchBoxProps> = (props): JSX.Element => {
  const { value, placeholder, onChange, onClickSearchButton } = props

  return (
    <InputGroup className="searchBox">
      <Form.Control
        type="text"
        role="searchbox"
        intc-id={props['intc-id']}
        data-wap_ref={props['intc-id']}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        aria-label={props['aria-label']}
      />
      <Button variant="icon-simple" id="search-icon" onClick={onClickSearchButton} aria-label="search-icon">
        <BsSearch />
      </Button>
    </InputGroup>
  )
}

export default SearchBox
