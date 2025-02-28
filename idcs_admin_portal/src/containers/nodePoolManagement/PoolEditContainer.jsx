import { React, useState, useEffect } from 'react'
import PoolEdit from '../../components/nodePoolManagement/PoolEdit'
import { useNavigate, useParams } from 'react-router-dom'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setFormValue
} from '../../utility/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import NodePoolService from '../../services/NodePoolService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useNodePoolStore from '../../store/nodePoolStore/NodePoolStore'

const PoolEditContainer = () => {
  // Navigation
  const navigate = useNavigate()
  const { poolId } = useParams()

  // Error Boundary
  const throwError = useErrorBoundary()

  // initial state
  const initialState = {
    title: poolId ? `Update existing pool (${poolId})` : 'Create New Pool',
    form: {
      poolName: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Compute Node Pool Name',
        placeholder: 'Pool Name',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        maxLength: 70,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '',
        helperMessage: ''
      },
      agsRole: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Pool Account Manager AGS Role',
        placeholder: 'AGS Role',
        value: '',
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        maxLength: 200,
        validationRules: {
          isRequired: false,
          checkMaxLength: true
        },
        validationMessage: '',
        helperMessage: '',
        hidden: true
      }
    },
    isValidForm: false,
    navigationTop: [
      {
        label: '⟵ Back to pools',
        buttonVariant: 'link',
        function: () => onCancel('/npm/pools')
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
        function: () => onCancel('/npm/pools')
      }
    ]
  }

  const [state, setState] = useState(initialState)
  const [showModal, setShowModal] = useState(false)
  const [nodeCount, setNodeCount] = useState(0)
  const [showAddNode, setShowAddNode] = useState(false)

  // Global Store
  const poolList = useNodePoolStore((state) => state.poolList)
  const setPoolList = useNodePoolStore((state) => state.setPoolList)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await setPoolList()
      } catch (error) {
        throwError(error)
      }
    }
    if (poolId) fetch()
  }, [])

  useEffect(() => {
    if (poolId && poolList.length > 0) setForm()
  }, [poolList])

  // functions
  function setForm() {
    const stateUpdated = { ...state }

    const poolDetails = poolList.find((pool) => pool.poolId === poolId)

    stateUpdated.form = setFormValue('poolName', poolDetails?.poolName, stateUpdated.form)
    stateUpdated.isValidForm = isValidForm(stateUpdated.form)

    setNodeCount(poolDetails?.numberOfNodes)

    setState(stateUpdated)
  }

  async function onSubmit() {
    const updatedState = { ...state }
    const updateForm = { ...updatedState.form }

    const poolName = getFormValue('poolName', updateForm)
    const poolAccountManagerAgsRole = ''

    const payload = {
      poolName,
      poolAccountManagerAgsRole
    }

    const pId = poolId ?? getFormValue('poolName', updateForm)

    try {
      setShowModal(true)
      await NodePoolService.createEditPool(pId, payload)
      setShowModal(false)
      navigate('/npm/pools')
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

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    const inputValue = event.target.value
    const updatedForm = UpdateFormHelper(inputValue, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function addNodeToPool() {
    setShowAddNode(true)
  }

  function cancelAddNode() {
    setShowAddNode(false)
    setShowModal('')
  }

  async function addNodeToPoolFn(payload, nodeDetail) {
    setShowModal(true)

    try {
      setShowModal(true)
      await NodePoolService.editNode(nodeDetail.nodeId, payload)

      await setPoolList()
      showSuccess('Node added to pool successfully.')
      setShowModal(false)
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
    } finally {
      cancelAddNode()
    }
  }

  return (
    <PoolEdit
      poolId={poolId}
      nodeCount={nodeCount}
      state={state}
      showModal={showModal}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      addNodeToPool={addNodeToPool}
      showAddNode={showAddNode}
      addNodeToPoolFn={addNodeToPoolFn}
      cancelAddNode={cancelAddNode}
    />
  )
}

export default PoolEditContainer
