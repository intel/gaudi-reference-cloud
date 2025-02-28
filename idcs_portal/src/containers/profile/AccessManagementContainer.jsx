// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import useUserStore from '../../store/userStore/UserStore'
import { useEffect, useState, useRef } from 'react'
import { BsEnvelope, BsPlusLg, BsPencilFill } from 'react-icons/bs'
import useToastStore from '../../store/toastStore/ToastStore'
import { UpdateFormHelper, getFormValue, isValidForm } from '../../utils/updateFormHelper/UpdateFormHelper'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import VerificationCodeContainer from './VerificationCodeContainer'
import AccessManagementView from '../../components/profile/accountSettings/AccessManagementView'
import CloudAccountService from '../../services/CloudAccountService'
import { InvitationState, specificErrorMessageEnum } from '../../utils/Enums'
import moment from 'moment'
import { useNavigate } from 'react-router-dom'
import { isFeatureFlagEnable, appFeatureFlags } from '../../config/configurator'
import useAuthorizationStore from '../../store/authorizationStore/AuthorizationStore'
import { friendlyErrorMessages } from '../../utils/apiError/apiError'

const getActionItemLabel = (text) => {
  let message = null

  switch (text) {
    case 'removePendingInvite':
      message = <> Remove </>
      break
    case 'removeAcceptedInvite':
      message = <> Revoke </>
      break
    case 'Resend':
      message = (
        <>
          {' '}
          <BsEnvelope /> {text}{' '}
        </>
      )
      break
    case 'editRoles':
      message = (
        <>
          {' '}
          <BsPencilFill /> Edit/View{' '}
        </>
      )
      break
    default:
      message = <> {text} </>
      break
  }

  return message
}

const AccessManagementContainer = () => {
  const addMemberFormInitialState = {
    modalOpened: false,
    fields: {
      emailAddress: {
        type: 'text',
        label: 'Email:',
        placeholder: 'example@domain.com',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        validationRules: {
          isRequired: true,
          emailAddress: true
        },
        validationMessage: 'Please enter a valid email address',
        helperMessage: 'An invitation will be sent to this email address.'
      },
      expirationDate: {
        type: 'date',
        label: 'Invitation expiration date:',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        validationRules: {
          isRequired: true,
          futureDate: true
        },
        validationMessage: 'Please enter a date in future',
        min: moment().format('YYYY-MM-DD')
      },
      note: {
        type: 'textarea',
        label: 'Note:',
        placeholder: 'Note for user',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 200,
        validationRules: {
          isRequired: true,
          isEmailNote: true,
          checkMaxLength: true
        },
        validationMessage: '',
        helperMessage: 'Note will be added to the invitation'
      }
    },
    isValidForm: false,
    cancelAddNewMember: () => {
      cancelAddNewMember()
    },
    sendInvite: () => {
      sendInvite()
    }
  }

  const removeMemberInitialState = {
    modalOpened: false,
    modalTitle: '',
    buttonTitle: '',
    emailAddress: '',
    invitationState: '',
    action: '',
    cancelRemoveMember: () => {
      cancelRemoveMember()
    },
    okRemoveMember: (args) => {
      okRemoveMember(args)
    }
  }

  const invitationGridOptions = [
    {
      id: 'removePendingInvite',
      name: getActionItemLabel('removePendingInvite'),
      label: 'Remove invitation',
      buttonLabel: 'Remove invitation',
      status: ['INVITE_STATE_PENDING_ACCEPT']
    },
    {
      id: 'removeAcceptedInvite',
      name: getActionItemLabel('removeAcceptedInvite'),
      label: 'Revoke access',
      buttonLabel: 'Revoke access',
      status: ['INVITE_STATE_ACCEPTED']
    },
    {
      id: 'resendInvitation',
      name: getActionItemLabel('Resend'),
      label: 'Resend invitation',
      buttonLabel: 'Resend invitation',
      status: ['INVITE_STATE_PENDING_ACCEPT']
    }
  ]

  const roleGridOptions = [
    {
      id: 'editRoles',
      name: getActionItemLabel('editRoles'),
      label: 'Edit/View',
      buttonLabel: 'Edit/View',
      status: ['INVITE_STATE_ACCEPTED']
    }
  ]

  const accessManagementColumns = [
    {
      columnName: 'Email',
      targetColumn: 'email'
    },
    {
      columnName: 'Status',
      targetColumn: 'status'
    },
    {
      columnName: 'Invitation',
      targetColumn: 'invitation',
      columnConfig: {
        behaviorType: 'Buttons',
        behaviorFunction: null
      }
    }
  ]

  if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_USER_ROLES)) {
    accessManagementColumns.push({
      columnName: 'Roles',
      targetColumn: 'roles',
      columnConfig: {
        behaviorType: 'Buttons',
        behaviorFunction: null
      }
    })

    addMemberFormInitialState.fields.roles = {
      type: 'multi-select-dropdown', // options = 'text ,'textArea'
      label: 'Select role (Optional)',
      placeholder: 'Please select role',
      value: [],
      isValid: true,
      isTouched: false,
      isReadOnly: false,
      hiddenLabel: true,
      validationRules: {
        isRequired: false
      },
      options: [],
      validationMessage: '',
      helperMessage:
        'Permissions associated with the selected role will be applied. If no role is provided, the member will not have any access. The role can be updated after the member accepts the invitation.',
      emptyOptionsMessage: 'No role found.'
    }
  }

  const accountManagementEmptyGrid = {
    subTitle: 'Your account currently has no other members.',
    action: {
      type: 'function',
      rightIcon: <BsPlusLg />,
      btnType: 'primary',
      href: () => addNewMember(),
      label: 'Grant access'
    }
  }

  const accountManagementEmptyGridNoResults = {
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  // Navigation
  const navigate = useNavigate()

  const [accessManagementGridInfo, setAccessManagementGridInfo] = useState([])
  const [filterText, setFilterText] = useState('')
  const [addMemberForm, setAddMemberForm] = useState(addMemberFormInitialState)
  const [removeMember, setRemoveMember] = useState(removeMemberInitialState)
  const [showVerificationModal, setShowVerificationModal] = useState(false)

  // Global State
  const user = useUserStore((state) => state.user)
  const isPremiumUser = useUserStore((state) => state.isPremiumUser)
  const isIntelUser = useUserStore((state) => state.isIntelUser)
  const invitationLoading = useUserStore((state) => state.invitationLoading)
  const adminInvitation = useUserStore((state) => state.adminInvitation)
  const setInvitations = useUserStore((state) => state.setInvitations)
  const roles = useAuthorizationStore((state) => state.roles)
  const setRoles = useAuthorizationStore((state) => state.setRoles)

  const showSuccess = useToastStore((state) => state.showSuccess)
  const showError = useToastStore((state) => state.showError)

  const memberRef = useRef(null)

  const throwError = useErrorBoundary()
  const invitationLimit = isPremiumUser() ? 10 : isIntelUser() ? 15 : 50

  // Hooks
  useEffect(() => {
    const fetchInvitations = async () => {
      try {
        await setInvitations()
        if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_USER_ROLES)) await setRoles()
      } catch (error) {
        throwError(error)
      }
    }
    fetchInvitations()
  }, [])

  useEffect(() => {
    initAccessManagementGridInfo()
  }, [adminInvitation])

  const initAccessManagementGridInfo = () => {
    const gridInfo = []

    adminInvitation.forEach((invitation) => {
      const gridData = {
        email: invitation.member_email,
        status: getInvitationStatus(invitation),
        invitation: {
          showField: true,
          type: 'Buttons',
          value: invitation,
          selectableValues: getActionButton(invitation.invitation_state, 'invitation'),
          function: setAction
        }
      }
      if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_USER_ROLES)) {
        gridData.roles = {
          showField: true,
          type: 'Buttons',
          value: invitation,
          selectableValues: getActionButton(invitation.invitation_state, 'roles'),
          function: setAction
        }
      }
      gridInfo.push(gridData)
    })
    setAccessManagementGridInfo(gridInfo)
  }

  function getInvitationStatus(invitation) {
    let status = InvitationState[invitation.invitation_state]
    if (status === 'Invited') {
      status += ` (expires on ${moment(invitation.expiry).utc().format('MM/DD/YYYY')})`
    }
    return status
  }

  function getActionButton(invitationState, column) {
    const result = []
    let gridOptions = []
    if (column === 'roles') gridOptions = [...roleGridOptions]
    else if (column === 'invitation') gridOptions = [...invitationGridOptions]

    for (const index in gridOptions) {
      const option = { ...gridOptions[index] }
      if (option.status.find((item) => item === invitationState)) {
        result.push(option)
      }
    }

    return result
  }

  const setAction = (action, item) => {
    switch (action.id) {
      case 'removePendingInvite':
        setRemoveMember({
          ...removeMemberInitialState,
          emailAddress: item.member_email,
          invitationState: item.invitation_state,
          modalOpened: true,
          modalTitle: 'Remove invitation',
          buttonTitle: 'Remove invitation',
          action: 'removePendingInvite'
        })
        break
      case 'removeAcceptedInvite':
        setRemoveMember({
          ...removeMemberInitialState,
          emailAddress: item.member_email,
          invitationState: item.invitation_state,
          modalOpened: true,
          modalTitle: 'Revoke access',
          buttonTitle: 'Revoke access',
          action: 'removeAcceptedInvite'
        })
        break
      case 'resendInvitation':
        resendInvitation(item.member_email)
        break
      case 'editRoles':
        navigate({ pathname: `/profile/accountAccessManagement/user-role/${item.member_email}` })
        break
      default:
        break
    }
  }

  const sendInvite = async () => {
    try {
      await CloudAccountService.createOtp(getFormValue('emailAddress', memberRef.current.fields))
      setShowVerificationModal(true)
      setAddMemberForm({ ...addMemberFormInitialState })
    } catch (error) {
      setAddMemberForm({ ...addMemberFormInitialState })
      let errMsg = ''
      if (error?.response?.data) {
        errMsg =
          error.response.data.code === 3 &&
          error.response.data?.message?.includes(specificErrorMessageEnum.maxMemberLimitReached)
            ? friendlyErrorMessages.maxMemberLimitReached
            : error.response.data?.message
      } else errMsg = error.message
      showError(errMsg)
    }
  }

  const okRemoveMember = (args) => {
    if (args.action === 'removeAcceptedInvite') {
      removeAcceptedInvite(args)
    } else {
      removePendingInvite(args)
    }
  }

  const removePendingInvite = async (args) => {
    try {
      await CloudAccountService.removePendingInvite(args.email, args.invitationState)
      showSuccess('Invitation has been removed.')
      setInvitations()
    } catch (error) {
      throwError(error)
    }
    setRemoveMember({ ...removeMemberInitialState })
  }

  const removeAcceptedInvite = async (args) => {
    try {
      await CloudAccountService.removeAcceptedInvite(args.email, args.invitationState)
      showSuccess('User removed from account.')
      setInvitations()
    } catch (error) {
      throwError(error)
    }
    setRemoveMember({ ...removeMemberInitialState })
  }

  const resendInvitation = async (email) => {
    try {
      const response = await CloudAccountService.resendInvite(email)
      if (response.data.blocked) {
        showError(response.data.message)
      } else {
        showSuccess(`Email sent to ${email}`)
      }
    } catch (error) {
      throwError(error)
    }
  }

  const addNewMember = () => {
    setAddMemberForm({
      ...addMemberFormInitialState,
      modalOpened: true
    })
  }

  const cancelAddNewMember = () => {
    setAddMemberForm({ ...addMemberFormInitialState })
  }

  const cancelRemoveMember = () => {
    setRemoveMember({ ...removeMemberInitialState })
  }

  const setFilter = (event, clear) => {
    if (clear) {
      setFilterText('')
    } else {
      setFilterText(event.target.value)
    }
  }

  function onChangeInput(event, formInputName) {
    let value = null
    value = event.target.value
    const addMemberFormCopy = {
      ...addMemberForm
    }

    const formCopy = { ...addMemberFormCopy.fields }

    const updatedForm = UpdateFormHelper(value, formInputName, formCopy)
    const isValid = isValidForm(updatedForm)

    addMemberFormCopy.fields = updatedForm
    addMemberFormCopy.isValidForm = isValid
    setAddMemberForm(addMemberFormCopy)
    memberRef.current = addMemberFormCopy
  }

  const onChangeDropdownMultiple = (values, formInputName) => {
    const addMemberFormCopy = { ...addMemberForm }
    const formCopy = { ...addMemberFormCopy.fields }

    const updatedForm = UpdateFormHelper(values, formInputName, formCopy)
    const isValid = isValidForm(updatedForm)

    addMemberFormCopy.fields = updatedForm
    addMemberFormCopy.isValidForm = isValid
    setAddMemberForm(addMemberFormCopy)
    memberRef.current = addMemberFormCopy
  }

  function cancelVerification() {
    setShowVerificationModal(false)
    setAddMemberForm(addMemberFormInitialState)
  }

  async function successVerification() {
    setShowVerificationModal(false)
    try {
      const email = getFormValue('emailAddress', memberRef.current.fields)
      const expiry = getFormValue('expirationDate', memberRef.current.fields)
      const note = getFormValue('note', memberRef.current.fields)

      const invitePayload = [
        {
          expiry: new Date(expiry).toISOString(),
          memberEmail: email,
          note
        }
      ]

      if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_USER_ROLES)) {
        const roles = getFormValue('roles', memberRef.current.fields)
        invitePayload[0].cloudAccountRoleIds = roles
      }

      await CloudAccountService.createInvite(invitePayload)
      showSuccess('User granted access.')
      setShowVerificationModal(false)
      setAddMemberForm(addMemberFormInitialState)
      setInvitations()
    } catch (error) {
      let message = ''

      if (error.response) {
        message = error.response.data.message
        if (message?.includes(specificErrorMessageEnum.maxMemberLimitReached)) {
          message = friendlyErrorMessages.maxMemberLimitReached
        } else if (message?.includes(specificErrorMessageEnum.memberExisted)) {
          message = friendlyErrorMessages.memeberExisted
        }
      } else {
        message = error.message
      }
      showError(message)
    }
  }

  return (
    <>
      {showVerificationModal ? (
        <VerificationCodeContainer
          showVerificationModal={showVerificationModal}
          emailAddress={getFormValue('emailAddress', memberRef.current.fields)}
          successVerification={successVerification}
          cancelVerification={cancelVerification}
        />
      ) : null}
      <AccessManagementView
        cloudAccountId={user.cloudAccountNumber}
        accessManagementColumns={accessManagementColumns}
        accessManagementGridInfo={accessManagementGridInfo}
        addNewMember={addNewMember}
        filterText={filterText}
        setFilter={setFilter}
        invitationLoading={invitationLoading}
        accountManagementEmptyGrid={accountManagementEmptyGrid}
        accountManagementEmptyGridNoResults={accountManagementEmptyGridNoResults}
        addMemberForm={addMemberForm}
        onChangeInput={onChangeInput}
        removeMember={removeMember}
        invitationLimit={invitationLimit}
        onChangeDropdownMultiple={onChangeDropdownMultiple}
        roles={roles}
      />
    </>
  )
}

export default AccessManagementContainer
