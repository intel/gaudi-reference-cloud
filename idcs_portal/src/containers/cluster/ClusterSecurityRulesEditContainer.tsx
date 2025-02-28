// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import useClusterStore from '../../store/clusterStore/ClusterStore'
import {
  UpdateFormHelper,
  getFormValue,
  isValidForm,
  setFormValue,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import ClusterSecurityRulesEdit from '../../components/cluster/clusterEditSecurityRules/ClusterSecurityRulesEdit'
import ClusterService from '../../services/ClusterService'

const ClusterSecurityRulesEditContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const { param: name } = useParams()
  const navigate = useNavigate()
  const securityRuleUuid = useClusterStore((state) => state.securityRuleUuid)
  const editSecurityRule = useClusterStore((state) => state.editSecurityRule)
  const clusterSecurityRules = useClusterStore((state) => state.clusterSecurityRules)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)
  // *****
  // Local state
  // *****

  const ipInput = {
    ip: {
      sectionGroup: 'configuration',
      type: 'text', // options = 'text ,'textArea'
      label: 'Source IP:',
      placeholder: 'e.g. 10.0.0.1 or 10.0.0.1 (/1 to /32) or any',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 50,
      validationRules: {
        isRequired: true,
        checkMaxLength: true,
        isSecuritySourceIP: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: (
        <div className="valid-feedback" intc-id="ipsValidMessage">
          Use &quot;any&quot; to allow access from anywhere. Specify a single IP (ex: 10.0.0.1) or CIDR-format (ex:
          10.0.0.1(/1 to /32))
        </div>
      )
    }
  }

  const initialState = {
    mainTitle: 'Edit cluster endpoint access',
    form: {
      ips: {
        label: 'Source Ips:',
        sectionGroup: 'sourceIps',
        items: [{ ...ipInput }],
        isValid: false,
        validationRules: {
          isRequired: true
        }
      },
      protocol: {
        sectionGroup: 'port',
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Protocol:',
        placeholder: 'Protocol',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [
          {
            name: 'TCP',
            value: 'TCP'
          },
          {
            name: 'UDP',
            value: 'UDP'
          }
        ],
        validationMessage: ''
      },
      port: {
        sectionGroup: 'port',
        type: 'integer', // options = 'text ,'textArea'
        label: 'Internal Port:',
        placeholder: 'e.g. 80',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: true, // Input create as read only
        maxLength: 5,
        validationRules: {
          isRequired: true,
          onlyCreditNumeric: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: ''
      },
      ipInternal: {
        sectionGroup: 'port',
        type: 'text', // options = 'text ,'textArea'
        label: 'Target Ip:',
        placeholder: 'e.g. 80',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: true, // Input create as read only
        maxLength: 5,
        validationRules: {
          isRequired: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: ''
      }
    },
    sourceIpsLimit: 20,
    isValidForm: false,
    servicePayload: {
      internalip: '',
      port: '',
      sourceip: [],
      protocol: []
    },
    navigationBottom: [
      {
        buttonAction: 'Submit',
        buttonLabel: 'Save',
        buttonVariant: 'primary'
      },
      {
        buttonAction: 'Function',
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => {
          onCancel()
        }
      }
    ]
  }

  const submitModalInitial = {
    show: false
  }

  const [state, setState] = useState(initialState)
  const [submitModal, setSubmitModal] = useState(submitModalInitial)
  // *****
  // use effect
  // *****
  useEffect(() => {
    if (clusterSecurityRules?.length === 0 && !editSecurityRule && !securityRuleUuid) {
      navigate({
        pathname: `/cluster/d/${name}`,
        search: 'tab=security'
      })
    }
    updateFormValues()
  }, [])

  // *****
  // Functions
  // *****
  const updateFormValues = (): void => {
    const stateUpdated = { ...state }
    const form = state.form
    const sourceIpsUpdated = { ...form.ips }
    const portValue = editSecurityRule?.port
    const sourceIps: any = editSecurityRule ? [...editSecurityRule?.sourceip] : []
    const protocols = [...(editSecurityRule?.protocol ?? [])]
    const protocol = protocols.length > 0 ? protocols[0] : null
    const internalIp = editSecurityRule?.destinationip
    for (let k = 0; k < sourceIps.length; k++) {
      if (sourceIps.length > sourceIpsUpdated.items.length) {
        sourceIpsUpdated.items.push({
          ip: { ...ipInput.ip }
        })
      }
      sourceIpsUpdated.items[k].ip.value = sourceIps[k].trim()
      sourceIpsUpdated.items[k].ip.isValid = true
    }

    sourceIpsUpdated.isValid = validateIpSourceRows(sourceIpsUpdated)
    form.ips = sourceIpsUpdated
    let formUpdated = setFormValue('port', portValue, form)
    formUpdated = setFormValue('protocol', protocol, formUpdated)
    formUpdated = setFormValue('ipInternal', internalIp, formUpdated)
    stateUpdated.isValidForm = isValidForm(formUpdated)
    setState(stateUpdated)
  }

  const goBack = (): void => {
    navigate({
      pathname: `/cluster/d/${name}`,
      search: 'tab=security'
    })
  }

  const onCancel = (): void => {
    // Navigates back to the page when this method triggers.
    goBack()
  }

  const onChangeInput = (event: any, formInputName: string, idParent: string = '', index: number): void => {
    const value = event.target.value
    const updatedState = {
      ...state
    }

    let updatedForm = updatedState.form

    if (idParent === 'ips') {
      const sourceIps = updatedForm.ips
      const sourceIpsItems = [...sourceIps.items]
      const sourceIpItem = sourceIpsItems[index]
      const updatedSource = UpdateFormHelper(value, formInputName, sourceIpItem)
      sourceIpsItems[index] = updatedSource
      updatedForm.ips.items = sourceIpsItems
      // // Validate rows
      updatedForm.ips.isValid = validateIpSourceRows(sourceIps)
    } else {
      updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)
    }

    updatedState.form = updatedForm
    updatedState.isValidForm = isValidForm(updatedForm)

    setState(updatedState)
  }

  const onClickActionSourceIp = (index: number, action: string): void => {
    const updatedState = {
      ...state
    }
    const form = state.form
    const sourceIpsUpdated = { ...form.ips }
    const itemsCopy = [...sourceIpsUpdated.items]
    switch (action) {
      case 'Delete':
        sourceIpsUpdated.items.splice(index, 1)
        break
      default: {
        const newSourceIp = { ...ipInput }
        itemsCopy.push(newSourceIp)
        sourceIpsUpdated.items = itemsCopy
        break
      }
    }
    sourceIpsUpdated.isValid = validateIpSourceRows(sourceIpsUpdated)
    form.ips = sourceIpsUpdated
    updatedState.isValidForm = isValidForm(form)
    updatedState.form = form
    setState(updatedState)
  }

  const validateIpSourceRows = (ipSources: any): boolean => {
    let isValidArray = true
    for (const index in ipSources.items) {
      const computeItem = { ...ipSources.items[index] }
      const isValidRow = isValidForm(computeItem)
      if (!isValidRow) {
        isValidArray = false
        break
      }
    }
    return isValidArray
  }

  const showRequiredFields = (): void => {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)
    // Mark array Inputs
    const sourceIps = updatedForm.ips
    const items = sourceIps.items
    const itemsUpdated = []
    for (const index in items) {
      const item = { ...items[index] }
      const updatedItem = showFormRequiredFields(item)
      itemsUpdated.push(updatedItem)
    }
    sourceIps.items = itemsUpdated
    updatedForm.sourceIps = sourceIps

    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  const onSubmit = async (): Promise<void> => {
    try {
      const isValidForm = state.isValidForm
      if (!isValidForm) {
        showRequiredFields()
        return
      }
      setSubmitModal({ ...submitModal, show: true })
      const payload = { ...state.servicePayload }
      payload.internalip = editSecurityRule?.destinationip ?? ''
      payload.port = getFormValue('port', state.form)
      const protocol = getFormValue('protocol', state.form)
      const sourceIps: any = []
      const protocols: any = []
      protocols.push(protocol)
      payload.protocol = protocols
      const formItems = { ...state.form.ips.items }
      for (const index in formItems) {
        const sourceIp = { ...formItems[index] }
        sourceIps.push(sourceIp.ip.value)
      }
      payload.sourceip = sourceIps
      if (editSecurityRule) {
        await ClusterService.putSecurityRules(securityRuleUuid, payload)
        showSuccess('Cluster Updated successfully', false)
        goBack()
      } else {
        showError('Unable to update security rule', false)
      }
      setSubmitModal({ ...submitModal, show: false })
    } catch (error: any) {
      const message = String(error.message)
      if (error.response) {
        const errData = error.response.data
        const errMessage = errData.message
        showError(errMessage, false)
      } else {
        showError(message, false)
      }
      setSubmitModal({ ...submitModal, show: false })
    }
  }

  return (
    <ClusterSecurityRulesEdit
      state={state}
      onChangeInput={onChangeInput}
      onSubmit={onSubmit}
      onClickActionSourceIp={onClickActionSourceIp}
      sourceIpsLimit={state.sourceIpsLimit}
      submitModal={submitModal}
    />
  )
}

export default ClusterSecurityRulesEditContainer
