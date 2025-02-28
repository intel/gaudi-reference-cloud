import React, { useState } from 'react'
import { Alert, Form } from 'react-bootstrap'
import Wrapper from '../../utility/wrapper/Wrapper'
import CloudAccountService from '../../services/CloudAccountService'
import CustomInput from '../../utility/customInput/CustomInput'
import { Link } from 'react-router-dom'
import HowToCreateSSHKey from './howToCreateSSHKey/HowToCreateSSHKey'
import { IoIosWarning } from 'react-icons/io'
import { BsArrowLeftShort } from 'react-icons/bs'
import { UpdateFormHelper, isValidForm, getFormValue } from '../../utility/updateFormHelper/UpdateFormHelper'

const KeyPairsForm = (props) => {
  // local state
  const formInitial = {
    keyPairName: {
      type: 'text', // options = 'text ,'textArea'
      fieldSize: 'medium', // options = 'small', 'medium', 'large'
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
      fieldSize: 'medium', // options = 'small', 'medium', 'large'
      label: 'key contents',
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
  const [showErrors, setShowErrors] = useState(false)
  const [alertMessage, setAlertMessage] = useState()
  const [showWarningMessage, setShowWarningMessage] = useState(true)
  const payload = props.payload
  const isModal = props.isModal
  const callAfterSuccess = props.callAfterSuccess

  // Errors Related functions
  function onChange(value, key) {
    const formUpdated = UpdateFormHelper(
      value,
      key,
      form
    )
    setValidForm(isValidForm(formUpdated))
    setState(formUpdated)
  }

  const handleSubmit = (e) => {
    e.preventDefault()

    const newPayload = { ...payload }

    const metadata = { ...newPayload.metadata }

    const spec = { ...newPayload.spec }

    metadata.name = getFormValue('keyPairName', form)

    spec.sshPublicKey = getFormValue('keyPairContent', form).trim()

    newPayload.metadata = metadata
    newPayload.spec = spec

    CloudAccountService.postSshByCloud(newPayload)
      .then((res) => {
        if (res.status === 200) {
          callAfterSuccess()
        }
      })
      .catch((error) => {
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

        setShowErrors(true)
        setAlertMessage(errorMessage)
      })
  }

  return (
    <Wrapper>
      <div className="container">
        {showErrors &&
          (<Alert
            variant="danger"
            className="col-4 center"
            onClose={() => setShowErrors(false)}
            dismissible
          >
            {alertMessage}
          </Alert>)
        }
        <div className={`mt-3 ${isModal ? 'm-2 mb-0' : 'main'}`}>
          {!isModal
            ? <>
              <Link to="/security/publickeys/" className="text-decoration-none">
                <BsArrowLeftShort />
                Back to keys list
              </Link>
              <h2 intc-id="myPublicKeysTitle" className="mt-5">Upload key</h2>
            </>
            : null}

          {showWarningMessage && <Alert
            variant="warning"
            className="col-12"
            onClose={() => setShowWarningMessage(false)}
            dismissible
          >
            <IoIosWarning className="me-2" style={{ fontSize: '32px' }} />
            <span className="h5 pe-3 align-middle">Warning</span>
            <span className="align-middle">Never share your private keys with anyone. Never create a SSH Private key without a passphrase</span>
          </Alert>
          }

          <div className={`row bg-white ${!isModal ? 'mt-3 p-1' : null}`}>
            {!isModal
              ? <h3 className="mt-3">SSH key details</h3>
              : null}
            <Form onSubmit={handleSubmit}>
              <div className="row">

                <div className="col-md-12">
                  <CustomInput
                    key={0}
                    type={form.keyPairName.type}
                    fieldSize={form.keyPairName.fieldSize}
                    placeholder={form.keyPairName.placeholder}
                    isRequired={form.keyPairName.validationRules.isRequired}
                    label={form.keyPairName.label}
                    value={form.keyPairName.value}
                    onChanged={(e) => onChange(e.target.value, 'keyPairName')}
                    isValid={form.keyPairName.isValid}
                    isTouched={form.keyPairName.isTouched}
                    isReadOnly={form.keyPairName.isReadOnly}
                    validationMessage={form.keyPairName.validationMessage}
                    maxLength={form.keyPairName.maxLength}
                  />
                </div>
              </div>

              <HowToCreateSSHKey />

              <div className="row">
                <div className="col-md-12">
                  <h3 className="h6 my-3 fw-bold">Key contents</h3>
                  <CustomInput
                    key={0}
                    type={form.keyPairContent.type}
                    fieldSize={form.keyPairContent.fieldSize}
                    placeholder={form.keyPairContent.placeholder}
                    isRequired={form.keyPairContent.validationRules.isRequired}
                    label={'Paste your key contents: *'}
                    value={form.keyPairContent.value}
                    onChanged={(e) => onChange(e.target.value, 'keyPairContent')}
                    isValid={form.keyPairContent.isValid}
                    isTouched={form.keyPairContent.isTouched}
                    isReadOnly={form.keyPairContent.isReadOnly}
                    validationMessage={form.keyPairContent.validationMessage}
                    maxLength={form.keyPairContent.maxLength}
                  />
                </div>
              </div>

              <div className={`${!isModal ? 'mb-3' : null}`}>

                {isModal
                  ? <>
                    <div className="d-flex flex-row-reverse">
                      <button className="btn btn-primary" intc-id="createPublicKeyButton" disabled={!validForm}>Upload key</button>
                      <a className="btn-sm btn btn-link text-decoration-none my-auto" intc-id="cancelPublicKeyButton" onClick={() => { props.handleClose() }} >Cancel</a>
                    </div>
                  </>
                  : <>
                    <button className="btn btn-primary" intc-id="createPublicKeyButton" disabled={!validForm}>Upload</button>
                    <Link className="btn-sm btn btn-link text-decoration-none" intc-id="cancelPublicKeyButton" to={'/security/publickeys/'}>Cancel</Link>
                  </>
                }
              </div>
            </Form>

          </div>

        </div>
      </div>
    </Wrapper>
  )
}

export default KeyPairsForm
