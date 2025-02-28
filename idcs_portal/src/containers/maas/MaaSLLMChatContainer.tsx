// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import MaaSLLMChat from '../../components/maas/llmChat/MaaSLLMChat'
import useSoftwareStore from '../../store/SoftwareStore/SoftwareStore'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import { useParams } from 'react-router-dom'
import useMaasStore from '../../store/maas/MaasStore'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import useAppStore from '../../store/appStore/AppStore'
import { isErrorInsufficientCredits } from '../../utils/apiError/apiError'

const MaaSLLMChatContainer = (): JSX.Element => {
  // Local state
  const initialState = {
    form: {
      maxToken: {
        sectionGroup: 'configuration',
        type: 'number',
        label: 'Max Response Token Length:',
        value: 512,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true,
          checkMaxValue: 512,
          checkMinValue: 1
        }
      },
      temperature: {
        sectionGroup: 'configuration',
        type: 'range',
        label: 'Temperature:',
        minRange: 0,
        maxRange: 0.99,
        step: 0.01,
        value: 0.5,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        }
      },
      chat: {
        sectionGroup: 'chat',
        type: 'text',
        label: 'Chat:',
        placeholder: 'Enter prompt',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        }
      }
    },
    isValidForm: false,
    servicePayload: {
      metadata: {
        name: null
      },
      spec: {
        sshPublicKeyNames: []
      }
    },
    navigationBottom: [
      {
        buttonLabel: 'Reset',
        buttonVariant: 'secondary',
        buttonFunction: () => {
          onReset()
        }
      }
    ],
    defaultExamples: [
      {
        buttonLabel: 'Tell me a joke'
      },
      {
        buttonLabel: 'What is Intel Tiber AI Cloud'
      },
      {
        buttonLabel: 'Show me a loop in Python'
      }
    ]
  }

  const errorModalInitial = {
    show: false,
    title: '',
    description: 'There was an error while processing your prompt',
    message: '',
    hideRetryMessage: true,
    onClose: () => {}
  }

  const actionOptions = [
    {
      id: 'clear',
      label: 'Clear',
      variant: 'outline-primary'
    },
    {
      id: 'settings',
      label: 'Settings',
      variant: 'primary'
    }
  ]

  const isAvailable = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_SOFTWARE)
  const comingMessage =
    'Get ready for Intel-optimized software stacks hosted on performance optimized Intel compute platforms'
  const noFoundSoftware = {
    title: 'No software found',
    subTitle: 'The page you are trying to access does not exist. \n You can go to any of the following links:',
    action: {
      type: 'redirect',
      btnType: 'link',
      href: '/software',
      label: 'Software catalog'
    }
  }

  // States
  const [state, setState] = useState(initialState)
  const [showSettings, setShowSettings] = useState(false)
  const [errorModal, setErrorModal] = useState(errorModalInitial)
  const [showUpgradeNeededModal, setShowUpgradeNeededModal] = useState(false)
  const softwareDetail = useSoftwareStore((state) => state.softwareDetail)
  const getSoftware = useSoftwareStore((state) => state.getSoftware)
  const loadingSoftware = useSoftwareStore((state) => state.loading)
  const loadingResponse = useMaasStore((state) => state.loading)
  const generatedMessage = useMaasStore((state) => state.generatedMessage)
  const generateStream = useMaasStore((state) => state.generateStream)
  const chat = useMaasStore((state) => state.chat)
  const addChatItem = useMaasStore((state) => state.addChatItem)
  const reset = useMaasStore((state) => state.reset)
  const showError = useToastStore((state) => state.showError)
  const addBreadcrumCustomTitle = useAppStore((state) => state.addBreadcrumCustomTitle)

  const { param: id } = useParams()

  useEffect(() => {
    if (isAvailable) {
      getSoftwareInfo()
      reset()
    }
  }, [])

  useEffect(() => {
    if (softwareDetail) {
      addBreadcrumCustomTitle(`/software/d/${softwareDetail.id}`, softwareDetail.displayName)
    }
  }, [softwareDetail])

  function getSoftwareInfo(): void {
    const fetch = async (): Promise<void> => {
      if (id) await getSoftware(id)
    }
    void fetch()
  }

  async function executeChatRequest(params: any): Promise<void> {
    try {
      await generateStream(params)
    } catch (error: any) {
      const errorModal = { ...errorModalInitial }
      if (error.response) {
        if (isErrorInsufficientCredits(error)) {
          // No Credits
          setShowUpgradeNeededModal(true)
        } else {
          errorModal.message = error.response.data.message
        }
      } else {
        errorModal.message = error.message
      }
      errorModal.onClose = () => {
        setErrorModal({ ...errorModalInitial })
      }
      errorModal.show = true
      errorModal.title = 'Could not fetch LLM response'
      setErrorModal({ ...errorModal })
    }
  }

  function setAction(action: any): void {
    switch (action.id) {
      case 'clear':
        onReset()
        break
      case 'settings':
        onSetShowSettings(true)
        break
      default: {
        break
      }
    }
  }

  const onReset = (): void => {
    const updatedState = {
      ...state
    }

    let updatedForm = UpdateFormHelper(512, 'maxToken', updatedState.form)
    updatedForm = UpdateFormHelper(0.5, 'temperature', updatedState.form)
    updatedForm = UpdateFormHelper('', 'chat', updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
    reset()
  }

  const onExampleSelected = (event: any, value: string): void => {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(value, 'chat', updatedState.form)

    updatedState.form = updatedForm
    updatedState.isValidForm = isValidForm(updatedState.form)

    setState(updatedState)
    onSubmit(event)
  }

  function showRequiredFields(): void {
    const stateCopy = { ...state }
    const updatedForm = showFormRequiredFields(stateCopy.form)
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  const handleSubmit = async (event: any): Promise<void> => {
    event.preventDefault()

    const validForm = isValidForm(state.form)
    if (!validForm) {
      showRequiredFields()
      return
    }

    const params = {
      prompt: getFormValue('chat', state.form),
      model: softwareDetail?.model,
      productName: softwareDetail?.name,
      productId: softwareDetail?.id,
      maxNewTokens: parseInt(getFormValue('maxToken', state.form), 10),
      temperature: parseFloat(getFormValue('temperature', state.form))
    }

    const updatedState = {
      ...state
    }
    const updatedForm = UpdateFormHelper('', 'chat', updatedState.form)
    updatedState.form = updatedForm
    setState(updatedState)

    addChatItem({ isResponse: false, text: params.prompt })

    await executeChatRequest(params)
  }

  const onSubmit = (event: any): void => {
    handleSubmit(event).catch((error) => {
      throw error
    })
  }

  const onChangeInput = (event: any, formInputName: string): void => {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function onSetShowSettings(value: boolean): void {
    setShowSettings(value)
  }

  return (
    <MaaSLLMChat
      onChangeInput={onChangeInput}
      onSubmit={onSubmit}
      onExampleSelected={onExampleSelected}
      generatedMessage={generatedMessage}
      chat={chat}
      state={state}
      errorModal={errorModal}
      loadingSoftware={loadingSoftware}
      loadingResponse={loadingResponse}
      actions={actionOptions}
      setAction={setAction}
      showSettings={showSettings}
      setShowSettings={onSetShowSettings}
      noFoundSoftware={noFoundSoftware}
      comingMessage={comingMessage}
      isAvailable={isAvailable}
      software={softwareDetail}
      showUpgradeNeededModal={showUpgradeNeededModal}
      setShowUpgradeNeededModal={setShowUpgradeNeededModal}
    />
  )
}

export default MaaSLLMChatContainer
