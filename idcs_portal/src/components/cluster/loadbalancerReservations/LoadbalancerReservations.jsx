// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import Button from 'react-bootstrap/Button'
import CustomInput from '../../../utils/customInput/CustomInput'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import Spinner from '../../../utils/spinner/Spinner'

const LoadbalancerReservations = (props) => {
  // Props
  const state = props.state
  const errorMessage = props.errorMessage
  const form = props.form
  const showReservationModal = props.showReservationModal
  const showErrorModal = props.showErrorModal
  const errorHideRetryMessage = props.errorHideRetryMessage
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const loading = props.loadingal

  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  // functions
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
        question={element.configInput.question}
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
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
      />
    )
  }

  // variables
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
      <ReservationSubmit showReservationCreateModal={showReservationModal} />
      <ErrorModal
        showModal={showErrorModal}
        description={state.errorDescription}
        message={errorMessage}
        titleMessage={state.titleMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
        hideRetryMessage={errorHideRetryMessage}
      />
      {loading ? (
        <Spinner />
      ) : (
        <>
          <div className="section">
            <h2>{state.mainTitle}</h2>
          </div>
          <div className="section">
            {formElementsClusterLoadbalancers.map((element) => buildCustomInput(element))}

            <ButtonGroup>
              {state.navigationBottom.map((item, index) => (
                <Button
                  key={index}
                  intc-id={`btn-iksLaunchVip-${item.buttonLabel}`}
                  data-wap_ref={`btn-iksLaunchVip-${item.buttonLabel}`}
                  variant={item.buttonVariant}
                  className="btn"
                  onClick={item.buttonLabel === 'Launch' ? onSubmit : item.buttonFunction}
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

export default LoadbalancerReservations
