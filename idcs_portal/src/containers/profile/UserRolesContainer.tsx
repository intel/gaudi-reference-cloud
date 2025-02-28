// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import UserRoles from '../../components/profile/accountRoles/UserRoles'
import { useNavigate, useParams } from 'react-router-dom'
import { BsXCircle } from 'react-icons/bs'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import AuthorizationService from '../../services/AuthorizationService'
import {
  UpdateFormHelper,
  getFormValue,
  setFormValue,
  setSelectOptions,
  showFormRequiredFields,
  isValidForm
} from '../../utils/updateFormHelper/UpdateFormHelper'
import useAuthorizationStore from '../../store/authorizationStore/AuthorizationStore'
import useUserStore from '../../store/userStore/UserStore'
import { toastMessageEnum } from '../../utils/Enums'

const UserRolesContainer = (): JSX.Element => {
  // Navigation and search params
  const navigate = useNavigate()
  const { param: user } = useParams()

  // local state

  const emptyGrid = {
    title: 'No assigned user roles found',
    subTitle: 'Assign a role to the user to set the appropriate permissions.'
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

  const columns = [
    {
      columnName: 'Roles',
      targetColumn: 'roles',
      columnConfig: {
        behaviorType: 'hyperLink',
        behaviorFunction: 'setDetails'
      }
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'Buttons',
        behaviorFunction: null
      }
    }
  ]

  const actionsOption = [
    {
      id: 'remove',
      name: (
        <>
          {' '}
          <BsXCircle /> Remove{' '}
        </>
      ),
      label: 'Remove Assigned Role',
      question: 'Do you want to remove the role $<Role> from the user $<User>?',
      buttonLabel: 'Remove'
    }
  ]

  const modalContent = {
    action: '',
    label: '',
    buttonLabel: '',
    question: '',
    name: '',
    roleId: ''
  }

  const initialForm = {
    role: {
      type: 'multi-select-dropdown', // options = 'text ,'textArea'
      label: 'Select role',
      placeholder: 'Please select role',
      value: [],
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      hiddenLabel: true,
      validationRules: {
        isRequired: true
      },
      options: [],
      validationMessage: '',
      maxWidth: '30rem',
      emptyOptionsMessage: 'No role found.'
    }
  }

  const initialErrorModal = {
    showErrorModal: false,
    errorMessage: '',
    errorHideRetryMessage: null,
    errorDescription: null,
    errorTitleMessage: 'Could not complete your request.'
  }

  const allowedStatus = ['INVITE_STATE_ACCEPTED']

  // Global States
  const throwError = useErrorBoundary()
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  const loading = useAuthorizationStore((state) => state.loading)
  const roles = useAuthorizationStore((state) => state.roles)
  const setRoles = useAuthorizationStore((state) => state.setRoles)
  const userRoles = useAuthorizationStore((state) => state.userRoles)
  const setUserRoles = useAuthorizationStore((state) => state.setUserRoles)
  const adminInvitation = useUserStore((state) => state.adminInvitation)
  const setInvitations = useUserStore((state) => state.setInvitations)

  // State declaration
  const [isPageReady, setIsPageReady] = useState(false)
  const [showActionModal, setShowActionModal] = useState(false)
  const [isAssignRole, setIsAssignRole] = useState(false)
  const [isRemoveRole, setIsRemoveRole] = useState(false)

  const [myReservations, setMyReservations] = useState<any[]>([])
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [form, setForm] = useState(initialForm)
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [errorModal, setErrorModal] = useState(initialErrorModal)

  // Hooks
  useEffect(() => {
    if (!user) navigate({ pathname: '/profile/accountAccessManagement' })

    const fetchData = async (): Promise<void> => {
      try {
        const promises = [setUserRoles(user as string), setInvitations()]
        if (roles.length === 0) promises.push(setRoles())
        await Promise.all(promises)
        setIsPageReady(true)
      } catch (error) {
        throwError(error)
      }
    }

    fetchData().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    if (isPageReady) {
      updateForm()
    }
  }, [isPageReady, userRoles])

  useEffect(() => {
    if (userRoles) {
      setGridInfo()
    }
  }, [userRoles])

  // *****
  // Functions
  // *****

  const updateForm = (): void => {
    const acceptedInvites = adminInvitation
      .filter((x: any) => allowedStatus.includes(x.invitation_state))
      .map((x: any) => x.member_email)

    if (!acceptedInvites.includes(user)) {
      navigate({ pathname: '/profile/accountAccessManagement' })
    }

    const formattedRoles = formatRoleOptions(roles)
    const updatedForm = setSelectOptions('role', formattedRoles, form)
    setForm(updatedForm)
  }

  const setGridInfo = (): void => {
    const gridInfo = []

    if (userRoles && userRoles.length > 0) {
      for (const role of userRoles) {
        gridInfo.push({
          roles: {
            showField: true,
            type: 'HyperLink',
            value: role.alias,
            function: () => {
              setDetails(role.alias)
            }
          },
          actions: {
            showField: true,
            type: 'Buttons',
            value: role,
            selectableValues: actionsOption,
            function: setAction
          }
        })
      }
    }

    setMyReservations(gridInfo)
  }

  const setDetails = (roleName: string | null = null): void => {
    if (roleName) navigate(`/profile/roles/d/${roleName}`)
  }

  const formatRoleOptions = (roles: any): any[] => {
    const unAssignedRoles = roles.filter(
      (role: any) => !userRoles.some((assignedRole) => assignedRole.cloudAccountRoleId === role.id)
    )
    const options = []
    for (const role of unAssignedRoles) {
      const option = { name: role.alias, value: role.id }
      options.push(option)
    }
    return options
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      case 'remove': {
        const copyModalContent = { ...modalContent }
        copyModalContent.action = action.id
        copyModalContent.roleId = item.cloudAccountRoleId
        copyModalContent.label = action.label
        copyModalContent.buttonLabel = action.buttonLabel
        const question = action.question.replace('$<Role>', item.alias).replace('$<User>', user)
        copyModalContent.question = question
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
          case 'remove':
            setShowActionModal(false)
            await removeUserFromRole(actionModalContent.roleId, user)
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

  const removeUserFromRole = async (roleId: string, userId: string | undefined): Promise<void> => {
    setIsRemoveRole(true)
    try {
      await AuthorizationService.removeUserFromRole(roleId, userId)
      showSuccess('Role has been removed.', false)
      setIsRemoveRole(false)
      await setUserRoles(user as string)
    } catch (error: any) {
      setIsRemoveRole(false)
      const errorModalUpdated = { ...errorModal }
      errorModalUpdated.errorDescription = null
      errorModalUpdated.showErrorModal = true
      errorModalUpdated.errorMessage = error.response ? error.response.data.message : error.message
      setErrorModal(errorModalUpdated)
    }
  }

  const resetSelectedRoles = (): void => {
    const updatedForm = setFormValue('role', [], form)
    updatedForm.role.isValid = false
    updatedForm.role.isTouched = false
    setForm(updatedForm)
  }

  function showRequiredFields(): void {
    let updatedForm = { ...form }
    // Mark regular Inputs
    updatedForm = showFormRequiredFields(updatedForm)
    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    setForm(updatedForm)
  }

  const addUserToRole = async (event: any): Promise<void> => {
    event.preventDefault()
    setIsAssignRole(true)
    const validForm = isValidForm(form)
    if (!validForm) {
      setIsAssignRole(false)
      showRequiredFields()
      return
    }
    try {
      const userId = user
      const payload = {
        cloudAccountRoleIds: getFormValue('role', form)
      }
      await AuthorizationService.addRolesToUser(userId, payload)
      setIsAssignRole(false)
      resetSelectedRoles()
      showSuccess('Role has been added.', false)
      await setUserRoles(user as string)
    } catch (error: any) {
      setIsAssignRole(false)
      const errorModalUpdated = { ...errorModal }
      errorModalUpdated.errorDescription = null
      errorModalUpdated.showErrorModal = true
      errorModalUpdated.errorMessage = error.response ? error.response.data.message : error.message
      setErrorModal(errorModalUpdated)
    }
  }

  const onChangeInput = (event: any, formInputName: string): void => {
    const value = event.target.value
    let updatedForm = {
      ...form
    }

    updatedForm = UpdateFormHelper(value, formInputName, updatedForm)

    setForm(updatedForm)
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

  const onChangeDropdownMultiple = (values: string[] | [], formInputName: string): void => {
    let updatedForm = { ...form }

    updatedForm = UpdateFormHelper(values, formInputName, updatedForm)
    setForm(updatedForm)
  }

  const onClickCloseErrorModal = (): void => {
    const errorModalCopy = { ...errorModal }

    errorModalCopy.showErrorModal = false

    setErrorModal(errorModalCopy)
  }

  return (
    <>
      <UserRoles
        columns={columns}
        isPageReady={isPageReady}
        emptyGrid={emptyGridObject}
        userRoles={myReservations}
        user={user}
        form={form}
        roles={roles}
        showActionModal={showActionModal}
        actionModalContent={actionModalContent}
        filterText={filterText}
        setFilter={setFilter}
        actionOnModal={actionOnModal}
        onChangeInput={onChangeInput}
        addUserToRole={addUserToRole}
        loading={loading}
        isAssignRole={isAssignRole}
        isRemoveRole={isRemoveRole}
        onChangeDropdownMultiple={onChangeDropdownMultiple}
        onClickCloseErrorModal={onClickCloseErrorModal}
        errorModal={errorModal}
      />
    </>
  )
}
export default UserRolesContainer
