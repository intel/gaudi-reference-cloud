// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import AccountsView from '../../components/profile/accounts/AccountsView'
import useUserStore from '../../store/userStore/UserStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import AccountsActionContainer from './AccountsActionContainer'
import AppSettingsService from '../../services/AppSettingsService'
import idcConfig, { setDefaultCloudAccount } from '../../config/configurator'
import AuthenticationSpinner from '../../utils/authenticationSpinner/AuthenticationSpinner'
import CloudAccountService from '../../services/CloudAccountService'
import useAppStore from '../../store/appStore/AppStore'
import TermsAndConditions from '../../pages/error/TermsAndConditions'
import { BsCopy } from 'react-icons/bs'
import { useCopy } from '../../hooks/useCopy'

const AccountsContainer = ({ goBack }) => {
  // local state
  const columns = [
    {
      columnName: 'Account ID',
      targetColumn: 'cloudAccountId'
    },
    {
      columnName: 'Account Owner',
      targetColumn: 'email'
    },
    {
      columnName: '',
      targetColumn: 'action',
      isSort: false
    }
  ]

  const modalContent = {
    label: 'Create a personal Cloud Account',
    buttonLabel: 'Confirm',
    uuid: '',
    name: '',
    resourceId: '',
    question: 'Do you want to create a personal cloud account?',
    feedback: '',
    resourceType: ''
  }

  const isAccountSelectionPath = goBack !== undefined

  const [listOwnedAccounts, setListOwnedAccounts] = useState([])
  const [listCloudAccounts, setListCloudAccounts] = useState([])
  const [filterText, setFilterText] = useState('')
  const [actionModalContent] = useState(modalContent)
  const [showModalActionConfirmation, setShowModalActionConfirmation] = useState(false)
  const [showTermsAndConditions, setShowTermsAndConditions] = useState(false)

  const throwError = useErrorBoundary()
  const { copyToClipboard } = useCopy()

  // Global State
  const user = useUserStore((state) => state.user)
  const loading = useUserStore((state) => state.loading)
  const cloudAccounts = useUserStore((state) => state.cloudAccounts)
  const setcloudAccounts = useUserStore((state) => state.setcloudAccounts)
  const setSelectedCloudAccount = useUserStore((state) => state.setSelectedCloudAccount)
  const resetStores = useAppStore((state) => state.resetStores)

  // Hooks
  useEffect(() => {
    const fetchAccounts = async () => {
      try {
        await setcloudAccounts()
      } catch (error) {
        throwError(error)
      }
    }
    fetchAccounts()
  }, [])

  useEffect(() => {
    if (!isAccountSelectionPath && !showTermsAndConditions) {
      reloadAccount()
    }
    setOwnedGridInfo()
    setMembersGridInfo()
  }, [cloudAccounts])

  useEffect(() => {
    if (!showModalActionConfirmation && showTermsAndConditions) {
      selectCloudAccount(user?.ownCloudAccountNumber)
      setShowTermsAndConditions(false)
    }
  }, [showModalActionConfirmation])

  // functions
  function setSelectedAccount() {
    if (cloudAccounts?.length > 0) {
      if (idcConfig.REACT_APP_SELECTED_CLOUD_ACCOUNT) {
        const selectedAccount = cloudAccounts.find(
          (x) => x.cloudAccountId === idcConfig.REACT_APP_SELECTED_CLOUD_ACCOUNT
        )
        if (selectedAccount !== undefined) {
          setSelectedCloudAccount(selectedAccount)
        } else {
          AppSettingsService.clearDefaultCloudAccount()
        }
      }
    }
  }

  function setOwnedGridInfo() {
    const gridInfo = []
    const allowCreatePersonalCloudAccount = !user.ownCloudAccountNumber

    gridInfo.push({
      cloudAccountId: allowCreatePersonalCloudAccount
        ? ''
        : {
            showField: true,
            type: 'Buttons',
            value: user.ownCloudAccountNumber,
            selectableValues: generateSelectableValues(user.ownCloudAccountNumber, true),
            function: (item, value) => {
              if (item?.id === 'copy') copyToClipboard(user.ownCloudAccountNumber)
              else selectCloudAccount(user.ownCloudAccountNumber)
            }
          },
      email: allowCreatePersonalCloudAccount
        ? {
            showField: true,
            type: 'hyperlink',
            value: 'Create a personal cloud account',
            function: () => {
              onClickCreateAccount()
            }
          }
        : user.email,
      action: null
    })
    setListOwnedAccounts(gridInfo)
  }

  const generateSelectableValues = (cloudId, enableRedirect) => {
    return [
      {
        id: 'hyperLink',
        label: `hyperLink-${cloudId}`,
        type: enableRedirect ? 'button' : 'icon',
        className: 'minWidth150',
        name: <div className={enableRedirect ? 'text-decoration-underline' : ''}>{cloudId}</div>
      },
      {
        id: 'copy',
        label: `copyBtn-${cloudId}`,
        variant: 'icon-simple',
        name: <BsCopy />
      }
    ]
  }
  function setMembersGridInfo() {
    const gridInfo = []
    const notOwnedCloudAccounts = cloudAccounts.filter((x) => x.email !== user.email)

    for (const index in notOwnedCloudAccounts) {
      const account = { ...notOwnedCloudAccounts[index] }
      gridInfo.push({
        cloudAccountId: account.isApproved
          ? {
              showField: true,
              type: 'Buttons',
              value: account.cloudAccountId,
              selectableValues: generateSelectableValues(account.cloudAccountId, account.isApproved),
              function: (item, value) => {
                if (item?.id === 'copy') copyToClipboard(account.cloudAccountId)
                else selectCloudAccount(account.cloudAccountId)
              }
            }
          : {
              showField: true,
              type: 'Buttons',
              value: account.cloudAccountId,
              selectableValues: generateSelectableValues(account.cloudAccountId, account.isApproved),
              function: () => {
                copyToClipboard(account.cloudAccountId)
              }
            },
        email: account.email,
        action: !account.isApproved ? (
          <AccountsActionContainer account={account} confirmInvite={confirmInvite} rejectInvite={rejectInvite} />
        ) : null
      })
    }

    setListCloudAccounts(gridInfo)
  }

  const reloadAccount = () => {
    setSelectedAccount()
    resetStores()
    setTimeout(() => {
      if (isAccountSelectionPath && goBack) {
        goBack()
      }
    }, 250)
  }

  const selectCloudAccount = (cloudAccountId) => {
    AppSettingsService.setDefaultCloudAccount(cloudAccountId)
    setDefaultCloudAccount(idcConfig)
    reloadAccount()
  }

  function setFilter(event, clear) {
    if (clear) {
      setFilterText('')
    } else {
      setFilterText(event.target.value)
    }
  }

  function onClickCreateAccount() {
    setShowModalActionConfirmation(true)
  }

  const onClickModalConfirmation = (status) => {
    if (status) {
      setShowTermsAndConditions(true)
    } else {
      setShowModalActionConfirmation(false)
    }
  }

  const confirmInvite = async (adminCloudAccountId, inviteCode) => {
    if (inviteCode !== '') {
      await CloudAccountService.checkInviteCodeForUserCloudAccount(adminCloudAccountId, inviteCode)
      setcloudAccounts()
    }
  }

  const rejectInvite = async (adminCloudAccountId, inviteState) => {
    await CloudAccountService.rejectUserCloudAccount(adminCloudAccountId, inviteState)
    setcloudAccounts()
  }

  if (idcConfig.REACT_APP_SELECTED_CLOUD_ACCOUNT && !user.cloudAccountNumber && !isAccountSelectionPath) {
    return <AuthenticationSpinner />
  }

  if (showTermsAndConditions) {
    return (
      <TermsAndConditions
        acceptCallback={() => {
          setShowModalActionConfirmation(false)
        }}
      />
    )
  }

  return (
    <AccountsView
      loading={loading}
      listCloudAccounts={listCloudAccounts}
      listOwnedAccounts={listOwnedAccounts}
      columns={columns}
      filterText={filterText}
      setFilter={setFilter}
      createPersonalAccount={!user.ownCloudAccountNumber}
      onClickCreateAccount={onClickCreateAccount}
      actionModalContent={actionModalContent}
      onClickModalConfirmation={onClickModalConfirmation}
      showModalActionConfirmation={showModalActionConfirmation}
    />
  )
}

export default AccountsContainer
