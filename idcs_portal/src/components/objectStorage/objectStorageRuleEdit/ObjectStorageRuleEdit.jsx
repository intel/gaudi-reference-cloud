// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import ButtonGroup from 'react-bootstrap/ButtonGroup'
import Button from 'react-bootstrap/Button'
import CustomInput from '../../../utils/customInput/CustomInput'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import Spinner from '../../../utils/spinner/Spinner'

const ObjectStorageRuleEdit = (props) => {
  // props
  const state = props.state
  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const loading = props.loading

  // State Variables
  const mainTitle = state.mainTitle
  const form = state.form
  const navigationBottom = state.navigationBottom
  const showReservationModal = state.showReservationModal
  const showErrorModal = state.showErrorModal
  const errorMessage = state.errorMessage

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
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        maxLength={element.configInput.maxLength}
        customClass={element.configInput.customClass}
      />
    )
  }

  const buildForm = (sectionGroup = true) => {
    const formInputs = []
    for (const key in form) {
      const formItem = { ...form[key] }
      const element = {
        id: key,
        configInput: formItem
      }

      if (formItem.sectionGroup === sectionGroup) {
        formInputs.push(buildCustomInput(element, key))
      }
    }
    return formInputs
  }

  return (
    <>
      <ReservationSubmit showReservationCreateModal={showReservationModal} />
      <ErrorModal
        showModal={showErrorModal}
        titleMessage="Could not update your Lifecycle Rule"
        description={'There was an error while processing your lifecycle rule.'}
        message={errorMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      {loading ? (
        <Spinner />
      ) : (
        <>
          <div className="section">
            <h2 intc-id="ObjectStorageEditTitle">{mainTitle}</h2>
          </div>
          <div className="section">
            {buildForm('configuration')}
            {buildForm('deleteMarker')}
            {buildForm('noncurrentExpireDays')}
            <ButtonGroup>
              {navigationBottom.map((item, index) => (
                <Button
                  intc-id={`btn-ObjectStorageRuleEdit-navigationBottom ${item.buttonLabel}`}
                  data-wap_ref={`btn-ObjectStorageRuleEdit-navigationBottom ${item.buttonLabel}`}
                  aria-label={item.buttonLabel}
                  key={index}
                  variant={item.buttonVariant}
                  onClick={item.buttonLabel === 'Save' ? onSubmit : item.buttonFunction}
                >
                  {item.buttonLabel}
                </Button>
              ))}
            </ButtonGroup>
          </div>
        </>
      )}
    </>
  )
}

export default ObjectStorageRuleEdit
