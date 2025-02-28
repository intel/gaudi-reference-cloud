import { useEffect, useState } from 'react'

import useNodePoolStore from '../../store/nodePoolStore/NodePoolStore'
import AddNodeToPool from '../../components/nodePoolManagement/AddNodeToPool'

const AddNodeToPoolContainer = (props) => {
  const selectedPool = props.selectedPool
  const cancelAddNode = props.cancelAddNode
  const addNodeToPoolFn = props.addNodeToPoolFn
  // local state

  const nodeListForm = {
    type: 'dropdown', // options = 'text ,'textArea'
    fieldSize: 'small', // options = 'small', 'medium', 'large'
    label: '',
    placeholder: 'Please select',
    value: [],
    isValid: true,
    isTouched: false,
    isMultiple: false,
    isReadOnly: false,
    validationRules: {
      isRequired: false
    },
    options: [],
    validationMessage: '',
    helperMessage: ''
  }

  const addNodeObject = {
    show: false,
    title: '',
    isLoader: false
  }

  // Local State

  const [addNodeModal, setAddNodeModal] = useState(addNodeObject)
  const [nodeListFormItem, setNodeListFormItem] = useState(nodeListForm)
  const [selectedNode, setSelectedNode] = useState(null)

  // Global State
  const nodeList = useNodePoolStore((state) => state.nodeList)

  const setNodeList = useNodePoolStore((state) => state.setNodeList)

  // Hooks
  useEffect(() => {
    if (selectedPool) {
      addNodeToPool()
    }
  }, [selectedPool])

  // functions

  async function addNodeToPool() {
    const addNodeModalCopy = { ...addNodeObject }
    addNodeModalCopy.show = true
    addNodeModalCopy.isLoader = true
    addNodeModalCopy.title = 'Add Node to Pool ' + selectedPool
    setAddNodeModal(addNodeModalCopy)

    const data = await setNodeList(null, true)
    const nodeOptions = getNodeListOptions(data)

    const formCopy = { ...nodeListForm }
    formCopy.options = nodeOptions
    formCopy.value = nodeOptions.length > 0 ? nodeOptions[0].value : ''

    setNodeListFormItem(formCopy)
    setSelectedNode(nodeOptions[0].value)
    addNodeModalCopy.isLoader = false
    setAddNodeModal(addNodeModalCopy)
  }

  function getNodeListOptions(data) {
    return data.map((item) => ({
      name: item.nodeName,
      value: item.nodeId
    }))
  }

  function onSelectNodeFromList(event) {
    const inputValue = event.target.value
    setSelectedNode(inputValue)

    const formCopy = { ...nodeListFormItem }
    formCopy.value = inputValue

    setNodeListFormItem(formCopy)
  }

  function cancelNodeAddition() {
    setAddNodeModal({ ...addNodeObject })
    setNodeListFormItem({ ...nodeListForm })
    setSelectedNode(null)
    cancelAddNode()
  }

  async function addNodeFn() {
    const addNodeModalCopy = { ...addNodeObject }
    addNodeModalCopy.show = false
    addNodeModalCopy.isLoader = false
    addNodeModalCopy.title = ''
    setAddNodeModal(addNodeModalCopy)

    const nodeDetail = nodeList.filter((x) => x.nodeId === selectedNode)[0]

    const pools = [...new Set([...nodeDetail.poolIds, selectedPool])]

    const payload = {
      availabilityZone: nodeDetail.availabilityZone,
      instanceTypesOverride: {
        overridePolicies: true,
        overrideValues: nodeDetail.instanceTypes
      },
      computeNodePoolsOverride: {
        overridePolicies: true,
        overrideValues: pools
      },
      region: nodeDetail.region
    }

    addNodeToPoolFn(payload, nodeDetail)
  }

  return (
    <AddNodeToPool
      addNodeModal={addNodeModal}
      nodeListFormItem={nodeListFormItem}
      selectedNode={selectedNode}
      onSelectNodeFromList={onSelectNodeFromList}
      cancelNodeAddition={cancelNodeAddition}
      addNodeFn={addNodeFn}
    />
  )
}

export default AddNodeToPoolContainer
