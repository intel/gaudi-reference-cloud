// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useAuthorizationStore from '../../store/authorizationStore/AuthorizationStore'
import AccountRolesCreate from '../../components/profile/accountRoles/AccountRolesCreate'
import { useNavigate } from 'react-router-dom'
import {
  getFormValue,
  isValidForm,
  showFormRequiredFields,
  UpdateFormHelper
} from '../../utils/updateFormHelper/UpdateFormHelper'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import useBucketStore from '../../store/bucketStore/BucketStore'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import useUserStore from '../../store/userStore/UserStore'
import AuthorizationService from '../../services/AuthorizationService'

const getCustomMessage = (messageType) => {
  let message = null

  switch (messageType) {
    case 'name':
      message = (
        <div className="valid-feedback">
          Max length 50 characters. Letters, numbers and ‘- ‘ accepted.
          <br />
          Name should start and end with an alphanumeric character.
        </div>
      )
      break
    case 'resourceDropdown':
      message = (
        <div className="valid-feedback">
          Begin typing to generate a list of resources. Multiple resources can be selected from this list.
        </div>
      )
      break
    default:
      break
  }

  return message
}

const AccountRolesCreateContainer = () => {
  // *****
  // Local Variables
  // *****

  const initialState = {
    mainTitle: 'Create Role and Set Permissions',
    keyId: 'UserRoleCreate',
    navigationBottom: [
      {
        buttonLabel: 'Create',
        buttonVariant: 'primary'
      },
      {
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => onCancel()
      }
    ]
  }

  const initialForm = {
    name: {
      type: 'text', // options = 'text ,'textArea'
      label: 'Role Name:',
      placeholder: 'Name',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 50,
      validationRules: {
        isRequired: true,
        onlyAlphaNumLower: true,
        checkMaxLength: true
      },
      validationMessage: '',
      helperMessage: getCustomMessage('name')
    },
    selectAllCheckbox: {
      type: 'checkbox', // options = 'text ,'textArea'
      label: '',
      placeholder: '',
      value: false, // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      isChecked: false,
      validationRules: {
        isRequired: false
      },
      options: [
        {
          name: 'Select All',
          value: false,
          defaultChecked: false
        }
      ]
    },
    permissionCheckbox: {
      type: 'checkbox', // options = 'text ,'textArea'
      label: '',
      placeholder: '',
      value: false, // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      isChecked: false,
      validationRules: {
        isRequired: false
      },
      options: [
        {
          name: 'Permission',
          value: false,
          defaultChecked: false
        }
      ],
      maxWidth: '20rem'
    },
    permissionSelectionRadio: {
      type: 'radio', // options = 'text ,'textArea'
      label: '',
      placeholder: '',
      value: 'all', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      isChecked: false,
      radioGroupHorizontal: true,
      validationRules: {
        isRequired: false
      },
      options: [
        {
          name: 'Select All',
          value: 'all'
        },
        {
          name: 'Select Resources',
          value: 'select'
        }
      ],
      maxWidth: '15rem'
    },
    resourceDropdown: {
      type: 'multi-select-dropdown', // options = 'text ,'textArea'
      label: '',
      placeholder: 'Select Resource',
      value: [],
      isValid: true,
      isTouched: false,
      isReadOnly: false,
      validationRules: {
        isRequired: false
      },
      options: [],
      validationMessage: '',
      helperMessage: '',
      hiddenLabel: true,
      hidden: true,
      emptyOptionsMessage: 'No resource found.'
    }
  }

  const initialErrorModal = {
    showErrorModal: false,
    errorMessage: '',
    errorHideRetryMessage: null,
    errorDescription: null,
    errorTitleMessage: 'Could not create your role'
  }

  // *****
  // Local State
  // *****

  const [state] = useState(initialState)
  const [form, setForm] = useState(initialForm)
  const [errorModal, setErrorModal] = useState(initialErrorModal)
  const [showReservationModal, setShowReservationModal] = useState(false)
  const [isPageReady, setIsPageReady] = useState(false)
  const [isFormValid, setIsFormValid] = useState(false)

  const [servicePermissions, setServicePermissions] = useState({})
  const [resourcePermissions, setResourcePermissions] = useState({})
  const [resourcesList, setResourcesList] = useState({})
  const [resourcesListLoader, setResourcesListLoader] = useState({})

  // *****
  // Global State
  // *****
  const throwError = useErrorBoundary()
  const navigate = useNavigate()

  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  const resources = useAuthorizationStore((state) => state.resources)
  const loading = useAuthorizationStore((state) => state.loading)
  const setResources = useAuthorizationStore((state) => state.setResources)
  const cloudAccountNumber = useUserStore((state) => state.user.cloudAccountNumber)
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)

  const instancesOptionsList = useCloudAccountStore((state) => state.setInstancesOptionsList)
  const setStorageOptionsList = useCloudAccountStore((state) => state.setStorageOptionsList)
  const setBucketsOptionsList = useBucketStore((state) => state.setBucketsOptionsList)
  const setBucketUsersOptionsList = useBucketStore((state) => state.setBucketUsersOptionsList)
  const setLifecycleRulesOptionsList = useBucketStore((state) => state.setLifecycleRulesOptionsList)

  // *****
  // Hooks
  // *****

  useEffect(() => {
    const fetch = async () => {
      try {
        await setResources()
        setIsPageReady(true)
      } catch (error) {
        throwError(error)
      }
    }
    if (resources.length === 0) {
      fetch().catch((error) => {
        throwError(error)
      })
    }
  }, [])

  // *****
  // Functions
  // *****

  const setResourcesOptionsList = (list, service) => {
    const resourcesListCopy = structuredClone(resourcesList)
    resourcesListCopy[service] = list
    setResourcesList(resourcesListCopy)
  }

  const onCancel = () => {
    // Navigates back to the page when this method triggers.
    navigate({
      pathname: '/profile/roles'
    })
  }

  function onClickCloseErrorModal() {
    const errorModalCopy = { ...errorModal }

    errorModalCopy.showErrorModal = false

    setErrorModal(errorModalCopy)
  }

  function onChangeInput(event, element) {
    if (element.type === 'selectAll') {
      onSelectAll(event, element)
    } else if (element.type === 'selectPermission') {
      onCheckboxChange(event, element)
    } else if (element.type === 'selectRadio') {
      onRadioSelect(event, element)
    } else {
      const value = event.target.value
      const formCopy = { ...form }
      const updatedForm = UpdateFormHelper(value, element.id, formCopy)
      setIsFormValid(isValidForm(updatedForm))
      setForm(updatedForm)
    }
  }

  const onCheckboxChange = (event, element) => {
    const service = element.resource.type
    const permission = element.action.name
    const checked = event.target.checked

    const servicePermissionsCopy = structuredClone(servicePermissions)

    if (!Object.prototype.hasOwnProperty.call(servicePermissionsCopy, service)) {
      servicePermissionsCopy[service] = []
    }

    if (!checked) {
      const index = servicePermissionsCopy[service].indexOf(permission, 0)
      if (index > -1) {
        servicePermissionsCopy[service].splice(index, 1)
      }
    } else {
      servicePermissionsCopy[service].push(permission)
    }

    setServicePermissions(servicePermissionsCopy)
  }

  const onSelectAll = (event, element) => {
    const checked = event.target.checked
    const service = element.resource.type

    const servicePermissionsCopy = structuredClone(servicePermissions)

    if (!Object.prototype.hasOwnProperty.call(servicePermissionsCopy, service)) {
      servicePermissionsCopy[service] = []
    }

    if (!checked) {
      servicePermissionsCopy[service] = []
    } else {
      servicePermissionsCopy[service] = element.resource.actions.map((x) => x.name)
    }

    setServicePermissions(servicePermissionsCopy)
  }

  const onRadioSelect = (event, element) => {
    const service = element.resource.type
    const permission = element.action.name
    const value = event.target.value

    const resourcePermissionsCopy = structuredClone(resourcePermissions)

    if (!Object.prototype.hasOwnProperty.call(resourcePermissionsCopy, service)) {
      resourcePermissionsCopy[service] = {}
    }

    if (!Object.prototype.hasOwnProperty.call(resourcePermissionsCopy[service], permission)) {
      resourcePermissionsCopy[service][permission] = {
        selectType: '',
        selectResources: []
      }
    }

    resourcePermissionsCopy[service][permission].selectType = value

    if (value === 'select') {
      if (!Object.prototype.hasOwnProperty.call(resourcesList, service) || resourcesList[service].length === 0) {
        getResourcesList(element)
      }
    }

    setResourcePermissions(resourcePermissionsCopy)
  }

  const getResourcesList = async (element) => {
    const service = element.resource.type
    const permission = element.action.name

    const resourcesListLoaderCopy = structuredClone(resourcesListLoader)

    if (!Object.prototype.hasOwnProperty.call(resourcesListLoaderCopy, service)) {
      resourcesListLoaderCopy[service] = {}
    }

    if (!Object.prototype.hasOwnProperty.call(resourcesListLoaderCopy[service], permission)) {
      resourcesListLoaderCopy[service][permission] = true
    }

    setResourcesListLoader(resourcesListLoaderCopy)

    try {
      let res = null

      if (service === 'instance') {
        res = await instancesOptionsList()
      }

      if (service === 'objectstorage') {
        res = await setBucketsOptionsList()
      }

      if (service === 'principal') {
        res = await setBucketUsersOptionsList()
      }

      if (service === 'lifecyclerule') {
        res = await setLifecycleRulesOptionsList()
      }

      if (service === 'filestorage') {
        res = await setStorageOptionsList()
      }

      if (res) setResourcesOptionsList(res, service)
    } catch (error) {
      throwError(error)
    }
    resourcesListLoaderCopy[service][permission] = false
    setResourcesListLoader(resourcesListLoaderCopy)
  }

  const showRequiredFields = () => {
    const formCopy = { ...form }

    const updatedForm = showFormRequiredFields(formCopy)
    showError(toastMessageEnum.formValidationError, false)

    setForm(updatedForm)
  }

  const getServicePayload = () => {
    const permissions = []

    for (const type in servicePermissions) {
      const res = servicePermissions[type]

      if (res.length > 0) {
        const dbActions = resources.find((x) => x.type === type).actions
        const colActions = dbActions.filter((x) => x.type === 'collection').map((x) => x.name)

        let collectionActions = []
        const resourceActions = {}

        if (!Object.prototype.hasOwnProperty.call(resourcePermissions, type)) {
          collectionActions = res
        } else {
          for (const act of res) {
            if (colActions.includes(act)) {
              collectionActions.push(act)
            } else if (Object.prototype.hasOwnProperty.call(resourcePermissions[type], act)) {
              if (resourcePermissions[type][act].selectType === 'all') {
                collectionActions.push(act)
              } else if (resourcePermissions[type][act].selectResources.length > 0) {
                for (const resourceId of resourcePermissions[type][act].selectResources) {
                  if (!Object.prototype.hasOwnProperty.call(resourceActions, resourceId)) {
                    resourceActions[resourceId] = [act]
                  } else {
                    resourceActions[resourceId].push(act)
                  }
                }
              }
            } else {
              collectionActions.push(act)
            }
          }
        }

        permissions.push({
          resourceType: type,
          resourceId: '*',
          actions: collectionActions
        })

        for (const resourceId in resourceActions) {
          const resPer = resourceActions[resourceId]
          permissions.push({
            resourceType: type,
            resourceId,
            actions: resPer
          })
        }
      }
    }

    return permissions
  }

  const onSubmit = () => {
    if (!isFormValid) {
      showRequiredFields()
      return
    }

    submitForm()
  }

  const submitForm = async () => {
    setShowReservationModal(true)

    try {
      const roleName = getFormValue('name', form)
      const permissions = getServicePayload()

      const servicePayload = {
        alias: roleName,
        cloudAccountId: cloudAccountNumber,
        effect: 'allow',
        permissions
      }
      await AuthorizationService.createCloudAccountRoles(servicePayload)
      showSuccess('Role created successfully.', false)
      navigate({
        pathname: '/profile/roles'
      })
    } catch (error) {
      const errorModalUpdated = { ...errorModal }

      errorModalUpdated.errorHideRetryMessage = false
      errorModalUpdated.errorDescription = null
      errorModalUpdated.showErrorModal = true
      errorModalUpdated.errorMessage = error.response ? error.response.data.message : error.message

      setErrorModal(errorModalUpdated)
    } finally {
      setShowReservationModal(false)
    }
  }

  function onChangeDropdownMultiple(values, element) {
    const service = element.resource.type
    const permission = element.action.name

    const resourcePermissionsCopy = structuredClone(resourcePermissions)
    resourcePermissionsCopy[service][permission].selectResources = values

    setResourcePermissions(resourcePermissionsCopy)
  }

  return (
    <>
      <AccountRolesCreate
        isPageReady={isPageReady}
        loading={loading}
        state={state}
        form={form}
        errorModal={errorModal}
        showReservationModal={showReservationModal}
        onSubmit={onSubmit}
        onChangeInput={onChangeInput}
        resources={resources}
        servicePermissions={servicePermissions}
        resourcePermissions={resourcePermissions}
        resourcesList={resourcesList}
        resourcesListLoader={resourcesListLoader}
        isOwnCloudAccount={isOwnCloudAccount}
        onChangeDropdownMultiple={onChangeDropdownMultiple}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
    </>
  )
}
export default AccountRolesCreateContainer
