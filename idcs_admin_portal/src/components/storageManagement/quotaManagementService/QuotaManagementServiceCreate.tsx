// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { Button, ButtonGroup } from 'react-bootstrap'
import { BsNodePlusFill } from 'react-icons/bs'
import CustomInput from '../../../utility/customInput/CustomInput'
import OnSubmitModal from '../../../utility/modals/onSubmitModal/OnSubmitModal'
import SearchBox from '../../../utility/searchBox/SearchBox'
import EmptyView from '../../../utility/emptyView/EmptyView'
import SelectedAccountCard from '../../../utility/selectedAccountCard/SelectedAccountCard'

const QuotaManagementServiceCreate = (props: any): JSX.Element => {
  // *****
  // Variables
  // *****
  const state = props.state
  const mainTitle = state.mainTitle
  const form = state.form
  const onCancel = props.onCancel
  const onChangeInput = props.onChangeInput
  const onClickActionResourceItem = props.onClickActionResourceItem
  const serviceResourceLimit = props.serviceResourceLimit
  const navigationBottom = state.navigationBottom
  const submitModal = props.submitModal
  const onSubmit = props.onSubmit
  const formElements = []
  const formResourceItemsElements = []
  const moduleName = props.moduleName
  const onSearchCloudAccount = props.onSearchCloudAccount
  const cloudAccountSelected = props.cloudAccountSelected
  const searchModal = props.searchModal
  const isPageReady = props.isPageReady
  const emptyView = props.emptyView

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (key === 'serviceResources') {
      const serviceResourcesItems = formItem.items
      for (const index in serviceResourcesItems) {
        const item = { ...serviceResourcesItems[index] }
        const serviceResourceElements: any = []
        for (const itemkey in item) {
          const itemElement = { ...item[itemkey] }
          serviceResourceElements.push({
            id: itemkey,
            idParent: key,
            nodeIndex: index,
            configInput: itemElement
          })
        }
        formResourceItemsElements.push({
          id: key,
          items: serviceResourceElements
        })
      }
    } else {
      if (!formItem.hidden) {
        formElements.push({
          id: key,
          configInput: formItem
        })
      }
    }
  }

  // *****
  // Functions
  // *****
  const buildCustomInput = (element: any, key: number): JSX.Element => {
    let response = null

    if (element.id === 'cloudAccount') {
      response = <div className="d-flex-customInput" key={key}>
        <div className="customInputLabel gap-s6" intc-id="filterText">
          Cloud Account: *
        </div>
        <div className="input-group">
          <SearchBox
            intc-id="searchCloudAccounts"
            classCustom={
              !element.configInput.isValid && element.configInput.isTouched ? 'form-control is-invalid' : ''
            }
            placeholder={`${String(element.configInput.placeholder)}`}
            aria-label="Type to search cloud account..."
            value={element.configInput.value || ''}
            onChange={(e) => onChangeInput(e, element.id)}
            onClickSearchButton={onSearchCloudAccount}
          />
          {!element.configInput.isValid && element.configInput.isTouched
            ? (
              <div className="invalid-feedback mt-2" intc-id="terminateInstanceError">
                {element.configInput.validationMessage}
              </div>
            )
            : null}
        </div>
        {
          cloudAccountSelected?.show
            ? <div className="d-flex flex-row">
              <SelectedAccountCard selectedCloudAccount={cloudAccountSelected}/>
            </div> : null
        }
      </div>
    } else {
      response = <CustomInput
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
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        refreshButton={element.configInput.refreshButton}
        extraButton={element.configInput.extraButton}
        selectAllButton={element.configInput.selectAllButton}
        labelButton={element.configInput.labelButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
      />
    }
    return response
  }

  let content = <>
    <div className="section">
      <h2 intc-id="maintitle">{mainTitle}</h2>
    </div>
    <div className="section">
      {formElements.map((element, index) => buildCustomInput(element, index))}
      {
        formResourceItemsElements.length > 0 ? <ButtonGroup>
          <Button
            intc-id="btn-quota-resource"
            data-wap_ref="btn-quota-resource"
            variant="outline-primary"
            disabled={formResourceItemsElements.length >= serviceResourceLimit}
            onClick={() => onClickActionResourceItem(null, 'Add')}
          >
            <BsNodePlusFill />
            Add service resource
          </Button>
          <span className="feedback">
            {formResourceItemsElements.length} of {serviceResourceLimit} Service resources
          </span>
        </ButtonGroup> : null
      }
      {formResourceItemsElements.map((element, index) => {
        return (
          <React.Fragment key={index}>
            <div className="d-flex flex-column flex-md-row gap-s6 py-s3 w-100">
              {element.items.map((item: any, indexItem: number) => {
                return buildCustomInput(item, indexItem)
              })}
              <div className="d-flex align-self-center">
                <Button
                  onClick={() => onClickActionResourceItem(index, 'Delete')}
                  disabled={index === 0}
                  className="mt-auto"
                  variant="close"
                  type="button"
                  aria-label="Close"
                ></Button>
              </div>
            </div>
            <hr />
          </React.Fragment>
        )
      })}
      <ButtonGroup>
        {navigationBottom.map((item: any, index: string) => (
          <Button
            intc-id={`btn-quota-navigationBottom ${item.buttonLabel}`}
            data-wap_ref={`btn-quota-navigationBottom ${item.buttonLabel}`}
            aria-label={item.buttonLabel}
            key={index}
            variant={item.buttonVariant}
            onClick={item.buttonAction === 'Submit' ? onSubmit : item.buttonFunction}
          >
            {item.buttonLabel}
          </Button>
        ))}
      </ButtonGroup>
    </div>
  </>
  if (emptyView.show) {
    content = <EmptyView title={emptyView.title} subTitle={emptyView.label} action={emptyView.action}/>
  }

  return <>
    <OnSubmitModal showModal={searchModal?.show} message={searchModal?.message}></OnSubmitModal>
    <OnSubmitModal showModal={submitModal.show} message={submitModal.message}></OnSubmitModal>
    {
      !isPageReady ? <div className="section">
        <div className="col-12 row mt-s2">
          <div className="spinner-border text-primary center"></div>
        </div>
      </div>
        : <>
          <div className="section">
            <Button variant="link" className="p-s0" onClick={() => onCancel()}>
              ⟵ Back to {moduleName}
            </Button>
          </div>
          {content}
        </>
    }
  </>
}

export default QuotaManagementServiceCreate
