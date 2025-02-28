// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useProductStore from '../../store/productStore/ProductStore'
import { UpdateFormHelper, isValidForm, getFormValue, setFormValue, setSelectOptions } from '../../utility/updateFormHelper/UpdateFormHelper'
import CloudAccountService from '../../services/CloudAccountService'
import SkuQuotaCreate from '../../components/skuManagement/skuQuotaCreate/SkuQuotaCreate'
import SkuQuotaService from '../../services/SkuQuotaService'
import useUserStore from '../../store/userStore/UserStore'
import useToastStore from '../../store/toastStore/ToastStore'
import useVendorStore from '../../store/vendorStore/VendorStore'

const SkuQuotaCreateContainer = () => {
  // Initial state for form and validation
  const initialState = {
    mainTitle: 'Cloud Account Assignment',
    mainSubtitle: 'Specify the needed Information',
    form: {
      serviceType: {
        sectionGroup: 'configuration',
        type: 'select', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Service Type:',
        placeholder: 'Please select',
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
      families: {
        sectionGroup: 'configuration',
        type: 'select', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Product Family:',
        placeholder: 'Please select',
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
      cloudAccount: {
        sectionGroup: 'cloudAccount',
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Cloud Account:',
        placeholder: 'Enter Cloud Account Email or ID',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '' // Errror message to display to the user
      }
    },
    isValidForm: false,
    timeoutMiliseconds: 5000
  }

  // Initial state for loader
  const initialLoaderData = {
    isShow: false,
    message: ''
  }

  // Error Boundry
  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()

  const [state, setState] = useState(initialState)
  const [showLoader, setShowLoader] = useState(initialLoaderData)
  const [selectedProducts, setSelectedProducts] = useState([])
  const [filteredProducts, setFilteredProducts] = useState([])
  const [selectedCloudAccount, setSelectedCloudAccount] = useState(null)
  const [idcVendor, setIDCVendor] = useState(null)

  // Global States)
  const vendors = useVendorStore((state) => state.vendors)
  const getIDCVendor = useVendorStore((state) => state.getIDCVendor)
  const controlledProducts = useProductStore((state) => state.controlledProducts)
  const controlledFamilies = useProductStore((state) => state.controlledFamilies)
  const setProducts = useProductStore((state) => state.setProducts)
  const user = useUserStore((state) => state.user)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  const backButtonLabel = '⟵ Back to Previous Page'
  // Hooks
  useEffect(() => {
    const fetchVendors = async () => {
      try {
          await getIDCVendor()
        } catch (error) {
          throwError(error)
        }
    }
    const fetchProducts = async () => {
      try {
        await setProducts()
      } catch (error) {
        throwError(error)
      }
    }
    if (!vendors) fetchVendors()
    if (controlledFamilies.length === 0) fetchProducts()
  }, [])

  useEffect(() => {
    if (vendors) {
      const idcVendor = vendors.find((x) => x.name === 'idc')
      if (idcVendor) setIDCVendor(idcVendor)
    }
  }, [vendors])

  useEffect(() => {
    setForm()
  }, [controlledFamilies, idcVendor])

  // functions
  function setForm() {
    const stateUpdated = {
      ...state
    }
    // Load servicess
    const selectableServices = getServiceTypeOptions()
    if (selectableServices.length > 0) {
      stateUpdated.form = setSelectOptions('serviceType', selectableServices, stateUpdated.form)
      stateUpdated.form = setFormValue('serviceType', selectableServices[0].value, stateUpdated.form)

      // Load families
      const selectableFamilies = getFamiliesOptions(selectableServices[0].value)
      if (selectableFamilies.length > 0) {
        stateUpdated.form = setSelectOptions('families', selectableFamilies, stateUpdated.form)
        stateUpdated.form = setFormValue('families', selectableFamilies[0].value, stateUpdated.form)

        // Filter associated products
        productsFilter(selectableFamilies[0].value, selectableServices[0].value)
      }
    }

    setState(stateUpdated)
  }

  function getServiceTypeOptions() {
    if (!idcVendor) return []
    const services = idcVendor?.families
    return services.map(service => {
      return {
        name: service.description,
        value: service.id
      }
    })
  }

  function productsFilter(familyDisplayName, familyId) {
    let filteredProducts = []
    if (familyDisplayName) filteredProducts = controlledProducts.filter((product) => product.familyDisplayName === familyDisplayName && product.familyId === familyId)
    setFilteredProducts(filteredProducts)
    setSelectedProducts([])
  }

  function getFamiliesOptions(familyId) {
    const families = controlledFamilies.filter(family => family.familyId === familyId)
    return families.map((family) => {
      return {
        name: family.familyDisplayName,
        value: family.familyDisplayName
      }
    })
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)
    updatedState.form = updatedForm

    if (formInputName === 'cloudAccount') {
      setSelectedCloudAccount(null)
      updatedState.isValidForm = false
    }

    if (formInputName === 'serviceType') {
      // Clear value
      updatedState.form = setFormValue('families', '', updatedState.form)
      updatedState.form = setSelectOptions('families', [], updatedState.form)

      const selectableFamilies = getFamiliesOptions(event.target.value)
      updatedState.form = setSelectOptions('families', selectableFamilies, updatedState.form)
      let familyValue = null
      if (selectableFamilies.length > 0) {
        familyValue = selectableFamilies[0].value
        updatedState.form = setFormValue('families', familyValue, updatedState.form)
      }
      productsFilter(familyValue, event.target.value)
    }

    if (formInputName === 'families') {
      productsFilter(event.target.value, getFormValue('serviceType', updatedState.form))
    }

    if (!selectedCloudAccount) {
      updatedState.isValidForm = false
    }

    setState(updatedState)
  }

  async function onSearchCloudAccount() {
    const cloudAccount = getFormValue('cloudAccount', state.form)
    setCloudAccountError('')
    setSelectedCloudAccount(null)
    if (cloudAccount !== '') {
      try {
        let data
        if (cloudAccount.includes('@') || /[a-zA-Z]/.test(cloudAccount)) {
          data = await CloudAccountService.getCloudAccountDetailsByName(cloudAccount)
        } else {
          data = await CloudAccountService.getCloudAccountDetailsById(cloudAccount)
        }

        setSelectedCloudAccount(data?.data)
        const updatedState = { ...state }
        updatedState.form.cloudAccount.validationMessage = ''
        updatedState.isValidForm = isValidForm(updatedState.form)
        setState(updatedState)
      } catch (e) {
        const code = e.response.data?.code
        const errorMsg = e.response.data?.message
        const message =
          code && [3, 5].includes(code)
            ? String(errorMsg.charAt(0).toUpperCase()) + String(errorMsg.slice(1))
            : 'Cloud Account ID is not found'
        setCloudAccountError(message)
        setSelectedCloudAccount(null)
      }
    } else {
      setCloudAccountError('Cloud Account Number is required')
    }
  }

  function setCloudAccountError(errorMessage) {
    const updatedState = { ...state }
    const updatedForm = { ...updatedState.form }
    const updatedFormElement = { ...updatedForm.cloudAccount }
    updatedFormElement.isValid = false
    updatedFormElement.validationMessage = errorMessage
    updatedForm.cloudAccount = updatedFormElement
    updatedState.form = updatedForm
    updatedState.isValidForm = false
    setState(updatedState)
  }

  function onCancel() {
    navigate(-1)
  }

  function onSubmit() {
    setShowLoader({ isShow: true, message: 'Working on your request' })
    submitForm()
  }

  async function submitForm() {
    try {
      const selectedProduct = filteredProducts.find(x => x.id === selectedProducts[0])
      const servicePayload = {
        cloudaccountId: selectedCloudAccount.id,
        productId: selectedProduct.id,
        vendorId: selectedProduct.vendorId,
        familyId: selectedProduct.familyId,
        adminName: user.email
      }

      await postAcl(servicePayload)
    } catch (error) {
      let message = ''
      if (error?.response?.data?.message) {
        const isInsertionError = error.response.data.code === 13
        message = isInsertionError ? 'Account is already whitelisted for this Product' : error.response.data.message
      } else {
        message = error.message
      }
      setShowLoader({ isShow: false })
      showError(message)
    }
  }

  async function postAcl(servicePayload) {
    await SkuQuotaService.postAcl(servicePayload)
    setTimeout(() => {
      setShowLoader(initialLoaderData)
      showSuccess('Cloud account whitelisted.')
    }, state.timeoutMiliseconds)
  }

  return (
    <SkuQuotaCreate
      state={state}
      products={filteredProducts}
      showLoader={showLoader}
      backButtonLabel={backButtonLabel}
      onCancel={onCancel}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      selectedCloudAccount={selectedCloudAccount}
      selectedProducts={selectedProducts}
      setSelectedProducts={setSelectedProducts}
      onSearchCloudAccount={onSearchCloudAccount}
    />
  )
}

export default SkuQuotaCreateContainer
