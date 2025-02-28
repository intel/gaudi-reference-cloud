// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { Button, ButtonGroup } from 'react-bootstrap'
import CustomInput from '../../utility/customInput/CustomInput'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'

const ProductCatalogCreate = (props: any): JSX.Element => {
  const onCancel = props.onCancel
  const state = props.state
  const title = state.title
  const form = state.form
  const onChangeInput = props.onChangeInput
  const navigationBottom = state.navigationBottom
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const onSubmit = props.onSubmit
  const selectAllButton = props.selectAllButton
  const submitModal = props.submitModal
  const onClickAddMetaData = props.onClickAddMetaData
  const newMetaDataSets = props.newMetaDataSets
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

  // *****
  // Functions
  // *****
  const buildCustomInput = (element: any, key: number): JSX.Element => {
    return <CustomInput
      key={key}
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
      onChanged={(event: any) => onChangeInput(event, element.id, element.idParent, element.nodeIndex)}
      onChangeDropdownMultiple={onChangeDropdownMultiple}
      isValid={element.configInput.isValid}
      isTouched={element.configInput.isTouched}
      helperMessage={element.configInput.helperMessage}
      isReadOnly={element.configInput.isReadOnly}
      options={element.configInput.options}
      validationMessage={element.configInput.validationMessage}
      refreshButton={element.configInput.refreshButton}
      extraButton={element.configInput.extraButton}
      selectAllButton={selectAllButton}
      labelButton={element.configInput.labelButton}
      emptyOptionsMessage={element.configInput.emptyOptionsMessage}
    />
  }
  return <>
    <OnSubmitModal showModal={submitModal.show} message={submitModal.message}></OnSubmitModal>
    <div className="section">
      <Button variant="link" className="p-s0" onClick={() => onCancel()}>
        ⟵ Back to products
      </Button>
    </div>
    <div className="section">
      <h2 intc-id="maintitle">{title}</h2>
    </div>
    <div className="section">
      {formElements.map((element, index) => buildCustomInput(element, index))}
    </div>
    <div className='section'>
      <Button
        intc-id={'btn-add-metadata'}
        data-wap_ref={'btn-add-metadata'}
        aria-label={'btn-add-metadata'}
        variant={'outline-primary'}
        onClick={() => onClickAddMetaData()}
      >
        {newMetaDataSets.length === 0 ? 'Add metadata' : 'Edit metadata'}
      </Button>
    </div>
    <div className='section'>
      <ButtonGroup>
        {navigationBottom.map((item: any, index: string) => (
          <Button
            intc-id={`btn-navigationBottom ${item.buttonLabel}`}
            data-wap_ref={`btn-navigationBottom ${item.buttonLabel}`}
            aria-label={item.buttonLabel}
            key={index}
            variant={item.buttonVariant}
            onClick={item.buttonAction === 'create' ? onSubmit : onCancel}
          >
            {item.buttonLabel}
          </Button>
        ))}
      </ButtonGroup>
    </div>
  </>
}

export default ProductCatalogCreate
