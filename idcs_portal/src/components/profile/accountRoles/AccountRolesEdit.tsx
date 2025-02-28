// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import CustomInput from '../../../utils/customInput/CustomInput'
import { Accordion, Button, ButtonGroup } from 'react-bootstrap'
import Spinner from '../../../utils/spinner/Spinner'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import PermissionHandler from './PermissionHandler'

const AccountRolesEdit = (props: any): JSX.Element => {
  const isPageReady = props.isPageReady
  const isOwnCloudAccount = props.isOwnCloudAccount

  const state = props.state
  const form = props.form

  const showReservationModal = props.showReservationModal
  const errorModal = props.errorModal
  const onClickCloseErrorModal = props.onClickCloseErrorModal

  const onSubmit = props.onSubmit
  const onChangeInput = props.onChangeInput
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple

  const resources = props.resources
  const servicePermissions = props.servicePermissions
  const resourcePermissions = props.resourcePermissions
  const resourcesList = props.resourcesList
  const resourcesListLoader = props.resourcesListLoader

  const buildCustomInput = (element: any): JSX.Element => {
    return (
      <CustomInput
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired
            ? String(element.configInput.label) + ' *'
            : String(element.configInput.label)
        }
        value={element.configInput.value}
        onChanged={(event) => onChangeInput(event, element)}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        validationMessage={element.configInput.validationMessage}
        maxLength={element.configInput.maxLength}
        options={element.configInput.options}
        hidden={element.configInput.hidden}
        hiddenLabel={element.configInput.hiddenLabel}
        maxWidth={element.configInput.maxWidth}
        radioGroupHorizontal={element.configInput.radioGroupHorizontal}
        onChangeDropdownMultiple={(values) => onChangeDropdownMultiple(values, element)}
      />
    )
  }

  const getSelectAllCheckbox = (resource: any): JSX.Element => {
    const selectAllInput = structuredClone(form.selectAllCheckbox)
    const type = resource.type
    const id = `selectAll-${resource.type}`

    const isChecked =
      Object.prototype.hasOwnProperty.call(servicePermissions, type) &&
      servicePermissions[type].length === resource.actions.length

    selectAllInput.options[0].name = !isChecked ? 'Select All' : 'Deselect All'

    selectAllInput.isChecked = isChecked
    selectAllInput.value = isChecked

    const element = {
      id,
      configInput: selectAllInput,
      resource,
      action: null,
      type: 'selectAll'
    }

    return buildCustomInput(element)
  }

  const getResources = (): JSX.Element => {
    const items = resources.map((resource: any) => {
      if (resource.actions.length === 0) {
        return null
      }
      return (
        <Accordion.Item key={resource.type} eventKey={resource.type} className="w-100 border-top">
          <Accordion.Header intc-id={`${state.keyId}ServiceTitle-${resource.type}`}>
            <h3>{resource.description}</h3>
          </Accordion.Header>
          <Accordion.Body>
            <div className="section p-0">
              {getSelectAllCheckbox(resource)}
              <PermissionHandler
                resource={resource}
                form={form}
                servicePermissions={servicePermissions}
                resourcePermissions={resourcePermissions}
                resourcesListLoader={resourcesListLoader}
                resourcesList={resourcesList}
                isOwnCloudAccount={isOwnCloudAccount}
                buildCustomInput={buildCustomInput}
              />
            </div>
          </Accordion.Body>
        </Accordion.Item>
      )
    })

    return (
      <div className="section">
        <Accordion intc-id={`${state.keyId}-accordion-`} className="d-flex-customInput border-0">
          {items}
        </Accordion>
      </div>
    )
  }

  return (
    <>
      <div className="section">
        <h2 intc-id={`${state.keyId}Title`}>{state.mainTitle}</h2>
      </div>

      {!isPageReady ? (
        <Spinner />
      ) : (
        <>
          <div className="section">
            {buildCustomInput({
              id: 'name',
              configInput: form.name,
              resource: null,
              action: null,
              type: 'name'
            })}
          </div>

          {getResources()}

          <div className="section">
            <ButtonGroup>
              {state.navigationBottom.map((item: any, index: number) => (
                <Button
                  intc-id={`${state.keyId}-btn-navigationBottom ${item.buttonLabel}`}
                  data-wap_ref={`${state.keyId}-btn-navigationBottom ${item.buttonLabel}`}
                  aria-label={item.buttonLabel}
                  key={index}
                  variant={item.buttonVariant}
                  className="btn"
                  onClick={item.buttonLabel === 'Update' ? onSubmit : item.buttonFunction}
                >
                  {item.buttonLabel}
                </Button>
              ))}
            </ButtonGroup>
          </div>
        </>
      )}

      <ErrorModal
        showModal={errorModal.showErrorModal}
        titleMessage={errorModal.errorTitleMessage}
        description={errorModal.errorDescription}
        message={errorModal.errorMessage}
        hideRetryMessage={errorModal.errorHideRetryMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      <ReservationSubmit showReservationCreateModal={showReservationModal} />
    </>
  )
}
export default AccountRolesEdit
