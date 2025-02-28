// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import { BsPencilFill, BsXCircle } from 'react-icons/bs'
import useErrorBoundary from '../../hooks/useErrorBoundary'

import AccountRolesDetails from '../../components/profile/accountRoles/AccountRolesDetails'
import useToastStore from '../../store/toastStore/ToastStore'
import AuthorizationService from '../../services/AuthorizationService'
import { useNavigate, useParams } from 'react-router-dom'
import useAppStore from '../../store/appStore/AppStore'
import useAuthorizationStore from '../../store/authorizationStore/AuthorizationStore'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import useBucketStore from '../../store/bucketStore/BucketStore'
import useUserStore from '../../store/userStore/UserStore'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

const AccountRolesDetailsContainer = (): JSX.Element => {
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)

  const actionsOption = [
    {
      id: 'delete',
      name: (
        <>
          <BsXCircle /> Delete{' '}
        </>
      ),
      label: 'Delete role',
      buttonLabel: 'Delete'
    }
  ]

  if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_USER_ROLE_EDIT)) {
    actionsOption.unshift({
      id: 'edit',
      name: (
        <>
          {' '}
          <BsPencilFill /> Edit{' '}
        </>
      ),
      label: 'Edit role',
      buttonLabel: 'Edit'
    })
  }

  const modalContent = {
    action: '',
    label: '',
    buttonLabel: '',
    question: '',
    name: '',
    roleId: ''
  }

  const tabDetailsInitial: any = [
    {
      tapTitle: 'Permission information',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'Users information',
      tapConfig: { type: 'custom' },
      fields: [{ label: 'Assigned Users:', field: 'users', value: '' }]
    }
  ]

  const initialTabs = [
    {
      label: 'Permission',
      id: 'permission'
    },
    {
      label: 'Users',
      id: 'users'
    }
  ]

  const emptyGrid = {
    title: 'No users found',
    subTitle: 'No users associated with this role'
  }

  const initialForm = {
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
      hiddenLabel: true,
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
          value: 'all',
          disabled: true
        },
        {
          name: 'Select Resources',
          value: 'select',
          disabled: true
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

  const [isPageReady, setIsPageReady] = useState(false)
  const [isRoleReady, setIsRoleReady] = useState(false)
  const [activeTab, setActiveTab] = useState<string | number>(0)
  const [tabDetails, setTabDetails] = useState<any>(tabDetailsInitial)
  const [reserveDetails, setReserveDetails] = useState<any>(null)
  const [actionsReserveDetails, setActionsReserveDetails] = useState<any[]>([])
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [servicePermissions, setServicePermissions] = useState<any>({})
  const [resourcePermissions, setResourcePermissions] = useState<any>({})
  const [resourcesReady, setResourcesReady] = useState(false)

  // Global State
  const user = useUserStore((state) => state.user)
  const loading = useAuthorizationStore((state) => state.loading)
  const roles = useAuthorizationStore((state) => state.roles)
  const setRoles = useAuthorizationStore((state) => state.setRoles)
  const userRoles = useAuthorizationStore((state) => state.userRoles)
  const setUserRoles = useAuthorizationStore((state) => state.setUserRoles)
  const addBreadcrumCustomTitle = useAppStore((state) => state.addBreadcrumCustomTitle)
  const resources = useAuthorizationStore((state) => state.resources)
  const setResources = useAuthorizationStore((state) => state.setResources)

  const [resourcesList, setResourcesList] = useState<any>({})
  const instancesOptionsList = useCloudAccountStore((state) => state.setInstancesOptionsList)
  const setStorageOptionsList = useCloudAccountStore((state) => state.setStorageOptionsList)
  const setBucketsOptionsList = useBucketStore((state) => state.setBucketsOptionsList)
  const setBucketUsersOptionsList = useBucketStore((state) => state.setBucketUsersOptionsList)
  const setLifecycleRulesOptionsList = useBucketStore((state) => state.setLifecycleRulesOptionsList)

  const adminInvitation = useUserStore((state) => state.adminInvitation)
  const setInvitations = useUserStore((state) => state.setInvitations)

  const throwError = useErrorBoundary()
  const showError = useToastStore((state) => state.showError)

  // Navigation
  const navigate = useNavigate()
  const { param: roleName } = useParams()

  // Hooks
  useEffect(() => {
    const fetchData = async (): Promise<void> => {
      try {
        if (isOwnCloudAccount) {
          await setInvitations()
          if (roles.length === 0) await setRoles()
        } else {
          if (userRoles.length === 0) await setUserRoles(user?.email as string)
        }
        if (resources.length === 0) await setResources()
        setIsRoleReady(true)
      } catch (error) {
        throwError(error)
      }
    }
    fetchData().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    const callUpdate = async (): Promise<void> => {
      await updateDetails()
    }

    if (isRoleReady) {
      void callUpdate()
    }
  }, [isRoleReady])

  useEffect(() => {
    if (reserveDetails) {
      addBreadcrumCustomTitle(`/profile/roles/d/${reserveDetails.id}`, reserveDetails.alias)
    }
  }, [reserveDetails])

  useEffect(() => {
    if (resourcesReady) {
      setResourceActions()
    }
  }, [resourcesReady])

  // Functions

  const setResourcesOptionsList = (list: any, service: string): void => {
    setResourcesList((oldResourcesList: any) => {
      const newList = {
        [service]: list
      }

      return { ...oldResourcesList, ...newList }
    })
  }

  const getResourcesList = async (service: string): Promise<any> => {
    let res = null

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
    return res
  }

  const updateDetails = async (): Promise<void> => {
    const roleData: any[] = isOwnCloudAccount ? roles : userRoles
    const role: any = roleData.find((role) => role?.alias === roleName)
    if (role === undefined) {
      if (isRoleReady) navigate('/profile/roles')
      setActionsReserveDetails([])
      setReserveDetails(null)
      return
    }

    const tabDetailsCopy = []
    for (const tapIndex in tabDetails) {
      const tapDetail = { ...tabDetails[tapIndex] }
      const updateFields = []
      for (const index in tapDetail.fields) {
        const field = { ...tapDetail.fields[index] }
        field.value = role[field.field]
        updateFields.push(field)
      }
      tapDetail.fields = updateFields
      tabDetailsCopy.push(tapDetail)
    }

    setTabDetails(tabDetailsCopy)
    setActionsReserveDetails(actionsOption)
    setReserveDetails(role)

    const perTypes = role.permissions
      .map((x: any) => x.resourceType)
      .filter((value: string, index: number, array: any) => array.indexOf(value) === index)

    for (const type of perTypes) {
      if (isOwnCloudAccount) await getResourcesList(type)
      setServiceActions(type, role)
    }

    setResourcesReady(true)
  }

  const setServiceActions = (service: string, role: any): void => {
    setServicePermissions((oldList: any) => {
      const uniqueActions = role?.permissions
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
    const roleData: any[] = isOwnCloudAccount ? roles : userRoles
    const role: any = roleData.find((role) => role?.alias === roleName)
    if (role) {
      const perTypes = role.permissions
        .map((x: any) => x.resourceType)
        .filter((value: string, index: number, array: any) => array.indexOf(value) === index)

      const resourcePermissionsCopy: any = {}

      for (const service of perTypes) {
        resourcePermissionsCopy[service] = {}

        const resourcesActions = resources.find((x: any) => x.type === service && x.actions.length > 0)
        const resourceAction = resourcesActions?.actions.filter((x: any) => x.type === 'resource').map((x) => x.name)

        const allActions = role?.permissions
          .filter((x: any) => x.resourceType === service && x.resourceId === '*')[0]
          .actions.filter((x: any) => resourceAction?.includes(x))
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

        const resActions = role?.permissions.filter((x: any) => x.resourceType === service && x.resourceId !== '*')
        const resList = resourcesList[service]
        if (resActions && resActions.length > 0) {
          for (let j = 0; j < resActions.length; j++) {
            const dbResAction = resActions[j]
            const loopActions = dbResAction.actions

            const resResult = resList?.find((x: any) => String(x.value) === String(dbResAction.resourceId))

            if (resResult || !isOwnCloudAccount) {
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

  const deleteRole = async (roleId: string): Promise<void> => {
    try {
      await AuthorizationService.deleteRole(roleId)
      await setRoles()
      setTimeout(() => {
        navigate({ pathname: '/profile/roles' })
      }, 1000)
    } catch (error) {
      throwError(error)
    }
  }

  const actionOnModal = async (result: boolean): Promise<void> => {
    if (result) {
      try {
        switch (actionModalContent.action) {
          case 'delete':
            await deleteRole(actionModalContent.roleId)
            setShowActionModal(false)
            break
          default:
            setShowActionModal(false)
            break
        }
      } catch (error) {
        setShowActionModal(false)
        showError('Unable to perform action', false)
      }
    } else {
      setShowActionModal(result)
    }
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      case 'edit': {
        navigate({
          pathname: `/profile/roles/d/${item.alias}/edit`,
          search: '?backTo=detail'
        })
        break
      }
      case 'delete': {
        const copyModalContent = { ...modalContent }
        copyModalContent.action = action.id
        copyModalContent.roleId = item.id
        copyModalContent.label = action.label
        copyModalContent.buttonLabel = action.buttonLabel
        copyModalContent.name = item.alias
        setActionModalContent(copyModalContent)
        setShowActionModal(true)
        break
      }
      default: {
        break
      }
    }
  }

  return (
    <>
      <AccountRolesDetails
        form={initialForm}
        isPageReady={isPageReady}
        tabs={initialTabs}
        reserveDetails={reserveDetails}
        tabDetails={tabDetails}
        loading={loading}
        activeTab={activeTab}
        actionsReserveDetails={actionsReserveDetails}
        showActionModal={showActionModal}
        actionModalContent={actionModalContent}
        emptyGrid={emptyGrid}
        setActiveTab={setActiveTab}
        setAction={setAction}
        setShowActionModal={actionOnModal}
        resources={resources}
        resourcesList={resourcesList}
        adminInvitation={adminInvitation}
        servicePermissions={servicePermissions}
        resourcePermissions={resourcePermissions}
        isOwnCloudAccount={isOwnCloudAccount}
      />
    </>
  )
}
export default AccountRolesDetailsContainer
