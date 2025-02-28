import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import CloudAccountList from '../../components/nodePoolManagement/CloudAccountList'
import useNodePoolStore from '../../store/nodePoolStore/NodePoolStore'
import useToastStore from '../../store/toastStore/ToastStore'
import NodePoolService from '../../services/NodePoolService'

const CloudAccountListContainer = () => {
  const navigate = useNavigate()

  const { poolId } = useParams()
  if (!poolId) navigate('/npm/pools')

  // local state
  const columns = [
    {
      columnName: 'Compute Node Pool',
      targetColumn: 'poolId'
    },
    {
      columnName: 'Cloud Account',
      targetColumn: 'cloudAccountId'
    },
    {
      columnName: 'Assignment Admin',
      targetColumn: 'createAdmin',
      isSort: false
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'buttons',
        behaviorFunction: null
      }
    }
  ]

  const emptyGrid = {
    title: `No cloud accounts found for ${poolId ?? ''}`,
    subTitle: 'No cloud accounts attached to pool',
    action: {
      type: 'redirect',
      href: '/npm/pools/accounts/add/' + poolId,
      label: 'Add Cloud Account to ' + poolId
    }
  }

  const emptyGridByFilter = {
    title: 'No cloud accounts found',
    subTitle: 'The applied filter criteria did not match any cloud accounts',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  // Initial state for confirm modal
  const initialConfirmData = {
    isShow: false,
    title: '',
    data: [],
    id: null,
    isBlocked: null,
    onClose: closeConfirmModal
  }

  // Initial state for loader
  const initialLoaderData = {
    isShow: false,
    message: ''
  }

  // Local State
  const [cloudAccounts, setCloudAccounts] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [confirmModalData, setConfirmModalData] = useState(initialConfirmData)
  const [showLoader, setShowLoader] = useState(initialLoaderData)

  // Global State
  const loading = useNodePoolStore((state) => state.loading)
  const cloudAccountList = useNodePoolStore((state) => state.cloudAccountList)
  const setCloudAccountList = useNodePoolStore((state) => state.setCloudAccountList)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // Error Boundary
  const throwError = useErrorBoundary()

  const refreshCloudAccounts = async () => {
    await setCloudAccountList(poolId)
  }

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await refreshCloudAccounts()
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    setGridInfo()
  }, [cloudAccountList])

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in cloudAccountList) {
      const account = { ...cloudAccountList[index] }

      gridInfo.push({
        poolId: account.poolId,
        cloudAccountId: account.cloudAccountId,
        createAdmin: account.createAdmin,
        actions: getActionButton(account)
      })
    }

    setCloudAccounts(gridInfo)
  }

  function closeConfirmModal() {
    setConfirmModalData(initialConfirmData)
  }

  function getActionButton(account) {
    return {
      showField: true,
      type: 'buttons',
      function: setAction,
      value: account,
      selectableValues: [
        {
          value: account,
          label: 'Remove button',
          name: 'Remove'
        }
      ]
    }
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

  function setAction(item, account) {
    const data = [
      {
        col: 'Compute Node Pool',
        value: account.poolId
      },
      {
        col: 'Cloud Account',
        value: account.cloudAccountId
      },
      {
        col: 'Assignment Admin',
        value: account.createAdmin
      }
    ]

    setConfirmModalData({
      title: `Remove cloud account: ${account.cloud_account_id}?`,
      data,
      isShow: true,
      id: account.cloudAccountId,
      isBlocked: false,
      onClose: closeConfirmModal
    })
  }

  function onSubmit() {
    closeConfirmModal()
    setShowLoader({ isShow: true, message: 'Working on your request' })
    submitForm()
  }

  async function submitForm() {
    try {
      await NodePoolService.removeCloudAccounts(poolId, confirmModalData.id)
      showSuccess(`Cloud account ${confirmModalData.id} has been removed from pool ${poolId}.`)
      await refreshCloudAccounts()
    } catch (error) {
      let message = ''
      if (error.response) {
        if (error.response.data.message !== '') {
          message = error.response.data.message
        } else {
          message = error.message
        }
      } else {
        message = error.message
      }
      showError(message)
    }
    setShowLoader({ isShow: false })
  }

  return (
    <CloudAccountList
      loading={loading}
      cloudAccounts={cloudAccounts}
      columns={columns}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      confirmModalData={confirmModalData}
      showLoader={showLoader}
      setFilter={setFilter}
      onSubmit={onSubmit}
      poolId={poolId}
    />
  )
}

export default CloudAccountListContainer
