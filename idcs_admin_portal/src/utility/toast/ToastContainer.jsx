// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { BsCheck2Circle, BsExclamationOctagon, BsInfoCircle, BsExclamationCircle } from 'react-icons/bs'
import Toast from 'react-bootstrap/Toast'
import ToastContainerRb from 'react-bootstrap/ToastContainer'
import useToastStore from '../../store/toastStore/ToastStore'

const ToastWithCounter = ({ toast }) => {
  const [timeCounter, setTimeCounter] = useState(0)
  const removeToast = useToastStore((state) => state.removeToast)

  useEffect(() => {
    const interval = setInterval(() => {
      setTimeCounter((state) => state + 5)
    }, 5000)

    return () => clearInterval(interval)
  }, [])

  const getCounter = () => {
    if (timeCounter < 5) {
      return 'Just now'
    }
    if (timeCounter < 60) {
      return `${timeCounter} seconds ago`
    }
    if (timeCounter >= 60) {
      return `${Math.trunc(timeCounter / 60)} minutes ago`
    }
  }

  const getTitle = (variant) => {
    switch (variant) {
      case 'success':
        return 'Success'
      case 'warning':
        return 'Warning'
      case 'info':
        return 'Information'
      case 'danger':
        return 'Error'
      default:
        return null
    }
  }

  const getTitleIcon = (variant) => {
    switch (variant) {
      case 'success':
        return <BsCheck2Circle />
      case 'warning':
        return <BsExclamationCircle />
      case 'info':
        return <BsInfoCircle />
      case 'danger':
        return <BsExclamationOctagon />
      default:
        return null
    }
  }

  return (
    <Toast
      key={toast.id}
      bg={toast.variant}
      autohide={toast.autohide}
      delay={toast.delay}
      onClose={() => {
        removeToast(toast.id)
      }}
    >
      <Toast.Header>
        {getTitleIcon(toast.variant)}
        <span className="fw-bold">{getTitle(toast.variant)}</span>
        <span className="small pe-s3 ms-auto">{getCounter()}</span>
      </Toast.Header>
      <hr className="toast-divider"></hr>
      <Toast.Body className="d-flex flex-row justify-content-between">
        <span>{toast.bodyMessage}</span>
      </Toast.Body>
    </Toast>
  )
}

const ToastContainer = () => {
  const toasts = useToastStore((state) => state.toasts)
  const unixTime = React.useMemo(() => Date.now(), [])
  const id = `toast-container-${unixTime}`

  useEffect(() => {
    const element = document.getElementById(id)
    if (element) {
      const parent = element.parentNode
      const isInsideOfModal = parent.className.indexOf('modal-content') !== -1
      if (isInsideOfModal) {
        const dialogContainer = parent.parentNode.parentNode
        dialogContainer.appendChild(element)
      }
    }
  }, [])

  return (
    <ToastContainerRb
      id={id}
      className="p-3 mb-5"
      style={{
        position: 'fixed',
        display: 'block',
        bottom: 0,
        right: 0,
        zIndex: 9999
      }}
      containerPosition="body"
    >
      {toasts && toasts.map((toast) => <ToastWithCounter key={toast.id} toast={toast} />)}
    </ToastContainerRb>
  )
}

export default ToastContainer
