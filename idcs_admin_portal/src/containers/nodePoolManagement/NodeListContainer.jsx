import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import NodeList from '../../components/nodePoolManagement/NodeList'
import useNodePoolStore from '../../store/nodePoolStore/NodePoolStore'
import { formatNumber } from '../../utility/numberFormatHelper/NumberFormatHelper'

const NodeListContainer = () => {
  const navigate = useNavigate()

  const { poolId } = useParams()

  // local state
  const columns = [
    {
      columnName: 'Node Id',
      targetColumn: 'nodeId',
      hideField: true
    },
    {
      columnName: 'Node Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Region',
      targetColumn: 'region'
    },
    {
      columnName: 'Cluster Id',
      targetColumn: 'clusterId'
    },
    {
      columnName: 'Availability Zone (AZ)',
      targetColumn: 'az'
    },
    {
      columnName: 'Instance Types',
      width: '20rem',
      targetColumn: 'instanceTypes'
    },
    {
      columnName: 'Compute Node Pools',
      targetColumn: 'pools',
      width: '20rem',
      columnConfig: {
        behaviorType: 'Buttons',
        behaviorFunction: onRedirect
      }
    },
    {
      columnName: '% Used',
      targetColumn: 'resources'
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
    title: 'No nodes found',
    subTitle: 'There are no nodes available yet'
  }

  const emptyGridByFilter = {
    title: 'No nodes found',
    subTitle: 'The applied filter criteria did not match any nodes',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  // Local State
  const [nodes, setNodes] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)

  // Global State
  const loading = useNodePoolStore((state) => state.loading)
  const nodeList = useNodePoolStore((state) => state.nodeList)
  const setNodeList = useNodePoolStore((state) => state.setNodeList)

  // Error Boundary
  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await setNodeList(poolId)
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [poolId])

  useEffect(() => {
    setGridInfo()
  }, [nodeList])

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in nodeList) {
      const node = { ...nodeList[index] }

      gridInfo.push({
        nodeId: node.nodeId,
        name: node.nodeName,
        region: node.region,
        clusterId: node.clusterId,
        az: node.availabilityZone,
        instanceTypes: node.instanceTypes.join(', '),
        pools: getHyperLinks(node.poolIds),
        resources: formatNumber(node.percentageResourcesUsed, 2),
        actions: getActionButton(node)
      })
    }
    setNodes(gridInfo)
  }

  function getHyperLinks(pools) {
    return {
      showField: true,
      type: 'Buttons',
      function: onRedirect,
      value: pools,
      selectableValues: pools.map((pool) => {
        return {
          id: pool,
          value: pool,
          label: `Edit pool ${pool}`,
          name: pool
        }
      })
    }
  }
  function getActionButton(node) {
    return {
      showField: true,
      type: 'buttons',
      function: setAction,
      value: node,
      selectableValues: [
        {
          value: node,
          label: 'edit-button',
          name: 'Edit'
        },
        {
          value: node,
          label: 'states-button',
          name: 'Statistic'
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

  function onRedirect(action) {
    navigate('/npm/pools/edit/' + action.id)
  }

  function setAction(item, node) {
    const type = item.label
    if (type === 'edit-button') {
      navigate('/npm/nodes/edit/' + node.nodeName)
    } else {
      navigate('/npm/nodes/statistic/' + node.nodeName)
    }
  }

  function backToHome() {
    if (poolId) {
      navigate('/npm/pools/edit/' + poolId)
    } else {
      navigate('/')
    }
  }

  return (
    <NodeList
      loading={loading}
      nodes={nodes}
      columns={columns}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      setFilter={setFilter}
      backToHome={backToHome}
      poolId={poolId}
    />
  )
}

export default NodeListContainer
