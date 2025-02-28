import React from 'react'
import CustomInput from '../../utility/customInput/CustomInput'
import Button from 'react-bootstrap/Button'
import { ButtonGroup } from 'react-bootstrap'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'
import CustomAlerts from '../../utility/customAlerts/CustomAlerts'

const BannerCreate = (props) => {
  const state = props.state
  const title = state.title
  const description = state.description
  const form = state.form
  const navigationTop = state.navigationTop
  const navigationBottom = state.navigationBottom
  const isValidForm = state.isValidForm
  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  const showBanner = props.showBanner
  const setShowBanner = props.setShowBanner
  const showModal = props.showModal
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const onSelectAll = props.onSelectAll

  const bannerInfo = []
  const bannerConfig = []
  const bannerLink = []
  for (const key in form) {
    const formItem = { ...form[key] }

    if (key === 'userTypes' || key === 'regions') {
      formItem.selectAllButton.buttonFunction = () => onSelectAll(key)
    }

    if (formItem.section === 'banner-info') {
      bannerInfo.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.section === 'banner-link') {
      bannerLink.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.section === 'banner-config') {
      bannerConfig.push({
        id: key,
        configInput: formItem
      })
    }
  }

  const buildCustomInput = (element) => {
    return <CustomInput
            key={element.id}
            customContent={element.configInput.customContent}
            hideContent={false}
            type={element.configInput.type}
            fieldSize={element.configInput.fieldSize}
            placeholder={element.configInput.placeholder}
            isRequired={element.configInput.validationRules.isRequired}
            label={
              element.configInput.validationRules.isRequired
                ? element.configInput.label + ' *'
                : element.configInput.label
            }
            value={element.configInput.value}
            onChanged={(event) => onChangeInput(event, element.id)}
            isValid={element.configInput.isValid}
            isTouched={element.configInput.isTouched}
            helperMessage={element.configInput.helperMessage}
            isReadOnly={element.configInput.isReadOnly}
            options={element.configInput.options}
            validationMessage={element.configInput.validationMessage}
            readOnly={element.configInput.readOnly}
            refreshButton={element.configInput.refreshButton}
            borderlessDropdownMultiple={element.configInput.borderlessDropdownMultiple}
            onChangeSelectValue={element.configInput.onChangeSelectValue}
            extraButton={element.configInput.extraButton}
            emptyOptionsMessage={element.configInput.emptyOptionsMessage}
            selectAllButton={element.configInput.selectAllButton}
            onChangeDropdownMultiple={(value) => onChangeDropdownMultiple(value, element.id)}
          />
  }

  const getLink = () => {
    const linkLabel = state?.form?.linkLabel?.value
    const linkHref = state?.form?.linkHref?.value
    const linkNewTab = state?.form?.linkNewTab?.value

    const link =
      linkLabel && linkHref
        ? {
            label: linkLabel,
            href: linkHref,
            openInNewTab: linkNewTab
          }
        : undefined

    return link
  }

  return (
    <>
      <div className="section">
        {navigationTop.map((item, index) => (
          <div key={index}>
            {
              <Button
                intc-id={`navigationTop${item.label}`}
                className="p-s0"
                variant={item.buttonVariant}
                onClick={item.function}
              >
                {item.label}
              </Button>
            }
          </div>
        ))}
      </div>
      <OnSubmitModal showModal={showModal} message="Working on your request" />
      <div className="section">
        <h2 className="h4">{title}</h2>
        <p>{description}</p>
      </div>
      <div className="section">
        {showBanner ? (
          <CustomAlerts
            showAlert={showBanner}
            alertType={state?.form?.type?.value}
            title={state?.form?.title?.value}
            link={getLink()}
            message={state?.form?.message?.value}
            onCloseAlert={() => setShowBanner(false)}
            className="w-100"
          />
        ) : (
          ''
        )}
      </div>
      <div className="section">
        <h2 className="h5">Banner Information</h2>
        {bannerInfo.map((element) => buildCustomInput(element))}
      </div>

      <div className="section">
        <h2 className="h5">Banner Configuration</h2>
        {bannerConfig.map((element) => buildCustomInput(element))}
      </div>

      <div className="section">
        <h2 className="h5">Banner Link</h2>
        <p>
          To embed a hyperlink within the banner, please utilize the following fields provided below. Each field is
          designed <br /> to capture specific details necessary for creating an effective and functional link. Ensure
          that you enter the <br />
          correct information in the corresponding fields to guarantee the link directs users to the intended
          destination.
        </p>
        {bannerLink.map((element) => buildCustomInput(element))}
      </div>

      <div className="section">
        <p>
          Before finalizing the creation of your banner, it is recommended to use the <strong>Preview</strong> button.
          This feature allows you <br /> to view a mock-up of the banner as it will appear to users. Previewing is a
          crucial step that helps ensure all content <br /> and formatting are correct and meet your expectations.
          Please take a moment to review your banner thoroughly <br /> using this function to avoid any errors or
          necessary revisions after it goes live.
        </p>
        <ButtonGroup>
          {navigationBottom.map((item, index) => (
            <Button
              key={index}
              intc-id={`navigationTop${item.label}`}
              disabled={item.label === 'Create' || item.label === 'Update' ? !isValidForm : false}
              variant={item.buttonVariant}
              onClick={item.label === 'Create' || item.label === 'Update' ? onSubmit : item.function}
            >
              {item.label}
            </Button>
          ))}
        </ButtonGroup>
      </div>
    </>
  )
}

export default BannerCreate
