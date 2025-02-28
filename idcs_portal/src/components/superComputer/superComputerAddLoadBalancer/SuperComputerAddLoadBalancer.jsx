// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import CustomInput from '../../../utils/customInput/CustomInput'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import Spinner from '../../../utils/spinner/Spinner'

const SuperComputerAddLoadBalancer = ({
  navigationBottom,
  form,
  isValidForm,
  mainTitle,
  onChangeInput,
  onSubmit,
  errorModal,
  onCloseErrorModal,
  reservationModal,
  loading
}) => {
  // *****
  // functions
  // *****
  function buildCustomInput(element) {
    return (
      <CustomInput
        key={element.id}
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
        readOnly={element.configInput.readOnly}
        refreshButton={element.configInput.refreshButton}
        extraButton={element.configInput.extraButton}
        labelButton={element.configInput.labelButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
      />
    )
  }
  const formElementsClusterLoadbalancers = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.sectionGroup === 'lb') {
      formElementsClusterLoadbalancers.push({
        id: key,
        configInput: formItem
      })
    }
  }
  return (
    <>
      <ReservationSubmit showReservationCreateModal={reservationModal.show} />
      <ErrorModal showModal={errorModal.show} message={errorModal.message} onClickCloseErrorModal={onCloseErrorModal} />
      {loading ? (
        <Spinner />
      ) : (
        <div className="section">
          <h2>{mainTitle}</h2>
          {formElementsClusterLoadbalancers.map((element, index) => {
            return buildCustomInput(element, index)
          })}
          <ButtonGroup>
            {navigationBottom.map((item, index) => (
              <Button
                key={index}
                intc-id={`btn-iksLaunchCluster-${item.buttonLabel}`}
                data-wap_ref={`btn-iksLaunchCluster-${item.buttonLabel}`}
                aria-label={item.buttonLabel}
                variant={item.buttonVariant}
                onClick={item.buttonLabel === 'Launch' ? onSubmit : item.buttonFunction}
              >
                {item.buttonLabel}
              </Button>
            ))}
          </ButtonGroup>
        </div>
      )}
    </>
  )
}

export default SuperComputerAddLoadBalancer
