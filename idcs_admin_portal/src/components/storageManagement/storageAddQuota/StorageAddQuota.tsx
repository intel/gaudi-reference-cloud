// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { Button, Card, ButtonGroup } from 'react-bootstrap'
import SearchBox from '../../../utility/searchBox/SearchBox'
import OnSubmitModal from '../../../utility/modals/onSubmitModal/OnSubmitModal'
import CustomInput from '../../../utility/customInput/CustomInput'
import SelectedAccountCard from '../../../utility/selectedAccountCard/SelectedAccountCard'

const AddQuotaContainer = (props: any): JSX.Element => {
  // *****
  // Variables
  // *****
  const state = props.state
  const title = state.title
  const form = state.form
  const onCancel = props.onCancel
  const onChangeInput = props.onChangeInput
  const onSearchCloudAccount = props.onSearchCloudAccount
  const cloudAccountSelected = props.cloudAccountSelected
  const searchModal = props.searchModal
  const navigationBottom = state.navigationBottom
  const onSubmit = props.onSubmit

  const submitButtonLabels = ['Request']
  const formElementsCloudAccount = []
  const formElementsConfiguration = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.sectionGroup === 'configuration') {
      formElementsConfiguration.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.sectionGroup === 'cloudAccount') {
      formElementsCloudAccount.push({
        id: key,
        configInput: formItem
      })
    }
  }

  const buildCustomInput = (element: any, key: number): JSX.Element => {
    return (
      <CustomInput
        key={element.id}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired
            ? String(element.configInput.label) + ' *'
            : element.configInput.label
        }
        value={element.configInput.value}
        onChanged={(event) => onChangeInput(event, element.id)}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        refreshButton={element.configInput.refreshButton}
        extraButton={element.configInput.extraButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
      />
    )
  }

  return (
    <>
      <OnSubmitModal showModal={searchModal.show} message={searchModal.message}></OnSubmitModal>
      <div className="section">
        <Button intc-id="navigationTop-BackButton" variant="link" className="p-s0" onClick={onCancel}>
          ⟵ Back to Previous Page
        </Button>
      </div>
      <div className="section">
        <h2 className="h4">{title}</h2>
      </div>
      <div className="section">
        {formElementsCloudAccount.map((element, index) => (
          <div className="d-flex-customInput" key={index}>
            <div className="customInputLabel gap-s6" intc-id="filterText">
              Cloud Account: *
            </div>
            <div className="input-group">
              <SearchBox
                intc-id="searchCloudAccounts"
                classCustom={
                  !element.configInput.isValid && element.configInput.isTouched ? 'form-control is-invalid' : ''
                }
                placeholder={`${String(element.configInput.placeholder)}`}
                aria-label="Type to search cloud account..."
                value={element.configInput.value || ''}
                onChange={(e) => onChangeInput(e, element.id)}
                onClickSearchButton={onSearchCloudAccount}
              />
              {!element.configInput.isValid && element.configInput.isTouched
                ? (
                  <div className="invalid-feedback mt-2" intc-id="terminateInstanceError">
                    {element.configInput.validationMessage}
                  </div>
                  )
                : null}
            </div>
          </div>
        ))}
      </div>
      {cloudAccountSelected.show
        ? (
          <div className="section">
            <div className="d-flex flex-row gap-4">
              <SelectedAccountCard selectedCloudAccount={cloudAccountSelected}/>
              <Card>
                <Card.Header>Default Quota</Card.Header>
                <Card.Body>
                  <div className="d-flex flex-column">
                    <div>File Size Quota: {cloudAccountSelected?.defaultQuota.filesizeQuotaInTB}</div>
                    <div>Volume Quota: {cloudAccountSelected?.defaultQuota.filevolumesQuota}</div>
                    <div>Bucket Quota: {cloudAccountSelected?.defaultQuota.bucketsQuota}</div>
                  </div>
                </Card.Body>
              </Card>
              <Card>
                <Card.Header>Custom Quota</Card.Header>
                <Card.Body>
                  <div className="d-flex flex-column">
                    <div>File Size Quota: {cloudAccountSelected?.updatedQuota?.filesizeQuotaInTB ?? 'Not requested'}</div>
                    <div>Volume Quota: {cloudAccountSelected?.updatedQuota?.filevolumesQuota ?? 'Not requested'}</div>
                    <div>Bucket Quota: {cloudAccountSelected?.updatedQuota?.bucketsQuota ?? 'Not requested'}</div>
                  </div>
                </Card.Body>
              </Card>
            </div>
          </div>
          )
        : null}
      <div className="section">
        {formElementsConfiguration.map((element, index) => {
          return buildCustomInput(element, index)
        })}
      </div>
      <div className="section">
        <ButtonGroup>
          {navigationBottom.map((item: any, index: number) => (
            <Button
              intc-id={`btn-storageaddquota-navigationBottom-${String(item.buttonLabel)}`}
              data-wap_ref={`btn-storageaddquota-navigationBottom-${String(item.buttonLabel)}`}
              aria-label={item.buttonLabel}
              key={index}
              variant={item.buttonVariant}
              onClick={submitButtonLabels.includes(item.buttonLabel) ? onSubmit : item.buttonFunction}
            >
              {item.buttonLabel}
            </Button>
          ))}
        </ButtonGroup>
      </div>
    </>
  )
}

export default AddQuotaContainer
