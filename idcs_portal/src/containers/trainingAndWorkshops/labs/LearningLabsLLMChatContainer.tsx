// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState } from 'react'
import LLMChat from '../../../components/learningLabs/llmChat/LLMChat'
import { UpdateFormHelper, isValidForm, getFormValue } from '../../../utils/updateFormHelper/UpdateFormHelper'
import useLearningLabsStore from '../../../store/learningLabsStore/LearningLabsStore'

const DEFAULT_PROMPT =
  'You are an AI assistant, provide concise and precise answers for the user query within the specified token limit, do not generate lengthy answers. Do not hallucinate. Please do not generate incomplete sentences.'

const LearningLabsLLMChatContainer = (): JSX.Element => {
  // Local state
  const initialState = {
    mainTitle: 'LLM Chat',
    form: {
      model: {
        sectionGroup: 'configuration',
        type: 'dropdown',
        label: 'Model:',
        value: 'lama3',
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        options: [
          { name: 'Lama3', value: 'lama3' },
          { name: 'Mistral-7B-Instruct', value: 'Mistral-7B-Instruct' },
          { name: 'meta-llama/Llama-2-70b-chat-hf', value: 'meta-llama/Llama-2-70b-chat-hf' }
        ],
        validationRules: {
          isRequired: true
        }
      },
      maxToken: {
        sectionGroup: 'configuration',
        type: 'number',
        label: 'Max Token Len:',
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
        minRange: 0.1,
        maxValue: 1.0,
        step: 0.1,
        value: 0.5,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        }
      },
      top_p: {
        sectionGroup: 'configuration',
        type: 'range',
        label: 'Top P:',
        minRange: 0.1,
        maxValue: 1.0,
        step: 0.1,
        value: 0.5,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        }
      },
      top_k: {
        sectionGroup: 'configuration',
        type: 'range',
        label: 'Top K:',
        minRange: 0.1,
        maxValue: 1.0,
        step: 0.1,
        value: 0.5,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        }
      },
      prompt: {
        sectionGroup: 'configuration',
        type: 'textArea',
        label: 'Instructions:',
        value: DEFAULT_PROMPT,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        maxLength: 500,
        textAreaRows: 7,
        validationRules: {
          isRequired: false,
          checkMaxLength: true
        },
        validationMessage: ''
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
        maxLength: 200,
        validationRules: {
          isRequired: false,
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
    },
    navigationBottom: [
      {
        buttonLabel: 'Try',
        buttonVariant: 'primary',
        buttonFunction: (event: any) => {
          onSubmit(event)
        }
      },
      {
        buttonLabel: 'Reset',
        buttonVariant: 'secondary',
        buttonFunction: () => {
          onReset()
        }
      }
    ]
  }

  // States
  const [state, setState] = useState(initialState)
  const [error, setError] = useState(false)
  const loading = useLearningLabsStore((state) => state.loading)
  const getMessageFromText = useLearningLabsStore((state) => state.getMessageFromText)
  const generatedMessage = useLearningLabsStore((state) => state.generatedMessage)
  const reset = useLearningLabsStore((state) => state.reset)

  const onReset = (): void => {
    const updatedState = {
      ...state
    }

    let updatedForm = UpdateFormHelper('lama3', 'model', updatedState.form)
    updatedForm = UpdateFormHelper(512, 'maxToken', updatedState.form)
    updatedForm = UpdateFormHelper(0.5, 'temperature', updatedState.form)
    updatedForm = UpdateFormHelper(0.5, 'top_p', updatedState.form)
    updatedForm = UpdateFormHelper(0.5, 'top_k', updatedState.form)
    updatedForm = UpdateFormHelper(DEFAULT_PROMPT, 'prompt', updatedState.form)
    updatedForm = UpdateFormHelper('', 'chat', updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
    setError(false)
    reset()
  }

  const handleSubmit = async (event: any): Promise<void> => {
    event.preventDefault()

    const params = {
      inputs: getFormValue('chat', state.form) || DEFAULT_PROMPT,
      prompt: getFormValue('prompt', state.form),
      model: getFormValue('model', state.form),
      max_new_tokens: parseInt(getFormValue('maxToken', state.form), 10),
      temperature: parseFloat(getFormValue('temperature', state.form)),
      top_k: Math.max(1, parseInt(getFormValue('top_k', state.form), 10)),
      top_p: parseFloat(getFormValue('top_p', state.form))
    }

    try {
      await getMessageFromText(params)
      setError(false)
    } catch (error) {
      setError(true)
    }
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

  return (
    <LLMChat
      onChangeInput={onChangeInput}
      onSubmit={onSubmit}
      error={error}
      onReset={onReset}
      generatedMessage={generatedMessage}
      state={state}
      loading={loading}
    />
  )
}

export default LearningLabsLLMChatContainer
