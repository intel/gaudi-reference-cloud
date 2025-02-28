// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState } from 'react'
import { Button, ButtonGroup } from 'react-bootstrap'
import CustomInput from '../../utils/customInput/CustomInput'
import { Link } from 'react-router-dom'
import HowToCreateSSHKey from './howToCreateSSHKey/HowToCreateSSHKey'
import {
  UpdateFormHelper,
  getFormValue,
  isValidForm,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import CloudAccountService from '../../services/CloudAccountService'
import useToastStore from '../../store/toastStore/ToastStore'
import CustomAlerts from '../../utils/customAlerts/CustomAlerts'
import { toastMessageEnum } from '../../utils/Enums'

const KeyPairsForm = (props) => {
  // props variables
  const isModal = props.isModal
  const keysPagePath = props.keysPagePath
  const callAfterSuccess = props.callAfterSuccess

  // local state
  const formInitial = {
    keyPairName: {
      type: 'text', // options = 'text ,'textArea'
      label: 'Key Name: *',
      placeholder: 'Key Name',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 30,
      validationRules: {
        isRequired: true,
        onlyAlphaNumLower: true,
        checkMaxLength: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: ''
    },
    keyPairContent: {
      type: 'textArea', // options = 'text ,'textArea'
      label: 'key contents:',
      placeholder: 'Paste your key contents',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      validationRules: {
        isRequired: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: ''
    }
  }

  const [form, setState] = useState(formInitial)
  const [validForm, setValidForm] = useState(false)
  const showError = useToastStore((state) => state.showError)

  // Errors Related functions
  function onChange(event, key) {
    const value = event.target.value

    const formUpdated = UpdateFormHelper(value, key, form)
    setValidForm(isValidForm(formUpdated))

    setState(formUpdated)
  }

  function showRequiredFields() {
    let formCopy = { ...form }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(form)
    // Create toast
    if (!isModal) {
      showError(toastMessageEnum.formValidationError, false)
    }

    formCopy = updatedForm
    setState(formCopy)
  }

  const handleSubmit = async () => {
    if (!validForm) {
      showRequiredFields()
      return
    }
    const newPayload = {
      metadata: {
        name: getFormValue('keyPairName', form)
      },
      spec: {
        sshPublicKey: getFormValue('keyPairContent', form).trim()
      }
    }

    try {
      const res = await CloudAccountService.postSshByCloud(newPayload)
      if (res.status === 200) {
        callAfterSuccess()
      }
    } catch (error) {
      let errorMessage = ''
      if (error.response) {
        if (error.response.data.code === 2) {
          errorMessage = 'Duplicate keypair name'
        } else {
          errorMessage = error.response.data.message
        }
      } else {
        errorMessage = error.message
      }
      showError(errorMessage)
    }
  }

  return (
    <>
      <div className="section">
        <h2 intc-id="myPublicKeysTitle">Upload key </h2>
      </div>
      <div className="section">
        <CustomAlerts
          showAlert
          showIcon
          alertType="warning"
          title="Warning"
          message="Never share your private keys with anyone. Never create a SSH Private key without a passphrase"
        />
        <h3>SSH key details</h3>
        <CustomInput
          type={form.keyPairName.type}
          fieldSize={form.keyPairName.fieldSize}
          placeholder={form.keyPairName.placeholder}
          isRequired={form.keyPairName.validationRules.isRequired}
          label={form.keyPairName.label}
          value={form.keyPairName.value}
          onBlur={() => {}}
          onChanged={(e) => onChange(e, 'keyPairName')}
          isValid={form.keyPairName.isValid}
          isTouched={form.keyPairName.isTouched}
          isReadOnly={form.keyPairName.isReadOnly}
          validationMessage={form.keyPairName.validationMessage}
          maxLength={form.keyPairName.maxLength}
        />
        <HowToCreateSSHKey />
        <h3>Key contents</h3>
        <CustomInput
          type={form.keyPairContent.type}
          fieldSize={form.keyPairContent.fieldSize}
          placeholder={form.keyPairContent.placeholder}
          isRequired={form.keyPairContent.validationRules.isRequired}
          label={'Paste your key contents: *'}
          value={form.keyPairContent.value}
          onBlur={() => {}}
          onChanged={(e) => onChange(e, 'keyPairContent')}
          isValid={form.keyPairContent.isValid}
          isTouched={form.keyPairContent.isTouched}
          isReadOnly={form.keyPairContent.isReadOnly}
          validationMessage={form.keyPairContent.validationMessage}
          maxLength={form.keyPairContent.maxLength}
          textAreaRows={12}
          customClass="w-100"
        />
      </div>
      <div className={isModal ? 'modal-footer' : 'section'}>
        <ButtonGroup className={`d-flex ${isModal ? 'flex-row-reverse' : 'flex-row'} `}>
          <Button
            intc-id="btn-ssh-createpublickey"
            variant="primary"
            type="submit"
            aria-label="Upload key"
            onClick={handleSubmit}
            data-wap_ref="btn-ssh-createpublickey"
          >
            Upload key
          </Button>
          {isModal && (
            <Button
              intc-id="btn-ssh-cancelPublicKey"
              variant="link"
              onClick={() => {
                props.handleClose()
              }}
            >
              Cancel
            </Button>
          )}
          {!isModal && (
            <Link
              className="btn-sm btn btn-link text-decoration-none"
              intc-id="btn-ssh-cancelPublicKey"
              data-wap_ref="btn-ssh-cancelPublicKey"
              to={keysPagePath}
            >
              Cancel
            </Link>
          )}
        </ButtonGroup>
      </div>
    </>
  )
}

export default KeyPairsForm
