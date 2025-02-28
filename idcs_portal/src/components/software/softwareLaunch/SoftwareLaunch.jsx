// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import ComingSoonBanner from '../../../utils/comingSoonBanner/ComingSoonBanner'
import CustomInput from '../../../utils/customInput/CustomInput'
import ModalCreatePublicKey from '../../../utils/modals/modalCreatePublicKey/ModalCreatePublicKey'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import UpgradeNeededModal from '../../../utils/modals/upgradeNeededModal/UpgradeNeededModal'
import ModalInstanceCompare from '../../../utils/modals/modalInstanceCompare/ModalInstanceCompare'
import { ButtonGroup } from 'react-bootstrap'
import './SoftwareLaunch.scss'
import LineDivider from '../../../utils/lineDivider/LineDivider'

const SoftwareLaunch = (props) => {
  // props
  const state = props.state
  const isAvailable = props.isAvailable
  const comingMessage = props.comingMessage
  const onChangeInput = props.onChangeInput
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const monthlyEstimateRates = props.monthlyEstimateRates
  const showPublicKeyModal = props.showPublicKeyModal
  const onShowHidePublicKeyModal = props.onShowHidePublicKeyModal
  const afterPubliKeyCreate = props.afterPubliKeyCreate
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const showUpgradeNeededModal = props.showUpgradeNeededModal
  const setShowUpgradeNeededModal = props.setShowUpgradeNeededModal
  const category = props.category
  const onSubmit = props.onSubmit
  const showInstanceCompareModal = props.showInstanceCompareModal
  const onShowHideInstanceCompareModal = props.onShowHideInstanceCompareModal
  const afterInstanceSelected = props.afterInstanceSelected
  const products = props.products
  const mainTitle = props.mainTitle

  // state
  const showReservationModal = state.showReservationModal
  const showErrorModal = state.showErrorModal
  const errorMessage = state.errorMessage
  const errorHideRetryMessage = state.errorHideRetryMessage
  const errorTitleMessage = state?.errorTitleMessage
  const errorDescription = state.errorDescription
  const navigationBottom = state.navigationBottom
  const isValidForm = state.isValidForm
  const mainSubtitle = state.mainSubtitle
  const instanceConfigSectionTitle = state.instanceConfigSectionTitle
  const publicKeysMenuSection = state.publicKeysMenuSection
  const showMonthlyEstimate = false

  const osInputElements = []
  const configurationInputElements = []
  const publicKeyInputElements = []

  for (const key in state.form) {
    const formItem = {
      ...state.form[key]
    }

    if (formItem.sectionGroup === 'operationSystem') {
      osInputElements.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.sectionGroup === 'configuration') {
      configurationInputElements.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.sectionGroup === 'keys') {
      publicKeyInputElements.push({
        id: key,
        configInput: formItem
      })
    }
  }

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
        hidden={element.configInput.hidden}
      />
    )
  }

  let content = null
  if (!isAvailable) {
    return <ComingSoonBanner message={comingMessage} />
  } else {
    content = (
      <>
        <div className="section">
          <h2 intc-id="computeSoftwareReserveTitle">{mainTitle}</h2>
          <p>{mainSubtitle}</p>
        </div>
        <div className="flex-xs-column flex-md-row align-self-stretch gap-s8">
          <div className="section softwareLaunch">
            {osInputElements.map((element, index) => buildCustomInput(element))}
            <h3>{instanceConfigSectionTitle}</h3>
            {configurationInputElements.map((element, index) => buildCustomInput(element))}
            <h3>{publicKeysMenuSection}</h3>
            {publicKeyInputElements.map((element) => buildCustomInput(element))}
            <ButtonGroup>
              {navigationBottom.map((item, index) => (
                <Button
                  intc-id={`btn-softwarelaunch-navigationBottom ${item.buttonLabel} - ${category}`}
                  data-wap_ref={`btn-softwarelaunch-navigationBottom ${item.buttonLabel} - ${category}`}
                  key={index}
                  disabled={item.buttonLabel === 'Launch' ? !isValidForm : false}
                  variant={item.buttonVariant}
                  onClick={item.buttonLabel === 'Launch' ? onSubmit : item.buttonFunction}
                >
                  {item.buttonLabel}
                </Button>
              ))}
            </ButtonGroup>
          </div>
          {showMonthlyEstimate && (
            <>
              <LineDivider className="d-xs-none d-md-flex" vertical />
              <div className="section">
                <h4>Monthly estimate</h4>
                <h5>{monthlyEstimateRates.totalMonth}</h5>
                <p className="mb-s6">(about {monthlyEstimateRates.totalHour} hourly)</p>
                <span className="h5">{monthlyEstimateRates.totalMonth}</span>
                <p>(about {monthlyEstimateRates.totalHour} hourly)</p>
                <div className="d-flex flex-row gap-4">
                  <span className="fw-bold">Selected Software:</span>
                  <span>{monthlyEstimateRates.softwareRate}</span>
                </div>
                <div className="d-flex flex-row gap-4">
                  <span className="fw-bold">Selected Instance:</span>
                  <span>{monthlyEstimateRates.instanceRate}</span>
                </div>
                <div className="d-flex flex-row gap-4">
                  <span className="fw-bold">Total:</span>
                  <span>{monthlyEstimateRates.totalMonth}</span>
                </div>
              </div>
            </>
          )}
        </div>
      </>
    )
  }
  return (
    <>
      <ModalCreatePublicKey
        showModalActionConfirmation={showPublicKeyModal}
        closeCreatePublicKeyModal={onShowHidePublicKeyModal}
        afterPubliKeyCreate={afterPubliKeyCreate}
        isModal={true}
      />
      <ModalInstanceCompare
        showModalActionConfirmation={showInstanceCompareModal}
        closeInstanceCompareModal={onShowHideInstanceCompareModal}
        afterInstanceSelected={afterInstanceSelected}
        products={products}
      />
      <ReservationSubmit showReservationCreateModal={showReservationModal} />
      <ErrorModal
        showModal={showErrorModal}
        titleMessage={errorTitleMessage}
        description={errorDescription}
        message={errorMessage}
        hideRetryMessage={errorHideRetryMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      <UpgradeNeededModal showModal={showUpgradeNeededModal} onClose={() => setShowUpgradeNeededModal(false)} />
      {content}
    </>
  )
}

export default SoftwareLaunch
