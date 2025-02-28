// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React, { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import ProductCatalogEdit from '../../components/productCatalog/ProductCatalogEdit'
import useRegionStore from '../../store/regionStore/RegionStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { UpdateFormHelper, isValidForm, showFormRequiredFields, getFormValue, setSelectOptions, setFormValue } from '../../utility/updateFormHelper/UpdateFormHelper'
import { toastMessageEnum } from '../../utility/Enums'
import useToastStore from '../../store/toastStore/ToastStore'
import useProductV2Store, { type Product } from '../../store/productV2Store/ProductV2Store'
import useFamilyStore from '../../store/familyStore/FamilyStore'
import PublicService from '../../services/PublicService'

const ProductCatalogEditContainer = (): JSX.Element => {
  // *****
  // Params
  // *****
  const { param } = useParams()
  // *****
  // local state
  // *****
  const families = useFamilyStore((state) => state.families)
  const getFamilies = useFamilyStore((state) => state.getFamilies)
  const regions = useRegionStore((state) => state.regions)
  const setRegions = useRegionStore((state) => state.setRegions)
  const loading = useProductV2Store((state) => state.loading)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const product = useProductV2Store((state) => state.product)
  const editProduct = useProductV2Store((state) => state.editProduct)
  const setEditProduct = useProductV2Store((state) => state.setEditProduct)
  const getProductByName = useProductV2Store((state) => state.getProductByName)
  const editMode = useProductV2Store((state) => state.editMode)
  const setEditMode = useProductV2Store((state) => state.setEditMode)
  const productServices = useProductV2Store((state) => state.productServices)
  const setproductServices = useProductV2Store((state) => state.setproductServices)

  // *****
  // local state
  // *****
  const navigate = useNavigate()
  const throwError = useErrorBoundary()

  const initialState = {
    title: `Edit product ${param}`,
    form: {
      productName: {
        type: 'text', // options = 'text ,'textArea'
        label: 'Product name:',
        placeholder: 'Product name',
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
        helperMessage: <>Please provide a product name.</>,
        hidden: true
      },
      family: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Product family',
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
        </>
      },
      regions: {
        type: 'multi-select', // options = 'text ,'textArea'
        label: 'Regions:',
        placeholder: 'Please select',
        value: [], // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        validationRules: {
          isRequired: true
        },
        options: [], // Only required for select inputs, contains the seletable options
        validationMessage: '', // Errror message to display to the user
        selectAllButton: undefined,
        emptyOptionsMessage: 'No region found.'
      },
      serviceName: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Service',
        placeholder: 'Please select service',
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
          Choose the desired service
        </>
      },
      usage: {
        type: 'textArea', // options = 'text ,'textArea'
        label: 'Usage:',
        placeholder: 'Usage',
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
        helperMessage: <>Please provide a usage.</>,
        hidden: false
      }
    },
    isValidForm: false,
    servicePayload: {
      name: '',
      product_family_name: '',
      region_name: '',
      service_name: '',
      usage: '',
      metadatasets: [] as any[]
    },
    navigationBottom: [
      {
        buttonAction: 'create',
        buttonLabel: 'Save',
        buttonVariant: 'primary'
      },
      {
        buttonAction: 'cancel',
        buttonLabel: 'Cancel',
        buttonVariant: 'link'
      }
    ]
  }

  const selectAllButton = {
    label: 'Select/Deselect All',
    buttonFunction: () => {
      onSelectAll()
    }
  }

  const submitModalInitial = {
    show: false,
    message: 'Updating product'
  }

  const [state, setState] = useState(initialState)
  const [submitModal, setSubmitModal] = useState(submitModalInitial)
  const [isPageReady, setIsPageReady] = useState(false)

  // *****
  // Hooks
  // *****
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      if (param) {
        const promiseArray = []
        promiseArray.push(getProductByName(param))
        promiseArray.push(setRegions())
        promiseArray.push(getFamilies())
        promiseArray.push(setproductServices())
        await Promise.all(promiseArray)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    if (!editMode) {
      setEditProduct(product)
    }
  }, [product])

  useEffect(() => {
    loadProductForm()
  }, [regions, families, productServices, editProduct])

  // *****
  // functions
  // *****
  const loadProductForm = (): void => {
    if (families && regions && productServices) {
      const items = [...regions]
      const options: any = []
      const stateUpdate = { ...state }
      items.forEach((item) => {
        options.push({
          name: item.name,
          value: item.name
        })
      })
      const familyOptions: any = []
      const familyItems = [...families]
      familyItems.forEach((item) => {
        familyOptions.push({
          name: item.name,
          value: item.name
        })
      })
      const seriviceOptions: any = []
      const productServicesItems = [...productServices]
      productServicesItems.forEach((item) => {
        seriviceOptions.push({
          name: item.name,
          value: item.name
        })
      })
      let form = setSelectOptions('regions', options, stateUpdate.form)
      form = setSelectOptions('family', familyOptions, form)
      form = setSelectOptions('serviceName', seriviceOptions, form)
      if (editProduct) {
        const { name, familyName, regionName, serviceName, usage } = editProduct
        form = setFormValue('productName', name, form)
        form = setFormValue('family', familyName, form)
        const regions = regionName.split(',')
        form = setFormValue('regions', regions, form)
        form = setFormValue('serviceName', serviceName, form)
        form = setFormValue('usage', usage, form)
      }
      stateUpdate.isValidForm = isValidForm(form)
      stateUpdate.form = form
      setState(stateUpdate)
      setIsPageReady(true)
    }
  }

  const onChangeInput = (event: any, formInputName: string): void => {
    const stateUpdated = {
      ...state
    }
    const value = event.target.value
    const updatedForm = UpdateFormHelper(value, formInputName, stateUpdated.form)
    stateUpdated.isValidForm = isValidForm(updatedForm)
    stateUpdated.form = updatedForm
    setState(stateUpdated)
  }

  const onChangeDropdownMultiple = (values: any): void => {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(values, 'regions', updatedState.form)
    updatedState.isValidForm = isValidForm(updatedForm)
    updatedState.form = updatedForm
    setState(updatedState)
  }

  const onSelectAll = (): void => {
    if (regions) {
      const stateUpdated = {
        ...state
      }

      const allKeys = regions.map((key) => key.name)
      const selectedValues: any[] = [...stateUpdated.form.regions.value]
      const shouldDeselect = allKeys.every((x: any) => selectedValues.includes(x))
      onChangeDropdownMultiple(shouldDeselect ? [] : allKeys)
    }
  }

  const onSubmit = async (): Promise<void> => {
    try {
      const isValidForm = state.isValidForm
      if (!isValidForm) {
        showRequiredFields()
        return
      }
      setSubmitModal({ ...submitModal, show: true })
      const payloadCopy = { ...state.servicePayload }
      payloadCopy.name = getFormValue('productName', state.form)
      const regions = getFormValue('regions', state.form).join(',')
      payloadCopy.region_name = regions
      payloadCopy.service_name = getFormValue('serviceName', state.form)
      payloadCopy.usage = getFormValue('usage', state.form)
      payloadCopy.product_family_name = getFormValue('family', state.form)
      const metadatasets: any[] = []
      for (const index in editProduct?.metaDataSets) {
        const metaDataSet = {
          ...editProduct?.metaDataSets[Number(index)]
        }
        const metaData: any[] = []
        for (const subIndex in metaDataSet.metaData) {
          const item = { ...metaDataSet.metaData[Number(subIndex)] }
          metaData.push({
            id: item.id,
            key: item.key,
            value: item.value,
            type: item.type
          })
        }
        metadatasets.push({
          id: metaDataSet.id,
          name: metaDataSet.name,
          context: metaDataSet.context,
          metadata: metaData
        })
      }
      payloadCopy.metadatasets = [...metadatasets]
      await PublicService.putProduct(payloadCopy, product?.name)
      setSubmitModal({ ...submitModal, show: false })
      showSuccess('Product updated successfully', false)
      onCancel()
    } catch (error: any) {
      const message = String(error.message)
      setSubmitModal({ ...submitModal, show: false })
      if (error.response) {
        const errData = error.response.data
        const errMessage = errData.message
        showError(errMessage, false)
      } else {
        showError(message, false)
      }
    }
  }

  const showRequiredFields = (): void => {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)

    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  const onClickAddMetaData = (): void => {
    if (editProduct) {
      const regions = getFormValue('regions', state.form).join(',')
      const editUpdatedProduct: Product = {
        name: getFormValue('productName', state.form),
        familyName: getFormValue('family', state.form),
        regionName: regions,
        serviceName: getFormValue('serviceName', state.form),
        usage: getFormValue('usage', state.form),
        metaDataSets: [...editProduct?.metaDataSets]
      }
      setEditMode(true)
      setEditProduct(editUpdatedProduct)
      navigate(`/products/d/${param}/metadata`)
    }
  }

  const onCancel = (): void => {
    setEditMode(false)
    navigate('/products')
  }

  return <ProductCatalogEdit
    onCancel={onCancel}
    state={state}
    loading={loading}
    onChangeInput={onChangeInput}
    onChangeDropdownMultiple={onChangeDropdownMultiple}
    selectAllButton={selectAllButton}
    onClickAddMetaData={onClickAddMetaData}
    onSubmit={onSubmit}
    submitModal={submitModal}
    editProduct={editProduct}
    isPageReady={isPageReady} />
}

export default ProductCatalogEditContainer
