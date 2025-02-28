// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import Button from 'react-bootstrap/Button'
import CustomInput from '../../../utils/customInput/CustomInput'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import { formatCurrency, formatNumber } from '../../../utils/numberFormatHelper/NumberFormatHelper'
import UpgradeNeededModal from '../../../utils/modals/upgradeNeededModal/UpgradeNeededModal'
import EmptyCatalogModal from '../../../utils/modals/emptyCatalogModal/EmptyCatalogModal'
import { Link } from 'react-router-dom'
import { Spinner } from 'react-bootstrap'

const ClusterReserve = (props) => {
  // Props
  const mainTitle = props.mainTitle
  const navigationBottom = props.navigationBottom
  const form = props.form
  const clusterMenuSection = props.clusterMenuSection
  const showReservationModal = props.showReservationModal
  const showErrorModal = props.showErrorModal
  const errorMessage = props.errorMessage
  const errorHideRetryMessage = props.errorHideRetryMessage
  const errorDescription = props.errorDescription
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const titleMessage = props.titleMessage
  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  const clusterProduct = props.clusterProduct
  const showUpgradeNeededModal = props.showUpgradeNeededModal
  const setShowUpgradeNeededModal = props.setShowUpgradeNeededModal
  const emptyCatalogModal = props.emptyCatalogModal
  const isPageReady = props.isPageReady
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
        question={element.configInput.question}
        onChanged={(event) => onChangeInput(event, element.id)}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        readOnly={element.configInput.readOnly}
        maxLength={element.configInput.maxLength}
        refreshButton={element.configInput.refreshButton}
        extraButton={element.configInput.extraButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
        hidden={element.configInput.hidden}
      />
    )
  }

  // variables
  const formElementsClusterConfiguration = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.sectionGroup === 'clusterConfiguration') {
      formElementsClusterConfiguration.push({
        id: key,
        configInput: formItem
      })
    }
  }

  function getPricing() {
    return formatCurrency(formatNumber(clusterProduct.rate * 60, 2))
  }

  return (
    <>
      <ReservationSubmit showReservationCreateModal={showReservationModal} />
      <ErrorModal
        showModal={showErrorModal}
        description={errorDescription}
        message={errorMessage}
        titleMessage={titleMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
        hideRetryMessage={errorHideRetryMessage}
      />
      <UpgradeNeededModal showModal={showUpgradeNeededModal} onClose={() => setShowUpgradeNeededModal(false)} />
      <EmptyCatalogModal
        show={emptyCatalogModal}
        product="kubernetes clusters"
        goBackPath="/cluster"
        extraExplanation="You can still launch individual compute instances."
        extraActions={
          <Link to={'/compute/reserve'} className="btn btn-outline-primary">
            Launch instance
          </Link>
        }
      />
      {!isPageReady ? (
        <Spinner />
      ) : (
        <>
          <div className="section">
            <h2>{mainTitle}</h2>
          </div>
          <div className="section">
            <h3>{clusterMenuSection}</h3>
            {formElementsClusterConfiguration.map((element, index) => buildCustomInput(element, index))}
            {clusterProduct && (
              <div intc-id="IKS-Cluster-Cost-InputLabel" className="d-flex flex-row valid-feedback">
                Cluster cost: <span className="fw-semibold px-s4">{getPricing()} per hour </span>
              </div>
            )}
            <ButtonGroup>
              {navigationBottom.map((item, index) => (
                <Button
                  key={index}
                  intc-id={`btn-iksLaunchCluster-${item.buttonLabel}`}
                  data-wap_ref={`btn-iksLaunchCluster-${item.buttonLabel}`}
                  disabled={item.buttonLabel === 'Launch' ? !isPageReady : false}
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

export default ClusterReserve
