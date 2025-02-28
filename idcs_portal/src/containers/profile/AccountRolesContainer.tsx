// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { BsPencilFill, BsXCircle } from 'react-icons/bs'

import useAuthorizationStore from '../../store/authorizationStore/AuthorizationStore'
import useToastStore from '../../store/toastStore/ToastStore'
import AuthorizationService from '../../services/AuthorizationService'
import { useNavigate } from 'react-router-dom'
import AccountRoles from '../../components/profile/accountRoles/AccountRoles'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import useUserStore from '../../store/userStore/UserStore'

const AccountRolesContainer = (): JSX.Element => {
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)

  // *****
  // local state
  // *****
  const launchPagePath = '/profile/roles/reserve'

  const emptyGrid: any = {
    title: 'No roles found',
    subTitle: isOwnCloudAccount
      ? 'Create a role to manage permissions for your team members.'
      : 'Please contact your system administrator for further information regarding this permission.'
  }

  const emptyGridByFilter = {
    title: 'No roles found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => {
        setFilter('', true)
      },
      label: 'Clear filters'
    }
  }

  const columns: any[] = [
    {
      columnName: 'Roles',
      targetColumn: 'roles',
      columnConfig: {
        behaviorType: 'hyperLink',
        behaviorFunction: 'setDetails'
      }
    }
  ]

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

  if (isOwnCloudAccount) {
    columns.push({
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'Buttons',
        behaviorFunction: null
      }
    })

    emptyGrid.action = {
      type: 'redirect',
      href: launchPagePath,
      label: 'Create new role'
    }
  }

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

  // *****
  // State Declaration
  // *****
  const [myReservations, setMyReservations] = useState<any[] | null>(null)
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState<any>(emptyGrid)
  const [isPageReady, setIsPageReady] = useState(false)
  // *****
  // Global State
  // *****
  const navigate = useNavigate()
  const throwError = useErrorBoundary()
  const roles = useAuthorizationStore((state) => state.roles)
  const userRoles = useAuthorizationStore((state) => state.userRoles)
  const loading = useAuthorizationStore((state) => state.loading)
  const user = useUserStore((state) => state.user)
  const setRoles = useAuthorizationStore((state) => state.setRoles)
  const setUserRoles = useAuthorizationStore((state) => state.setUserRoles)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)
  // *****
  // Hooks
  // *****
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        if (isOwnCloudAccount) await setRoles()
        else await setUserRoles(user?.email as string)
        setIsPageReady(true)
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    if (!isPageReady) {
      return
    }
    setGridInfo()
  }, [roles, isPageReady])

  // *****
  // Functions
  // *****
  const setGridInfo = (): void => {
    const gridData: any[] = isOwnCloudAccount ? roles : userRoles
    const gridInfo: any[] = []
    for (const role of gridData) {
      const tableColumn: any = {
        roles: {
          showField: true,
          type: 'HyperLink',
          value: role.alias,
          function: () => {
            setDetails(role.alias)
          }
        }
      }
      if (isOwnCloudAccount) {
        tableColumn.actions = {
          showField: true,
          type: 'Buttons',
          value: role,
          selectableValues: actionsOption,
          function: setAction
        }
      }
      gridInfo.push(tableColumn)
    }
    setMyReservations(gridInfo)
  }

  const setDetails = (roleName: string | null = null): void => {
    if (roleName) navigate(`/profile/roles/d/${roleName}`)
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      case 'edit': {
        navigate({
          pathname: `/profile/roles/d/${item.alias}/edit`,
          search: '?backTo=grid'
        })
        break
      }
      case 'delete': {
        const copyModalContent = { ...modalContent }
        copyModalContent.action = action.id
        copyModalContent.roleId = item.id || item.cloudAccountRoleId
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

  const deleteRole = async (roleId: string): Promise<void> => {
    try {
      await AuthorizationService.deleteRole(roleId)
      await setRoles()
      showSuccess('Role has been removed.', false)
    } catch (error) {
      throwError(error)
    }
  }

  const setFilter = (event: any, clear: boolean): void => {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  return (
    <AccountRoles
      userRoles={myReservations ?? []}
      loading={loading || myReservations === null}
      emptyGrid={emptyGridObject}
      columns={columns}
      launchPagePath={launchPagePath}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      filterText={filterText}
      isOwnCloudAccount={isOwnCloudAccount}
      setFilter={setFilter}
      actionOnModal={actionOnModal}
    />
  )
}

export default AccountRolesContainer
