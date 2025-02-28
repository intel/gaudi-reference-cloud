import React, { useEffect, useState } from 'react'
import InstanceType from '../../components/instanceTypes/InstanceType'
import useInstanceTypeStore from '../../store/instanceTypeStore/InstanceTypeStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { UpdateFormHelper, isValidForm } from '../../utility/updateFormHelper/UpdateFormHelper'
import useIMIStore from '../../store/imiStore/IMIStore'
import { useNavigate } from 'react-router'
import IKSService from '../../services/IKSService'
import useProductStore from '../../store/productStore/ProductStore'
import useToastStore from '../../store/toastStore/ToastStore'

const InstanceTypesFormInitialState = {
  instancetypename: {
    id: 'instancetypename',
    sectionGroup: 'instanceType_form',
    type: 'text',
    fieldSize: 'small',
    label: 'Instance Type Name *',
    placeholder: 'Enter Instance Type Name',
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
  memory: {
    id: 'memory',
    sectionGroup: 'instanceType_form',
    type: 'number',
    fieldSize: 'small',
    label: 'Memory (GB) *',
    placeholder: 'Enter Memory (GB)',
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
  cpu: {
    id: 'cpu',
    sectionGroup: 'instanceType_form',
    type: 'number',
    fieldSize: 'small',
    label: 'CPU Cores *',
    placeholder: 'Enter CPU Cores',
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
  nodeprovidername: {
    id: 'nodeprovidername',
    sectionGroup: 'instanceType_form',
    type: 'select',
    fieldSize: 'small',
    label: 'Node Provider Name *',
    placeholder: 'Please Select Node Provider Name',
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
  storage: {
    id: 'storage',
    sectionGroup: 'instanceType_form',
    type: 'number',
    fieldSize: 'small',
    label: 'Storage (GB) *',
    placeholder: 'Enter Storage (GB)',
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
  status: {
    id: 'status',
    sectionGroup: 'instanceType_form',
    type: 'select',
    fieldSize: 'small',
    label: 'Status *',
    placeholder: 'Please Select Status',
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
  displayname: {
    id: 'displayname',
    sectionGroup: 'instanceType_form',
    type: 'text',
    fieldSize: 'small',
    label: 'Display Name *',
    placeholder: 'Enter Display Name',
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
  imioverride: {
    id: 'imioverride',
    sectionGroup: 'instanceType_form',
    type: 'checkbox',
    fieldSize: 'small',
    label: '',
    placeholder: 'Enter IMI Override',
    value: false,
    isValid: true,
    isTouched: false,
    isReadOnly: false,
    maxLength: 63,
    validationRules: {
      isRequired: false
    },
    options: [
      {
        name: 'IMI Override *',
        value: '0'
      }
    ],
    validationMessage: ''
  },
  description: {
    id: 'description',
    sectionGroup: 'instanceType_form',
    type: 'text',
    fieldSize: 'small',
    label: 'Description *',
    placeholder: 'Enter Description',
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
  category: {
    id: 'category',
    sectionGroup: 'instanceType_form',
    type: 'text',
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
    validationMessage: ''
  },
  family: {
    id: 'family',
    sectionGroup: 'instanceType_form',
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
  allowManualInsert: {
    id: 'allowManualInsert',
    sectionGroup: 'instanceType_form',
    type: 'checkbox',
    fieldSize: 'small',
    label: '',
    placeholder: '',
    value: false,
    isValid: true,
    isTouched: false,
    isReadOnly: false,
    maxLength: 63,
    validationRules: {
      isRequired: false
    },
    options: [
      {
        name: 'Manual Insert',
        value: '0'
      }
    ],
    validationMessage: ''
  }
}

const InstanceTypesFormStateValueMapper = {
  active: '1',
  archived: '2',
  staged: '3'
}

const instanceTypeFormID = 'instanceType_form'

function InstanceTypeContainer() {
  const emptyGrid = {
    title: 'No Instance Type lists found',
    subTitle: 'There are currently no instance types'
  }

  const emptyGridByFilter = {
    title: 'No Instance Types found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  const imiEmptyGrid = {
    title: 'No IMI lists found matching family and category',
    subTitle: 'There are currently no imis'
  }

  const computeInstanceEmptyGrid = {
    title: 'No Compute Instance Type lists found',
    subTitle: 'There are currently no Compute Instance Types'
  }

  const computeInstanceEmptyGridByFilter = {
    title: 'No Compute Instance Types found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => setComputeInstanceFilter('', true),
      label: 'Clear filters'
    }
  }

  const imiK8sEmptyGrid = {
    title: 'No Compatable IMIs for K8s found',
    subTitle: 'There are currently no compatable IMIs'
  }

  // Instance Type States
  const [gridItems, setGridItems] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)

  // Compute Instance Type States
  const [computeInstanceGridItems, setComputeInstanceGridItems] = useState([])
  const [computeInstanceFilterText, setComputeInstanceFilterText] = useState('')
  const [computeInstanceEmptyGridObject, setComputeInstanceEmptyGridObject] = useState(computeInstanceEmptyGrid)
  const [computeInstanceTypeData, setComputeInstanceTypeData] = useState(null)
  const [selectedComputeInstanceType, setSelectedComputeInstanceType] = useState('')
  const [allowedProducts, setAllowedProducts] = useState([])

  // Modal States
  const [showModal, setShowModal] = useState(false)
  const [primaryLabel, setPrimaryLabel] = useState(null) // 'Edit', 'Create', 'Save', 'Update'
  const [instanceTypeFormData, setInstanceTypeFormData] = useState(InstanceTypesFormInitialState)
  const [isFormValid, setIsFormValid] = useState(false)
  const [formErrorMsg, setFormErrorMsg] = useState('')
  const [showConfirmationModal, setShowConfirmationModal] = useState(false)

  // IMI K8s Modal States
  const [imiK8sGridItems, setImiK8sGridItems] = useState([])

  // K8s Modal States
  const [showK8sModal, setShowK8sModal] = useState(false)
  const [k8sFormErrorMsg, setK8sFormErrorMsg] = useState('')
  const [isK8sFormValid, setIsK8sFormValid] = useState(false)
  const [selectedIMIs, setSelectedIMIs] = useState({})
  const [k8sCompatableIMIs, setk8sCompatableIMIs] = useState([])
  const [selectedK8sCompatableName, setSelectedK8sCompatableName] = useState({})

  // IMI Grid States
  const [imiGridItems, setImiGridItems] = useState([])
  const [imiEmptyGridObject, setImiEmptyGridObject] = useState(null)

  // Auth Modal States
  const [showAuthModal, setShowAuthModal] = useState(false)
  const [showUnAuthModal, setShowUnAuthModal] = useState(false)
  const [authPassword, setAuthPassword] = useState('')

  // Instance Type Store
  const loading = useInstanceTypeStore((state) => state.loading)
  const stopLoading = useInstanceTypeStore((state) => state.stopLoading)
  const instanceTypesData = useInstanceTypeStore((state) => state.instanceTypesData)
  const getInstanceTypesData = useInstanceTypeStore((state) => state.getInstanceTypesData)
  const instanceTypeData = useInstanceTypeStore((state) => state.instanceTypeData)
  const getInstanceTypeDataByID = useInstanceTypeStore((state) => state.getInstanceTypeDataByID)
  const clearInstanceTypeData = useInstanceTypeStore((state) => state.clearInstanceTypeData)
  const createInstanceTypeData = useInstanceTypeStore((state) => state.createInstanceTypeData)
  const updateInstanceTypeData = useInstanceTypeStore((state) => state.updateInstanceTypeData)
  const deleteInstanceTypeData = useInstanceTypeStore((state) => state.deleteInstanceTypeData)
  const instanceTypesInfoData = useInstanceTypeStore((state) => state.instanceTypesInfoData)
  const getInstanceTypesInfoData = useInstanceTypeStore((state) => state.getInstanceTypesInfoData)
  const updateInstanceTypeK8sData = useInstanceTypeStore((state) => state.updateInstanceTypeK8sData)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // IMI Store
  const imisData = useIMIStore((state) => state.imisData)
  const getIMISData = useIMIStore((state) => state.getIMISData)

  // Product Store
  const products = useProductStore((state) => state.products)
  const setProducts = useProductStore((state) => state.setProducts)

  // Navigation
  const navigate = useNavigate()

  // Error Boundary
  const throwError = useErrorBoundary()

  useEffect(() => {
    fetchInstanceTypesData()
    fetchInstanceTypesInfoData()
  }, [])

  const debounceInstanceTypesListRefresh = (delay = 2000) => {
    setTimeout(() => {
      fetchInstanceTypesData(true)
    }, delay)
  }

  const fetchInstanceTypesData = async (isBackground) => {
    try {
      await getInstanceTypesData(isBackground)
    } catch (error) {
      throwError(error)
    } finally {
      stopLoading()
    }
  }

  const fetchInstanceTypeDataByID = async (id, isBackground) => {
    try {
      await getInstanceTypeDataByID(id, isBackground)
    } catch (error) {
      throwError(error)
    } finally {
      stopLoading()
    }
  }

  const fetchInstanceTypesInfoData = async (isBackground = true) => {
    try {
      await getInstanceTypesInfoData(isBackground)
    } catch (error) {
      throwError(error)
    } finally {
      stopLoading()
    }
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

  const fetchProductsData = async () => {
    try {
      await setProducts()
    } catch (error) {
      throwError(error)
    }
  }

  useEffect(() => {
    setGridInfo()
  }, [instanceTypesData])

  const setGridInfo = () => {
    const items = []

    instanceTypesData.forEach((instanceType) => {
      const actionValues = [{ name: 'View', id: 'view' }, { name: 'Update K8s', id: 'updatek8' }]

      items.push({
        instancetypename: instanceType.instancetypename,
        memory: instanceType.memory,
        cpu: instanceType.cpu,
        nodeprovidername: instanceType.nodeprovidername,
        storage: instanceType.storage,
        status: instanceType.status,
        displayname: instanceType.displayname,
        actions: {
          showField: true,
          type: 'Buttons',
          value: instanceType,
          selectableValues: actionValues,
          function: (action, item) => {
            switch (action.id) {
              case 'view':
                openViewModal(item.instancetypename)
                break
              case 'updatek8':
                openUpdateModal(item.instancetypename)
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

  const setImiGridInfo = () => {
    const items = imisData.reduce((result, imi) => {
      if (
        imi.family === instanceTypeFormData.family.value &&
        imi.category === instanceTypeFormData.category.value
      ) {
        result.push({
          name: imi.name,
          family: imi.family,
          category: imi.category
        })
      }

      return result
    }, [])

    setImiGridItems(items)
    setImiEmptyGridObject(imiEmptyGrid)
  }

  useEffect(() => {
    // Update Grid Info everytime new Compute Instance is selected to retain checked state on navigating to page 2,3..
    setComputeInstanceGridInfo()
  }, [selectedComputeInstanceType])

  const setComputeInstanceGridInfo = () => {
    const items = allowedProducts.map((computeInstanceType) => ({
      actions: {
        showField: true,
        type: 'function',
        value: selectedComputeInstanceType,
        function: (value) => {
          return (
            <input
              key={computeInstanceType.instancetypename}
              className='form-check-input ml-2'
              type='radio'
              id={computeInstanceType.instancetypename}
              defaultChecked={value === computeInstanceType.instancetypename}
              name={'select-checkbox'}
              intc-id={'Select RadioboxTable'}
              onClick={() => {
                setSelectedComputeInstanceType(computeInstanceType.instancetypename)
                setFormValuesOnCreate(computeInstanceType.instancetypename)
              }}
            />
          )
        }
      },
      instancetypename: computeInstanceType.instancetypename,
      memory: computeInstanceType.memory,
      cpu: computeInstanceType.cpu,
      storage: computeInstanceType.storage
    }))

    setComputeInstanceGridItems(items)
  }

  const setInstanceTypeFormDropdownValues = () => {
    const formCopy = { ...instanceTypeFormData }

    const { states = [], nodeprovidername = [] } = instanceTypesInfoData || {}

    const stateOptions = states.map((data) => ({
      name: data,
      value: data
    }))

    const nodeprovidernameOptions = nodeprovidername.map((data) => ({
      name: data,
      value: data
    }))

    formCopy.status.options = stateOptions
    formCopy.nodeprovidername.options = nodeprovidernameOptions

    setInstanceTypeFormData(formCopy)
  }

  useEffect(() => {
    if (instanceTypesInfoData) {
      setInstanceTypeFormDropdownValues()
    }
  }, [instanceTypesInfoData])

  useEffect(() => {
    const { computeResponse = [] } = instanceTypesInfoData || {}

    if (products.length > 0 && computeResponse.length > 0) {
      // Calculating Allowed Products to Create
      const items = computeResponse.filter((computeInstanceType) => (
        products.findIndex((x) => x.instanceType === computeInstanceType.instancetypename) > -1
      ))

      setAllowedProducts(items)
    }
  }, [products])

  useEffect(() => {
    setComputeInstanceGridInfo()
  }, [allowedProducts])

  const setFormValuesOnCreate = (instancetypename) => {
    // Logic to set the form values when selecting from compute list
    const computeInstanceTypeData = instanceTypesInfoData.computeResponse.find(
      (value) => value.instancetypename === instancetypename
    )

    let formCopy = structuredClone(instanceTypeFormData)

    for (const key in formCopy) {
      const value = computeInstanceTypeData[key.toLowerCase()]

      if (value) {
        formCopy = { ...UpdateFormHelper(value, key, formCopy) }
      }
    }

    setInstanceTypeFormData(formCopy)
  }

  useEffect(() => {
    if (
      instanceTypeData &&
      instanceTypeData.iksInstanceType &&
      instanceTypeData.iksInstanceType.instancetypename
    ) {
      if (primaryLabel === 'Edit') {
        setFormValuesOnView()
      } else if (primaryLabel === 'Update') {
        setInstacetypeIMIK8s()
      }
    }
  }, [instanceTypeData])

  const setFormValuesOnView = () => {
    // Logic to set the form values while editing
    let formCopy = structuredClone(instanceTypeFormData)

    for (const key in formCopy) {
      const value = instanceTypeData.iksInstanceType[key.toLowerCase()]

      if (value) {
        formCopy = { ...UpdateFormHelper(value, key, formCopy) }
      }
      formCopy[key].isReadOnly = true
    }

    const items = []
    if (Array.isArray(instanceTypeData.iksInstanceType?.imiResponse)) {
      instanceTypeData.iksInstanceType.imiResponse.forEach((imi) => {
        const gridInfo = {
          name: imi.name,
          cposimageinstances: imi.cposimageinstances[0],
          family: imi.family,
          category: imi.category,
          isTagged: !imi.iscompatabilityactiveimi ? 'UnTagged' : 'Tagged'
        }

        items.push(gridInfo)
      })
    }

    setInstanceTypeFormData(formCopy)
    setComputeInstanceTypeData(instanceTypeData.computeInstanceType)
    setImiGridItems(items)
    setImiEmptyGridObject(imiEmptyGrid)
    setShowModal(true)
  }

  const setInstacetypeIMIK8s = () => {
    const k8sCompatableGridData = []
    const copySelectedIMIS = { ...selectedIMIs }

    Array.isArray(instanceTypeData?.iksInstanceType?.instacetypeimik8scompatibilityresponse) &&
      instanceTypeData.iksInstanceType.instacetypeimik8scompatibilityresponse.forEach(imi => {
        copySelectedIMIS[imi.upstreamreleasename] = null

        const duplicateIndex = k8sCompatableGridData.findIndex((K8sIMI) => K8sIMI.upstreamreleasename === imi.upstreamreleasename)

        if (duplicateIndex === -1) {
          k8sCompatableGridData.push({ ...imi })
        } else {
          let imiName = k8sCompatableGridData[duplicateIndex].name

          if (Array.isArray(imiName)) {
            imiName.push(imi.name)
          } else {
            imiName = [imiName, imi.name]
          }

          k8sCompatableGridData[duplicateIndex].name = imiName
        }
      })

    setSelectedIMIs({ ...copySelectedIMIS })
    setk8sCompatableIMIs(k8sCompatableGridData)
  }

  useEffect(() => {
    if (k8sCompatableIMIs.length > 0) {
      setFormValuesOnUpdate(true)
    }
  }, [k8sCompatableIMIs])

  useEffect(() => {
    const selectedIMIsKeys = Object.keys(selectedIMIs)
    const selectedK8sCompatableNameKeys = Object.keys(selectedK8sCompatableName)

    if (instanceTypeData?.iksInstanceType) {
      // Update Grid Info everytime new IMI is selected to retain checked state on navigating to page 2,3..
      let isValidK8sForm = false
      if (selectedIMIsKeys.length > 0) {
        isValidK8sForm = !selectedIMIsKeys.some((key) => selectedIMIs[key] === null)
      }
      if (isValidK8sForm && selectedK8sCompatableNameKeys.length > 0) {
        isValidK8sForm = !selectedK8sCompatableNameKeys.some((key) => selectedK8sCompatableName[key] === null)
      }

      setIsK8sFormValid(isValidK8sForm)
      setK8sFormErrorMsg('')
      setFormValuesOnUpdate(false)
    }
  }, [selectedIMIs, selectedK8sCompatableName])

  const setFormValuesOnUpdate = (isFirstCall) => {
    const items = []
    const copySelectedK8sCompatableName = { ...selectedK8sCompatableName }

    k8sCompatableIMIs.forEach((imi) => {
      let imiName = imi.name
      if (Array.isArray(imiName)) {
        copySelectedK8sCompatableName[imi.upstreamreleasename] = null

        imiName = {
          showField: true,
          type: 'function',
          value: selectedK8sCompatableName,
          function: (value) => {
            const placeholder = 'Select Instance Name'

            const imiNameOptions = imi.name.map((name) => ({
              name,
              value: name
            }))

            return (
              <select
                className={'form-select-sm w-100'}
                onChange={(event) => {
                  const copy = { ...selectedK8sCompatableName }

                  copy[imi.upstreamreleasename] = event.target.value

                  setSelectedK8sCompatableName({ ...copy })
                }}
                value={value?.[imi.upstreamreleasename] || placeholder}
              >
                <option
                  value={placeholder}
                  disabled
                >
                  {placeholder}
                </option>

                {imiNameOptions.map((option, index) => (
                  <option
                    key={index}
                    value={option.value}
                  >
                    {option.name}
                  </option>
                ))}
              </select>
            )
          }
        }
      }

      items.push({
        name: imiName,
        upstreamreleasename: imi.upstreamreleasename,
        category: imi.category,
        cposimageinstances: {
          showField: true,
          type: 'function',
          value: selectedIMIs,
          function: (value) => {
            const placeholder = 'Select CP OS Instance'

            const cposimageinstancesOptions = imi.cposimageinstances.map((instance) => ({
              name: instance,
              value: instance
            }))

            return (
              <select
                className={'form-select-sm w-100'}
                onChange={(event) => {
                  const copy = { ...selectedIMIs }

                  copy[imi.upstreamreleasename] = event.target.value

                  setSelectedIMIs({ ...copy })
                }}
                value={value?.[imi.upstreamreleasename] || placeholder}
              >
                <option
                  value={placeholder}
                  disabled
                >
                  {placeholder}
                </option>

                {cposimageinstancesOptions.map((option, index) => (
                  <option
                    key={index}
                    value={option.value}
                  >
                    {option.name}
                  </option>
                ))}
              </select>
            )
          }
        }
      })
    })

    if (isFirstCall) {
      setSelectedK8sCompatableName({ ...copySelectedK8sCompatableName })
    }
    setImiK8sGridItems(items)
    setShowK8sModal(true)
  }

  const openCreateModal = () => {
    fetchIMISData()
    fetchInstanceTypesInfoData()
    fetchProductsData()
    setPrimaryLabel('Create')
    setShowModal(true)
  }

  const openViewModal = (id) => {
    setPrimaryLabel('Edit')
    fetchInstanceTypeDataByID(id, true)
  }

  const openUpdateModal = (id) => {
    setPrimaryLabel('Update')
    fetchInstanceTypeDataByID(id, true)
  }

  const closeModal = () => {
    // Close the Modal
    setShowModal(false)
    // Reset Primary Button Label
    setPrimaryLabel(null)
    // Reset Form Error Messsage
    setFormErrorMsg('')
    // Reset Instance Type Data in Store
    clearInstanceTypeData()
    // Reset Form Data
    setInstanceTypeFormData(InstanceTypesFormInitialState)
    // Reset IMI Grid Data
    setImiGridItems([])
    // Reset Empty Grid Object
    setImiEmptyGridObject(null)
    // Reset Compute Instance Type Data
    setComputeInstanceTypeData(null)
    // Reset Compute Instance Type Filter Text
    setComputeInstanceFilterText('')
    // Reset Compute Instance Grid Object
    setComputeInstanceEmptyGridObject(computeInstanceEmptyGrid)
  }

  const closeK8sModal = () => {
    // Close the K8s Modal
    setShowK8sModal(false)
    // Reset Form Error Messsage
    setK8sFormErrorMsg('')
    // Reset IMI K8s Grid Data
    setImiK8sGridItems([])
    // Reset K8s Modal Validity
    setIsK8sFormValid(false)
    // Clear Selected IMIs
    setSelectedIMIs({})
    // Reset Instance Type Data in Store
    clearInstanceTypeData()
    // Reset Modified K8s compatable IMIs
    setk8sCompatableIMIs([])
    // Reset Selected IMI Names
    setSelectedK8sCompatableName({})
  }

  const onEditInstanceTypeForm = () => {
    const formCopy = structuredClone(instanceTypeFormData)

    for (const key in formCopy) {
      formCopy[key].isReadOnly = false
    }
    formCopy.instancetypename.isReadOnly = true

    setInstanceTypeFormData(formCopy)
    setPrimaryLabel('Save')
  }

  const onSubmitInstanceTypesData = async (isCreate) => {
    let allowManualInsert = false
    if (isCreate) {
      allowManualInsert = instanceTypeFormData.allowManualInsert.value

      if (
        !allowManualInsert &&
        allowedProducts.findIndex((x) => x.instancetypename === instanceTypeFormData.instancetypename.value) === -1
      ) {
        setFormErrorMsg('Instance Type name does not match with any Compute Instance and Product Catalog Instance Type')
        return
      }
    } else {
      allowManualInsert = !(computeInstanceTypeData?.instancetypename)
    }

    const payload = {
      instancetypename: instanceTypeFormData.instancetypename.value,
      memory: Number(instanceTypeFormData.memory.value),
      cpu: Number(instanceTypeFormData.cpu.value),
      nodeprovidername: instanceTypeFormData.nodeprovidername.value,
      storage: Number(instanceTypeFormData.storage.value),
      status: InstanceTypesFormStateValueMapper[instanceTypeFormData.status.value.toLowerCase()],
      displayname: instanceTypeFormData.displayname.value,
      imioverride: instanceTypeFormData.imioverride.value,
      description: instanceTypeFormData.description.value,
      category: instanceTypeFormData.category.value,
      family: instanceTypeFormData.family.value,
      iksDB: true,
      allowManualInsert,
      iksadminkey: btoa(authPassword)
    }

    try {
      if (isCreate) {
        await createInstanceTypeData(payload, true)
        showSuccess('Instance Type Created Successfully')
      } else {
        delete payload.instancetypename
        await updateInstanceTypeData(instanceTypeData.iksInstanceType.instancetypename, payload, true)
        showSuccess('Instance Type Updated Successfully')
      }
      debounceInstanceTypesListRefresh()
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

  const onDeleteInstanceTypeData = async () => {
    try {
      const payload = {
        iksadminkey: btoa(authPassword)
      }

      await deleteInstanceTypeData(instanceTypeData.iksInstanceType.instancetypename, payload, true)
      showSuccess('Instance Type Deleted Successfully')
      debounceInstanceTypesListRefresh()
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
    const iksInstanceType = instanceTypeData.iksInstanceType

    const instacetypeimik8scompatibilityresponse = k8sCompatableIMIs.map(k8sCompatableIMI => {
      let imiName = k8sCompatableIMI.name
      if (Array.isArray(imiName)) {
        imiName = selectedK8sCompatableName[k8sCompatableIMI.upstreamreleasename]
      }

      return {
        name: imiName,
        artifact: k8sCompatableIMI.artifact,
        category: k8sCompatableIMI.category,
        family: k8sCompatableIMI.family,
        os: k8sCompatableIMI.os,
        provider: k8sCompatableIMI.provider,
        runtime: k8sCompatableIMI.runtime,
        type: k8sCompatableIMI.type,
        upstreamreleasename: k8sCompatableIMI.upstreamreleasename,
        cposimageinstances: selectedIMIs[k8sCompatableIMI.upstreamreleasename]
      }
    })

    const payload = {
      category: iksInstanceType.category,
      family: iksInstanceType.family,
      iksadminkey: btoa(authPassword),
      nodeprovidername: iksInstanceType.nodeprovidername,
      instacetypeimik8scompatibilityresponse
    }

    try {
      await updateInstanceTypeK8sData(iksInstanceType.instancetypename, payload)
      showSuccess('Instance Type K8s Updated Successfully')
      debounceInstanceTypesListRefresh()
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
    // Update form validity everytime FormData is updated
    const isValidOptions = isValidForm(instanceTypeFormData)

    setIsFormValid(isValidOptions)
    if (formErrorMsg !== '') {
      setFormErrorMsg('')
    }
  }, [instanceTypeFormData])

  const formEvents = {
    onChanged: (event, id) => {
      const formCopy = structuredClone(instanceTypeFormData)

      let value = event.target.value
      if (formCopy[id].type === 'checkbox') {
        value = event.target.checked
      }

      const updatedForm = UpdateFormHelper(value, id, formCopy)

      setInstanceTypeFormData(updatedForm)
    }
  }

  const formFooterEvents = {
    onPrimaryBtnClick: () => {
      if (primaryLabel === 'Edit') {
        onEditInstanceTypeForm()
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

  const closeAuthModal = () => {
    // Close Auth Modal
    setShowAuthModal(false)
    // Reset Auth Password
    setAuthPassword('')
  }

  const setUserAuthorization = () => {
    const payload = {
      iksadminkey: btoa(authPassword)
    }

    IKSService.authenticateIMIS(payload)
      .then(({ data }) => {
        if (data?.isAuthenticatedUser) {
          if (primaryLabel === 'Save') {
            onSubmitInstanceTypesData(false)
          } else if (primaryLabel === 'Create') {
            onSubmitInstanceTypesData(true)
          } else if (primaryLabel === 'Edit') {
            onDeleteInstanceTypeData()
          } else if (primaryLabel === 'Update') {
            onSubmitK8sData()
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

  const k8sFormFooterEvents = {
    onPrimaryBtnClick: () => {
      setShowAuthModal(true)
    },
    onSecondaryBtnClick: () => {
      closeK8sModal()
    }
  }

  useEffect(() => {
    setInstanceTypeModalVisbility(!(showAuthModal || showUnAuthModal))
  }, [showAuthModal, showUnAuthModal])

  const setInstanceTypeModalVisbility = (visibility = true) => {
    const element = document.getElementById(instanceTypeFormID)
    if (element) {
      element.style.display = visibility ? 'block' : 'none'
    }
  }

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

  function setComputeInstanceFilter(event, clear) {
    if (clear) {
      setComputeInstanceEmptyGridObject(computeInstanceEmptyGrid)
      setComputeInstanceFilterText('')
    } else {
      setComputeInstanceEmptyGridObject(computeInstanceEmptyGridByFilter)
      setComputeInstanceFilterText(event.target.value)
    }
  }

  const gridColumns = [
    {
      columnName: 'Instance Name',
      targetColumn: 'instancetypename'
    },
    {
      columnName: 'Memory (GB)',
      targetColumn: 'memory'
    },
    {
      columnName: 'CPU Cores',
      targetColumn: 'cpu'
    },
    {
      columnName: 'Node Provider',
      targetColumn: 'nodeprovidername'
    },
    {
      columnName: 'Storage (GB)',
      targetColumn: 'storage'
    },
    {
      columnName: 'Status',
      targetColumn: 'status'
    },
    {
      columnName: 'Display Name',
      targetColumn: 'displayname',
      width: '25rem'
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

  const imiGridColumns = [
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

  if (primaryLabel === 'Edit') {
    imiGridColumns.push({
      columnName: 'Tag Status',
      targetColumn: 'isTagged'
    })

    imiGridColumns.splice(1, 0, {
      columnName: 'CP OS Image Instance',
      targetColumn: 'cposimageinstances'
    })
  }

  const computeInstanceGridColumns = [
    {
      columnName: 'Select',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'checkbox',
        behaviorFunction: null
      }
    },
    {
      columnName: 'Instance Name',
      targetColumn: 'instancetypename'
    },
    {
      columnName: 'Memory (GB)',
      targetColumn: 'memory'
    },
    {
      columnName: 'CPU Cores',
      targetColumn: 'cpu'
    },
    {
      columnName: 'Storage (GB)',
      targetColumn: 'storage'
    }
  ]

  const imiK8sGridColumns = [
    {
      columnName: 'Name',
      targetColumn: 'name'
    },
    {
      columnName: 'K8s Version',
      targetColumn: 'upstreamreleasename'
    },
    {
      columnName: 'Category',
      targetColumn: 'category'
    },
    {
      columnName: 'CP OS Image Instances',
      targetColumn: 'cposimageinstances',
      columnConfig: {
        behaviorType: 'dropdown',
        behaviorFunction: null
      }
    }
  ]

  return (
    <InstanceType
      gridData={{
        data: gridItems,
        columns: gridColumns,
        emptyGridObject,
        loading
      }}
      modalData={{
        id: instanceTypeFormID,
        showModal,
        closeModal,
        primaryLabel,
        formErrorMsg,
        isFormValid,
        instanceTypeFormData,
        computeInstanceFilterText,
        setComputeInstanceFilter,
        formEvents,
        formFooterEvents,
        showConfirmationModal,
        setShowConfirmationModal
      }}
      imiGridData={{
        data: imiGridItems,
        columns: imiGridColumns,
        emptyGrid: imiEmptyGridObject,
        loading: false,
        onRefreshGridData: () => {
          setImiGridInfo()
        }
      }}
      computeInstanceTypeData={computeInstanceTypeData}
      computeInstanceGridData={{
        data: computeInstanceGridItems,
        columns: computeInstanceGridColumns,
        emptyGrid: computeInstanceEmptyGridObject,
        loading: false
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
      k8sModalData={{
        showK8sModal,
        closeK8sModal,
        k8sFormErrorMsg,
        isK8sFormValid,
        k8sFormFooterEvents
      }}
      iksInstanceTypeData={instanceTypeData?.iksInstanceType}
      imiK8sGridData={{
        data: imiK8sGridItems,
        columns: imiK8sGridColumns,
        emptyGrid: imiK8sEmptyGrid,
        loading: false
      }}
      backToHome={backToHome}
      filterText={filterText}
      setFilter={setFilter}
      openCreateModal={openCreateModal}
    />
  )
}

export default InstanceTypeContainer
