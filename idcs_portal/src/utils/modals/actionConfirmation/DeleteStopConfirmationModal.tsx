// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import { Modal, Button } from 'react-bootstrap'
import {
  UpdateFormHelper,
  getFormValue,
  isValidForm,
  showFormRequiredFields
} from '../../updateFormHelper/UpdateFormHelper'
import CustomInput from '../../customInput/CustomInput'
import { type ActionConfirmationProps } from './ActionConfirmation'
import CustomAlerts from '../../customAlerts/CustomAlerts'
import useToastStore from '../../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../Enums'
import ToastContainer from '../../toast/ToastContainer'
import SpinnerIcon from '../../spinner/SpinnerIcon'

export interface DeleteStopConfirmationModalProps extends ActionConfirmationProps {
  loading: boolean
  toggleLoading: (show: boolean) => void
}

const DeleteStopConfirmationModal: React.FC<DeleteStopConfirmationModalProps> = (props): JSX.Element => {
  const modalType = props.isDeleteModal ? 'Delete' : 'Stop'
  const actionModalContent = props.actionModalContent
  const label = actionModalContent.label
  const name = actionModalContent.name || 'delete'

  // local variables
  const item = label.replace(`${modalType} `, '')
  const invalidInput = 'Please provide a valid name.'
  const initialForm: any = {
    form: {
      name: {
        type: 'text',
        label: 'Name',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        maxLength: 63,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: invalidInput,
        helperMessage: ''
      }
    },
    isValidForm: false
  }

  const [formState, setFormState] = useState(initialForm)

  useEffect(() => {
    if (props.showModalActionConfirmation) setFormState(initialForm)
  }, [props.showModalActionConfirmation])

  // Global State
  const showError = useToastStore((state) => state.showError)

  // functions
  const onSubmitWhenEnter = (event: any): void => {
    if (event.key === 'Enter') {
      onSubmit(true)
    }
  }

  const onChangeInput = (value: any, formInputName: string): void => {
    const updatedState = {
      ...formState
    }
    const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)
    if (value.length > 0) {
      updatedForm.name.isValid = value === name
      updatedForm.name.validationMessage = value !== name ? invalidInput : ''
    }
    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setFormState(updatedState)
  }

  const onSubmit = (status: boolean): void => {
    if (status) {
      const isValidForm = formState.isValidForm
      if (!isValidForm) {
        showRequiredFields()
        return
      }
      props.toggleLoading(true)
    }
    props.onClickModalConfirmation(status)
  }

  const showRequiredFields = (): void => {
    const stateCopy = { ...formState }
    const nameValue = getFormValue('name', stateCopy.form)
    if (nameValue.length === 0) {
      // Mark regular Inputs
      const updatedForm = showFormRequiredFields(stateCopy.form)
      stateCopy.form = updatedForm
      setFormState(stateCopy)
    }
    // Create toast
    showError(toastMessageEnum.formValidationError, false)
  }

  const isVowelWord = (word: string): boolean => {
    const vowelRegex = /^[aeiouAEIOU]/
    return vowelRegex.test(word)
  }

  const getID = (label: string, type: string): string => {
    const id = `btn-confirm-${label}-${type.toLowerCase()}`
    return id.replace(' ', '')
  }

  const deleteModalMessage = (
    <>
      <CustomAlerts
        showAlert={true}
        alertType="warning"
        message={`Deleting ${isVowelWord(item) ? 'an' : 'a'} ${item} cannot be undone.`}
        showIcon={true}
      />
      <div className="d-flex flex-column">
        <span>
          Are you sure you want to delete the{' '}
          <span className="fw-semibold text-break" intc-id="deleteConfirmationName">
            {name}
          </span>{' '}
          {item}?
        </span>
        <span>To confirm deletion enter the name of the {item} below.</span>
      </div>
    </>
  )

  const stopModalMessage = (
    <>
      <CustomAlerts
        showAlert={true}
        alertType="warning"
        message={`The ${item} will stay reserved in your account and
          keep incurring charges.`}
        showIcon={true}
      />
      <div className="d-flex flex-column gap-s6">
        <span>
          Are you sure you want to stop the{' '}
          <span className="fw-semibold" intc-id="deleteConfirmationName">
            {name}
          </span>{' '}
          {item}?
        </span>
        <span>
          Stopping an {item} will pause it. You can restart it later. To permanently delete the {item} and avoid
          charges, please use the &quot;Delete&quot; option.
        </span>
        <span>To confirm enter the name of the {item} below.</span>
      </div>
    </>
  )

  return (
    <Modal
      show={props.showModalActionConfirmation}
      onHide={() => {
        onSubmit(false)
      }}
      backdrop="static"
      keyboard={false}
      intc-id={`${modalType.toLowerCase()}ConfirmModal`}
      aria-label={`${modalType} confirmation modal`}
    >
      <ToastContainer />
      <Modal.Header closeButton>
        <Modal.Title className="text-break">{`${label} "${name}"`}</Modal.Title>
      </Modal.Header>
      <Modal.Body className="d-flex flex-column gap-s6">
        {props.isDeleteModal ? deleteModalMessage : stopModalMessage}
        <CustomInput
          type={formState.form.name.type}
          placeholder={`Enter ${item} name`}
          label={formState.form.name.label}
          value={formState.form.name.value}
          onChanged={(e) => {
            onChangeInput(e.target.value, 'name')
          }}
          onKeyDown={(e) => {
            onSubmitWhenEnter(e)
          }}
          maxLength={formState.form.name.maxLength}
          fieldSize={formState.form.name.fieldSize}
          isRequired={formState.form.name.validationRules.isRequired}
          isValid={formState.form.name.isValid}
          isTouched={formState.form.name.isTouched}
          isReadOnly={formState.form.name.isReadOnly}
          validationMessage={formState.form.name.validationMessage}
          helperMessage={formState.form.name.helperMessage}
        />
      </Modal.Body>
      <Modal.Footer>
        <Button
          intc-id={getID(label, 'cancel')}
          data-wap_ref={getID(label, 'cancel')}
          aria-label="Cancel"
          variant="outline-primary"
          onClick={() => {
            onSubmit(false)
          }}
        >
          Cancel
        </Button>
        <Button
          intc-id={getID(label, modalType)}
          data-wap_ref={getID(label, modalType)}
          aria-label={modalType}
          variant={'danger'}
          disabled={props.loading}
          onClick={() => {
            onSubmit(true)
          }}
        >
          {props.loading && <SpinnerIcon />}
          {modalType}
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default DeleteStopConfirmationModal
