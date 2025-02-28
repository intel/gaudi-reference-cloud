// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Spinner from '../../../utils/spinner/Spinner'

interface PermissionHandlerProps {
  resource: any
  form: any
  servicePermissions: any
  resourcePermissions: any
  resourcesListLoader?: any
  resourcesList: any
  buildCustomInput: (element: any) => JSX.Element
  isOwnCloudAccount: boolean
  viewMode?: boolean
}

const PermissionHandler: React.FC<PermissionHandlerProps> = (props): JSX.Element => {
  const resource = props.resource
  const form = props.form
  const servicePermissions = props.servicePermissions
  const resourcePermissions = props.resourcePermissions
  const resourcesListLoader = props.resourcesListLoader
  const resourcesList = props.resourcesList
  const buildCustomInput = props.buildCustomInput
  const isOwnCloudAccount = props.isOwnCloudAccount
  const viewMode = props.viewMode

  const actions = resource.actions
  const collectionAction = actions.filter((x: any) => x.type === 'collection')
  const resourceAction = actions.filter((x: any) => x.type === 'resource')

  const getPermissionCheckbox = (resource: any, action: any): any => {
    const form = props.form
    const servicePermissions = props.servicePermissions

    const permissionInput = structuredClone(form.permissionCheckbox)
    const id = `${resource.type}-${action.name}`
    const type = resource.type
    const name = action.name

    permissionInput.options[0].name = action.description

    const isChecked =
      Object.prototype.hasOwnProperty.call(servicePermissions, type) && servicePermissions[type].includes(name)

    permissionInput.isChecked = isChecked
    permissionInput.value = isChecked

    return {
      id,
      configInput: permissionInput,
      resource,
      action,
      type: 'selectPermission'
    }
  }

  const collectionsInputs = collectionAction.map((action: any) => (
    <React.Fragment key={`${resource.type}-${action.name}`}>
      {buildCustomInput(getPermissionCheckbox(resource, action))}
    </React.Fragment>
  ))

  const resourcesInput = resourceAction.map((action: any) => {
    const isChecked =
      Object.prototype.hasOwnProperty.call(servicePermissions, resource.type) &&
      servicePermissions?.[resource.type]?.includes(action.name)

    const displayStyle = isChecked ? 'd-block' : 'd-none'

    const isRadioSelected =
      Object.prototype.hasOwnProperty.call(resourcePermissions, resource.type) &&
      Object.prototype.hasOwnProperty.call(resourcePermissions[resource.type], action.name) &&
      resourcePermissions[resource.type][action.name].selectType

    const permissionSelectionRadioCopy = structuredClone(form.permissionSelectionRadio)
    permissionSelectionRadioCopy.value = isRadioSelected === 'select' ? 'select' : 'all'

    const dropdownDisplay = isRadioSelected === 'select' ? 'd-block' : 'd-none'

    const permissionSelectionRadioCopyElement = {
      id: `resourceDropdown-${resource.type}-${action.name}`,
      configInput: permissionSelectionRadioCopy,
      resource,
      action,
      type: 'selectRadio'
    }

    const resourceDropdownCopy = structuredClone(form.resourceDropdown)
    resourceDropdownCopy.hidden = !isChecked

    if (
      Object.prototype.hasOwnProperty.call(resourcesList, resource.type) &&
      resourcesList?.[resource.type]?.length > 0
    ) {
      resourceDropdownCopy.options = resourcesList[resource.type]
    }

    if (isRadioSelected) {
      resourceDropdownCopy.value = resourcePermissions[resource.type][action.name].selectResources
    }

    if (viewMode) {
      const optionIdx = isRadioSelected === 'select' ? 1 : 0
      permissionSelectionRadioCopy.options[optionIdx].disabled = false

      if (isRadioSelected) {
        resourceDropdownCopy.options = resourceDropdownCopy.options.map((option: any) => {
          option.disabled = !resourceDropdownCopy.value.some((data: string) => data === option.value)
          return option
        })
      }
    }

    const resourceDropdownCopyElement = {
      id: `resourceDropdown-${resource.type}-${action.name}`,
      configInput: resourceDropdownCopy,
      resource,
      action,
      type: 'selectDropdown'
    }

    const showLoader =
      resourcesListLoader &&
      Object.prototype.hasOwnProperty.call(resourcesListLoader, resource.type) &&
      Object.prototype.hasOwnProperty.call(resourcesListLoader[resource.type], action.name) &&
      resourcesListLoader[resource.type][action.name]

    const resourceInfo = isOwnCloudAccount ? (
      buildCustomInput(resourceDropdownCopyElement)
    ) : (
      <div className="valid-feedback">
        Please contact your system administrator for further information regarding this permission.
      </div>
    )

    return (
      <React.Fragment key={`${resource.type}-${action.name}`}>
        <div className={'d-flex w-100'}>
          {buildCustomInput(getPermissionCheckbox(resource, action))}
          <div className={`${displayStyle}`}>{buildCustomInput(permissionSelectionRadioCopyElement)}</div>
        </div>
        <div className={`${dropdownDisplay} mb-s3`}>{showLoader ? <Spinner /> : resourceInfo}</div>
      </React.Fragment>
    )
  })

  return (
    <>
      <div className="d-flex flex-column gap-s4 w-100">
        <h4>Collection specific permissions</h4>
        <div className="d-flex w-100">{collectionsInputs}</div>
      </div>
      <div className="d-flex flex-column gap-s4 w-100">
        <h4>Resources specific permissions</h4>
        <div className="d-flex flex-column gap-s3">{resourcesInput}</div>
      </div>
    </>
  )
}

export default PermissionHandler
