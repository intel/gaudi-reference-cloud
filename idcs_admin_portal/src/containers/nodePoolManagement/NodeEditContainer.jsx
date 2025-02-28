import { React, useEffect, useState } from 'react'
import NodeEdit from '../../components/nodePoolManagement/NodeEdit'
import { useNavigate, useParams } from 'react-router-dom'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setSelectOptions,
  setFormValue
} from '../../utility/updateFormHelper/UpdateFormHelper'
import useNodePoolStore from '../../store/nodePoolStore/NodePoolStore'
import useToastStore from '../../store/toastStore/ToastStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import NodePoolService from '../../services/NodePoolService'
import useProductStore from '../../store/productStore/ProductStore'

const NodeEditContainer = () => {
  // Navigation
  const navigate = useNavigate()
  const { nodeName } = useParams()

  // local state
  const initialState = {
    title: 'Edit Node',
    description: '',
    form: {
      instanceTypes: {
        sectionGroup: 'instanceType',
        type: 'multi-select-dropdown', // options = 'text ,'textArea'
        fieldSize: 'small', // options = 'small', 'medium', 'large'
        label: '',
        placeholder: 'Please select',
        value: [],
        isValid: true,
        isTouched: false,
        isMultiple: true,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        borderlessDropdownMultiple: true,
        options: [],
        validationMessage: 'At least one instance policy is required.',
        helperMessage: '',
        maxWidth: '65rem'
      },
      pools: {
        sectionGroup: 'pool',
        type: 'multi-select', // options = 'text ,'textArea'
        fieldSize: 'small', // options = 'small', 'medium', 'large'
        label: '',
        placeholder: 'Please select',
        value: [],
        isValid: true,
        isTouched: false,
        isMultiple: true,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        borderlessDropdownMultiple: true,
        options: [],
        validationMessage: 'At least one pool policy is required.',
        helperMessage: ''
      }
    },
    isValidForm: true,
    navigationTop: [
      {
        label: '⟵ Back to nodes',
        buttonVariant: 'link',
        function: () => onCancel('/npm/nodes')
      }
    ],
    navigationBottom: [
      {
        label: 'Save',
        buttonVariant: 'primary'
      },
      {
        label: 'Cancel',
        buttonVariant: 'link',
        function: () => onCancel('/npm/nodes')
      }
    ]
  }

  const [state, setState] = useState(initialState)
  const [showModal, setShowModal] = useState(false)
  const [nodeDetails, setNodeDetails] = useState(null)
  const [apiCallsCompleted, setApiCallsCompleted] = useState(false)

  // Global Store
  const showError = useToastStore((state) => state.showError)
  const poolList = useNodePoolStore((state) => state.poolList)
  const setPoolList = useNodePoolStore((state) => state.setPoolList)
  const nodeList = useNodePoolStore((state) => state.nodeList)
  const setNodeList = useNodePoolStore((state) => state.setNodeList)

  const products = useProductStore((state) => state.products)
  const setProducts = useProductStore((state) => state.setProducts)

  // Error Boundary
  const throwError = useErrorBoundary()

  useEffect(() => {
    const fetch = async () => {
      try {
        await Promise.all([setPoolList(), setNodeList(), products.length === 0 && setProducts()])
        setApiCallsCompleted(true)
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    setForm()
  }, [nodeList, products, poolList])

  // functions

  function setForm() {
    if (nodeList.length !== 0) {
      const stateUpdated = { ...state }

      const nodeDetail = nodeList.find((node) => node.nodeName.toString() === nodeName.toString())

      if (!nodeDetail) {
        navigate('/npm/nodes')
        return false
      }

      setNodeDetails(nodeDetail)

      let poolOptions = poolList.map((x) => x.poolId)

      poolOptions = [...new Set([...nodeDetail.poolIds, ...poolOptions])]

      const poolOptionsSelect = generateSelectOptions(poolOptions)

      // initial local data
      const instanceTypeOptions = getProductsOption()

      stateUpdated.form = setSelectOptions('pools', poolOptionsSelect, stateUpdated.form)
      stateUpdated.form = setFormValue('pools', structuredClone(nodeDetail.poolIds), stateUpdated.form)

      stateUpdated.form = setSelectOptions('instanceTypes', instanceTypeOptions, stateUpdated.form)
      stateUpdated.form = setFormValue('instanceTypes', nodeDetail.instanceTypes, stateUpdated.form)

      stateUpdated.title = 'Edit Node ' + nodeDetail.nodeName

      setState(stateUpdated)
    }
  }

  function getProductsOption() {
    if (products.length === 0) return []
    return products
      .sort((a, b) => {
        if (a.familyId < b.familyId) return -1
        if (a.familyId > b.familyId) return 1
        if (a.instanceType < b.instanceType) return -1
        if (a.instanceType > b.instanceType) return 1
        return 0
      })
      .filter((product) => product.instanceType !== undefined)
      .map((product) => {
        return {
          name: `${product.name} - ${product.familyDisplayName}`,
          value: product.name
        }
      })
  }

  async function onSubmit() {
    const updatedState = {
      ...state
    }
    const updateForm = { ...updatedState.form }
    const zone = nodeDetails.availabilityZone
    const region = nodeDetails.region
    const instanceTypes = getFormValue('instanceTypes', updateForm)
    const pools = getFormValue('pools', updateForm)

    const payload = {
      availabilityZone: zone,
      instanceTypesOverride: {
        overridePolicies: true,
        overrideValues: instanceTypes
      },
      computeNodePoolsOverride: {
        overridePolicies: true,
        overrideValues: pools
      },
      region
    }

    try {
      setShowModal(true)
      await NodePoolService.editNode(nodeDetails.nodeId, payload)
      setShowModal(false)
      navigate('/npm/nodes')
    } catch (error) {
      setShowModal(false)
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
  }

  function onCancel(location) {
    navigate(location)
  }

  function onSelectAll(formInputName) {
    const stateUpdated = {
      ...state
    }

    let options = []

    if (formInputName === 'pools') {
      options = poolList.map((x) => x.poolId)
    } else {
      options = getProductsOption().map((x) => x.value)
    }

    const selectedUserTypes = stateUpdated.form[formInputName].value
    const shouldDeselect = options.every((x) => selectedUserTypes.includes(x))
    onChangeDropdownMultiple(shouldDeselect ? [] : options, formInputName)
  }

  function onChangeDropdownMultiple(values, formInputName) {
    const updatedState = {
      ...state
    }

    updatedState.form = setFormValue(formInputName, values, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedState.form)

    setState(updatedState)
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    const inputValue = event.target.checked
    const updatedForm = UpdateFormHelper(inputValue, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function generateSelectOptions(data) {
    const result = []
    for (const e of data) {
      const option = { name: e, value: e }
      result.push(option)
    }
    return result
  }

  return (
    <NodeEdit
      state={state}
      nodeDetails={nodeDetails}
      showModal={showModal}
      onSubmit={onSubmit}
      onChangeDropdownMultiple={onChangeDropdownMultiple}
      onSelectAll={onSelectAll}
      onChangeInput={onChangeInput}
      apiCallsCompleted={apiCallsCompleted}
    />
  )
}

export default NodeEditContainer
