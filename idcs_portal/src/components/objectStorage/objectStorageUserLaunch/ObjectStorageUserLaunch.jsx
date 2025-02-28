// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import CustomInput from '../../../utils/customInput/CustomInput'
import Button from 'react-bootstrap/Button'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import ObjectStorageUsersPermissionsSelection from '../objectStorageUsersPermissionsManagement/ObjectStorageUsersPermissionsSelection'
import ButtonGroup from 'react-bootstrap/ButtonGroup'

const ObjectStorageUserLaunch = (props) => {
  // props
  const state = props.state
  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  const onClickCloseErrorModal = props.onClickCloseErrorModal

  // state variables
  const mainTitle = state.mainTitle
  const form = state.form
  const navigationBottom = state.navigationBottom
  const showReservationModal = state.showReservationModal
  const showErrorModal = state.showErrorModal
  const errorMessage = state.errorMessage
  const errorTitleMessage = state.errorTitleMessage
  const errorDescription = state.errorDescription

  // Variables
  let content = null

  // functions
  function buildCustomInput(element, key) {
    return (
      <CustomInput
        key={key}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired ? element.configInput.label + ' *' : element.configInput.label
        }
        value={element.configInput.value}
        onChanged={(event) => onChangeInput(event, element.id)}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        validationMessage={element.configInput.validationMessage}
        maxLength={element.configInput.maxLength}
      />
    )
  }

  const buildForm = () => {
    const formInputs = []
    for (const key in form) {
      const formItem = { ...form[key] }
      const element = {
        id: key,
        configInput: formItem
      }
      formInputs.push(element)
    }
    return formInputs.map((x, index) => buildCustomInput(x, index))
  }

  content = (
    <>
      <div className="section">
        <h2 intc-id="computeReserveTitle">{mainTitle}</h2>
      </div>
      <div className="section">
        {buildForm()}
        <h3>Permission management</h3>
        <ObjectStorageUsersPermissionsSelection buckets={props.objectStorages} />
        <ButtonGroup>
          {navigationBottom.map((item, index) => (
            <Button
              intc-id={`btn-objectStorageUsersLaunch-navigationBottom ${item.buttonLabel}`}
              data-wap_ref={`btn-objectStorageUsersLaunch-navigationBottom ${item.buttonLabel}`}
              aria-label={item.buttonLabel}
              key={index}
              variant={item.buttonVariant}
              onClick={item.buttonLabel === 'Create' ? onSubmit : item.buttonFunction}
            >
              {item.buttonLabel}
            </Button>
          ))}
        </ButtonGroup>
      </div>
    </>
  )

  return (
    <>
      <ReservationSubmit showReservationCreateModal={showReservationModal} />
      <ErrorModal
        showModal={showErrorModal}
        titleMessage={errorTitleMessage}
        description={errorDescription}
        message={errorMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      {content}
    </>
  )
}

export default ObjectStorageUserLaunch
