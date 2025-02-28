// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import CustomInput from '../../../utils/customInput/CustomInput'
import { Button, ButtonGroup, Card, Modal, ModalBody } from 'react-bootstrap'
import '../LearningLabs.scss'
import imagePlaceholder from '../../../assets/images/imagePlaceholder.png'
import imagePlaceholderDark from '../../../assets/images/imagePlaceholderDark.png'
import idcConfig from '../../../config/configurator'
import SpinnerIcon from '../../../utils/spinner/SpinnerIcon'

interface TextToImageConverterProps {
  loading: boolean
  openSettings: boolean
  isDarkMode: boolean
  generatedImage: any
  onChangeInput: (event: any, formInputName: string) => void
  onSubmit: (event: any) => void
  setOpenSettings: (status: boolean) => void
  state: any
  navigationButtons: any
  modalButtons: any
}

const TextToImageConverter: React.FC<TextToImageConverterProps> = (props): JSX.Element => {
  const onChangeInput = props.onChangeInput
  const state = props.state
  const mainTitle = state.mainTitle
  const form = state.form
  const generatedImage = props.generatedImage
  const loading = props.loading
  const openSettings = props.openSettings
  const setOpenSettings = props.setOpenSettings
  const modalButtons = props.modalButtons
  const navigationButtons = props.navigationButtons
  const isDarkMode = props.isDarkMode
  const onSubmit = props.onSubmit

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
        onKeyDown={(e) => {
          if (element.id === 'prompt' && e.key === 'Enter') {
            onSubmit(e)
          }
        }}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        validationMessage={element.configInput.validationMessage}
        options={element.configInput.options}
        maxWidth={element.configInput.maxWidth}
        customClass={element.configInput.customClass}
      />
    )
  }

  // variables
  const formElementsConfiguration = []
  const formElementPrompt = {
    id: 'prompt',
    configInput: form.prompt
  }

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.sectionGroup === 'configuration') {
      formElementsConfiguration.push({
        id: key,
        configInput: formItem
      })
    }
  }

  const spinner = (
    <>
      <div className="position-absolute"></div>
    </>
  )

  const initialMessage = <div className="position-absolute"></div>

  const imageSection = (
    <>
      <img src={generatedImage} className="imgShadow" style={{ width: '100%' }} />
    </>
  )

  const placeHolderSection = (
    <>
      <img src={isDarkMode ? imagePlaceholderDark : imagePlaceholder} className="imgShadow" style={{ width: '100%' }} />
      {loading ? spinner : initialMessage}
    </>
  )

  return (
    <>
      {/* configuration form */}
      <Modal
        show={openSettings}
        backdrop="static"
        keyboard={false}
        onHide={() => {
          setOpenSettings(false)
        }}
      >
        <Modal.Header closeButton>
          <Modal.Title>Options</Modal.Title>
        </Modal.Header>
        <ModalBody>
          <div className="d-flex flex-column align-self-xs-stretch">
            <Card>
              <Card.Body className="gap-s6">
                {formElementsConfiguration.map((element) => buildCustomInput(element))}
                <ButtonGroup className="pt-s6">
                  {modalButtons.map((item: any, index: number) => (
                    <Button
                      intc-id={`btn-textToImage-modalButtons ${String(item.buttonLabel)}`}
                      data-wap_ref={`btn-textToImage-modalButtons ${String(item.buttonLabel)}`}
                      key={index}
                      variant={item.buttonVariant}
                      onClick={(e) => item.buttonFunction(e)}
                    >
                      {item.buttonLabel}
                    </Button>
                  ))}
                </ButtonGroup>
              </Card.Body>
            </Card>
          </div>
        </ModalBody>
      </Modal>
      <div className="section">
        <h2 intc-id="TextToImageTitle">{mainTitle}</h2>
        <span className="maw-900">
          Leverage {idcConfig.REACT_APP_COMPANY_SHORT_NAME}&apos;s advanced AI capabilities to generate high-quality
          images from text. Utilize hosted, scalable models to enhance the performance of AI-driven applications,
          seamlessly integrating into your existing workflows.
        </span>
      </div>
      <div className="d-flex flex-row gap-s4 w-100 flex-column flex-md-row p-s5 my-s6">
        {buildCustomInput(formElementPrompt)}

        <ButtonGroup className="justify-content-center justify-content-sm-center justify-content-md-start miw-80 mt-s9 buttonGroupGap">
          <Button
            intc-id={'btn-textToImage-navigationButtons Generate'}
            data-wap_ref={'btn-textToImage-navigationButtons Generate'}
            variant="primary"
            onClick={(e) => {
              onSubmit(e)
            }}
            disabled={loading}
          >
            {loading ? (
              <>
                <SpinnerIcon /> Generating...
              </>
            ) : (
              'Generate'
            )}
          </Button>
          {navigationButtons.map((item: any, index: number) => (
            <Button
              intc-id={`btn-textToImage-navigationButtons ${String(item.buttonLabel)}`}
              data-wap_ref={`btn-textToImage-navigationButtons ${String(item.buttonLabel)}`}
              key={index}
              variant={item.buttonVariant}
              onClick={(e) => item.buttonFunction(e)}
            >
              {item.buttonLabel}
            </Button>
          ))}
        </ButtonGroup>
      </div>
      {/* prompt and image */}
      <div className="section w-100 h-100 flex-column-reverse flex-sm-column pt-s8">
        <div
          className={'bannerBackground  d-flex align-self-center align-items-center h-100 w-100 justify-content-center'}
        >
          {generatedImage && !loading ? imageSection : placeHolderSection}
        </div>
      </div>
    </>
  )
}

export default TextToImageConverter
