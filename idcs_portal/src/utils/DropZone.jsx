// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useCallback } from 'react'

import { Form } from 'react-bootstrap'
import { useDropzone } from 'react-dropzone'
// import { BsUpload } from "react-icons/bs";

const DropZone = (props) => {
  const [fileContent, setFileContent] = useState('')

  const onDrop = useCallback(
    (acceptedFiles) => {
      acceptedFiles.forEach((file) => {
        const reader = new FileReader()

        reader.onabort = () => console.log('file reading was aborted')
        reader.onerror = () => console.log('file reading has failed')
        reader.onload = () => {
          // Read file content.
          const readContent = reader.result
          if (readContent.length > 0) {
            setFileContent(readContent)
            // Making sure onload passes the textarea content to the parent component.
            props.onChange('keyContent', readContent)
          }
        }
        reader.readAsText(file)
      })
    },
    [props]
  )

  const { acceptedFiles } = useDropzone({
    // Disable click and keydown behavior
    onDrop,
    multiple: false,
    accept: '.pub'
  })

  const files = acceptedFiles.map((file) => (
    <span key={file.path}>
      <i className="fa fa-thin fa-file text-primary">&nbsp;&nbsp;{file.path}</i>
      <i className="fa fa-trash fa-file float-right pt-1" onClick={() => setFileContent('')}>
        &nbsp;
      </i>
    </span>
  ))

  const onChangeSetField = (e) => {
    setFileContent(e.target.value)
    props.onChange('keyContent', e.target.value)
  }

  return (
    <>
      <section className="mb-2 pb-3">
        {fileContent !== '' && (
          <aside>
            <div className="bg-light mt-2 pl-1">{files}</div>
          </aside>
        )}

        <Form.Group className="mb-1 mt-3" controlId="keypair">
          <Form.Label>Paste your key contents: *</Form.Label>
          <Form.Control
            value={fileContent || ''}
            as="textarea"
            onChange={(e) => onChangeSetField(e)}
            rows={3}
            placeholder={'Paste the contents of your key here'}
            isInvalid={!!props.keycontenterror}
            intc-id="FileDropZoneTextAreaInput"
          />
          <Form.Control.Feedback type="invalid" intc-id="FileDropZoneInvalidError">
            {props.keycontenterror}
          </Form.Control.Feedback>
        </Form.Group>
      </section>
    </>
  )
}

export default DropZone
