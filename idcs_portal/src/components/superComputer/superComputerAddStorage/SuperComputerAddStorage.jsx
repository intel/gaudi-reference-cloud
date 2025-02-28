// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import CustomInput from '../../../utils/customInput/CustomInput'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import Spinner from '../../../utils/spinner/Spinner'

const SuperComputerAddStorage = ({
  mainTitle,
  form,
  onChangeInput,
  navigationBottom,
  onSubmit,
  submitModal,
  errorModal,
  loading
}) => {
  // *****
  // functions
  // *****
  function buildCustomInput(element, index) {
    return (
      <CustomInput
        key={index}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={element.configInput.label}
        value={element.configInput.value}
        onChanged={(event) => onChangeInput(event, element.id)}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        readOnly={element.configInput.readOnly}
        refreshButton={element.configInput.refreshButton}
        extraButton={element.configInput.extraButton}
        labelButton={element.configInput.labelButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
      />
    )
  }

  // *****
  // Variables
  // *****
  const formElementsStorage = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.sectionGroup === 'storage') {
      formElementsStorage.push({
        id: key,
        configInput: formItem
      })
    }
  }
  return (
    <>
      <ReservationSubmit showReservationCreateModal={submitModal.show} />
      <ErrorModal
        showModal={errorModal.show}
        titleMessage={errorModal.title}
        description={errorModal.description}
        message={errorModal.message}
        hideRetryMessage={errorModal.hideRetryMessage}
        onClickCloseErrorModal={errorModal.onClose}
      />
      {loading ? (
        <Spinner />
      ) : (
        <>
          <div className="section">
            <h2 intc-id="computeReserveTitle">{mainTitle}</h2>
          </div>
          <div className="section">
            <h3>Storage</h3>
            {formElementsStorage.map((element, index) => {
              return buildCustomInput(element, index)
            })}
            <ButtonGroup>
              {navigationBottom.map((item, index) => (
                <Button
                  key={index}
                  intc-id={`btn-iksLaunchCluster-${item.buttonLabel}`}
                  data-wap_ref={`btn-iksLaunchCluster-${item.buttonLabel}`}
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

export default SuperComputerAddStorage
