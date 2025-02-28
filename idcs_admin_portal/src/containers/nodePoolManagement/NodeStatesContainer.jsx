import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useNodePoolStore from '../../store/nodePoolStore/NodePoolStore'
import NodeStates from '../../components/nodePoolManagement/NodeStates'

const NodeStatesContainer = () => {
  const navigate = useNavigate()

  const { nodeName } = useParams()

  // local state
  const columns = [
    {
      columnName: 'Instance Type',
      targetColumn: 'instanceType'
    },
    {
      columnName: 'Running Instances',
      targetColumn: 'runningInstances'
    },
    {
      columnName: 'Max New Instances',
      targetColumn: 'maxNewInstances'
    }
  ]

  const emptyGrid = {
    title: 'No states found',
    subTitle: 'There are no states are available'
  }

  const emptyGridByFilter = {
    title: 'No instance types found',
    subTitle: 'The applied filter criteria did not match any instance types',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  // Local State
  const [nodeStatesList, setNodeStates] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [nodeDetails, setNodeDetails] = useState(null)
  const [isPageReady, setIsPageReady] = useState(false)

  // Global State
  const loading = useNodePoolStore((state) => state.loading)
  const nodeList = useNodePoolStore((state) => state.nodeList)
  const setNodeList = useNodePoolStore((state) => state.setNodeList)
  const nodeStates = useNodePoolStore((state) => state.nodeStates)
  const setNodeStatesList = useNodePoolStore((state) => state.setNodeStatesList)

  // Error Boundary
  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
      if (nodeList.length === 0) await setNodeList()
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    const fetchOptions = async () => {
      try {
        if (nodeList.length > 0) {
          await setOptions()
        }
      } catch (error) {
        throwError(error)
      }
    }
    fetchOptions().catch(() => {})
  }, [nodeList])

  useEffect(() => {
    if (nodeStates.length > 0 && !isPageReady) {
      setGridInfo()
    }
  }, [nodeStates])

  async function setOptions() {
    const nodeDetail = nodeList.find((node) => node.nodeName.toString() === nodeName.toString())
    setNodeDetails(nodeDetail)
    await setNodeStatesList(nodeDetail.nodeId)
  }

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in nodeStates) {
      const node = { ...nodeStates[index] }

      for (const j in node.instanceTypeStats) {
        const instanceType = { ...node.instanceTypeStats[j] }
        gridInfo.push({
          instanceType: instanceType.instanceType,
          runningInstances: instanceType.runningInstances,
          maxNewInstances: instanceType.maxNewInstances
        })
      }
    }
    setNodeStates(gridInfo)
    setIsPageReady(true)
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

  function backToHome() {
    navigate('/npm/nodes/')
  }

  return (
    <NodeStates
      loading={loading}
      nodeStatesList={nodeStatesList}
      columns={columns}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      setFilter={setFilter}
      backToHome={backToHome}
      nodeDetails={nodeDetails}
      isPageReady={isPageReady}
    />
  )
}

export default NodeStatesContainer
