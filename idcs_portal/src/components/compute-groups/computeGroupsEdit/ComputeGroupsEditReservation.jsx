// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import CustomInput from '../../../utils/customInput/CustomInput'
import ModalCreatePublicKey from '../../../utils/modals/modalCreatePublicKey/ModalCreatePublicKey'
import CustomAlerts from '../../../utils/customAlerts/CustomAlerts'
import Spinner from '../../../utils/spinner/Spinner'

const ComputeGroupsEditReservation = (props) => {
  // Props
  const state = props.state
  const mainTitle = state.mainTitle
  const instanceDetailsMenuSection = state.instanceDetailsMenuSection
  const publicKeysMenuSection = state.publicKeysMenuSection
  const form = state.form
  const onChangeInput = props.onChangeInput
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const navigationBottom = state.navigationBottom
  const showPublicKeyModal = props.showPublicKeyModal
  const onShowHidePublicKeyModal = props.onShowHidePublicKeyModal
  const onSubmit = props.onSubmit
  const afterPubliKeyCreate = props.afterPubliKeyCreate
  const loading = props.loading

  // Functions
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
        selectAllButton={element.configInput.selectAllButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
      />
    )
  }

  // variables
  const formELementsInstanceDetails = []
  const formElementskeys = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.sectionGroup === 'instanceDetails') {
      formELementsInstanceDetails.push({
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
  }

  return (
    <>
      <ModalCreatePublicKey
        showModalActionConfirmation={showPublicKeyModal}
        closeCreatePublicKeyModal={onShowHidePublicKeyModal}
        afterPubliKeyCreate={afterPubliKeyCreate}
        isModal={true}
      />
      {loading ? (
        <Spinner />
      ) : (
        <>
          <div className="section">
            <h2>{mainTitle}</h2>
          </div>
          <div className="section">
            <h3>{instanceDetailsMenuSection}</h3>
            {formELementsInstanceDetails.map((element) => buildCustomInput(element))}
            <h3>{publicKeysMenuSection}</h3>
            <CustomAlerts
              showAlert
              showIcon
              alertType="secondary"
              message="After saving your changes, follow the next steps to complete the instance key update process"
            />
            {formElementskeys.map((element) => buildCustomInput(element))}
            <ButtonGroup>
              {navigationBottom.map((item, index) => (
                <Button
                  key={index}
                  aria-label={item.buttonLabel}
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

export default ComputeGroupsEditReservation
