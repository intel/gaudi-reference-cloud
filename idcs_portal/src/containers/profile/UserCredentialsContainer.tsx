// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import UserCredentialsView from '../../components/profile/userCredentials/UserCredentialsView'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import CloudAccountService from '../../services/CloudAccountService'
import { BsTrash3, BsCopy } from 'react-icons/bs'
import useToastStore from '../../store/toastStore/ToastStore'
import moment from 'moment'
import { useCopy } from '../../hooks/useCopy'
import idcConfig from '../../config/configurator'

const UserCredentialsContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)
  // *****
  // local state
  // *****
  const emptyGrid = {
    title: 'No Client Secrets found',
    subTitle: (
      <>
        {`Use Client Secrets to authenticate to the ${idcConfig.REACT_APP_CONSOLE_SHORT_NAME} API and`} <br />
        work with your account programmatically.
      </>
    ),
    action: {
      type: 'redirect',
      href: '/profile/credentials/launch',
      label: 'Generate Client Secret'
    }
  }

  const columns = [
    {
      columnName: 'Name',
      targetColumn: 'appClientName'
    },
    {
      columnName: 'Client Id',
      targetColumn: 'clientId'
    },
    {
      columnName: 'Create Date',
      targetColumn: 'created'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions'
    }
  ]

  const actionsOption = [
    {
      id: 'delete',
      name: (
        <>
          <BsTrash3 /> Delete
        </>
      ),
      label: 'Delete client secret',
      buttonLabel: 'Delete'
    }
  ]

  const actionModalInitial = {
    id: '',
    name: '',
    clientId: '',
    label: '',
    buttonLabel: ''
  }

  const [loading, setLoading] = useState(false)
  const [actionModal, setActionModal] = useState(actionModalInitial)
  const [showActionModal, setShowActionModal] = useState(false)
  const [userCredentials, setUserCredentials] = useState<any[] | null>(null)
  const { copyToClipboard } = useCopy()

  // *****
  // Global State
  // *****
  const throwError = useErrorBoundary()

  // *****
  // Hooks
  // *****
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await getUserCredentials()
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  // *****
  // Functions
  // *****
  const getUserCredentials = async (): Promise<void> => {
    setLoading(true)
    const result = await CloudAccountService.getUserCredentials()
    const data = result.data
    const appClients = data.appClients
    const gridInfo: any[] = []
    for (const index in appClients) {
      const item = { ...appClients[index] }
      const stringDate = item.created
      const createdDt = moment(stringDate).format('MM/DD/YYYY h:mm a')
      gridInfo.push({
        appClientName: item.appClientName,
        clientId: {
          showField: true,
          type: 'hyperlink',
          noHyperLinkValue: item.clientId,
          value: '',
          icon: <BsCopy className="text-primary" size="20" />,
          function: () => {
            KubeConfigCopy(item.clientId)
          }
        },
        created: createdDt,
        actions: {
          showField: true,
          type: 'Buttons',
          value: item,
          selectableValues: actionsOption,
          function: setAction
        }
      })
    }
    setUserCredentials(gridInfo)
    setLoading(false)
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      default: {
        const name = item.appClientName
        const clientId = item.clientId
        setActionModal({ ...action, name, clientId })
        setShowActionModal(true)
        break
      }
    }
  }

  const KubeConfigCopy = (clientId: string): void => {
    copyToClipboard(clientId)
  }

  const onClickDeleteModal = async (action: string): Promise<void> => {
    try {
      if (!action) {
        setActionModal(actionModalInitial)
        setShowActionModal(false)
      } else {
        await CloudAccountService.deleteCredentialUser(actionModal.clientId)
        setActionModal(actionModalInitial)
        setShowActionModal(false)
        showSuccess('Client secret deleted successfully', false)
        await getUserCredentials()
      }
    } catch (error: any) {
      const message = String(error.message)
      showError(message, false)
      setActionModal(actionModalInitial)
      setShowActionModal(false)
    }
  }

  return (
    <UserCredentialsView
      credentials={userCredentials ?? []}
      loading={loading || userCredentials === null}
      emptyGrid={emptyGrid}
      columns={columns}
      actionModal={actionModal}
      showActionModal={showActionModal}
      onClickDeleteModal={onClickDeleteModal}
    />
  )
}

export default UserCredentialsContainer
