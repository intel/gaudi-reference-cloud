import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import PoolList from '../../components/nodePoolManagement/PoolList'
import useNodePoolStore from '../../store/nodePoolStore/NodePoolStore'
import NodePoolService from '../../services/NodePoolService'
import useToastStore from '../../store/toastStore/ToastStore'

const PoolListContainer = () => {
  // local state
  const columns = [
    {
      columnName: 'Pool ID',
      targetColumn: 'poolId'
    },
    {
      columnName: 'Pool Name',
      targetColumn: 'poolName'
    },
    {
      columnName: '# of Nodes',
      targetColumn: 'numberOfNodes',
      columnConfig: {
        behaviorType: 'hyperlink',
        behaviorFunction: onRedirect
      }
    },
    {
      columnName: 'Pool Account Manager',
      targetColumn: 'poolAccountManagerAgsRole',
      hideField: true
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
    title: 'No pools found',
    subTitle: 'There are no pools created yet',
    action: {
      type: 'redirect',
      href: '/npm/pools/create',
      label: 'Create New Pool'
    }
  }

  const emptyGridByFilter = {
    title: 'No pools found',
    subTitle: 'The applied filter criteria did not match any pools',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  // Local State
  const [pools, setPools] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [selectedPool, setSelectedPool] = useState(null)
  const [showRequestModal, setShowRequestModal] = useState(false)

  // Global State
  const showError = useToastStore((state) => state.showError)
  const loading = useNodePoolStore((state) => state.loading)
  const poolList = useNodePoolStore((state) => state.poolList)
  const setPoolList = useNodePoolStore((state) => state.setPoolList)

  // Navigation
  const navigate = useNavigate()

  // Error Boundary
  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await refreshPools()
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    setGridInfo()
  }, [poolList])

  // functions

  async function refreshPools() {
    await setPoolList()
  }
  function setGridInfo() {
    const gridInfo = []

    for (const index in poolList) {
      const pool = { ...poolList[index] }
      gridInfo.push({
        poolId: pool.poolId,
        poolName: pool.poolName,
        numberOfNodes:
          pool.numberOfNodes > 0
            ? {
                showField: true,
                type: 'HyperLink',
                value: pool.numberOfNodes,
                function: () => {
                  onRedirect('nodes', pool.poolId)
                }
              }
            : pool.numberOfNodes,
        poolAccountManagerAgsRole: pool.poolAccountManagerAgsRole,
        actions: getActionButton(pool)
      })
    }

    setPools(gridInfo)
  }

  function getActionButton(pool) {
    return {
      showField: true,
      type: 'buttons',
      function: setAction,
      value: pool,
      selectableValues: [
        {
          value: pool,
          label: 'edit-button',
          name: 'Edit'
        },
        {
          value: pool,
          label: 'view-vl-button',
          name: 'View cloud accounts'
        },
        {
          value: pool,
          label: 'view-nodes-button',
          name: 'View nodes'
        },
        {
          value: pool,
          label: 'add-node-button',
          name: 'Add nodes'
        }
      ]
    }
  }

  function onRedirect(location, poolId = null) {
    navigate(`/npm/${location}` + (poolId ? `/${poolId}` : ''))
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

  function setAction(item, pool) {
    const type = item.label
    if (type === 'edit-button') {
      navigate('/npm/pools/edit/' + pool.poolId)
    } else if (type === 'view-vl-button') {
      navigate('/npm/pools/accounts/' + pool.poolId)
    } else if (type === 'add-node-button') {
      addNodeToPool(pool.poolId)
    } else {
      navigate('/npm/nodes/' + pool.poolId)
    }
  }

  function backToHome() {
    navigate('/')
  }

  function addNodeToPool(poolId) {
    setSelectedPool(poolId)
  }

  function cancelAddNode() {
    setSelectedPool(null)
    setShowRequestModal('')
  }

  async function addNodeToPoolFn(payload, nodeDetail) {
    setShowRequestModal(true)

    try {
      setShowRequestModal(true)
      await NodePoolService.editNode(nodeDetail.nodeId, payload)
      setShowRequestModal(false)
      await refreshPools()
    } catch (error) {
      setShowRequestModal(false)
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
    } finally {
      cancelAddNode()
    }
  }

  return (
    <PoolList
      loading={loading}
      pools={pools}
      columns={columns}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      setFilter={setFilter}
      backToHome={backToHome}
      showRequestModal={showRequestModal}
      selectedPool={selectedPool}
      cancelAddNode={cancelAddNode}
      addNodeToPoolFn={addNodeToPoolFn}
    />
  )
}

export default PoolListContainer
