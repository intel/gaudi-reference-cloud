import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import IMI from '../../components/imis/IMI'
import useIMIStore from '../../store/imiStore/IMIStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { UpdateFormHelper, isValidForm } from '../../utility/updateFormHelper/UpdateFormHelper'
import useInstanceTypeStore from '../../store/instanceTypeStore/InstanceTypeStore'
import IKSService from '../../services/IKSService'
import useToastStore from '../../store/toastStore/ToastStore'

const IMIFormInitialState = {
  name: {
    id: 'name',
    sectionGroup: 'imi_form',
    type: 'text',
    fieldSize: 'small',
    label: 'Name *',
    placeholder: 'Enter Name',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    maxLength: 63,
    validationRules: {
      isRequired: true
    },
    validationMessage: ''
  },
  upstreamReleaseName: {
    id: 'upstreamReleaseName',
    sectionGroup: 'imi_form',
    type: 'text',
    fieldSize: 'small',
    label: 'Upstream Release Name *',
    placeholder: 'Enter Upstream Release Name',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    validationMessage: ''
  },
  provider: {
    id: 'provider',
    sectionGroup: 'imi_form',
    type: 'select',
    fieldSize: 'small',
    label: 'Provider *',
    placeholder: 'Please Select Provider',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    options: [],
    validationMessage: ''
  },
  type: {
    id: 'type',
    sectionGroup: 'imi_form',
    type: 'select',
    fieldSize: 'small',
    label: 'Type *',
    placeholder: 'Please Select Type',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    options: [],
    validationMessage: ''
  },
  runtime: {
    id: 'runtime',
    sectionGroup: 'imi_form',
    type: 'select',
    fieldSize: 'small',
    label: 'Runtime *',
    placeholder: 'Please Select Runtime',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    options: [],
    validationMessage: ''
  },
  os: {
    id: 'os',
    sectionGroup: 'imi_form',
    type: 'select',
    fieldSize: 'small',
    label: 'OS *',
    placeholder: 'Please Select OS',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    options: [],
    validationMessage: ''
  },
  state: {
    id: 'state',
    sectionGroup: 'imi_form',
    type: 'select',
    fieldSize: 'small',
    label: 'State *',
    placeholder: 'Please Select State',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    options: [],
    validationMessage: ''
  },
  artifact: {
    id: 'artifact',
    sectionGroup: 'imi_form',
    type: 'text',
    fieldSize: 'small',
    label: 'Artifact *',
    placeholder: 'Enter Artifact',
    value: '',
    isValid: true,
    isTouched: false,
    isReadOnly: false,
    maxLength: 63,
    validationRules: {
      isRequired: false
    },
    validationMessage: ''
  },
  family: {
    id: 'family',
    sectionGroup: 'imi_form',
    type: 'text',
    fieldSize: 'small',
    label: 'Family',
    placeholder: 'Enter Family',
    value: '',
    isValid: true,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: false
    },
    validationMessage: ''
  },
  category: {
    id: 'category',
    sectionGroup: 'imi_form',
    type: 'select',
    fieldSize: 'small',
    label: 'Category',
    placeholder: 'Please Select Category',
    value: '',
    isValid: true,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: false
    },
    options: [],
    validationMessage: ''
  }
}

const IMIComponentFormFields = {
  name: {
    id: 'name',
    sectionGroup: 'imi_form',
    type: 'text',
    fieldSize: 'small',
    label: '',
    placeholder: 'Enter Name',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    maxLength: 63,
    validationRules: {
      isRequired: true
    },
    validationMessage: ''
  },
  version: {
    id: 'version',
    sectionGroup: 'imi_form',
    type: 'text',
    fieldSize: 'small',
    label: '',
    placeholder: 'Enter Version',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    validationMessage: ''
  },
  artifact: {
    id: 'artifact',
    sectionGroup: 'imi_form',
    type: 'text',
    fieldSize: 'small',
    label: '',
    placeholder: 'Enter Artifact',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    maxLength: 63,
    validationRules: {
      isRequired: true
    },
    validationMessage: ''
  }
}

const IMIFormStateValueMapper = {
  active: '1',
  archived: '2',
  staged: '3'
}

const IMIK8sFormInitialState = {
  name: {
    id: 'name',
    sectionGroup: 'k8s_form',
    type: 'text',
    fieldSize: 'small',
    label: 'IMI Name *',
    placeholder: 'Enter Name',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    maxLength: 63,
    validationRules: {
      isRequired: true
    },
    validationMessage: ''
  },
  upstreamReleaseName: {
    id: 'upstreamReleaseName',
    sectionGroup: 'k8s_form',
    type: 'text',
    fieldSize: 'small',
    label: 'K8s Version *',
    placeholder: 'Enter K8s Version',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    validationMessage: ''
  },
  provider: {
    id: 'provider',
    sectionGroup: 'k8s_form',
    type: 'select',
    fieldSize: 'small',
    label: 'Provider *',
    placeholder: 'Please Select Provider',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    options: [],
    validationMessage: ''
  },
  type: {
    id: 'type',
    sectionGroup: 'k8s_form',
    type: 'select',
    fieldSize: 'small',
    label: 'Type *',
    placeholder: 'Please Select Type',
    value: '',
    isValid: true,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: false
    },
    options: [],
    validationMessage: ''
  },
  runtime: {
    id: 'runtime',
    sectionGroup: 'k8s_form',
    type: 'select',
    fieldSize: 'small',
    label: 'Runtime *',
    placeholder: 'Please Select Runtime',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    options: [],
    validationMessage: ''
  },
  os: {
    id: 'os',
    sectionGroup: 'k8s_form',
    type: 'select',
    fieldSize: 'small',
    label: 'OS *',
    placeholder: 'Please Select OS',
    value: '',
    isValid: false,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: true
    },
    options: [],
    validationMessage: ''
  },
  family: {
    id: 'family',
    sectionGroup: 'k8s_form',
    type: 'text',
    fieldSize: 'small',
    label: 'Family',
    placeholder: 'Enter Family',
    value: '',
    isValid: true,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: false
    },
    validationMessage: ''
  },
  category: {
    id: 'category',
    sectionGroup: 'k8s_form',
    type: 'select',
    fieldSize: 'small',
    label: 'Category',
    placeholder: 'Please Select Category',
    value: '',
    isValid: true,
    isTouched: false,
    isReadOnly: false,
    validationRules: {
      isRequired: false
    },
    options: [],
    validationMessage: ''
  }
}

const IMIFormID = 'imi_form'

function IMIContainer() {
  const emptyGrid = {
    title: 'No IMI lists found',
    subTitle: 'There are currently no imis'
  }

  const emptyGridByFilter = {
    title: 'No IMIs found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  const instanceTypeEmptyGrid = {
    title: 'No Instance Type lists found matching family and category',
    subTitle: 'There are currently no instance types'
  }

  // IMI States
  const [gridItems, setGridItems] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [modalType, setModalType] = useState('imi') // 'imi' or 'k8s'

  // Modal States
  const [showModal, setShowModal] = useState(false)
  const [primaryLabel, setPrimaryLabel] = useState(null) // 'Edit', 'Create', 'Save'
  const [imiFormData, setImiFormData] = useState(IMIFormInitialState)
  const [imiComponentsData, setImiComponentsData] = useState([])
  const [isFormValid, setIsFormValid] = useState(false)
  const [formErrorMsg, setFormErrorMsg] = useState('')
  const [showConfirmationModal, setShowConfirmationModal] = useState(false)

  // K8s Modal States
  const [imiK8sFormData, setImiK8sFormData] = useState(IMIK8sFormInitialState)
  const [showK8sModal, setShowK8sModal] = useState(false)
  const [k8sFormErrorMsg, setK8sFormErrorMsg] = useState('')
  const [isK8sFormValid, setIsK8sFormValid] = useState(false)
  const [selectedInstanceTypes, setSelectedInstanceTypes] = useState([])

  // Instance Type States
  const [instanceTypesGridItems, setInstanceTypesGridItems] = useState([])
  const [instanceTypeEmptyGridObject, setInstanceTypeEmptyGridObject] = useState(null)

  // Auth Modal States
  const [showAuthModal, setShowAuthModal] = useState(false)
  const [showUnAuthModal, setShowUnAuthModal] = useState(false)
  const [authPassword, setAuthPassword] = useState('')

  // IMI Store
  const loading = useIMIStore((state) => state.loading)
  const stopLoading = useIMIStore((state) => state.stopLoading)
  const imisData = useIMIStore((state) => state.imisData)
  const getIMISData = useIMIStore((state) => state.getIMISData)
  const imiData = useIMIStore((state) => state.imiData)
  const clearIMIData = useIMIStore((state) => state.clearIMIData)
  const getIMIDataByID = useIMIStore((state) => state.getIMIDataByID)
  const createIMIData = useIMIStore((state) => state.createIMIData)
  const updateIMIData = useIMIStore((state) => state.updateIMIData)
  const deleteIMIData = useIMIStore((state) => state.deleteIMIData)
  const imiInfoData = useIMIStore((state) => state.imiInfoData)
  const getIMIInfoData = useIMIStore((state) => state.getIMIInfoData)
  const updateIMIK8sData = useIMIStore((state) => state.updateIMIK8sData)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // InstanceTypeStore
  const instanceTypesData = useInstanceTypeStore((state) => state.instanceTypesData)
  const getInstanceTypesData = useInstanceTypeStore((state) => state.getInstanceTypesData)

  // Error Boundary
  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()

  useEffect(() => {
    fetchIMISData()
  }, [])

  const debounceIMIListRefresh = (delay = 2000) => {
    setTimeout(() => {
      fetchIMISData(true)
    }, delay)
  }

  const fetchIMISData = async (isBackground) => {
    try {
      await getIMISData(isBackground)
    } catch (error) {
      throwError(error)
    } finally {
      stopLoading()
    }
  }

  useEffect(() => {
    setGridInfo()
  }, [imisData])

  const fetchIMISInfoData = async (isBackground = true) => {
    try {
      await getIMIInfoData(isBackground)
    } catch (error) {
      throwError(error)
    } finally {
      stopLoading()
    }
  }

  const fetchIMIDataByID = async (id, isBackground = true) => {
    try {
      await getIMIDataByID(id, isBackground)
    } catch (error) {
      throwError(error)
    } finally {
      stopLoading()
    }
  }

  const fetchInstanceTypesData = async (isBackground = true) => {
    try {
      await getInstanceTypesData(isBackground)
    } catch (error) {
      throwError(error)
    } finally {
      stopLoading()
    }
  }

  useEffect(() => {
    if (imiInfoData) {
      updateIMIFormDropdowns()
    }
  }, [imiInfoData])

  const updateIMIFormDropdowns = () => {
    const formCopy = { ...imiFormData }
    const k8sformCopy = { ...imiK8sFormData }

    const { provider, runtime, osimage, state } = imiInfoData

    const providerOptions = provider.map((data) => ({
      name: data,
      value: data
    }))

    const typeOptions = [
      {
        name: 'Control Plane',
        value: 'controlplane'
      },
      {
        name: 'Worker',
        value: 'worker'
      }
    ]

    const runtimeOptions = runtime.map((data) => ({
      name: data,
      value: data
    }))

    const osOptions = osimage.map((data) => ({
      name: data,
      value: data
    }))

    const stateOptions = state.map((data) => ({
      name: data,
      value: data
    }))

    const categoryOptions = [
      {
        name: 'Bare Metal Host',
        value: 'BareMetalHost'
      },
      {
        name: 'Virtual Machine',
        value: 'VirtualMachine'
      }
    ]

    formCopy.provider.options = providerOptions
    formCopy.type.options = typeOptions
    formCopy.runtime.options = runtimeOptions
    formCopy.os.options = osOptions
    formCopy.state.options = stateOptions
    formCopy.category.options = categoryOptions

    k8sformCopy.provider.options = providerOptions
    k8sformCopy.type.options = typeOptions
    k8sformCopy.runtime.options = runtimeOptions
    k8sformCopy.os.options = osOptions
    k8sformCopy.category.options = categoryOptions

    setImiFormData(formCopy)
    setImiK8sFormData(k8sformCopy)
  }

  const setGridInfo = () => {
    const items = []

    imisData.forEach((imi) => {
      const actionValues = [{ name: 'View', id: 'view' }]
      if (imi.type === 'worker' || (imi.type === 'controlplane' && !imi.iscompatabilityactiveimi)) {
        actionValues.push({ name: 'Update K8s', id: 'updatek8' })
      }

      items.push({
        name: imi.name,
        provider: imi.provider,
        type: imi.type,
        state: imi.state,
        version: imi.upstreamreleasename,
        artiface: imi.artifact,
        iscompatabilityactiveimi: imi.iscompatabilityactiveimi.toString(),
        actions: {
          showField: true,
          type: 'Buttons',
          value: imi,
          selectableValues: actionValues,
          function: (action, item) => {
            switch (action.id) {
              case 'view':
                openViewModal(true, item.name)
                break
              case 'updatek8':
                openUpdateK8sModal(true, item.name)
                break
              default:
                break
            }
          }
        }
      })
    })

    setGridItems(items)
  }

  useEffect(() => {
    // Update Grid Info everytime new Instance is selected to retain checked state on navigating to page 2,3..
    setK8sInstanceTypeGridInfo()
  }, [selectedInstanceTypes])

  const setK8sInstanceTypeGridInfo = () => {
    const items = []

    imiData && imiData.instacetypeimik8scompatibilityresponse.forEach((instanceType) => {
      items.push({
        actions: {
          showField: true,
          type: 'function',
          value: selectedInstanceTypes,
          function: (value) => {
            return (
              <input
                key={instanceType.instancetypename}
                className='form-check-input ml-2'
                type='checkbox'
                id={instanceType.instancetypename}
                defaultChecked={value.includes(instanceType.instancetypename)}
                name={'select-checkbox'}
                intc-id={'Select CheckboxTable'}
                onClick={() => {
                  const copy = [...selectedInstanceTypes]

                  const index = selectedInstanceTypes.indexOf(instanceType.instancetypename)
                  if (index === -1) {
                    copy.push(instanceType.instancetypename)
                  } else {
                    copy.splice(index, 1)
                  }

                  setSelectedInstanceTypes(copy)
                }}
              />
            )
          }
        },
        name: instanceType.instancetypename,
        family: instanceType.family,
        category: instanceType.category
      })
    })

    setInstanceTypesGridItems(items)
    setInstanceTypeEmptyGridObject(instanceTypeEmptyGrid)
  }

  const setInstanceTypeGridInfo = () => {
    const items = instanceTypesData.reduce((result, instanceType) => {
      if (
        instanceType.family === imiFormData.family.value &&
        instanceType.category === imiFormData.category.value
      ) {
        const gridInfo = {
          name: instanceType.instancetypename,
          family: instanceType.family,
          category: instanceType.category
        }

        result.push(gridInfo)
      }

      return result
    }, [])

    setInstanceTypesGridItems(items)
    setInstanceTypeEmptyGridObject(instanceTypeEmptyGrid)
  }

  useEffect(() => {
    if (imiData && imiData.name) {
      fetchIMISInfoData()
      if (modalType === 'imi') {
        openViewModal(false)
      } else {
        openUpdateK8sModal(false)
      }
    }
  }, [imiData])

  const openViewModal = (isNewData, id) => {
    if (isNewData) {
      getIMIDataByID(id, true)
      setModalType('imi')
    } else {
      // Logic to set the form values while editing
      let formCopy = structuredClone(imiFormData)

      for (const key in formCopy) {
        const value = imiData[key.toLowerCase()]

        if (value) {
          formCopy = { ...UpdateFormHelper(value, key, formCopy) }
        }
        formCopy[key].isReadOnly = true
      }

      // Logic to set the component form values while editing
      const components = []
      imiData.components.forEach((component) => {
        let newComponent = structuredClone(IMIComponentFormFields)

        for (const key in newComponent) {
          const value = component[key.toLowerCase()]

          newComponent = { ...UpdateFormHelper(value, key, newComponent) }
          newComponent[key].isReadOnly = true
        }

        components.push(newComponent)
      })

      // Logic to set Instance Types data
      const instanceTypeItems = []
      if (Array.isArray(imiData?.instanceTypeResponse)) {
        imiData.instanceTypeResponse.forEach((instanceType) => {
          const gridInfo = {
            name: instanceType.instancetypename,
            family: instanceType.family,
            category: instanceType.category,
            isTagged: !instanceType.iscompatabilityactiveinstance ? 'UnTagged' : 'Tagged'
          }

          instanceTypeItems.push(gridInfo)
        })
      }

      setImiFormData(formCopy)
      setImiComponentsData(components)
      setInstanceTypesGridItems(instanceTypeItems)
      setInstanceTypeEmptyGridObject(instanceTypeEmptyGrid)
      setPrimaryLabel('Edit')
      setShowModal(true)
    }
  }

  const closeModal = () => {
    // Close the Modal
    setShowModal(false)
    // Reset Primary Button Label
    setPrimaryLabel(null)
    // Reset Form Error Messsage
    setFormErrorMsg('')
    // Reset IMI Data in Store
    clearIMIData()
    // Reset Form Data
    setImiFormData(IMIFormInitialState)
    // Reset Component Form Data
    setImiComponentsData([])
    // Reset InstanceType Grid Data
    setInstanceTypesGridItems([])
    // Reset Empty Grid Object
    setInstanceTypeEmptyGridObject(null)
    // Reset Modal Type
    setModalType('imi')
  }

  const openCreateModal = () => {
    setPrimaryLabel('Create')
    fetchIMISInfoData()
    fetchInstanceTypesData()
    setShowModal(true)
  }

  const closeK8sModal = () => {
    // Close the K8s Modal
    setShowK8sModal(false)
    // Reset Form Error Messsage
    setK8sFormErrorMsg('')
    // Reset IMI Data in Store
    clearIMIData()
    // Reset Form Data
    setImiK8sFormData(IMIK8sFormInitialState)
    // Reset InstanceType Grid Data
    setInstanceTypesGridItems([])
    // Reset Empty Grid Object
    setInstanceTypeEmptyGridObject(null)
    // Reset Modal Type
    setModalType('imi')
    // Clear Selected Instance Types
    setSelectedInstanceTypes([])
  }

  const openUpdateK8sModal = (isNewData, id) => {
    if (isNewData) {
      fetchIMIDataByID(id)
      setModalType('k8s')
    } else {
      let formCopy = structuredClone(imiK8sFormData)

      for (const key in formCopy) {
        const value = imiData[key.toLowerCase()]

        if (value) {
          formCopy = { ...UpdateFormHelper(value, key, formCopy) }
        }

        formCopy[key].isReadOnly = true
      }

      setImiK8sFormData(formCopy)
      setK8sInstanceTypeGridInfo()
      setShowK8sModal(true)
    }
  }

  useEffect(() => {
    if (showK8sModal && !imiData.isk8sActive) {
      setTimeout(() => {
        setK8sFormErrorMsg('K8s Version is not yet deployed or Not Available to use')
      }, 500)
    }
  }, [showK8sModal])

  const editIMIForm = () => {
    setPrimaryLabel('Save')

    // Logic to enable Form Values
    const imiFormCopy = structuredClone(imiFormData)
    for (const key in imiFormCopy) {
      imiFormCopy[key].isReadOnly = false
    }
    imiFormCopy.name.isReadOnly = true
    imiFormCopy.artifact.isReadOnly = true

    // Logic to enable Components Values
    const componentFormCopy = structuredClone(imiComponentsData)
    componentFormCopy.forEach((component) => {
      for (const key in component) {
        component[key].isReadOnly = false
      }
    })

    setImiFormData(imiFormCopy)
    setImiComponentsData(componentFormCopy)
  }

  const onSubmitIMIData = async (isCreate) => {
    const payload = {
      artifact: imiFormData.artifact.value,
      components: [],
      name: imiFormData.name.value,
      os: imiFormData.os.value,
      provider: imiFormData.provider.value,
      runtime: imiFormData.runtime.value,
      state: IMIFormStateValueMapper[imiFormData.state.value.toLowerCase()],
      type: imiFormData.type.value,
      upstreamreleasename: imiFormData.upstreamReleaseName.value,
      family: imiFormData.family.value,
      category: imiFormData.category.value,
      iksadminkey: btoa(authPassword)
    }

    imiComponentsData.forEach((component) => {
      payload.components.push({
        name: component.name.value,
        version: component.version.value,
        artifact: component.artifact.value
      })
    })

    try {
      if (isCreate) {
        payload.artifact = imiFormData.name.value

        await createIMIData(payload, true)
        showSuccess('IMI Created Successfully')
      } else {
        delete payload.name
        await updateIMIData(imiData.name, payload, true)
        showSuccess('IMI Updated Successfully')
      }
      debounceIMIListRefresh()
      closeModal()
    } catch (error) {
      if (error.response?.status === 500) {
        setFormErrorMsg(error.response?.data?.message ?? '')
      } else {
        closeModal()
        throwError(error.response)
      }
    }
  }

  const onDeleteIMIData = async () => {
    try {
      const payload = {
        iksadminkey: btoa(authPassword)
      }

      await deleteIMIData(imiData.name, payload, true)
      showSuccess('IMI Deleted Successfully')
      debounceIMIListRefresh()
      closeModal()
    } catch (error) {
      if (error.response?.status === 500) {
        setFormErrorMsg(error.response?.data?.message ?? '')
      } else {
        closeModal()
        throwError(error.response)
      }
    }
  }

  const onSubmitK8sData = async () => {
    const payload = {
      upstreamreleasename: imiK8sFormData.upstreamReleaseName.value,
      provider: imiK8sFormData.provider.value,
      type: imiK8sFormData.type.value,
      runtime: imiK8sFormData.runtime.value,
      os: imiK8sFormData.os.value,
      state: imiData.state,
      family: imiK8sFormData.family.value,
      category: imiK8sFormData.category.value,
      instancetypes: [...selectedInstanceTypes]
    }

    try {
      await updateIMIK8sData(imiK8sFormData.name.value, payload, true)
      showSuccess('IMI K8s Updated Successfully')
      debounceIMIListRefresh()
      closeK8sModal()
    } catch (error) {
      if (error.response?.status === 500) {
        setK8sFormErrorMsg(error.response?.data?.message ?? '')
      } else {
        closeK8sModal()
        throwError(error.response)
      }
    }
  }

  useEffect(() => {
    // Update form validity everytime FormData or Components Data is updated
    let isValidComponentForm = true
    imiComponentsData.forEach((component) => {
      isValidComponentForm = isValidComponentForm ? isValidForm(component) : false
    })

    const isValidOptions = isValidForm(imiFormData)

    setIsFormValid(isValidOptions && isValidComponentForm)
    if (formErrorMsg !== '') {
      setFormErrorMsg('')
    }
  }, [imiFormData, imiComponentsData])

  useEffect(() => {
    // Update K8s Form Validity everytime FormData Data is updated
    const isValidOptions = isValidForm(imiK8sFormData)

    let isValidK8sForm = true
    if (imiK8sFormData.type.value === 'worker') {
      isValidK8sForm = (selectedInstanceTypes.length > 0)
    }

    let isK8sActive = true
    if (imiData && !imiData.isk8sActive) {
      isK8sActive = false
    }

    setIsK8sFormValid(isValidOptions && isValidK8sForm && isK8sActive)
    if (k8sFormErrorMsg !== '') {
      setK8sFormErrorMsg('')
    }
  }, [imiK8sFormData, selectedInstanceTypes])

  const onAddComponentClick = () => {
    const componentFormCopy = structuredClone(imiComponentsData)
    const newComponent = structuredClone(IMIComponentFormFields)

    componentFormCopy.push({ ...newComponent })

    setImiComponentsData(componentFormCopy)
  }

  const onRemoveComponentClick = (index) => {
    const componentFormCopy = structuredClone(imiComponentsData)

    componentFormCopy.splice(index, 1)

    setImiComponentsData(componentFormCopy)
  }

  const gridColumns = [
    {
      columnName: 'ID',
      targetColumn: 'name',
      width: '20rem'
    },
    {
      columnName: 'Provider',
      targetColumn: 'provider'
    },
    {
      columnName: 'Type',
      targetColumn: 'type'
    },
    {
      columnName: 'State',
      targetColumn: 'state'
    },
    {
      columnName: 'K8 Version',
      targetColumn: 'version'
    },
    {
      columnName: 'Artifact',
      targetColumn: 'artifact'
    },
    {
      columnName: 'K8s Active',
      targetColumn: 'iscompatabilityactiveimi'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'Buttons',
        behaviorFunction: null
      }
    }
  ]

  function backToHome() {
    navigate('/')
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

  const formEvents = {
    onChanged: (event, id) => {
      const imiFormCopy = structuredClone(imiFormData)

      const updatedForm = { ...UpdateFormHelper(event.target.value, id, imiFormCopy) }

      setImiFormData(updatedForm)
    }
  }

  const componentFormEvents = {
    onChanged: (event, id, index) => {
      const componentFormCopy = structuredClone(imiComponentsData)

      let updatedComponent
      componentFormCopy.forEach((component, entry) => {
        if (entry === index) {
          updatedComponent = UpdateFormHelper(event.target.value, id, component)
        }
      })

      componentFormCopy[index] = updatedComponent

      setImiComponentsData(componentFormCopy)
    }
  }

  const closeAuthModal = () => {
    // Close Auth Modal
    setShowAuthModal(false)
    // Reset Auth Password
    setAuthPassword('')
  }

  const formFooterEvents = {
    onPrimaryBtnClick: () => {
      if (primaryLabel === 'Edit') {
        editIMIForm()
      } else {
        setShowAuthModal(true)
      }
    },
    onSecondaryBtnClick: () => {
      closeModal()
    },
    onDangerBtnClick: () => {
      setShowAuthModal(true)
    }
  }

  const setUserAuthorization = () => {
    const payload = {
      iksadminkey: btoa(authPassword)
    }

    IKSService.authenticateIMIS(payload)
      .then(({ data }) => {
        if (data?.isAuthenticatedUser) {
          if (primaryLabel === 'Save') {
            onSubmitIMIData(false)
          } else if (primaryLabel === 'Create') {
            onSubmitIMIData(true)
          } else if (primaryLabel === 'Edit') {
            onDeleteIMIData()
          }
        } else {
          setShowUnAuthModal(true)
        }
        closeAuthModal()
      })
      .catch((error) => {
        closeAuthModal()
        throwError(error.response)
      })
  }

  const authFormFooterEvents = {
    onPrimaryBtnClick: () => {
      setUserAuthorization()
    },
    onSecondaryBtnClick: () => {
      closeAuthModal()
    }
  }

  useEffect(() => {
    setIMIModalVisbility(!(showAuthModal || showUnAuthModal))
  }, [showAuthModal, showUnAuthModal])

  const setIMIModalVisbility = (visibility = true) => {
    const element = document.getElementById(IMIFormID)
    if (element) {
      element.style.display = visibility ? 'block' : 'none'
    }
  }

  const k8sFormEvents = {
    onChanged: (event, id) => {
      const formCopy = structuredClone(imiK8sFormData)

      const updatedForm = { ...UpdateFormHelper(event.target.value, id, formCopy) }

      setImiK8sFormData(updatedForm)
    }
  }

  const k8sFormFooterEvents = {
    onPrimaryBtnClick: () => {
      onSubmitK8sData()
    },
    onSecondaryBtnClick: () => {
      closeK8sModal()
    }
  }

  const instanceTypeGridColumns = [
    {
      columnName: 'Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Family',
      targetColumn: 'family'
    },
    {
      columnName: 'Category',
      targetColumn: 'category'
    }
  ]

  if (modalType === 'k8s') {
    instanceTypeGridColumns.unshift({
      columnName: 'Select',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'checkbox',
        behaviorFunction: null
      }
    })
  }

  if (modalType === 'imi' && primaryLabel === 'Edit') {
    instanceTypeGridColumns.push({
      columnName: 'Tag Status',
      targetColumn: 'isTagged'
    })
  }

  return (
    <IMI
      gridData={{
        data: gridItems,
        columns: gridColumns,
        emptyGridObject,
        loading
      }}
      modalData={{
        id: IMIFormID,
        showModal,
        closeModal,
        primaryLabel,
        formErrorMsg,
        isFormValid,
        imiFormData,
        imiComponentsData,
        onAddComponentClick,
        onRemoveComponentClick,
        formEvents,
        componentFormEvents,
        formFooterEvents,
        showConfirmationModal,
        setShowConfirmationModal
      }}
      k8sModalData={{
        showK8sModal,
        closeK8sModal,
        k8sFormErrorMsg,
        isK8sFormValid,
        imiK8sFormData,
        k8sFormEvents,
        k8sFormFooterEvents
      }}
      instanceTypeGridData={{
        data: instanceTypesGridItems,
        columns: instanceTypeGridColumns,
        emptyGrid: instanceTypeEmptyGridObject,
        loading: false,
        onRefreshGridData: () => { setInstanceTypeGridInfo() }
      }}
      authModalData={{
        showAuthModal,
        closeAuthModal,
        authPassword,
        setAuthPassword,
        authFormFooterEvents,
        showUnAuthModal,
        setShowUnAuthModal
      }}
      backToHome={backToHome}
      filterText={filterText}
      setFilter={setFilter}
      openCreateModal={openCreateModal}
    />
  )
}

export default IMIContainer
