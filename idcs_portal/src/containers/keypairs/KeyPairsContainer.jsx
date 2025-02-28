// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import KeyPairsView from '../../components/keypairs/KeyPairsView'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import CloudAccountService from '../../services/CloudAccountService'
import { BsTrash3, BsCopy } from 'react-icons/bs'
import useToastStore from '../../store/toastStore/ToastStore'
import { useCopy } from '../../hooks/useCopy'

const getActionItemLabel = (text) => {
  let message = null

  switch (text) {
    case 'Delete':
      message = (
        <>
          {' '}
          <BsTrash3 /> {text}{' '}
        </>
      )
      break
    case 'Copy':
      message = (
        <>
          {' '}
          <BsCopy /> {' Copy Key'}
        </>
      )
      break
  }
  return message
}

const KeyPairsContainer = () => {
  // local state

  const onAction = (action, item) => {
    const selectedKey = item.name
    switch (action.id) {
      case 'terminate': {
        setSelectedKeyPairs(selectedKey)
        const copyModalContent = { ...modalContent }
        copyModalContent.label = action.label
        copyModalContent.name = item.name
        copyModalContent.buttonLabel = action.buttonLabel
        copyModalContent.action = action.id
        setActionModalContent(copyModalContent)
        setShowActionModal(true)
        break
      }
      case 'copy':
        copyToClipboard(item.sshPublicKey)
        break
      default:
        break
    }
  }

  const columns = [
    {
      columnName: 'Id',
      targetColumn: 'resourceId',
      hideField: true
    },
    { columnName: 'Name', targetColumn: 'name' },
    { columnName: 'Owner', targetColumn: 'ownerEmail' },
    { columnName: 'Type', targetColumn: 'type', hideField: true },
    {
      columnName: 'Actions',
      targetColumn: 'actions'
    }
  ]

  const emptyGrid = {
    title: 'No keys found',
    subTitle: 'Your account currently has no keys.',
    action: {
      type: 'redirect',
      href: '/security/publickeys/import',
      label: 'Upload key'
    }
  }

  const emptyGridByFilter = {
    title: 'No keys found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  const modalContent = {
    label: '',
    buttonLabel: '',
    name: '',
    action: ''
  }

  const [myPublicKeys, setMyPublicKeys] = useState(null)
  const [filteredKeys, setFilteredKeys] = useState(null)
  const [selectedKeyPairs, setSelectedKeyPairs] = useState()
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [isPageReady, setIsPageReady] = useState(false)

  const throwError = useErrorBoundary()
  const { copyToClipboard } = useCopy()

  const loading = useCloudAccountStore((state) => state.loading)
  const publicKeys = useCloudAccountStore((state) => state.publicKeys)
  const setPublickeys = useCloudAccountStore((state) => state.setPublickeys)

  const showSuccess = useToastStore((state) => state.showSuccess)
  const showError = useToastStore((state) => state.showError)

  // Hooks
  useEffect(() => {
    const fetchKeys = async () => {
      try {
        await setPublickeys()
        setIsPageReady(true)
      } catch (error) {
        throwError(error)
      }
    }
    fetchKeys()
  }, [])

  useEffect(() => {
    if (!isPageReady) {
      return
    }
    setGridInfo()
  }, [publicKeys, isPageReady])

  useEffect(() => {
    if (myPublicKeys === null) {
      return
    }
    if (filterText === '') {
      setFilteredKeys([...myPublicKeys])
    } else {
      const newValue = myPublicKeys.filter((x) => !filterText || x.name.indexOf(filterText) !== -1)
      setFilteredKeys(newValue)
      setEmptyGridObject(emptyGridByFilter)
    }
  }, [filterText])

  // functions
  function getActions(allowDelete) {
    const actions = [
      {
        id: 'copy',
        name: getActionItemLabel('Copy'),
        label: 'Copy key',
        buttonLabel: 'Copy'
      }
    ]
    if (allowDelete) {
      actions.push({
        id: 'terminate',
        name: getActionItemLabel('Delete'),
        label: 'Delete key',
        buttonLabel: 'Delete'
      })
    }
    return actions
  }

  function setGridInfo() {
    const gridInfo = []

    for (const index in publicKeys) {
      const publicKey = { ...publicKeys[index] }

      gridInfo.push({
        resourceId: publicKey.resourceId,
        name: publicKey.name,
        owner: publicKey.ownerEmail,
        type: publicKey.type,
        actions: {
          showField: true,
          type: 'Buttons',
          value: publicKey,
          selectableValues: getActions(publicKey.allowDelete),
          function: onAction
        }
      })
    }
    setMyPublicKeys(gridInfo)
    setFilteredKeys(gridInfo)
  }

  function setFilter(event, clear) {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  // Hide the modal
  const hideConfirmationModal = () => {
    setShowActionModal(false)
  }

  function actionOnModal(result) {
    if (!result) {
      hideConfirmationModal()
      return
    }
    switch (actionModalContent.action) {
      case 'terminate':
        deleteKeyPairsHandler()
        break
      default:
        break
    }
  }

  // Handle the actual deletion of the item
  const deleteKeyPairsHandler = () => {
    const result = selectedKeyPairs

    // Calls delete Public key service to perform delete action.
    CloudAccountService.deleteSshByCloud(selectedKeyPairs)
      .then((res) => {
        if (res) {
          getAvailableKeyPairs()
          setShowActionModal(false)
          showSuccess(`The key(s) '${result}' was deleted successfully.`)
        }
      })
      .catch((error) => {
        setShowActionModal(false)
        let errorMessage = ''
        let errorCode = ''
        let errorStatus = -1
        const isApiErrorWithErrorMessage = Boolean(error.response && error.response.data && error.response.data.message)

        if (isApiErrorWithErrorMessage) {
          errorStatus = error.response.status
          errorCode = error.response.data.code
          errorMessage = error.response.data.message
        } else {
          errorMessage = error.toString()
        }

        if (errorStatus === 403 && errorCode === 7) {
          showError(errorMessage)
        } else {
          throwError(error)
        }
      })
  }

  const getAvailableKeyPairs = async () => {
    // Calling KeyPairs service to fetch public keys.
    try {
      await setPublickeys()
    } catch (error) {
      throwError(error)
    }
  }

  return (
    <KeyPairsView
      loading={loading || filteredKeys === null}
      getAvailableKeyPairs={getAvailableKeyPairs}
      myPublicKeys={filteredKeys}
      columns={columns}
      actionModalContent={actionModalContent}
      showActionModal={showActionModal}
      actionOnModal={actionOnModal}
      importPagePath="/security/publickeys/import"
      filterText={filterText}
      setFilter={setFilter}
      emptyGrid={emptyGridObject}
    />
  )
}

export default KeyPairsContainer
