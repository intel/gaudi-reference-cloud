// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { Modal } from 'react-bootstrap'

const CustomModal = (props: any): JSX.Element => {
  const show = props.show
  const onHide = props.onHide
  const title = props.title
  const body = props.body
  const footer = props.footer
  const size = props.size
  return (
    <Modal
      show={show}
      onHide={() => {
        onHide(false)
      }}
      backdrop="static"
      size={size}
      aria-label="Custom content modal"
    >
      <Modal.Header closeButton closeLabel="Close request extention modal">
        <Modal.Title>{title}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {body}
      </Modal.Body>
      <Modal.Footer>
        {footer}
      </Modal.Footer>
    </Modal>
  )
}

export default CustomModal
