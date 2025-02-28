import React from 'react'
import CustomInput from '../../../utility/customInput/CustomInput'
import Button from 'react-bootstrap/Button'
import { ButtonGroup } from 'react-bootstrap'
import OnSubmitModal from '../../../utility/modals/onSubmitModal/OnSubmitModal'

import OnCreateCoupon from '../../../utility/modals/onCreateCoupon/OnCreateCoupon'

const CloudCreditCreate = (props) => {
  const state = props.state
  const desciption = state.desciption
  const form = state.form
  const navigationTop = state.navigationTop
  const navigationBottom = state.navigationBottom
  const isValidForm = state.isValidForm
  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  const showModal = props.showModal
  const onCreateOkModal = props.onCreateOkModal
  const formElements = []
  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    formElements.push({
      id: key,
      configInput: formItem
    })
  }

  return (
        <>
            <div className="section">
                {
                    navigationTop.map((item, index) => (
                        <div key={index}>
                            {
                                <Button
                                    intc-id={`navigationTop${item.label}`}
                                    className='p-s0'
                                    variant={item.buttonVariant}
                                    onClick={item.function}>
                                    {item.label}
                                </Button>
                            }
                        </div>
                    ))
                }
            </div>
            <OnSubmitModal showModal={showModal} message="Working on your request" />
            <OnCreateCoupon modal={onCreateOkModal}/>
            <div className="section">
                <h2 className='h4'>{desciption}</h2>
            </div>
            <div className="section">
                {formElements.map((element, index) => (
                        <CustomInput
                            key={index}
                            customContent={element.configInput.customContent}
                            type={element.configInput.type}
                            fieldSize={element.configInput.fieldSize}
                            placeholder={element.configInput.placeholder}
                            isRequired={element.configInput.validationRules.isRequired}
                            label={element.configInput.validationRules.isRequired ? element.configInput.label + ' *' : element.configInput.label}
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
                            isMultiple={element.configInput.isMultiple}
                            hiddenLabel={element.configInput.hiddenLabel}
                            onChangeSelectValue={element.configInput.onChangeSelectValue}
                            extraButton={element.configInput.extraButton}
                            emptyOptionsMessage={element.configInput.emptyOptionsMessage}
                        />
                ))}
                <ButtonGroup>
                    {
                        navigationBottom.map((item, index) => (
                            <Button
                                key={index}
                                intc-id={`navigationTop${item.label}`}
                                disabled={item.label === 'Create' ? !isValidForm : false}
                                variant={item.buttonVariant}
                                onClick={item.label === 'Create' ? onSubmit : item.function}>
                                {item.label}
                            </Button>
                        ))
                    }
                </ButtonGroup>
            </div>
        </>
  )
}

export default CloudCreditCreate
