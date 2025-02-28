// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import { Link } from 'react-router-dom'
import CustomInput from '../../../utils/customInput/CustomInput'
import ModalCreatePublicKey from '../../../utils/modals/modalCreatePublicKey/ModalCreatePublicKey'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import ModalUpgradeAccount from '../../../utils/modals/modalUpgradeAccount/ModalUpgradeAccount'
import UpgradeNeededModal from '../../../utils/modals/upgradeNeededModal/UpgradeNeededModal'
import EmptyCatalogModal from '../../../utils/modals/emptyCatalogModal/EmptyCatalogModal'
import { ButtonGroup } from 'react-bootstrap'
import Spinner from '../../../utils/spinner/Spinner'
import CostEstimateCard from '../../../utils/costEstimate/CostEstimateCard'
import { BsCalculator } from 'react-icons/bs'
import CostEstimateModal from '../../../utils/costEstimate/CostEstimateModal'

const ComputeLaunch = (props) => {
  // props
  const state = props.state
  const mainTitle = state.mainTitle
  const form = state.form
  const instanceConfigSectionTitle = state.instanceConfigSectionTitle
  const requestSectionTitle = state.requestSectionTitle
  const publicKeysMenuSection = state.publicKeysMenuSection
  const instancLabelsMenuSection = state.instancLabelsMenuSection
  const showPublicKeyModal = props.showPublicKeyModal
  const showCostEstimateModal = props.showCostEstimateModal
  const onChangeInput = props.onChangeInput
  const onShowHidePublicKeyModal = props.onShowHidePublicKeyModal
  const onShowHideCostEstimateModal = props.onShowHideCostEstimateModal
  const afterPubliKeyCreate = props.afterPubliKeyCreate
  const navigationBottom = state.navigationBottom
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const showReservationModal = state.showReservationModal
  const onSubmit = props.onSubmit
  const showErrorModal = state.showErrorModal
  const errorMessage = state.errorMessage
  const errorHideRetryMessage = state.errorHideRetryMessage
  const errorTitleMessage = state?.errorTitleMessage
  const errorDescription = state.errorDescription
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const showPremiumModal = props.showPremiumModal
  const premiumCancelButtonOptions = props.premiumCancelButtonOptions
  const premiumFormActions = props.premiumFormActions
  const premiumError = props.premiumError
  const showUpgradeNeededModal = props.showUpgradeNeededModal
  const setShowUpgradeNeededModal = props.setShowUpgradeNeededModal
  const emptyCatalogModal = props.emptyCatalogModal
  const category = props.category
  const computeReservationsPagePath = props.computeReservationsPagePath
  const onChangeTagValue = props.onChangeTagValue
  const onClickActionTag = props.onClickActionTag
  const loading = props.loading
  const quickConnectMenuSection = state.quickConnectMenuSection
  const costEstimate = state.costEstimate

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
          element.configInput.validationRules.isRequired && element.configInput.label
            ? element.configInput.label + ' *'
            : element.configInput.label
        }
        subLabel={element.configInput.subLabel}
        hiddenLabel={element.configInput.hiddenLabel}
        value={element.configInput.value}
        maxWidth={element.configInput.maxWidth}
        maxInputWidth={element.configInput.maxInputWidth}
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
        selectAllButton={element.configInput.selectAllButton}
        labelButton={element.configInput.labelButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
        dictionaryOptions={element.configInput.dictionaryOptions}
        onChangeTagValue={onChangeTagValue}
        onClickActionTag={onClickActionTag}
        hidden={element.configInput.hidden}
        idField={element.id}
        columns={element.configInput.columns}
        singleSelection={element.configInput.singleSelection}
        selectedRecords={element.configInput.selectedRecords}
        setSelectedRecords={element.configInput.setSelectedRecords}
        emptyGridMessage={element.configInput.emptyGridMessage}
        gridBreakpoint={element.configInput.gridBreakpoint}
        gridOptions={element.configInput.options}
      />
    )
  }

  // variables
  const formElementsInstanceConfiguration = []
  const formElementsRequestInformation = []
  const formElementskeys = []
  const formElementsQuickConnect = []
  const formElementsInstanceLabels = []
  const submitButtonLabels = ['Launch instance', 'Launch instance group', 'Launch node group']

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

    if (formItem.sectionGroup === 'request') {
      formElementsRequestInformation.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.sectionGroup === 'keys') {
      formElementskeys.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.sectionGroup === 'quickConnect' && formItem.hidden === false) {
      formElementsQuickConnect.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.sectionGroup === 'instancesLabels') {
      formElementsInstanceLabels.push({
        id: key,
        configInput: formItem
      })
    }
  }

  return (
    <>
      <ModalCreatePublicKey
        showModalActionConfirmation={showPublicKeyModal}
        closeCreatePublicKeyModal={onShowHidePublicKeyModal}
        afterPubliKeyCreate={afterPubliKeyCreate}
        isModal={true}
      />
      {costEstimate ? (
        <CostEstimateModal
          showModal={showCostEstimateModal}
          title={costEstimate.title}
          description={costEstimate.description}
          costArray={costEstimate.costArray}
          onHide={() => onShowHideCostEstimateModal(false)}
        />
      ) : null}
      <ReservationSubmit showReservationCreateModal={showReservationModal} />
      <ErrorModal
        showModal={showErrorModal}
        titleMessage={errorTitleMessage}
        description={errorDescription}
        message={errorMessage}
        hideRetryMessage={errorHideRetryMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      <ModalUpgradeAccount
        showPremiumModal={showPremiumModal}
        cancelButtonOptions={premiumCancelButtonOptions}
        formActions={premiumFormActions}
        error={premiumError}
      />
      <UpgradeNeededModal showModal={showUpgradeNeededModal} onClose={() => setShowUpgradeNeededModal(false)} />
      {emptyCatalogModal ? (
        <EmptyCatalogModal
          show={emptyCatalogModal.show}
          product={emptyCatalogModal.product}
          goBackPath={emptyCatalogModal.goBackPath}
          extraExplanation={emptyCatalogModal.extraExplanation}
          extraActions={
            <Link to={emptyCatalogModal.launchPath || computeReservationsPagePath} className="btn btn-outline-primary">
              {emptyCatalogModal.launchLabel || 'Launch instance'}
            </Link>
          }
        />
      ) : null}
      <div className="section">
        <h2 intc-id="computeReserveTitle">{mainTitle}</h2>
      </div>
      {loading ? (
        <Spinner />
      ) : (
        <>
          <div className="d-flex w-100">
            <div className="section col-xs-12 col-md-9">
              <h3>{instanceConfigSectionTitle}</h3>
              {formElementsInstanceConfiguration.map((element, index) => buildCustomInput(element, index))}
              {requestSectionTitle && <h3>{requestSectionTitle}</h3>}
              {formElementsRequestInformation.map((element, index) => buildCustomInput(element, index))}
              <h3>{publicKeysMenuSection}</h3>
              {formElementskeys.map((element, index) => buildCustomInput(element, index))}
              <h3>{instancLabelsMenuSection}</h3>
              {formElementsInstanceLabels.map((element, index) => buildCustomInput(element, index))}
              {formElementsQuickConnect.length > 0 ? (
                <>
                  <h3>{quickConnectMenuSection}</h3>
                  {formElementsQuickConnect.map((element, index) => buildCustomInput(element, index))}
                </>
              ) : null}
              {costEstimate ? (
                <Button
                  className="d-sm-none"
                  intc-id="btn-show-cost-estimate-modal"
                  data-wap_ref="btn-show-cost-estimate-modal"
                  aria-label="Show cost estimate modal"
                  variant="link"
                  onClick={() => onShowHideCostEstimateModal(true)}
                >
                  <BsCalculator />
                  Cost estimate
                </Button>
              ) : null}
              <ButtonGroup className="w-100">
                {navigationBottom.map((item, index) => (
                  <Button
                    intc-id={`btn-computelaunch-navigationBottom ${item.buttonLabel} - ${category ?? ''}`}
                    data-wap_ref={`btn-computelaunch-navigationBottom ${item.buttonLabel} - ${category ?? ''}`}
                    aria-label={item.buttonLabel}
                    key={index}
                    variant={item.buttonVariant}
                    onClick={submitButtonLabels.includes(item.buttonLabel) ? onSubmit : item.buttonFunction}
                  >
                    {item.buttonLabel}
                  </Button>
                ))}
                {costEstimate ? (
                  <Button
                    className="d-none d-sm-flex d-md-none ms-auto"
                    intc-id="btn-show-cost-estimate-modal"
                    data-wap_ref="btn-show-cost-estimate-modal"
                    aria-label="Show cost estimate modal"
                    variant="link"
                    onClick={() => onShowHideCostEstimateModal(true)}
                  >
                    <BsCalculator />
                    Cost estimate
                  </Button>
                ) : null}
              </ButtonGroup>
            </div>
            {costEstimate ? (
              <div className="section d-none d-md-flex">
                <CostEstimateCard
                  title={costEstimate.title}
                  description={costEstimate.description}
                  costArray={costEstimate.costArray}
                ></CostEstimateCard>
              </div>
            ) : null}
          </div>
        </>
      )}
    </>
  )
}

export default ComputeLaunch
