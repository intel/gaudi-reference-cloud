import { React, useState, useEffect } from 'react'
import BannerCreate from '../../components/bannerManagement/BannerCreate'
import { useNavigate, useLocation } from 'react-router-dom'
import { UpdateFormHelper, isValidForm, getFormValue, showFormRequiredFields } from '../../utility/updateFormHelper/UpdateFormHelper'
import useBannerStore from '../../store/bannerStore/BannerStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { EnrollAccountType, AppRoutesEnum, BannerType, toastMessageEnum } from '../../utility/Enums'
import moment from 'moment'
import useToastStore from '../../store/toastStore/ToastStore'
import idcConfig from '../../config/configurator'

const DEFAULT_VALUE_ALL = 'all'

const getCustomMessage = (messageType) => {
  let message = null

  switch (messageType) {
    case 'isTop':
      message = (
        <div className="valid-feedback" intc-id={'BannerCreationAdminIsTopValidMessage'}>
          Selecting the option will update the banner&apos;s timestamp, ensuring that it appears at the forefront.
          <br />
          By default, any newly created banner will automatically appear at the top of the display sequence.
          <br />
          Note: The maintenance banner will always have priority and will be displayed at the top
        </div>
      )
      break
    case 'isMaintenance':
      message = (
        <div className="valid-feedback" intc-id={'BannerCreationAdminIsMaintenanceValidMessage'}>
          Only one maintenance banner can be created for each route per region at any given time.
          <br />
          The maintenance banner will always have priority and will be displayed at the top
        </div>
      )
      break
    default:
      break
  }

  return message
}

const BannerCreateContainer = () => {
  // Navigation
  const navigate = useNavigate()
  const location = useLocation()
  const initialBanner = location?.state?.banner || null
  const isEditBanner = Boolean(location.pathname === '/bannermanagement/update' && initialBanner)

  // initial local data
  const userTypeOptions = generateSelectOptions(EnrollAccountType)
  const routeOptions = generateSelectOptions(AppRoutesEnum)
  const bannerOptions = generateSelectOptions(BannerType)
  const regionOptions = generateSelectOptions(idcConfig.REACT_APP_DEFAULT_REGIONS)

  // local state
  const initialState = {
    title: isEditBanner
      ? `Update existing alert (${initialBanner.id}) for Intel Tiber AI Cloud Console`
      : 'Generate new alert for Intel Tiber AI Cloud Console',
    description:
      'Multiple banners can be associated with a single route, with the most recently created banner displaying at the top.',
    form: {
      title: {
        section: 'banner-info',
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Banner Title',
        placeholder: 'Title',
        value: initialBanner ? initialBanner.title : '',
        isValid: !!(initialBanner && initialBanner.title),
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
      message: {
        section: 'banner-info',
        type: 'textarea', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Banner Description',
        placeholder: 'Provide a description for this alert',
        value: initialBanner ? initialBanner.message : '',
        isValid: !!(initialBanner && initialBanner.message),
        isTouched: false,
        isReadOnly: false,
        maxLength: 200,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '',
        helperMessage: ''
      },
      linkLabel: {
        section: 'banner-link',
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Link Label (Optional)',
        placeholder: 'Link Label',
        value: initialBanner && initialBanner.link ? initialBanner.link.label : '',
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        maxLength: 100,
        validationRules: {
          checkMaxLength: true
        },
        validationMessage: '',
        helperMessage: 'Provide a label for the link at the end of the message.'
      },
      linkHref: {
        section: 'banner-link',
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Link URL (Optional)',
        placeholder: 'Link URL',
        value: initialBanner && initialBanner.link ? initialBanner.link.href : '',
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          checkMaxLength: true
        },
        validationMessage: '',
        helperMessage: 'Provide a url for the link at the end of the message.'
      },
      linkNewTab: {
        section: 'banner-link',
        type: 'checkbox', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Open link in new tab',
        placeholder: '',
        value: initialBanner && initialBanner.link ? initialBanner.link.openInNewTab : true,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        options: [
          {
            name: 'Open link at the end in new tab',
            value: '1'
          }
        ],
        validationMessage: '',
        helperMessage: ''
      },
      type: {
        section: 'banner-info',
        type: 'select', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Banner Type',
        placeholder: 'Please select',
        value: initialBanner && initialBanner.type ? initialBanner.type : 'info', // set default value as info
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: bannerOptions,
        validationMessage: '',
        helperMessage: ''
      },
      status: {
        section: 'banner-config',
        type: 'checkbox', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: '',
        placeholder: '',
        value: initialBanner ? initialBanner?.status === 'active' : false,
        isValid: true,
        isTouched: true,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        options: [
          {
            name: 'Is an Active Banner',
            value: '0'
          }
        ],
        validationMessage: '',
        helperMessage: ''
      },
      isMaintenance: {
        section: 'banner-config',
        type: 'checkbox', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: '',
        placeholder: '',
        value: initialBanner ? initialBanner?.isMaintenance === 'True' : false,
        isValid: true,
        isTouched: true,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        options: [
          {
            name: 'Is a Maintenance Banner',
            value: '0'
          }
        ],
        validationMessage: '',
        helperMessage: getCustomMessage('isMaintenance')
      },
      isTop: {
        section: 'banner-config',
        type: 'checkbox', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: '',
        placeholder: '',
        value: false,
        isValid: true,
        isTouched: true,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        options: [
          {
            name: 'Show Banner on Top',
            value: '0'
          }
        ],
        validationMessage: '',
        helperMessage: getCustomMessage('isTop')
      },
      userTypes: {
        section: 'banner-config',
        type: 'multi-select', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'User Types',
        placeholder: 'Please select',
        value:
          initialBanner ? allValueConversion(initialBanner?.userTypes, 'userTypes') : Object.values(EnrollAccountType), // set default value as all
        borderlessDropdownMultiple: true,
        options: userTypeOptions,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        selectAllButton: {
          label: 'Select/Deselect All'
        },
        validationMessage: 'User type is required.',
        helperMessage: ''
      },
      routes: {
        section: 'banner-config',
        type: 'select', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'UI Sections',
        placeholder: 'Please select',
        value: initialBanner ? initialBanner.routes : DEFAULT_VALUE_ALL, // set default value as all
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: routeOptions,
        validationMessage: '',
        helperMessage: ''
      },
      regions: {
        section: 'banner-config',
        type: 'multi-select-dropdown', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Regions',
        placeholder: 'Please select',
        value: initialBanner ? allValueConversion(initialBanner?.regions, 'regions') : idcConfig.REACT_APP_DEFAULT_REGIONS,
        options: regionOptions,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        selectAllButton: {
          label: 'All'
        },
        validationMessage: '',
        helperMessage: '',
        emptyOptionsMessage: 'No regions found.'
      },
      expirationDatetime: {
        section: 'banner-config',
        type: 'date', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Expiration Date',
        placeholder: 'Expiration Date',
        value:
          initialBanner && initialBanner.expirationDatetime
            ? moment(new Date(initialBanner.expirationDatetime)).format('YYYY-MM-DD')
            : '',
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        customContent: [],
        validationMessage: '',
        helperMessage: ''
      }
    },
    isValidForm: !!(initialBanner && initialBanner.title && initialBanner.message && initialBanner.type && initialBanner.routes && initialBanner.userTypes.length > 0 && initialBanner?.regions?.length > 0),
    payloadService: {
      amount: null,
      creator: null,
      start: null,
      expires: null,
      numUses: null,
      isStandard: null
    },
    navigationTop: [
      {
        label: '⟵ Back to home',
        buttonVariant: 'link',
        function: () => onCancel('home')
      }
    ],
    navigationBottom: [
      {
        label: 'Preview',
        buttonVariant: 'primary',
        function: () => onShowBanner()
      },
      {
        label: isEditBanner ? 'Update' : 'Create',
        buttonVariant: 'primary'
      },
      {
        label: 'Cancel',
        buttonVariant: 'link',
        function: () => onCancel('cancel')
      }
    ]
  }

  // Error Boundary
  const throwError = useErrorBoundary()

  const [state, setState] = useState(initialState)
  const [showModal, setShowModal] = useState(false)
  const [showBanner, setShowBanner] = useState(false)
  // Global Store
  const setBannerList = useBannerStore((state) => state.setBannerList)
  const addBanner = useBannerStore((state) => state.addBanner)
  const updateBanner = useBannerStore((state) => state.updateBanner)
  const showError = useToastStore((state) => state.showError)

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await setBannerList()
      } catch (error) {
        throwError(error)
      }
    }

    fetch()
  }, [])

  async function onSubmit() {
    const updatedState = { ...state }
    const updateForm = { ...updatedState.form }

    if (!isValidForm(updateForm)) {
      showRequiredFields()
      return
    }

    const type = getFormValue('type', updateForm)
    const title = getFormValue('title', updateForm)
    const status = getFormValue('status', updateForm) ? 'active' : 'inactive'
    const isMaintenance = getFormValue('isMaintenance', updateForm) ? 'True' : 'False'
    const isTop = getFormValue('isTop', updateForm) ? 'True' : 'False'
    const message = getFormValue('message', updateForm)
    const routes = [getFormValue('routes', updateForm)]
    const regions = allValueConversion(getFormValue('regions', updateForm), 'regions')
    const userTypes = allValueConversion(getFormValue('userTypes', updateForm), 'userTypes')
    const expirationDatetime = getFormValue('expirationDatetime', updateForm)

    const linkLabel = getFormValue('linkLabel', updateForm)
    const linkHref = getFormValue('linkHref', updateForm)
    const linkNewTab = getFormValue('linkNewTab', updateForm)

    const link =
      linkLabel && linkHref
        ? {
            label: linkLabel,
            href: linkHref,
            openInNewTab: linkNewTab
          }
        : undefined

    const payload = {
      id: isEditBanner ? initialBanner.id : Date.now(),
      type,
      title,
      status,
      message,
      userTypes,
      routes,
      regions,
      expirationDatetime,
      link,
      isMaintenance,
      updatedTimestamp:
        isEditBanner && isTop === 'False' && initialBanner?.updatedTimestamp
          ? initialBanner.updatedTimestamp
          : Date.now()
    }

    try {
      setShowModal(true)

      if (isEditBanner) {
        await updateBanner(payload)
      } else {
        await addBanner(payload)
      }
      setShowModal(false)
      navigate('/bannermanagement')
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

  function onShowBanner() {
    setShowBanner(true)
    window.scrollTo(0, 0)
  }

  function onCancel(route) {
    if (route === 'home') {
      navigate('/')
    } else {
      navigate('/bannermanagement')
    }
  }

  function onChangeDropdownMultiple(values, field) {
    const updatedState = {
      ...state
    }
    const updatedForm = UpdateFormHelper(values, field, updatedState.form)
    updatedForm[field].isValid = values.length > 0
    updatedState.isValidForm = isValidForm(updatedForm)
    updatedState.form = updatedForm

    setState(updatedState)
  }

  function onSelectAll(field) {
    const stateUpdated = {
      ...state
    }

    const typeOptions = field === 'userTypes' ? userTypeOptions : regionOptions
    const allTypes = typeOptions.map((type) => type.value)
    const selectedTypes = stateUpdated.form[field].value

    const shouldDeselect = allTypes.every((x) => selectedTypes.includes(x))
    onChangeDropdownMultiple(shouldDeselect ? [] : allTypes, field)
  }

  function onChangeInput(event, formInputName) {
    const updatedState = { ...state }

    let inputValue = event.target.value
    if (formInputName === 'status') {
      inputValue = event.target.checked
    }

    if (formInputName === 'isMaintenance') {
      inputValue = event.target.checked
    }

    if (formInputName === 'isTop') {
      inputValue = event.target.checked
    }

    if (formInputName === 'linkNewTab') {
      inputValue = event.target.checked | ''
    }
    const updatedForm = UpdateFormHelper(inputValue, formInputName, updatedState.form)
    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function generateSelectOptions(data) {
    const result = []
    if (Array.isArray(data)) {
      for (const e of data) {
        const option = { name: e, value: e }
        result.push(option)
      }
    } else {
      for (const key in data) {
        const option = { name: key, value: data[key] }
        result.push(option)
      }
    }
    return result
  }

  function allValueConversion(value, field) {
    let allTypes = []
    if (field === 'regions') allTypes = idcConfig.REACT_APP_DEFAULT_REGIONS
    else if (field === 'userTypes') allTypes = Object.values(EnrollAccountType)

    if (value.includes('all')) return allTypes
    else if (allTypes.every((x) => value.includes(x))) return ['all']

    return value
  }

  const showRequiredFields = async () => {
      const stateCopy = { ...state }
      // Mark regular Inputs
      const updatedForm = showFormRequiredFields(stateCopy.form)
      // Create toast
      showError(toastMessageEnum.formValidationError, false)
      stateCopy.form = updatedForm
      setState(stateCopy)
  }

  return (
    <BannerCreate
      state={state}
      showModal={showModal}
      showBanner={showBanner}
      setShowBanner={setShowBanner}
      onSubmit={onSubmit}
      onChangeDropdownMultiple={onChangeDropdownMultiple}
      onSelectAll={onSelectAll}
      onChangeInput={onChangeInput}
    />
  )
}

export default BannerCreateContainer
