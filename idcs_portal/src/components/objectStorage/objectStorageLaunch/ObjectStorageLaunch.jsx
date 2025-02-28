// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import CustomInput from '../../../utils/customInput/CustomInput'
import Button from 'react-bootstrap/Button'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import UpgradeNeededModal from '../../../utils/modals/upgradeNeededModal/UpgradeNeededModal'
import EmptyCatalogModal from '../../../utils/modals/emptyCatalogModal/EmptyCatalogModal'
import ButtonGroup from 'react-bootstrap/ButtonGroup'

const ObjectStorageLaunch = (props) => {
  // props
  const state = props.state
  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const costPerHour = props.costPerHour
  const showUpgradeNeededModal = props.showUpgradeNeededModal
  const setShowUpgradeNeededModal = props.setShowUpgradeNeededModal
  const emptyCatalogModal = props.emptyCatalogModal
  const storageUsageUnit = props.storageUsageUnit

  // State Variables
  const mainTitle = state.mainTitle
  const form = state.form
  const navigationBottom = state.navigationBottom
  const showReservationModal = state.showReservationModal
  const showErrorModal = state.showErrorModal
  const errorMessage = state.errorMessage

  // Variables
  let Content = null

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
        options={element.configInput.options}
        prepend={element.configInput.prepend}
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

  Content = (
    <>
      <div className="section">
        <h2 intc-id="ObjectStorageLaunchTitle">{mainTitle}</h2>
      </div>
      <div className="section">
        {buildForm()}
        <div intc-id="Object-Storage-Cost-InputLabel" className="d-flex flex-row valid-feedback">
          Storage Cost:{' '}
          <span className="fw-semibold px-s4">
            {costPerHour} {storageUsageUnit}{' '}
          </span>
        </div>
        <ButtonGroup>
          {navigationBottom.map((item, index) => (
            <Button
              intc-id={`btn-ObjectStorage-navigationBottom ${item.buttonLabel}`}
              data-wap_ref={`btn-ObjectStorage-navigationBottom ${item.buttonLabel}`}
              aria-label={item.buttonLabel}
              key={index}
              variant={item.buttonVariant}
              className="btn"
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
        titleMessage="Could not create your bucket"
        description={'There was an error while processing your storage bucket.'}
        message={errorMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      <UpgradeNeededModal showModal={showUpgradeNeededModal} onClose={() => setShowUpgradeNeededModal(false)} />
      <EmptyCatalogModal show={emptyCatalogModal} product="Object Storage" goBackPath="/buckets" />
      {Content}
    </>
  )
}

export default ObjectStorageLaunch
