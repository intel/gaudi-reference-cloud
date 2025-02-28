// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import CustomInput from '../../../utils/customInput/CustomInput'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import Spinner from '../../../utils/spinner/Spinner'
import EmptyCatalogModal from '../../../utils/modals/emptyCatalogModal/EmptyCatalogModal'
import CustomAlerts from '../../../utils/customAlerts/CustomAlerts'

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

const ClusterAddStorage = ({
  mainTitle,
  mainSubtitle,
  form,
  onChangeInput,
  navigationBottom,
  onSubmit,
  submitModal,
  errorModal,
  loading,
  costPerMonth,
  storageUsageUnit,
  emptyCatalogModal
}) => {
  // *****
  // functions
  // *****
  function buildCustomInput(element) {
    return (
      <div key={element.id}>
        <CustomInput
          type={element.configInput.type}
          fieldSize={element.configInput.fieldSize}
          placeholder={element.configInput.placeholder}
          isRequired={element.configInput.validationRules.isRequired}
          label={element.configInput.label}
          value={element.configInput.value}
          minLength={element.configInput.minLength}
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
          maxLength={element.configInput.maxLength}
          emptyOptionsMessage={element.configInput.emptyOptionsMessage}
        />
      </div>
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
      <EmptyCatalogModal show={emptyCatalogModal} product="storage volumes" goBackPath="/cluster" />
      {loading ? (
        <Spinner />
      ) : (
        <>
          <div className="section">
            <h2 intc-id="computeReserveTitle">{mainTitle}</h2>
          </div>
          <div className="section">
            <CustomAlerts
              showAlert
              showIcon
              alertType="warning"
              title="Warning"
              message="Storage is only supported for clusters with bare metal worker nodes."
            />
            <h3>Storage</h3>
            {formElementsStorage.map((element, index) => buildCustomInput(element))}
            <CostPerMonthElement costPerMonth={costPerMonth} storageUsageUnit={storageUsageUnit} />
            <ButtonGroup>
              {navigationBottom.map((item, index) => (
                <Button
                  key={index}
                  intc-id={`btn-iksLaunchCluster-${item.buttonLabel}`}
                  data-wap_ref={`btn-iksLaunchCluster-${item.buttonLabel}`}
                  variant={item.buttonVariant}
                  className="btn"
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

export default ClusterAddStorage
