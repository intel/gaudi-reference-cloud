// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useAuthorizationStore, { type RoleDefinition } from '../../store/authorizationStore/AuthorizationStore'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import {
  isValidForm,
  setFormValue,
  showFormRequiredFields,
  UpdateFormHelper
} from '../../utils/updateFormHelper/UpdateFormHelper'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import useBucketStore from '../../store/bucketStore/BucketStore'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import useUserStore from '../../store/userStore/UserStore'
import AuthorizationService from '../../services/AuthorizationService'
import AccountRolesEdit from '../../components/profile/accountRoles/AccountRolesEdit'

const AccountRolesEditContainer = (): JSX.Element => {
  const { param: paramRoleName } = useParams()
  const [searchParams] = useSearchParams()
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)
  // *****
  // Local Variables
  // *****

  const initialState = {
    mainTitle: `Update Role and Set Permissions for ${paramRoleName}`,
    keyId: 'UserRoleUpdate',
    navigationBottom: [
      {
        buttonLabel: 'Update',
        buttonVariant: 'primary'
      },
      {
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => {
          onCancel()
        }
      }
    ]
  }

  const initialForm = {
    name: {
      type: 'text', // options = 'text ,'textArea'
      label: 'Role Name:',
      placeholder: 'Name',
      value: '', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: true, // Input create as read only
      maxLength: 50,
      validationRules: {
        isRequired: true,
        onlyAlphaNumLower: true,
        checkMaxLength: true
      },
      validationMessage: '',
      helperMessage: ''
    },
    selectAllCheckbox: {
      type: 'checkbox', // options = 'text ,'textArea'
      label: 'selectAllCheckbox',
      placeholder: '',
      value: false, // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      isChecked: false,
      hiddenLabel: true,
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
      label: 'permissionCheckbox',
      placeholder: '',
      value: false, // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      isChecked: false,
      hiddenLabel: true,
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
      label: 'permissionSelectionRadio',
      placeholder: '',
      value: 'all', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      isChecked: false,
      radioGroupHorizontal: true,
      hiddenLabel: true,
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
      label: 'resourceDropdown',
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
    errorTitleMessage: 'Could not update your role'
  }

  // *****
  // Local State
  // *****

  const [state] = useState(initialState)
  const [form, setForm] = useState(initialForm)
  const [errorModal, setErrorModal] = useState(initialErrorModal)
  const [showReservationModal, setShowReservationModal] = useState(false)
  const [isPageReady, setIsPageReady] = useState(false)
  const [isRoleReady, setIsRoleReady] = useState(false)
  const [resourcesReady, setResourcesReady] = useState(false)
  const [isFormValid, setIsFormValid] = useState(true)

  const [roleToEdit, setRoleToEdit] = useState<RoleDefinition | undefined>(undefined)

  const [servicePermissions, setServicePermissions] = useState<any>({})
  const [resourcePermissions, setResourcePermissions] = useState<any>({})
  const [resourcesList, setResourcesList] = useState<any>({})
  const [resourcesListLoader, setResourcesListLoader] = useState<any>({})

  // *****
  // Global State
  // *****
  const throwError = useErrorBoundary()
  const navigate = useNavigate()

  const showSuccess = useToastStore((state) => state.showSuccess)
  const showError = useToastStore((state) => state.showError)

  const cloudAccountNumber = useUserStore((state) => state?.user?.cloudAccountNumber)

  const loading = useAuthorizationStore((state) => state.loading)
  const roles = useAuthorizationStore((state) => state.roles)
  const setRoles = useAuthorizationStore((state) => state.setRoles)
  const resources = useAuthorizationStore((state) => state.resources)
  const setResources = useAuthorizationStore((state) => state.setResources)

  const instancesOptionsList = useCloudAccountStore((state) => state.setInstancesOptionsList)
  const setStorageOptionsList = useCloudAccountStore((state) => state.setStorageOptionsList)
  const setBucketsOptionsList = useBucketStore((state) => state.setBucketsOptionsList)
  const setBucketUsersOptionsList = useBucketStore((state) => state.setBucketUsersOptionsList)
  const setLifecycleRulesOptionsList = useBucketStore((state) => state.setLifecycleRulesOptionsList)

  // *****
  // Hooks
  // *****

  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        if (roles.length === 0) await setRoles()
        if (resources.length === 0) await setResources()
        setIsRoleReady(true)
      } catch (error) {
        throwError(error)
      }
    }

    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    let shouldExit = false
    if (isRoleReady) {
      if (roles.length > 0) {
        const roleToEdit = roles?.find((item) => item.alias === paramRoleName)
        if (roleToEdit !== undefined) {
          setRoleToEdit(roleToEdit)
        } else {
          shouldExit = true
        }
      } else {
        shouldExit = true
      }
    }

    if (shouldExit) {
      navigate({
        pathname: '/profile/roles'
      })
    }
  }, [isRoleReady])

  useEffect(() => {
    const callUpdate = async (): Promise<void> => {
      await updateDetails()
    }

    if (roleToEdit) {
      void callUpdate()
    }
  }, [roleToEdit])

  useEffect(() => {
    if (resourcesReady) {
      setResourceActions()
    }
  }, [resourcesReady])

  // *****
  // Functions
  // *****

  const updateDetails = async (): Promise<void> => {
    if (roleToEdit) {
      const updateForm = { ...form }
      const updatedForm = setFormValue('name', roleToEdit.alias, updateForm)
      setForm(updatedForm)

      const perTypes = roleToEdit.permissions
        .map((x: any) => x.resourceType)
        .filter((value: string, index: number, array: any) => array.indexOf(value) === index)

      for (const type of perTypes) {
        await getResourcesList(type)
        setServiceActions(type)
      }

      setResourcesReady(true)
    }
  }

  const getResourcesList = async (service: string, permission: string = ''): Promise<void> => {
    let res = null

    const resourcesListLoaderCopy = structuredClone(resourcesListLoader)
    if (permission) {
      if (!Object.prototype.hasOwnProperty.call(resourcesListLoaderCopy, service)) {
        resourcesListLoaderCopy[service] = {}
      }

      if (!Object.prototype.hasOwnProperty.call(resourcesListLoaderCopy[service], permission)) {
        resourcesListLoaderCopy[service][permission] = true
      }

      setResourcesListLoader(resourcesListLoaderCopy)
    }

    try {
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
      setResourcesOptionsList(res, service)
    } catch (error) {
      throwError(error)
    }

    if (permission) {
      resourcesListLoaderCopy[service][permission] = false
      setResourcesListLoader(resourcesListLoaderCopy)
    }
  }

  const setResourcesOptionsList = (list: any, service: string): void => {
    setResourcesList((oldResourcesList: any) => {
      const newList = {
        [service]: list
      }

      return { ...oldResourcesList, ...newList }
    })
  }

  const setServiceActions = (service: string): void => {
    setServicePermissions((oldList: any) => {
      const uniqueActions = roleToEdit?.permissions
        .filter((x: any) => x.resourceType === service)
        .map((x: any) => x.actions)
        .flat()
        .filter((value: string, index: number, array: any) => array.indexOf(value) === index)

      const newList = {
        [service]: uniqueActions
      }

      return { ...oldList, ...newList }
    })
  }

  const setResourceActions = (): void => {
    if (roleToEdit) {
      const perTypes = roleToEdit.permissions
        .map((x: any) => x.resourceType)
        .filter((value: string, index: number, array: any) => array.indexOf(value) === index)

      const resourcePermissionsCopy = structuredClone(resourcePermissions)

      for (const service of perTypes) {
        resourcePermissionsCopy[service] = {}

        const resourcesActions = resources.find((x: any) => x.type === service && x.actions.length > 0)
        const resourceAction = resourcesActions?.actions.filter((x: any) => x.type === 'resource').map((x) => x.name)

        const allActions = roleToEdit?.permissions
          .filter((x: any) => x.resourceType === service && x.resourceId === '*')[0]
          .actions.filter((x) => resourceAction?.includes(x))
          .map((x: any) => x.name)

        if (allActions) {
          for (let i = 0; i < allActions.length; i++) {
            const act = allActions[i]
            resourcePermissionsCopy[service][act] = {
              selectType: 'all',
              selectResources: []
            }
          }
        }

        const resActions = roleToEdit?.permissions.filter(
          (x: any) => x.resourceType === service && x.resourceId !== '*'
        )
        const resList = resourcesList[service]

        if (resActions && resActions.length > 0) {
          for (let j = 0; j < resActions.length; j++) {
            const dbResAction = resActions[j]
            const loopActions = dbResAction.actions

            const resResult = resList.find((x: any) => String(x.value) === String(dbResAction.resourceId))

            if (resResult) {
              for (const la of loopActions) {
                if (Object.prototype.hasOwnProperty.call(resourcePermissionsCopy[service], la)) {
                  resourcePermissionsCopy[service][la].selectResources.push(dbResAction.resourceId)
                } else {
                  resourcePermissionsCopy[service][la] = {
                    selectType: 'select',
                    selectResources: [dbResAction.resourceId]
                  }
                }
              }
            }
          }
        }
      }
      setResourcePermissions(resourcePermissionsCopy)
      setIsPageReady(true)
    }
  }

  const goBack = (): void => {
    const backTo = searchParams.get('backTo')
    switch (backTo) {
      case 'grid':
        navigate({
          pathname: '/profile/roles'
        })
        break
      default:
        navigate({
          pathname: `/profile/roles/d/${paramRoleName}`
        })
        break
    }
  }

  const onCancel = (): void => {
    // Navigates back to the page when this method triggers.
    goBack()
  }

  const onClickCloseErrorModal = (): void => {
    const errorModalCopy = { ...errorModal }

    errorModalCopy.showErrorModal = false

    setErrorModal(errorModalCopy)
  }

  const onChangeInput = (event: any, element: any): void => {
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

  const onCheckboxChange = (event: any, element: any): void => {
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

  const onSelectAll = (event: any, element: any): void => {
    const checked = event.target.checked
    const service = element.resource.type

    const servicePermissionsCopy = structuredClone(servicePermissions)

    if (!Object.prototype.hasOwnProperty.call(servicePermissionsCopy, service)) {
      servicePermissionsCopy[service] = []
    }

    if (!checked) {
      servicePermissionsCopy[service] = []
    } else {
      servicePermissionsCopy[service] = element.resource.actions.map((x: any) => x.name)
    }

    setServicePermissions(servicePermissionsCopy)
  }

  const onRadioSelect = (event: any, element: any): void => {
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
      if (!Object.prototype.hasOwnProperty.call(resourcesList, service) || resourcesList?.[service]?.length === 0) {
        void getResourcesList(service, permission)
      }
    }

    setResourcePermissions(resourcePermissionsCopy)
  }

  const showRequiredFields = (): void => {
    const formCopy = { ...form }

    const updatedForm = showFormRequiredFields(formCopy)
    showError(toastMessageEnum.formValidationError, false)

    setForm(updatedForm)
  }

  const getServicePayload = (): any => {
    const permissions = []

    for (const type in servicePermissions) {
      const res = servicePermissions[type]

      if (res.length > 0) {
        const dbActions = resources.find((x) => x.type === type)?.actions
        const colActions = dbActions ? dbActions.filter((x) => x.type === 'collection').map((x) => x.name) : []

        let collectionActions = []
        const resourceActions: any = {}

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

  const onSubmit = (): void => {
    if (!isFormValid) {
      showRequiredFields()
      return
    }

    void submitForm()
  }

  const submitForm = async (): Promise<void> => {
    setShowReservationModal(true)

    try {
      const permissions = getServicePayload()

      const servicePayload = {
        cloudAccountId: cloudAccountNumber,
        alias: roleToEdit?.alias,
        effect: 'allow',
        permissions
      }
      await AuthorizationService.updateCloudAccountRoles(roleToEdit?.id, servicePayload)
      navigate({
        pathname: '/profile/roles'
      })
      showSuccess('Role updated successfully', false)
    } catch (error: any) {
      const errorModalUpdated = { ...errorModal }

      errorModalUpdated.errorDescription = null
      errorModalUpdated.showErrorModal = true
      errorModalUpdated.errorMessage = error.response ? error.response.data.message : error.message

      setErrorModal(errorModalUpdated)
    } finally {
      setShowReservationModal(false)
    }
  }

  const onChangeDropdownMultiple = (values: any, element: any): void => {
    const service = element.resource.type
    const permission = element.action.name

    const resourcePermissionsCopy = structuredClone(resourcePermissions)
    resourcePermissionsCopy[service][permission].selectResources = values

    setResourcePermissions(resourcePermissionsCopy)
  }

  return (
    <>
      <AccountRolesEdit
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
        onChangeDropdownMultiple={onChangeDropdownMultiple}
        onClickCloseErrorModal={onClickCloseErrorModal}
        resourcesListLoader={resourcesListLoader}
        isOwnCloudAccount={isOwnCloudAccount}
      />
    </>
  )
}
export default AccountRolesEditContainer
