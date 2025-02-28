// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useMemo, useRef } from 'react'
import CustomInput from '../../../utils/customInput/CustomInput'
import { Button, ButtonGroup, Form, FormControl, InputGroup, Modal } from 'react-bootstrap'
import Spinner from '../../../utils/spinner/Spinner'
import './MaaSLLMChat.scss'
import { type ChatItem } from '../../../store/maas/MaasStore'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import { BsArrowRightSquare } from 'react-icons/bs'
import EmptyView from '../../../utils/emptyView/EmptyView'
import ComingSoonBanner from '../../../utils/comingSoonBanner/ComingSoonBanner'
import type Software from '../../../store/models/Software/Software'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import UpgradeNeededModal from '../../../utils/modals/upgradeNeededModal/UpgradeNeededModal'

interface MaaSLLMChatProps {
  loadingSoftware: boolean
  loadingResponse: boolean
  generatedMessage: string
  chat: ChatItem[]
  actions: any[]
  setAction: (action: any) => void
  onChangeInput: (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>,
    formInputName: string
  ) => void
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void
  onExampleSelected: (event: any, value: string) => void
  state: {
    form: Record<string, any>
    navigationBottom: Array<{
      buttonLabel: string
      buttonVariant: string
      buttonFunction: (event: React.MouseEvent<HTMLButtonElement>) => void
    }>
    defaultExamples: Array<{
      buttonLabel: string
    }>
  }
  errorModal: {
    show: boolean
    title: string
    description: string
    message: string
    hideRetryMessage: boolean
    onClose: () => void
  }
  showSettings: boolean
  setShowSettings: (value: boolean) => void
  noFoundSoftware: {
    title: string
    subTitle: string
    action: any
  }
  isAvailable: boolean
  comingMessage: string
  software: Software | null
  showUpgradeNeededModal: boolean
  setShowUpgradeNeededModal: React.Dispatch<React.SetStateAction<boolean>>
}

const MaaSLLMChat: React.FC<MaaSLLMChatProps> = ({
  loadingSoftware,
  loadingResponse,
  generatedMessage,
  chat,
  onChangeInput,
  onSubmit,
  onExampleSelected,
  state,
  errorModal,
  actions,
  setAction,
  showSettings,
  setShowSettings,
  noFoundSoftware,
  isAvailable,
  comingMessage,
  software,
  showUpgradeNeededModal,
  setShowUpgradeNeededModal
}) => {
  const { form, defaultExamples } = state
  const divRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = (): void => {
    if (divRef.current) {
      divRef.current.scrollTop = divRef.current.scrollHeight
    }
  }

  useEffect(() => {
    scrollToBottom()
  }, [chat, generatedMessage])

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
        maxWidth={element.configInput.maxWidth}
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
    <div className="section align-self-center justify-content-center h-100">
      <span className="lead text-center">Try out Large Language Models using this chat experience</span>
    </div>
  )

  const header = (
    <div className=" section d-none d-sm-flex w-100">
      <h2>{software?.displayName}</h2>
    </div>
  )

  const actionButtons = (size: string): JSX.Element => {
    const display = size === 'xs' ? 'd-flex d-sm-none' : 'd-none d-sm-flex'
    const alignment = size === 'xs' ? 'align-self-center' : 'align-self-end'

    return (
      <ButtonGroup className={`${display} ${alignment}`}>
        {actions.length > 0
          ? actions.map((action: any, index: number) => (
              <Button
                intc-id={`btn-llm-action ${action.label}`}
                data-wap_ref={`btn-llm-action ${action.label}`}
                key={index}
                variant={action.variant}
                onClick={() => {
                  setAction(action)
                }}
              >
                {action.label}
              </Button>
            ))
          : null}
      </ButtonGroup>
    )
  }

  const settingsModal = (
    <Modal
      show={showSettings}
      onHide={() => {
        setShowSettings(false)
      }}
      backdrop="static"
      keyboard={false}
      centered
      aria-label="Progress bar modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>Settings</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="section">{formElementsConfiguration.map((element) => buildCustomInput(element))}</div>
      </Modal.Body>
      <Modal.Footer>
        <Button
          variant="outline-primary"
          onClick={() => {
            setShowSettings(false)
          }}
          intc-id="btn-llm-action-save-settings"
        >
          Save
        </Button>
      </Modal.Footer>
    </Modal>
  )

  const notFound = (
    <EmptyView title={noFoundSoftware.title} subTitle={noFoundSoftware.subTitle} action={noFoundSoftware.action} />
  )

  const formatChatText = (text: string): JSX.Element[] => {
    return text.split('\n').map((line, index) => (
      <React.Fragment key={index}>
        {line} <br />
      </React.Fragment>
    ))
  }

  const formatChatResponseText = (text: string): JSX.Element => {
    return (
      <ReactMarkdown linkTarget="_blank" remarkPlugins={[remarkGfm]}>
        {text}
      </ReactMarkdown>
    )
  }

  return !isAvailable ? (
    <ComingSoonBanner message={comingMessage} />
  ) : (
    <>
      <ErrorModal
        showModal={errorModal.show}
        titleMessage={errorModal.title}
        description={errorModal.description}
        message={errorModal.message}
        hideRetryMessage={errorModal.hideRetryMessage}
        onClickCloseErrorModal={errorModal.onClose}
      />
      <UpgradeNeededModal
        showModal={showUpgradeNeededModal}
        onClose={() => {
          setShowUpgradeNeededModal(false)
        }}
      />
      {settingsModal}
      <div className="section fixedContainer">
        {loadingSoftware ? (
          <div className="section align-self-center justify-content-center h-100">
            <Spinner />
          </div>
        ) : software ? (
          <>
            {header}
            {actionButtons('sm')}
            {chat.length > 0 || generatedMessage ? (
              <div ref={divRef} className="section align-self-center chatContainer">
                {chat.map((item: ChatItem, index: number) =>
                  item.isResponse ? (
                    <div className="chatResponseItem p-s6" key={index}>
                      <span>{formatChatResponseText(item.text)}</span>
                    </div>
                  ) : (
                    <p className="chatRequestItem p-s6" key={index}>
                      <span>{formatChatText(item.text)}</span>
                    </p>
                  )
                )}
                {loadingResponse && !generatedMessage ? (
                  <div className="d-flex w-100 justify-content-center">{<Spinner />}</div>
                ) : null}
                {generatedMessage ? (
                  <div className="chatResponseItem p-s6">
                    <span>{formatChatResponseText(generatedMessage)}</span>
                  </div>
                ) : null}
              </div>
            ) : (
              initialMessage
            )}
            <Form onSubmit={onSubmit} className="d-flex w-100 justify-content-center">
              <InputGroup className="chatInputs gap-s5">
                {chat.length === 0 ? (
                  <>
                    <span className="align-self-start">Examples</span>
                    <div className="d-flex flex-wrap w-100 gap-s5">
                      {defaultExamples.map((item: any, index: number) => (
                        <Button
                          className="col"
                          intc-id={`btn-llm-examples ${item.buttonLabel}`}
                          data-wap_ref={`btn-llm-examples ${item.buttonLabel}`}
                          key={index}
                          variant="tag"
                          onClick={(event) => {
                            onExampleSelected(event, item.buttonLabel)
                          }}
                        >
                          {item.buttonLabel}
                        </Button>
                      ))}
                    </div>
                  </>
                ) : null}
                <div className="d-flex w-100">
                  <FormControl
                    intc-id={'input-llm-chat'}
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
                    variant="link"
                    type="submit"
                    aria-label="Generate response"
                    intc-id={'btn-llm-submit'}
                    data-wap_ref={'btn-llm-submit'}
                    disabled={loadingResponse}
                  >
                    <BsArrowRightSquare />
                  </Button>
                </div>
              </InputGroup>
            </Form>
            {actionButtons('xs')}
          </>
        ) : (
          notFound
        )}
      </div>
    </>
  )
}

export default MaaSLLMChat
