// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState } from 'react'
import TextToImageConverter from '../../../components/learningLabs/textToImage/TextToImageConverter'
import { UpdateFormHelper, isValidForm, getFormValue } from '../../../utils/updateFormHelper/UpdateFormHelper'
import useLearningLabsStore from '../../../store/learningLabsStore/LearningLabsStore'
import useDarkModeStore from '../../../store/darkModeStore/DarkModeStore'
import useToastStore from '../../../store/toastStore/ToastStore'

const LearningLabsTextToImageContainer = (): JSX.Element => {
  // Local State
  const initialState = {
    mainTitle: 'Text-to-Image with Stable Diffusion',
    form: {
      guidance: {
        sectionGroup: 'configuration',
        type: 'range',
        label: 'Guidance scale:',
        minRange: 0.5,
        maxRange: 31.5,
        step: 0.5,
        value: 7.5,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        helperMessage: 'How strongly the image should follow the prompt'
      },
      inference: {
        sectionGroup: 'configuration',
        type: 'range',
        label: 'Inference Steps:',
        minRange: 10,
        maxRange: 100,
        value: 50,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        }
      },
      seed: {
        sectionGroup: 'configuration',
        type: 'integer',
        label: 'Seed:',
        placeholder: '',
        value: 1000,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        maxLength: 5,
        validationRules: {
          isRequired: false,
          checkMaxValue: 10000,
          checkMinValue: 1
        },
        validationMessage: ''
      },
      negativePrompt: {
        sectionGroup: 'configuration',
        type: 'textArea',
        label: 'Negative Prompt:',
        placeholder: "List out the things which you don't need to be part of image",
        value: '',
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        maxLength: 200,
        validationRules: {
          isRequired: false,
          checkMaxLength: true
        },
        validationMessage: ''
      },
      prompt: {
        sectionGroup: 'prompt',
        type: 'text',
        placeholder: 'Enter your prompt',
        label: 'Prompt:',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        maxLength: 200,
        maxWidth: '100%',
        customClass: 'mw-100',
        validationMessage: '',
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        }
      }
    },
    isValidForm: true,
    servicePayload: {
      metadata: {
        name: null
      },
      spec: {
        sshPublicKeyNames: []
      }
    }
  }

  const actionButtons = [
    {
      buttonLabel: 'Reset',
      buttonVariant: 'outline-primary',
      buttonFunction: () => {
        onReset()
      }
    }
  ]

  const modalButtons = [
    {
      buttonLabel: 'Save',
      buttonVariant: 'primary',
      buttonFunction: () => {
        onSettingsDone()
      }
    },
    ...actionButtons
  ]

  const navigationButtons = [
    {
      buttonLabel: 'Options',
      buttonVariant: 'outline-primary',
      buttonFunction: () => {
        setOpenSettings(true)
      }
    },
    ...actionButtons
  ]

  // States
  const [state, setState] = useState(initialState)
  const [openSettings, setOpenSettings] = useState(false)
  const loading = useLearningLabsStore((state) => state.loading)
  const getImageFromText = useLearningLabsStore((state) => state.getImageFromText)
  const generatedImage = useLearningLabsStore((state) => state.generatedImage)
  const reset = useLearningLabsStore((state) => state.reset)
  const isDarkMode = useDarkModeStore((state) => state.isDarkMode)
  const showError = useToastStore((state) => state.showError)

  const onReset = (): void => {
    const updatedState = {
      ...state
    }

    let updatedForm = UpdateFormHelper(7.5, 'guidance', updatedState.form)
    updatedForm = UpdateFormHelper(50, 'inference', updatedState.form)
    updatedForm = UpdateFormHelper(1000, 'seed', updatedState.form)
    updatedForm = UpdateFormHelper('', 'prompt', updatedState.form)
    updatedForm = UpdateFormHelper('', 'negativePrompt', updatedState.form)

    updatedForm.prompt.validationMessage = ''
    updatedForm.prompt.isTouched = false

    updatedState.form = updatedForm

    setState(updatedState)
    reset()
    setOpenSettings(false)
  }

  const onSettingsDone = (): void => {
    setOpenSettings(false)
  }

  const handleSubmit = async (event: any): Promise<void> => {
    event.preventDefault()
    const params = {
      prompts: getFormValue('prompt', state.form),
      height: 512,
      width: 512,
      num_inference_steps: getFormValue('inference', state.form),
      guidance_scale: getFormValue('guidance', state.form),
      negative_prompts: getFormValue('negativePrompt', state.form),
      seed: getFormValue('seed', state.form)
    }
    try {
      await getImageFromText(params)
    } catch (error) {
      showError('Try again or change prompt or options.', false)
    }
  }

  const onSubmit = (event: any): void => {
    if (getFormValue('prompt', state.form) !== '') {
      handleSubmit(event).catch((error) => {
        throw error
      })
    } else {
      const updatedState = { ...state }
      updatedState.isValidForm = false
      updatedState.form.prompt.isTouched = true
      updatedState.form.prompt.isValid = false
      setState(updatedState)
    }
  }

  const onChangeInput = (event: any, formInputName: string): void => {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedForm.prompt.validationMessage = ''

    updatedState.form = updatedForm

    setState(updatedState)
  }

  return (
    <TextToImageConverter
      loading={loading}
      generatedImage={generatedImage}
      onChangeInput={onChangeInput}
      onSubmit={onSubmit}
      state={state}
      openSettings={openSettings}
      setOpenSettings={setOpenSettings}
      modalButtons={modalButtons}
      navigationButtons={navigationButtons}
      isDarkMode={isDarkMode}
    />
  )
}

export default LearningLabsTextToImageContainer
