// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import { BsCopy } from 'react-icons/bs'
import { useCopy } from '../hooks/useCopy'

const CodeLine = (props) => {
  const { copyToClipboard } = useCopy()
  const codeline = props.codeline
  const customCodeLine = props.customCodeLine
  const customSpacing = props.customSpacing || 'mt-s4'

  return (
    <div className={`section code-line rounded-3 ${customSpacing}`}>
      <div className="row mt-0 align-items-center">
        <pre className="col">
          <span className="ps-2">{codeline}</span>
        </pre>
        <div className="col-auto mt-0">
          <Button variant="secondary" onClick={() => copyToClipboard(customCodeLine || codeline)}>
            <BsCopy />
            Copy
          </Button>
        </div>
      </div>
    </div>
  )
}

export default CodeLine
