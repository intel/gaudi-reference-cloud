// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import GeneralConfirmationModal from './GeneralConfirmationModal'
import DeleteStopConfirmationModal from './DeleteStopConfirmationModal'

const DELETE_MODAL_KEYWORDS = ['Delete', 'Terminate']
const STOP_MODAL_KEYWORDS = ['Stop']

export interface ActionConfirmationProps {
  actionModalContent: any
  showModalActionConfirmation: boolean
  isDeleteModal?: boolean
  isStopModal?: boolean
  onClickModalConfirmation: (status: boolean) => void
}

const ActionConfirmation: React.FC<ActionConfirmationProps> = (props): JSX.Element => {
  const actionModalContent = props.actionModalContent
  const buttonLabel = actionModalContent.buttonLabel
  const isDeleteModal = DELETE_MODAL_KEYWORDS.includes(buttonLabel) || false
  const isStopModal = STOP_MODAL_KEYWORDS.includes(buttonLabel) || false

  const [loading, setLoading] = useState(false)

  const toggleLoading = (show: boolean): void => {
    setLoading(show)
  }

  useEffect(() => {
    if (!props.showModalActionConfirmation) {
      setLoading(false)
    }
  }, [props.showModalActionConfirmation])

  return (
    <>
      {isDeleteModal || isStopModal ? (
        <DeleteStopConfirmationModal
          {...props}
          isDeleteModal={isDeleteModal}
          isStopModal={isStopModal}
          loading={loading}
          toggleLoading={toggleLoading}
        />
      ) : (
        <GeneralConfirmationModal {...props} loading={loading} toggleLoading={toggleLoading} />
      )}
    </>
  )
}

export default ActionConfirmation
