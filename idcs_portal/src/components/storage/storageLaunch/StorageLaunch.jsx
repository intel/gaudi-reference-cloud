// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import ComingSoonBanner from '../../../utils/comingSoonBanner/ComingSoonBanner'
import CustomInput from '../../../utils/customInput/CustomInput'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import EmptyCatalogModal from '../../../utils/modals/emptyCatalogModal/EmptyCatalogModal'
import UpgradeNeededModal from '../../../utils/modals/upgradeNeededModal/UpgradeNeededModal'

const CostPerMonthElement = ({ costPerMonth, storageUsageUnit }) => {
  if (costPerMonth === '') {
    return (
      <div className="d-flex flex-row valid-feedback">
        Volume cost: <span className="fw-semibold px-s4">Enter size to calculate.</span>
      </div>
    )
  }

  if (costPerMonth === 'error') {
    return <div className="d-flex flex-row valid-feedback">Could not calculate cost.</div>
  }

  return (
    <div className="d-flex flex-row valid-feedback">
      Volume cost: &nbsp;<span className="fw-semibold">{costPerMonth}</span> &nbsp;{storageUsageUnit}
    </div>
  )
}

const StorageLaunch = (props) => {
  const state = props.state
  const emptyCatalogModal = props.emptyCatalogModal
  const mainTitle = state.mainTitle
  const configSectionTitle = state.configSectionTitle
  const isStorageFlagAvailable = props.isStorageFlagAvailable
  const messageBanner = props.messageBanner
  const form = state.form
  const onChangeInput = props.onChangeInput
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const navigationBottom = state.navigationBottom
  const onSubmit = props.onSubmit
  const showReservationModal = state.showReservationModal
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const showErrorModal = state.showErrorModal
  const errorMessage = state.errorMessage
  const costPerMonth = props.costPerMonth
  const storageUsageUnit = props.storageUsageUnit
  const showUpgradeNeededModal = props.showUpgradeNeededModal
  const setShowUpgradeNeededModal = props.setShowUpgradeNeededModal

  let Content = null

  // variables
  const formElementsInstanceConfiguration = []

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
        onChangeDropdownMultiple={onChangeDropdownMultiple}
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
        maxLength={element.configInput.maxLength}
      />
    )
  }

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.sectionGroup === 'configuration') {
      formElementsInstanceConfiguration.push({
        id: key,
        configInput: formItem
      })
    }
  }

  if (isStorageFlagAvailable) {
    Content = (
      <>
        <div className="section">
          <h2 intc-id="computeReserveTitle">{mainTitle}</h2>
        </div>
        <div className="section">
          <h3>{configSectionTitle}</h3>
          {formElementsInstanceConfiguration.map((element, index) => buildCustomInput(element, index))}
          <CostPerMonthElement costPerMonth={costPerMonth} storageUsageUnit={storageUsageUnit} />
          <ButtonGroup>
            {navigationBottom.map((item, index) => (
              <Button
                intc-id={`btn-storagelaunch-navigationBottom ${item.buttonLabel}`}
                data-wap_ref={`btn-storagelaunch-navigationBottom ${item.buttonLabel}`}
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
  } else {
    return <ComingSoonBanner message={messageBanner} />
  }

  return (
    <>
      <UpgradeNeededModal showModal={showUpgradeNeededModal} onClose={() => setShowUpgradeNeededModal(false)} />
      <ReservationSubmit showReservationCreateModal={showReservationModal} />
      <EmptyCatalogModal show={emptyCatalogModal} product="storage volumes" goBackPath="/storage" />
      <ErrorModal
        showModal={showErrorModal}
        titleMessage="Could not create your volume"
        description={'There was an error while processing your volume.'}
        message={errorMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      {Content}
    </>
  )
}

export default StorageLaunch
