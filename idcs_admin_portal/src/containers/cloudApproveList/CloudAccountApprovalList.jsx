import React, { useEffect, useState } from 'react'
import Wrapper from '../../utility/wrapper/Wrapper'
import { useNavigate } from 'react-router-dom'
import CloudAccountApprovalListComponent from '../../components/cloudApproveList/CloudAccountApproveListComponent'
import CloudAccountService from '../../services/CloudAccountService'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import { UpdateFormHelper, isValidForm } from '../../utility/updateFormHelper/UpdateFormHelper'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import IKSService from '../../services/IKSService'

const CloudAccApproveListFormInitialState = {
  account: {
    id: 'account',
    sectionGroup: 'approveList_form',
    type: 'text',
    fieldSize: 'small',
    label: 'Account',
    placeholder: 'Enter Account Details',
    value: '',
    isValid: true,
    isTouched: false,
    isReadOnly: true,
    maxLength: 63,
    validationRules: {
      isRequired: false
    },
    validationMessage: ''
  },
  providername: {
    id: 'providername',
    sectionGroup: 'approveList_form',
    type: 'text',
    fieldSize: 'small',
    label: 'Provider Name',
    placeholder: 'Enter Provider Name',
    value: '',
    isValid: true,
    isTouched: false,
    isReadOnly: true,
    maxLength: 63,
    validationRules: {
      isRequired: false
    },
    validationMessage: ''
  },
  status: {
    id: 'status',
    sectionGroup: 'approveList_form',
    type: 'checkbox',
    fieldSize: 'small',
    label: 'Status',
    placeholder: 'Enter Status',
    value: false,
    isValid: true,
    isTouched: false,
    isReadOnly: true,
    maxLength: 63,
    validationRules: {
      isRequired: false
    },
    validationMessage: ''
  },
  enableStorage: {
    id: 'enableStorage',
    sectionGroup: 'approveList_form',
    type: 'checkbox',
    fieldSize: 'small',
    label: 'Storage',
    placeholder: 'Enter Storage',
    value: false,
    isValid: true,
    isTouched: false,
    isReadOnly: true,
    maxLength: 63,
    validationRules: {
      isRequired: false
    },
    validationMessage: ''
  },
  maxclusterilb_override: {
    id: 'maxclusterilb_override',
    sectionGroup: 'approveList_form',
    type: 'number',
    fieldSize: 'small',
    label: 'Max Cluster ILB Override',
    placeholder: 'Enter Max Cluster ILB Override Value',
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
  maxclusterng_override: {
    id: 'maxclusterng_override',
    sectionGroup: 'approveList_form',
    type: 'number',
    fieldSize: 'small',
    label: 'Max Cluster NG Override',
    placeholder: 'Enter Max Cluster NG Override Value',
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
  maxclusters_override: {
    id: 'maxclusters_override',
    sectionGroup: 'approveList_form',
    type: 'number',
    fieldSize: 'small',
    label: 'Max Clusters Override',
    placeholder: 'Enter Max Clusters Override Value',
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
  maxclustervm_override: {
    id: 'maxclustervm_override',
    sectionGroup: 'approveList_form',
    type: 'number',
    fieldSize: 'small',
    label: 'Max Cluster VM Override',
    placeholder: 'Enter Max Cluster VM Override Value',
    value: '',
    isValid: true,
    isTouched: false,
    isReadOnly: true,
    maxLength: 63,
    validationRules: {
      isRequired: false
    },
    validationMessage: ''
  },
  maxnodegroupvm_override: {
    id: 'maxnodegroupvm_override',
    sectionGroup: 'approveList_form',
    type: 'number',
    fieldSize: 'small',
    label: 'Max Node Group VM Override',
    placeholder: 'Enter Max Node Group VM Override Value',
    value: '',
    isValid: true,
    isTouched: false,
    isReadOnly: false,
    maxLength: 63,
    validationRules: {
      isRequired: false
    },
    validationMessage: ''
  }
}

const cloudAccApproveListFormID = 'CloudAccountApprovalList_Form'

function CloudAccountApprovalList(props) {
  // Navigation
  const navigate = useNavigate()
  const throwError = useErrorBoundary()

  const AddCloudAccFormInitialState = {
    cloudAccountID: {
      sectionGroup: 'create_cloud_account',
      type: 'text', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Cloud Account ID',
      placeholder: 'Enter Cloud Account ID',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 63,
      validationRules: {
        isRequired: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: <div></div>
    },
    cloudAccountStatus: {
      sectionGroup: 'create_cloud_account',
      type: 'select', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Status',
      placeholder: 'Please select status',
      value: '',
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      validationRules: {
        isRequired: true
      },
      options: [
        { key: 1, name: 'active', value: 'Active' }
      ],
      validationMessage: ''
    },
    maxclusterilb_override: {
      id: 'maxclusterilb_override',
      sectionGroup: 'approveList_form',
      type: 'number',
      fieldSize: 'small',
      label: 'Max Cluster ILB Override',
      placeholder: 'Enter Max Cluster ILB Override Value',
      value: '',
      isValid: true,
      isTouched: false,
      isReadOnly: true,
      maxLength: 63,
      validationRules: {
        isRequired: false
      },
      validationMessage: ''
    },
    maxclusterng_override: {
      id: 'maxclusterng_override',
      sectionGroup: 'approveList_form',
      type: 'number',
      fieldSize: 'small',
      label: 'Max Cluster NG Override',
      placeholder: 'Enter Max Cluster NG Override Value',
      value: '',
      isValid: true,
      isTouched: false,
      isReadOnly: true,
      maxLength: 63,
      validationRules: {
        isRequired: false
      },
      validationMessage: ''
    },
    maxclusters_override: {
      id: 'maxclusters_override',
      sectionGroup: 'approveList_form',
      type: 'number',
      fieldSize: 'small',
      label: 'Max Clusters Override',
      placeholder: 'Enter Max Clusters Override Value',
      value: '',
      isValid: true,
      isTouched: false,
      isReadOnly: true,
      maxLength: 63,
      validationRules: {
        isRequired: false
      },
      validationMessage: ''
    },
    maxclustervm_override: {
      id: 'maxclustervm_override',
      sectionGroup: 'approveList_form',
      type: 'number',
      fieldSize: 'small',
      label: 'Max Cluster VM Override',
      placeholder: 'Enter Max Cluster VM Override Value',
      value: '',
      isValid: true,
      isTouched: false,
      isReadOnly: true,
      maxLength: 63,
      validationRules: {
        isRequired: false
      },
      validationMessage: ''
    },
    maxnodegroupvm_override: {
      id: 'maxnodegroupvm_override',
      sectionGroup: 'approveList_form',
      type: 'number',
      fieldSize: 'small',
      label: 'Max Node Group VM Override',
      placeholder: 'Enter Max Node Group VM Override Value',
      value: '',
      isValid: true,
      isTouched: false,
      isReadOnly: true,
      maxLength: 63,
      validationRules: {
        isRequired: false
      },
      validationMessage: ''
    }
  }

  const emptyGrid = {
    title: 'No Cloud Account Approval lists found',
    subTitle: 'There are currently no accounts'
  }

  const emptyGridByFilter = {
    title: 'No Cloud Account found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  // Cloud Accounts States
  const [cloudAccount, setCloudAccount] = useState(null)
  const [selectedCloudAccount, setSelectedCloudAccount] = useState(null)
  const [cloudAccountError, setCloudAccountError] = useState('')
  const [gridItems, setGridItems] = useState([])
  const [showAddCloudAccModal, setShowCloudAccModal] = useState(false)
  const [addCloudAccFormElements, setAddCloudAccFormElements] = useState(AddCloudAccFormInitialState)
  const [isFormValid, setIsFormValid] = useState(false)
  const [formErrorMessage, setFormErrorMessage] = useState('')
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [action, setAction] = useState('') // 'update', 'storage', 'status'
  const [resourceLimitsMap, setResourceLimitsMap] = useState(null)

  // Update Modal States
  const [cloudAccountData, setCloudAccountData] = useState(null)
  const [showUpdateModal, setShowUpdateModal] = useState(false)
  const [approveListFormData, setApproveListFormData] = useState(CloudAccApproveListFormInitialState)
  const [isUpdateFormValid, setIsUpdateFormValid] = useState(false)
  const [updateFormErrMsg, setUpdateFormErrMsg] = useState('')

  // Auth Modal States
  const [showAuthModal, setShowAuthModal] = useState(false)
  const [showUnAuthModal, setShowUnAuthModal] = useState(false)
  const [authPassword, setAuthPassword] = useState('')
  const [showSearchCloudAccount, setShowSearchCloudAccount] = useState(false)
  const [showSearchAccountLoader, setShowSearchAccountLoader] = useState('')

  // Cloud Account Stores
  const loading = useCloudAccountStore((state) => state.loading)
  const stopLoading = useCloudAccountStore((state) => state.stopLoading)
  const cloudAccountsData = useCloudAccountStore((state) => state.cloudAccountsData)
  const resourceLimits = useCloudAccountStore((state) => state.resourceLimits)
  const getCloudAccountsData = useCloudAccountStore((state) => state.getCloudAccountsData)
  const createCloudAccountsData = useCloudAccountStore((state) => state.createCloudAccountsData)
  const updateCloudAccountsData = useCloudAccountStore((state) => state.updateCloudAccountsData)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  useEffect(() => {
    fetchCloudAccounts()
  }, [])

  const fetchCloudAccounts = async (isBackground) => {
    try {
      await getCloudAccountsData(isBackground)
    } catch (error) {
      throwError(error)
    } finally {
      stopLoading()
    }
  }

  useEffect(() => {
    const limitMapper = {
      maxclusterilb_override: resourceLimits?.maxvipspercluster ?? 0,
      maxclusterng_override: resourceLimits?.maxnodegroupspercluster ?? 0,
      maxclusters_override: resourceLimits?.maxclusterpercloudaccount ?? 0,
      maxnodegroupvm_override: resourceLimits?.maxnodespernodegroup ?? 0,
      maxclustervm_override: resourceLimits?.maxclustervm ?? 0
    }

    setResourceLimitsMap({ ...limitMapper })
  }, [resourceLimits])

  useEffect(() => {
    setGridInfo()
  }, [cloudAccountsData])

  const onClearSearchInput = () => {
    // Making cloud account null.
    setCloudAccount(null)
    // Clearing / Making empty instance group state.
    setSelectedCloudAccount('')
    // Making error state with false.
    setCloudAccountError('')
  }

  const handleSearchInputChange = (e) => {
    // Update the state with the numeric value
    setCloudAccount(e.target.value)
  }

  const setFilter = (event, clear) => {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  // Function to handle form submission
  const handleSubmit = async (e) => {
    setCloudAccountError('')
    setSelectedCloudAccount('')
    if (cloudAccount !== '') {
      setShowSearchAccountLoader({ isShow: true, message: 'Searching for Details...' })
      try {
        let data
        // Calling the specific service based on ID to fetch the cloud account details.
        if (cloudAccount.includes('@') || /[a-zA-Z]/.test(cloudAccount)) {
          data = await CloudAccountService.getCloudAccountDetailsByName(
            cloudAccount
          )
        } else {
          data = await CloudAccountService.getCloudAccountDetailsById(
            cloudAccount
          )
        }

        setCloudAccount(cloudAccount)
        // Picks the selected searched data.
        setSelectedCloudAccount(data?.data)
        // Making error state with false.
        setCloudAccountError('')
      } catch (e) {
        const code = e.response.data?.code
        const errorMsg = e.response.data?.message
        const message =
          code && [3, 5].includes(code)
            ? errorMsg.charAt(0).toUpperCase() + errorMsg.slice(1)
            : 'Cloud Account ID is not found'
        // Assigning the error message.
        setCloudAccountError(message)
        // Clearing selected search data.
        setSelectedCloudAccount(null)
      }
    } else {
      // Assigning the error message.
      setCloudAccountError('Cloud Account Number is required')
    }
    setShowSearchAccountLoader('')
  }

  const openCreateCloudAccModel = (isOpen) => {
    // Resetting the Form Error and State
    const formInitialState = { ...AddCloudAccFormInitialState }

    Object.keys(formInitialState).forEach((key) => {
      const value = resourceLimitsMap[key]

      if (typeof value !== 'undefined') {
        formInitialState[key].value = value
      }
    })

    setAddCloudAccFormElements(formInitialState)
    setFormErrorMessage('')
    setShowCloudAccModal(isOpen)
  }

  const closeCreateModal = () => {
    // Closing the Modal
    setShowCloudAccModal(false)
    // Reset Form Data
    setAddCloudAccFormElements(AddCloudAccFormInitialState)
    // Reset Form Validity
    setIsFormValid(false)
    // Reset Form Error Message
    setFormErrorMessage('')
    // Reset Actions
    setAction('')
  }

  const onChangeCreateCloudAccInput = (event, formInputName) => {
    const formCopy = {
      ...addCloudAccFormElements
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, formCopy)
    const isValidoptions = isValidForm(updatedForm)

    setAddCloudAccFormElements(updatedForm)
    setIsFormValid(isValidoptions)
    setFormErrorMessage('')
  }

  const onCreateCloudAccFormSubmit = async () => {
    try {
      const data = await CloudAccountService.getCloudAccountDetailsById(addCloudAccFormElements.cloudAccountID.value)

      if (data?.data?.id === addCloudAccFormElements.cloudAccountID.value) {
        const payload = {
          account: addCloudAccFormElements.cloudAccountID.value,
          status: (addCloudAccFormElements.cloudAccountStatus.value.toLowerCase() === 'active'),
          enableStorage: false,
          maxclusterilb_override: convertToNum(addCloudAccFormElements.maxclusterilb_override.value),
          maxclusterng_override: convertToNum(addCloudAccFormElements.maxclusterng_override.value),
          maxclusters_override: convertToNum(addCloudAccFormElements.maxclusters_override.value),
          maxclustervm_override: convertToNum(addCloudAccFormElements.maxclustervm_override.value),
          maxnodegroupvm_override: convertToNum(addCloudAccFormElements.maxnodegroupvm_override.value),
          iksadminkey: btoa(authPassword)
        }

        try {
          await createCloudAccountsData(payload, true)
          debounceCloudAccListRefresh()
          closeCreateModal()
        } catch (error) {
          if (error.response.data?.message === 'Cloud Account already in use') {
            showError(error.response.data?.message)
          } else {
            closeCreateModal()
            throwError(error)
          }
        }
      } else {
        const message = 'Cloud Account ID is not found'

        // Assigning the error message.
        setFormErrorMessage(message)
      }
    } catch (e) {
      const code = e.response.data?.code
      const errorMsg = e.response.data?.message
      const message =
        code && [3, 5].includes(code)
          ? errorMsg.charAt(0).toUpperCase() + errorMsg.slice(1)
          : 'Cloud Account ID is not found'

      // Assigning the error message.
      setFormErrorMessage(message)
    }
  }

  const backToHome = () => {
    navigate('/')
  }

  const updateCloudAccountData = async (accountData, updatedData, field) => {
    const payload = {
      account: accountData.account,
      status: accountData.status,
      enableStorage: accountData.enableStorage,
      maxclusterilb_override: accountData.maxclusterilb_override,
      maxclusterng_override: accountData.maxclusterng_override,
      maxclusters_override: accountData.maxclusters_override,
      maxclustervm_override: accountData.maxclustervm_override,
      maxnodegroupvm_override: accountData.maxnodegroupvm_override,
      ...updatedData
    }

    if (!payload.status) {
      payload.enableStorage = false
    }

    try {
      await updateCloudAccountsData(payload, true)
      let message = 'Successfully Updated Cloud Account Data'
      if (field === 'account') {
        message = `Account ${payload.account} has been ${payload.status ? 'enabled' : 'disabled'} successfully.`
      } else if (field === 'storage') {
        message = `Storage has been ${payload.enableStorage ? 'enabled' : 'disabled'} successfully.`
      }
      showSuccess(message)
      debounceCloudAccListRefresh()
      resetDefaultStates()
      closeUpdateModal()
    } catch (error) {
      if (error.response?.status === 500) {
        setUpdateFormErrMsg(error.response?.data?.message ?? '')
      } else {
        showError(error.message)
        closeUpdateModal()
        throwError(error)
      }
    }
  }

  const openUpdateModal = (accountData) => {
    // Logic to set the form values for editing
    let formCopy = structuredClone(approveListFormData)
    for (const key in formCopy) {
      let value = accountData[key]
      if (Object.keys(resourceLimitsMap).includes(key) && value === 0) {
        value = resourceLimitsMap[key]
      }

      formCopy = { ...UpdateFormHelper(value, key, formCopy) }
    }

    setApproveListFormData(formCopy)
    setCloudAccountData(accountData)
    setUpdateFormErrMsg('')
    setShowUpdateModal(true)
  }

  const resetDefaultStates = () => {
    // Reset Actions
    setAction('')
    // Reset Selected Cloud Account Data
    setCloudAccountData(null)
  }

  const closeUpdateModal = () => {
    // Close the Update Modal
    setShowUpdateModal(false)
    // Reset Form Data
    setApproveListFormData(CloudAccApproveListFormInitialState)
    // Reset Selected Cloud Account Data
    setCloudAccountData(null)
    // Reset Form Validity
    setIsUpdateFormValid(false)
    // Reset Form Error Message
    setUpdateFormErrMsg('')
  }

  const updateFormEvents = {
    onChange: (event, id) => {
      const formCopy = structuredClone(approveListFormData)

      let value = event.target.value
      if (formCopy[id].type === 'checkbox') {
        value = event.target.checked
      }

      const updatedForm = UpdateFormHelper(value, id, formCopy)
      customInputValidation(updatedForm, id)
      const isValidOptions = isValidForm(updatedForm)

      setApproveListFormData({ ...updatedForm })
      setIsUpdateFormValid(isValidOptions)
      if (updateFormErrMsg !== '') {
        setUpdateFormErrMsg('')
      }
    }
  }

  const customInputValidation = (form, key) => {
    let validationMessage = ''
    let isValid = true

    const formAttr = form[key]

    if (Object.keys(resourceLimitsMap).includes(key)) {
      const value = parseInt(formAttr.value)
      const minValue = resourceLimitsMap[key]

      if (isNaN(value) || value < minValue) {
        validationMessage = formAttr.label + ' must be greated than ' + (minValue - 1)
        isValid = false
      }
    }

    formAttr.validationMessage = validationMessage
    formAttr.isValid = isValid
  }

  const setGridInfo = () => {
    const items = []

    cloudAccountsData.forEach((accountData) => {
      const statusBtnVal = accountData.status ? 'Disable account' : 'Enable account'
      const storageBtnVal = accountData.enableStorage ? 'Disable storage' : 'Enable storage'
      items.push({
        account: {
          showField: true,
          type: 'HyperLink',
          value: accountData.account,
          function: () => { getCloudAccountDetailsById(accountData.account) }
        },
        provider: accountData.providername,
        status: accountData.status ? 'Active' : 'In Active',
        enableStorage: accountData.enableStorage ? 'Active' : 'In Active',
        actions: {
          showField: true,
          type: 'Buttons',
          value: accountData,
          selectableValues: [
            { name: statusBtnVal, label: statusBtnVal },
            { name: storageBtnVal, label: storageBtnVal },
            { name: 'Update account', label: 'Update account', id: 'Update account' }
          ],
          function: (action, item) => {
            switch (action.name) {
              case statusBtnVal:
                setAction('status')
                setShowAuthModal(true)
                setCloudAccountData(item)
                break
              case storageBtnVal:
                setAction('storage')
                setShowAuthModal(true)
                setCloudAccountData(item)
                break
              case 'Update account':
                openUpdateModal(accountData)
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

  const convertToNum = (input, defaultValue = 0) => {
    let output = parseInt(input)
    output = isNaN(output) ? defaultValue : output

    return output
  }

  const updateFormFooterEvents = {
    onPrimaryBtnClick: () => {
      setAction('update')
      setShowAuthModal(true)
    },
    onSecondaryBtnClick: () => {
      closeUpdateModal()
    }
  }

  const setUserAuthorization = () => {
    const payload = {
      iksadminkey: btoa(authPassword)
    }

    IKSService.authenticateIMIS(payload)
      .then(({ data }) => {
        if (data?.isAuthenticatedUser) {
          if (action === 'add') {
            onCreateCloudAccFormSubmit()
          }
          if (action === 'status') {
            const payload = {
              status: !cloudAccountData.status,
              iksadminkey: btoa(authPassword)
            }

            updateCloudAccountData(cloudAccountData, payload)
          } else if (action === 'storage') {
            const payload = {
              enableStorage: !cloudAccountData.enableStorage,
              iksadminkey: btoa(authPassword)
            }

            updateCloudAccountData(cloudAccountData, payload)
          } else if (action === 'update') {
            const payload = {
              maxclusterilb_override: convertToNum(approveListFormData.maxclusterilb_override.value),
              maxclusterng_override: convertToNum(approveListFormData.maxclusterng_override.value),
              maxclusters_override: convertToNum(approveListFormData.maxclusters_override.value),
              maxnodegroupvm_override: convertToNum(approveListFormData.maxnodegroupvm_override.value),
              iksadminkey: btoa(authPassword)
            }

            updateCloudAccountData(cloudAccountData, payload)
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

  const closeAuthModal = () => {
    // Close Auth Modal
    setShowAuthModal(false)
    // Reset Auth Password
    setAuthPassword('')
  }

  useEffect(() => {
    setInstanceTypeModalVisbility(!(showAuthModal || showUnAuthModal))
  }, [showAuthModal, showUnAuthModal])

  const setInstanceTypeModalVisbility = (visibility = true) => {
    const element = document.getElementById(cloudAccApproveListFormID)
    if (element) {
      element.style.display = visibility ? 'block' : 'none'
    }
  }

  const debounceCloudAccListRefresh = () => {
    setTimeout(() => {
      fetchCloudAccounts(true)
    }, 2000)
  }

  const setShowSearchModal = (status) => {
    if (!status) onClearSearchInput()
    setShowSearchCloudAccount(status)
  }

  const getCloudAccountDetailsById = async (cloudAccount) => {
    if (cloudAccount !== '') {
      setShowSearchModal(true)
      setShowSearchAccountLoader({ isShow: true, message: 'Searching for Details...' })
      try {
        const data = await CloudAccountService.getCloudAccountDetailsById(cloudAccount)
        setSelectedCloudAccount(data?.data)
        setCloudAccountError(false)
      } catch (e) {
        const code = e.response.data?.code
        const errorMsg = e.response.data?.message
        const message = code && [3, 5].includes(code) ? errorMsg.charAt(0).toUpperCase() + errorMsg.slice(1) : 'Cloud Account ID is not found'
        setCloudAccountError(message)
        setSelectedCloudAccount(false)
      }
      setShowSearchAccountLoader('')
    }
  }

  const columns = [
    {
      columnName: 'Account',
      targetColumn: 'account',
      columnConfig: {
        behaviorType: 'hyperLink',
        behaviorFunction: 'getCloudAccountDetailsById'
      }
    },
    {
      columnName: 'Provider Name',
      targetColumn: 'provider'
    },
    {
      columnName: 'Status',
      targetColumn: 'status'
    },
    {
      columnName: 'Storage',
      targetColumn: 'enableStorage'
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

  const addCloudAccFormElementsArr = []
  for (const key in addCloudAccFormElements) {
    const formItem = {
      ...addCloudAccFormElements[key]
    }

    addCloudAccFormElementsArr.push({
      id: key,
      configInput: formItem
    })
  }

  const createformFooterEvents = {
    onPrimaryBtnClick: () => {
      setAction('add')
      setShowAuthModal(true)
    },
    onSecondaryBtnClick: () => {
      closeCreateModal()
    }
  }

  const createCloudAccCTA = [{
    buttonLabel: 'Create',
    buttonVariant: 'primary'
  },
  {
    buttonLabel: 'Cancel',
    buttonVariant: 'link',
    buttonFunction: () => { createformFooterEvents.onSecondaryBtnClick() }
  }]

  const approveListFormDataArr = []
  for (const key in approveListFormData) {
    const formItem = {
      ...approveListFormData[key]
    }

    approveListFormDataArr.push({
      id: key,
      configInput: formItem
    })
  }

  return (
    <Wrapper>
      <CloudAccountApprovalListComponent
        gridData={{
          data: gridItems,
          columns,
          emptyGrid: emptyGridObject,
          loading
        }}
        modalsID={cloudAccApproveListFormID}
        modalData={{
          isOpen: showAddCloudAccModal,
          openModal: openCreateCloudAccModel,
          addCloudAccFormElements: addCloudAccFormElementsArr,
          onChangeCreateCloudAccInput,
          isFormValid,
          formErrorMessage,
          createCloudAccCTA,
          onCreateCloudAccFormSubmit: createformFooterEvents.onPrimaryBtnClick
        }}
        updateModalData={{
          showUpdateModal,
          closeUpdateModal,
          approveListFormData: approveListFormDataArr,
          updateFormEvents,
          updateFormErrorMessage: updateFormErrMsg,
          isUpdateFormValid,
          updateFormFooterEvents
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
        filterData={{
          filterText,
          setFilter
        }}
        cloudAccount={cloudAccount}
        selectedCloudAccount={selectedCloudAccount}
        cloudAccountError={cloudAccountError}
        showSearchCloudAccount={showSearchCloudAccount}
        showSearchAccountLoader={showSearchAccountLoader}
        backToHome={backToHome}
        handleSearchInputChange={handleSearchInputChange}
        handleSubmit={handleSubmit}
        onClearSearchInput={onClearSearchInput}
        setShowSearchModal={setShowSearchModal}
      />
    </Wrapper>
  )
}

export default CloudAccountApprovalList
