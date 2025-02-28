// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo } from 'react'
import CustomInput from '../../../utils/customInput/CustomInput'
import { Button, ButtonGroup, Card, Form, FormControl, InputGroup } from 'react-bootstrap'
import { BsSend } from 'react-icons/bs'
import Spinner from '../../../utils/spinner/Spinner'

interface LLMChatProps {
  loading: boolean
  error: boolean
  generatedMessage: any
  onChangeInput: (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>,
    formInputName: string
  ) => void
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void
  onReset: () => void
  state: {
    mainTitle: string
    form: Record<string, any>
    navigationBottom: Array<{
      buttonLabel: string
      buttonVariant: string
      buttonFunction: (event: React.MouseEvent<HTMLButtonElement>) => void
    }>
  }
}

const LLMChat: React.FC<LLMChatProps> = ({
  loading,
  error,
  generatedMessage,
  onChangeInput,
  onSubmit,
  onReset,
  state
}) => {
  const { form, navigationBottom } = state

  const formElementChat = form.chat
  const formElementsConfiguration = useMemo(() => {
    return Object.entries(form)
      .filter(([_, value]) => value.sectionGroup === 'configuration')
      .map(([key, value]) => ({
        id: key,
        configInput: value
      }))
  }, [form])

  const initialMessage = (
    <div className="section align-self-center align-items-center">
      <h3>Start chatting</h3>
      <span className="lead">Enter a prompto to test your model behavior</span>
    </div>
  )
  const spinner = <Spinner />

  const imessageSection = <div className="section align-self-center align-items-center">{generatedMessage}</div>

  const errorMessage = (
    <div className="section align-self-center align-items-center">
      <h3>There was an error processing your prompt</h3>
      <span className="lead">Try again or change prompt and configurations</span>
    </div>
  )

  const buildCustomInput = (element: any): JSX.Element => {
    return (
      <CustomInput
        key={element.id}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired
            ? String(element.configInput.label) + ' *'
            : element.configInput.label
        }
        value={element.configInput.value}
        minRange={element.configInput.minRange}
        maxRange={element.configInput.maxRange}
        step={element.configInput.step}
        onChanged={(event: any) => {
          onChangeInput(event, element.id)
        }}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        validationMessage={element.configInput.validationMessage}
        options={element.configInput.options}
        textAreaRows={element.configInput.textAreaRows}
      />
    )
  }

  return (
    <>
      <div className="section">
        <h2 intc-id="LLMChatTitle">{state.mainTitle}</h2>
      </div>
      <div className="section flex-xs-column flex-lg-row">
        {/* configuration form */}
        <div className="d-flex flex-column align-self-xs-stretch">
          <Card className="chatPanel gap-s6">
            <Card.Body>
              {formElementsConfiguration.map((element) => buildCustomInput(element))}
              <ButtonGroup className="pt-s6">
                {navigationBottom.map((item: any, index: number) => (
                  <Button
                    intc-id={`btn-llm-navigationBottom ${String(item.buttonLabel)}`}
                    data-wap_ref={`btn-llm-navigationBottom ${String(item.buttonLabel)}`}
                    key={index}
                    variant="outline-primary"
                    onClick={(event) => item.buttonFunction(event)}
                  >
                    {item.buttonLabel}
                  </Button>
                ))}
              </ButtonGroup>
            </Card.Body>
          </Card>
        </div>
        {/* prompt and messages */}
        <div className="d-flex flex-column align-self-stretch flex-fill">
          <Card className="chatWindow w-100 h-100">
            <Card.Body className="gap-s6">
              <div className="d-flex flex-row flex-grow-1 align-self-center align-items-center">
                {loading ? spinner : error ? errorMessage : generatedMessage ? imessageSection : initialMessage}
              </div>
              <Form onSubmit={onSubmit} className="d-flex w-100">
                <InputGroup className="chatBox">
                  <FormControl
                    type="text"
                    role="searchbox"
                    placeholder="Enter prompt"
                    aria-label="Enter prompt"
                    value={formElementChat.value}
                    onChange={(event) => {
                      onChangeInput(event, 'chat')
                    }}
                  />
                  <Button
                    variant="icon-simple"
                    type="submit"
                    aria-label={`${String(formElementChat.label)} Submit Button`}
                    intc-id={`${String(formElementChat.label).replace(' ', '')}-SubmitButton`}
                    data-wap_ref={`${String(formElementChat.label).replace(' ', '')}-SubmitButton`}
                  >
                    <BsSend />
                  </Button>
                </InputGroup>
              </Form>
            </Card.Body>
          </Card>
        </div>
      </div>
    </>
  )
}

export default LLMChat
