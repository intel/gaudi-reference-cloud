// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import useProductV2Store, { type productMetaDataSet, type productMetaData } from '../../store/productV2Store/ProductV2Store'
import ProductCatalogCreateMetaDataset from '../../components/productCatalog/ProductCatalogCreateMetaDataset'
import useFamilyStore from '../../store/familyStore/FamilyStore'
import { setSelectOptions, UpdateFormHelper, isValidForm, showFormRequiredFields, getFormValue, setFormValue } from '../../utility/updateFormHelper/UpdateFormHelper'
import { BsPencilFill, BsTrash3 } from 'react-icons/bs'
import useToastStore from '../../store/toastStore/ToastStore'

const ProductCatalogCreateMetaDatasetContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const newProduct = useProductV2Store((state) => state.newProduct)
  const setNewProduct = useProductV2Store((state) => state.setNewProduct)
  const families = useFamilyStore((state) => state.families)
  const productsByFamily = useProductV2Store((state) => state.productsByFamily)
  const getProductByFamily = useProductV2Store((state) => state.getProductByFamily)
  const showSuccess = useToastStore((state) => state.showSuccess)
  // *****
  // local state
  // *****
  const navigate = useNavigate()

  const productColumns = [
    {
      columnName: '',
      targetColumn: 'productId',
      hideField: true
    },
    {
      columnName: 'Product name',
      targetColumn: 'name'
    },
    {
      columnName: 'Service',
      targetColumn: 'fields'
    },
    {
      columnName: 'Service',
      targetColumn: 'serviceName',
      showOnBreakpoint: true
    }
  ]

  const metadataSetColumns = [
    {
      columnName: '',
      targetColumn: 'id',
      hideField: true
    },
    {
      columnName: 'Metadataset name',
      targetColumn: 'name',
      hideField: true
    },
    {
      columnName: 'Context',
      targetColumn: 'context'
    }
  ]

  const metadataColumns = [
    {
      columnName: '',
      targetColumn: 'id',
      hideField: true
    },
    {
      columnName: 'Key',
      targetColumn: 'key'
    },
    {
      columnName: 'Value',
      targetColumn: 'value'
    },
    {
      columnName: 'Type',
      targetColumn: 'type'
    },
    {
      columnName: '',
      targetColumn: 'contextId',
      hideField: true
    },
    {
      columnName: 'Context',
      targetColumn: 'context'
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

  const addFormMetaSetStateInitial = {
    title: 'Add product information',
    form: {
      family: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Family',
        placeholder: 'Please select family',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: <>
          Choose the desired family
        </>,
        group: 'metadata'
      },
      productId: {
        type: 'grid',
        gridBreakpoint: 'xs',
        maxWidth: '100%',
        label: 'Metasets: ',
        hiddenLabel: true,
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        columns: productColumns,
        validationMessage: '',
        singleSelection: true,
        selectedRecords: [] as any[],
        setSelectedRecords: null,
        emptyGridMessage: {
          title: 'No available products',
          subTitle: 'Please choose a different family'
        },
        hidePaginationControl: false,
        group: 'metadata'
      }
    },
    isValidForm: false,
    actionsModal: [
      {
        buttonAction: 'close',
        buttonLabel: 'Cancel',
        buttonVariant: 'link'
      },
      {
        buttonAction: 'applyTemplate',
        buttonLabel: 'Apply',
        buttonVariant: 'primary'
      }
    ]
  }

  const addFormMetaSetManualStateInitial = {
    form: {
      name: {
        type: 'text', // options = 'text ,'textArea'
        label: 'Name:',
        placeholder: 'name',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 200,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: <>Please provide a name.</>,
        hidden: false
      },
      context: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Context:',
        placeholder: 'Please select',
        value: '',
        isValid: true,
        isTouched: false,
        isReadOnly: true,
        validationRules: {
          isRequired: true
        },
        options: [
          {
            name: 'Product',
            value: 'product'
          },
          {
            name: 'Service',
            value: 'service'
          }
        ],
        validationMessage: '',
        helperMessage: <>
          Choose the desired context
        </>
      }
    },
    isValidForm: false,
    actionsModal: [
      {
        buttonAction: 'close',
        buttonLabel: 'Cancel',
        buttonVariant: 'link'
      },
      {
        buttonAction: 'applyMetaDataSet',
        buttonLabel: 'Apply',
        buttonVariant: 'primary'
      }
    ]
  }

  const addFormMetaDataManualStateInitial = {
    form: {
      keyMetadata: {
        type: 'text', // options = 'text ,'textArea'
        label: 'Key:',
        placeholder: 'Key',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 200,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: <>Please provide a key.</>,
        hidden: false
      },
      typeMetadata: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Type:',
        placeholder: 'Please select',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [
          {
            name: 'String',
            value: 'string'
          }
        ],
        validationMessage: '',
        helperMessage: <>
          Please create a metadataset
        </>
      },
      valueMetadata: {
        type: 'text', // options = 'text ,'textArea'
        label: 'Value:',
        placeholder: 'Value',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 200,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: <>Please provide a value.</>,
        hidden: false
      },
      contextMetadata: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Context:',
        placeholder: 'Please select',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: <>
          Please create a metadataset
        </>
      }
    },
    isValidForm: false,
    actionsModal: [
      {
        buttonAction: 'close',
        buttonLabel: 'Cancel',
        buttonVariant: 'link'
      },
      {
        buttonAction: 'applyMetaData',
        buttonLabel: 'Apply',
        buttonVariant: 'primary'
      }
    ]
  }

  const generalFormOptions = [
    {
      buttonAction: 'cancelForm',
      buttonLabel: 'Cancel',
      buttonVariant: 'link'
    },
    {
      buttonAction: 'save',
      buttonLabel: 'Apply',
      buttonVariant: 'primary'
    }
  ]

  const discardOptions = [
    {
      buttonAction: 'cancelDiscard',
      buttonLabel: 'No',
      buttonVariant: 'link'
    },
    {
      buttonAction: 'applyDiscard',
      buttonLabel: 'Yes',
      buttonVariant: 'primary'
    }
  ]

  const deleteMetaOptions = [
    {
      buttonAction: 'cancel',
      buttonLabel: 'Cancel',
      buttonVariant: 'link'
    },
    {
      buttonAction: 'deleteMetaDatarow',
      buttonLabel: 'Delete',
      buttonVariant: 'primary'
    }
  ]

  const copyTemplateModalInitial = {
    show: false,
    title: 'Copy metadata sets from template'
  }

  const addManualModalInitial = {
    show: false,
    title: 'Add metadata sets'
  }

  const addManualMetadataModalInitial = {
    show: false,
    title: 'Add metadata'
  }

  const discardChangesModalInitial = {
    show: false,
    title: 'Discard changes',
    question: 'Are you sure you want to discard your unsaved changes?'
  }

  const deleteManualMetadataModalInitial = {
    show: false,
    question: 'Do you want to delete the row?',
    title: 'Delete metadata'
  }

  const actionsOptions = [
    {
      id: 'edit',
      name: <>
        <BsPencilFill /> Edit{' '}
      </>
    },
    {
      id: 'delete',
      name: <>
        <BsTrash3 /> Delete{' '}
      </>
    }
  ]

  const metaDataSetEmptyGridInitial = {
    title: 'No metadata sets found',
    subTitle: ''
  }

  const metaDataEmptyGrid = {
    title: 'No metadata found',
    subTitle: ''
  }

  const [addFormMetaSetState, setAddFormMetaSetState] = useState(addFormMetaSetStateInitial)
  const [addFormMetaSetManualState, setAddFormMetaSetManualState] = useState(addFormMetaSetManualStateInitial)
  const [addFormMetaDataManualState, setAddFormMetaDataManualState] = useState(addFormMetaDataManualStateInitial)
  const [copyTemplateModal, setCopyTemplateModal] = useState(copyTemplateModalInitial)
  const [addManualModal, setAddManualModal] = useState(addManualModalInitial)
  const [addManualMetadataModal, setAddManualMetadatalModal] = useState(addManualMetadataModalInitial)
  const [deleteManualMetadataModal, setDeleteManualMetadatalModal] = useState(deleteManualMetadataModalInitial)
  const [discardChangesModal, setDiscardChangesModal] = useState(discardChangesModalInitial)
  const [selectedFamily, setSelectedFamily] = useState('')
  const [selectedProductNames, setSelectedProductNames] = useState<any[]>([])
  const [selectedMetaDataSet, setSelectedMetaDataSet] = useState<any[]>([])
  const [showMetadataMessage, setShowMetadataMessage] = useState(false)
  const [metaDataSetItems, setMetaDataSetItems] = useState<any[]>([])
  const [metaDataSetItem, setMetaDataSetItem] = useState<any>(null)
  const [productMetaDataItems, setProductMetaDataItems] = useState<any[]>([])
  const [metaDataItems, setMetaDataItems] = useState<any[]>([])
  const [metaDataItem, setMetaDataItem] = useState<any>(null)
  const [filterMetadatasetText, setFilterMetadatasetText] = useState('')
  const [filterMetadataText, setFilterMetadataText] = useState('')
  const [metaDataSetEmptyGrid, setMetaDataSetEmptyGrid] = useState(metaDataSetEmptyGridInitial)
  const [isFormTouched, setIsFormTouched] = useState(false)

  // *****
  // Hooks
  // *****
  useEffect(() => {
    if (!newProduct) {
      navigate('/products/create')
    }
    updateMetaDataSetGridFromStore()
  }, [])

  useEffect(() => {
    if (selectedFamily) {
      getProductByFamily(selectedFamily)
    }
  }, [selectedFamily])

  useEffect(() => {
    if (selectedFamily) {
      loadProductGrid()
    }
  }, [productsByFamily])

  useEffect(() => {
    if (selectedProductNames.length > 0) {
      const productName = selectedProductNames.length > 0 ? selectedProductNames[0] : ''
      const event = { target: { value: productName } }
      onChangeInput(event, 'productId', 'formTemplate')
    }
  }, [selectedProductNames])

  useEffect(() => {
    updateMetaDataGridFromTemplate()
  }, [selectedMetaDataSet, productMetaDataItems])

  // *****
  // functions
  // *****
  const loadMetadataForm = (): void => {
    const familiesItems: any = []
    families?.forEach((family) => {
      familiesItems.push({
        name: family.name,
        value: family.name
      })
    })
    const stateUpdate = { ...addFormMetaSetStateInitial }
    const form = setSelectOptions('family', familiesItems, stateUpdate.form)
    stateUpdate.form = form
    setAddFormMetaSetState(stateUpdate)
  }

  const onChangeInput = (event: any, formInputName: string, formType: string): void => {
    const value = event.target.value
    switch (formType) {
      case 'formTemplate': {
        const updatedState = { ...addFormMetaSetState }
        const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)
        updatedForm.productId.selectedRecords = selectedProductNames
        updatedState.form = updatedForm
        if (formInputName === 'family') {
          setSelectedFamily(value)
          setShowMetadataMessage(false)
        }
        updatedState.isValidForm = isValidForm(updatedForm)
        setAddFormMetaSetState(updatedState)
        break
      }
      case 'addmetadasetForm': {
        const updatedState = { ...addFormMetaSetManualState }
        const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)
        updatedState.form = updatedForm
        updatedState.isValidForm = isValidForm(updatedForm)
        setAddFormMetaSetManualState(updatedState)
        break
      }
      case 'addmetadaForm': {
        const updatedState = { ...addFormMetaDataManualState }
        const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)
        updatedState.form = updatedForm
        updatedState.isValidForm = isValidForm(updatedForm)
        setAddFormMetaDataManualState(updatedState)
        break
      }
      default:
        break
    }
  }

  const loadProductGrid = (): void => {
    const productOptions: any = []
    productsByFamily?.forEach((product) => {
      productOptions.push({
        productId: product.id,
        name: {
          showField: true,
          type: 'function',
          canSelectRow: true,
          value: product,
          function: (value: any) => {
            return (
              <div className="d-flex flex-column">
                <div>
                  <span className="fw-semibold text-nowrap me-s3">{value.name.toUpperCase()}</span>
                </div>
              </div>
            )
          }
        },
        fields: product.serviceName,
        serviceName: product.serviceName
      })
    })
    const stateUpdate = { ...addFormMetaSetState }
    const form = setSelectOptions('productId', productOptions, stateUpdate.form)
    form.productId.selectedRecords = selectedProductNames
    form.productId.setSelectedRecords = setSelectedProductNames
    stateUpdate.form = form
    setAddFormMetaSetState(stateUpdate)
  }

  const updateMetaDataSetGridFromStore = (): void => {
    const gridInfo: any = []
    const metaDataItems: any = []
    for (const index in newProduct?.metaDataSets) {
      const metaDataSet = { ...newProduct?.metaDataSets[Number(index)] }
      gridInfo.push({
        id: metaDataSet.id,
        name: metaDataSet.name,
        context: metaDataSet.context
      })
      for (const indexMeta in metaDataSet.metaData) {
        const metaData = { ...metaDataSet.metaData[Number(indexMeta)] }
        metaDataItems.push({
          id: metaData.id,
          key: metaData.key,
          value: metaData.value,
          type: metaData.type,
          context: index
        })
      }
    }
    setProductMetaDataItems(metaDataItems)
    setMetaDataSetItems(gridInfo)
    if (gridInfo.length > 0) {
      setSelectedMetaDataSet([gridInfo[0].id])
    }
  }

  const updateMetaDataSetGridFromTemplate = (): void => {
    const gridInfo = [...metaDataSetItems]
    const metaDataItems: any = []
    const product = productsByFamily?.find((product) => product.id === selectedProductNames[0])
    for (const index in product?.metaDataSets) {
      const metaDataSet = { ...product?.metaDataSets[Number(index)] }
      metaDataSet.id = index
      gridInfo.push({
        id: index,
        name: metaDataSet.name,
        context: metaDataSet.context
      })
      if (metaDataSet.metaData) {
        metaDataSet.metaData.forEach(item => {
          const metaData = { ...item }
          metaDataItems.push({
            id: metaData.id,
            key: metaData.key,
            value: metaData.value,
            type: metaData.type,
            context: index
          })
        })
      }
    }
    setProductMetaDataItems(metaDataItems)
    setMetaDataSetItems(gridInfo)
    if (gridInfo.length > 0) {
      setSelectedMetaDataSet([gridInfo[0].id])
    }
    setCopyTemplateModal({ ...copyTemplateModal, show: false })
    setAddFormMetaSetState(addFormMetaSetStateInitial)
    setIsFormTouched(true)
  }

  const updateMetaDataGridFromTemplate = (): void => {
    const metaDataSelectedItems = productMetaDataItems.filter((metaData) => selectedMetaDataSet.includes(metaData.context))
    const metaDataItems: any = []
    metaDataSelectedItems.forEach(item => {
      const metaData = { ...item }
      const metaDataSet = metaDataSetItems.find((item) => item.id === metaData.context)
      metaData.contextId = metaDataSet.id
      metaData.context = metaDataSet.context
      metaDataItems.push({
        id: metaData.id,
        key: metaData.key,
        value: metaData.value,
        type: metaData.type,
        contextId: metaDataSet.id,
        context: metaDataSet.context,
        actions: {
          showField: true,
          type: 'Buttons',
          value: metaData,
          selectableValues: actionsOptions,
          function: setActionMetadata
        }
      })
    })
    setMetaDataItems(metaDataItems)
  }

  const updateMetadataSetGridFromManual = (): void => {
    if (metaDataSetItem) {
      editMetaDataSetItem()
    }
  }

  const updateMetadataGridFromManual = (): void => {
    if (metaDataItem) {
      editMetaDataItem()
    } else {
      addMetaDataItem()
    }
  }

  const deleteMetadataGridFromManual = (): void => {
    const updatedProductMetaData: any[] = []
    productMetaDataItems.forEach((meta) => {
      if (meta.id !== metaDataItem.id) {
        updatedProductMetaData.push(meta)
      }
    })
    setIsFormTouched(true)
    setProductMetaDataItems(updatedProductMetaData)
  }

  const updateMedataForm = (formUpdated: any): void => {
    const copyMetaDataSets = [...metaDataSetItems]
    const options: any = []
    copyMetaDataSets.forEach((item) => {
      options.push({
        name: item.context,
        value: item.id
      })
    })
    const stateUpdate = { ...formUpdated }
    const form = setSelectOptions('contextMetadata', options, stateUpdate.form)
    stateUpdate.form = form
    setAddFormMetaDataManualState(stateUpdate)
  }

  const addMetaDataItem = (): void => {
    const productMetaDataItemsUpdated = [...productMetaDataItems]
    productMetaDataItemsUpdated.push({
      id: `newItem-${productMetaDataItemsUpdated.length + 1}`,
      key: getFormValue('keyMetadata', addFormMetaDataManualState.form),
      value: getFormValue('valueMetadata', addFormMetaDataManualState.form),
      context: getFormValue('contextMetadata', addFormMetaDataManualState.form),
      type: getFormValue('typeMetadata', addFormMetaDataManualState.form)
    })
    setIsFormTouched(true)
    setProductMetaDataItems(productMetaDataItemsUpdated)
  }

  const editMetaDataSetItem = (): void => {
    const updatedGrid: any[] = []
    metaDataSetItems.forEach(item => {
      const metaDataSet = { ...item }
      if (metaDataSet.id === metaDataSetItem.id) {
        metaDataSet.name = getFormValue('name', addFormMetaSetManualState.form)
        metaDataSet.context = getFormValue('context', addFormMetaSetManualState.form)
        metaDataSet.actions = {
          showField: true,
          type: 'Buttons',
          value: metaDataSet,
          selectableValues: actionsOptions,
          function: setActionMetadataSet
        }
      }
      updatedGrid.push(metaDataSet)
    })
    setIsFormTouched(true)
    setMetaDataSetItems(updatedGrid)
  }

  const editMetaDataItem = (): void => {
    const updatedProductMetaData: any[] = []
    productMetaDataItems.forEach((meta) => {
      if (meta.id === metaDataItem.id) {
        const rowUpdated = { ...meta }
        rowUpdated.context = getFormValue('contextMetadata', addFormMetaDataManualState.form)
        rowUpdated.key = getFormValue('keyMetadata', addFormMetaDataManualState.form)
        rowUpdated.value = getFormValue('valueMetadata', addFormMetaDataManualState.form)
        rowUpdated.type = getFormValue('typeMetadata', addFormMetaDataManualState.form)
        updatedProductMetaData.push(rowUpdated)
      } else {
        updatedProductMetaData.push(meta)
      }
    })
    setIsFormTouched(true)
    setProductMetaDataItems(updatedProductMetaData)
  }

  const setFilterMetaDataset = (event: any, clear: boolean): void => {
    setMetaDataSetEmptyGrid(metaDataSetEmptyGridInitial)
    if (clear) {
      setFilterMetadatasetText('')
    } else {
      setFilterMetadatasetText(event.target.value)
    }
  }

  const setFilterMetaData = (event: any, clear: boolean): void => {
    setMetaDataSetEmptyGrid(metaDataEmptyGrid)
    if (clear) {
      setFilterMetadataText('')
    } else {
      setFilterMetadataText(event.target.value)
    }
  }

  const setActionMetadata = (action: any, item: any): void => {
    switch (action.id) {
      case 'edit': {
        setMetaDataItem(item)
        const updateState = { ...addFormMetaDataManualState }
        let form = { ...updateState.form }
        form = setFormValue('keyMetadata', item.key, form)
        form = setFormValue('valueMetadata', item.value, form)
        form = setFormValue('contextMetadata', item.contextId, form)
        form = setFormValue('typeMetadata', item.type, form)
        updateState.form = form
        updateMedataForm(updateState)
        setAddManualMetadatalModal({ ...addManualMetadataModal, show: true })
        break
      }
      case 'delete': {
        setMetaDataItem(item)
        setDeleteManualMetadatalModal({ ...deleteManualMetadataModal, show: true })
        break
      }
      default:
        break
    }
  }

  const setActionMetadataSet = (action: any, item: any): void => {
    switch (action.id) {
      case 'edit': {
        setMetaDataSetItem(item)
        const updateState = { ...addFormMetaSetManualState }
        let form = { ...updateState.form }
        form = setFormValue('name', item.name, form)
        form = setFormValue('context', item.context, form)
        updateState.form = form
        setAddFormMetaSetManualState(updateState)
        setAddManualModal({ ...addManualModal, show: true })
        break
      }
      default:
        break
    }
  }

  const onClickActionModals = (action: string): void => {
    switch (action) {
      case 'openMetaDataSetFromTemplate': {
        setShowMetadataMessage(false)
        loadMetadataForm()
        setCopyTemplateModal({ ...copyTemplateModal, show: true })
        break
      }
      case 'openMetaDataManual': {
        setMetaDataItem(null)
        setAddManualMetadatalModal({ ...addManualMetadataModal, show: true })
        updateMedataForm(addFormMetaDataManualStateInitial)
        break
      }
      case 'applyTemplate': {
        const stateCopy = { ...addFormMetaSetState }
        const isValid = isValidForm(stateCopy.form)
        if (selectedProductNames.length > 0 && isValid) {
          updateMetaDataSetGridFromTemplate()
        } else {
          setShowMetadataMessage(true)
          const form = showFormRequiredFields(stateCopy.form)
          stateCopy.form = form
          setAddFormMetaSetState(stateCopy)
        }
        break
      }
      case 'applyMetaDataSet': {
        const stateCopy = { ...addFormMetaSetManualState }
        const isValid = isValidForm(stateCopy.form)
        if (isValid) {
          updateMetadataSetGridFromManual()
          setAddManualModal({ ...addManualModal, show: false })
        } else {
          const form = showFormRequiredFields(stateCopy.form)
          stateCopy.form = form
          setAddFormMetaSetManualState(stateCopy)
        }
        break
      }
      case 'deleteMetaDatarow': {
        deleteMetadataGridFromManual()
        setDeleteManualMetadatalModal({ ...deleteManualMetadataModal, show: false })
        break
      }
      case 'applyMetaData': {
        const stateCopy = { ...addFormMetaDataManualState }
        const isValid = isValidForm(stateCopy.form)
        if (isValid) {
          updateMetadataGridFromManual()
          setAddManualMetadatalModal({ ...addManualMetadataModal, show: false })
          setAddFormMetaDataManualState(addFormMetaDataManualStateInitial)
        } else {
          const form = showFormRequiredFields(stateCopy.form)
          stateCopy.form = form
          setAddFormMetaDataManualState(stateCopy)
        }
        break
      }
      case 'save': {
        saveAllChanges()
        break
      }
      case 'cancelDiscard': {
        setDiscardChangesModal(discardChangesModalInitial)
        break
      }
      case 'applyDiscard': {
        onCancel()
        break
      }
      case 'cancelForm': {
        if (isFormTouched) {
          setDiscardChangesModal({ ...discardChangesModal, show: true })
        } else {
          onCancel()
        }
        break
      }
      default: {
        // Close modals
        setCopyTemplateModal({ ...copyTemplateModal, show: false })
        setAddManualModal({ ...addManualModal, show: false })
        setAddManualMetadatalModal({ ...addManualMetadataModal, show: false })
        break
      }
    }
  }

  const saveAllChanges = (): void => {
    const metaDataSets: productMetaDataSet[] = []
    metaDataSetItems.forEach((item, index) => {
      const metaDataArray: productMetaData[] = []
      const metaDataItems = productMetaDataItems.filter((metaData) => metaData.context === item.id)
      metaDataItems.forEach((metaData) => {
        metaDataArray.push({
          id: metaData.id,
          key: metaData.key,
          value: metaData.value,
          type: metaData.type,
          metadataSetId: metaData.context
        })
      })
      metaDataSets.push({
        id: String(index),
        name: item.name,
        context: item.context,
        metaData: metaDataArray
      })
    })
    const productUpdated = { ...newProduct }
    productUpdated.metaDataSets = [...metaDataSets]
    setNewProduct(productUpdated)
    showSuccess('Product metadata applied', false)
    onCancel()
  }

  const onCancel = (): void => {
    navigate(-1)
  }

  return <ProductCatalogCreateMetaDataset
    onCancel={onCancel}
    addFormMetaSetState={addFormMetaSetState}
    onChangeInput={onChangeInput}
    addFormMetaSetManualState={addFormMetaSetManualState}
    copyTemplateModal={copyTemplateModal}
    addManualModal={addManualModal}
    onClickActionModals={onClickActionModals}
    showMetadataMessage={showMetadataMessage}
    metadataSetColumns={metadataSetColumns}
    metaDataSetItems={metaDataSetItems}
    metaDataSetEmptyGrid={metaDataSetEmptyGrid}
    metadataColumns={metadataColumns}
    metaDataItems={metaDataItems}
    metaDataEmptyGrid={metaDataEmptyGrid}
    filterMetadatasetText={filterMetadatasetText}
    setFilterMetaDataset={setFilterMetaDataset}
    filterMetadataText={filterMetadataText}
    setFilterMetaData={setFilterMetaData}
    addManualMetadataModal={addManualMetadataModal}
    addFormMetaDataManualState={addFormMetaDataManualState}
    generalFormOptions={generalFormOptions}
    selectedMetaDataSet={selectedMetaDataSet}
    setSelectedMetaDataSet={setSelectedMetaDataSet}
    discardChangesModal={discardChangesModal}
    discardOptions={discardOptions}
    deleteManualMetadataModal={deleteManualMetadataModal}
    deleteMetaOptions={deleteMetaOptions} />
}

export default ProductCatalogCreateMetaDatasetContainer
