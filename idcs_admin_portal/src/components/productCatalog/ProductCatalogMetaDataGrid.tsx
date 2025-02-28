// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { Button, ButtonGroup } from 'react-bootstrap'
import CustomInput from '../../utility/customInput/CustomInput'
import SearchBox from '../../utility/searchBox/SearchBox'
import CustomModal from '../../utility/modals/customModal/CustomModal'
import GridPagination from '../../utility/gridPagination/gridPagination'

const ProductCatalogMetaDataGrid = (props: any): JSX.Element => {
  const state = props.state
  const form = state.form
  const onChangeInput = props.onChangeInput
  const filterText = props.filterText
  const setFilter = props.setFilter
  const setAction = props.setAction
  const setCopyTemplateModal = props.setCopyTemplateModal
  const copyTemplateModal = props.copyTemplateModal
  const actionsModal = state.actionsModal
  const onClickActionModal = props.onClickActionModal
  const showMetadataMessage = props.showMetadataMessage
  const onClickOpenMedataModal = props.onClickOpenMedataModal
  const columnsMetadata = props.columnsMetadata
  const metaDataItems = props.metaDataItems
  const emptyGrid = props.emptyGrid
  const metaDataModal = props.metaDataModal
  const setMetaDataModal = props.setMetaDataModal
  const addMetadataState = props.addMetadataState
  const formModal = addMetadataState.form
  const actionsMetaDataModal = addMetadataState.actionsModal
  const onChangeInputModal = props.onChangeInputModal
  const onClickActionMetadataModal = props.onClickActionMetadataModal
  const formElements = []
  const formModalElements = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }
    formElements.push({
      id: key,
      configInput: formItem
    })
  }

  for (const key in formModal) {
    const formItem = {
      ...formModal[key]
    }
    formModalElements.push({
      id: key,
      configInput: formItem
    })
  }

  // *****
  // Functions
  // *****
  const buildCustomInput = (element: any, key: number, formType: string): JSX.Element => {
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
      onChanged={formType === 'formTemplate' ? (event: any) => onChangeInput(event, element.id) : (event: any) => onChangeInputModal(event, element.id)}
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
      idField={element.id}
      columns={element.configInput.columns}
      singleSelection={element.configInput.singleSelection}
      selectedRecords={element.configInput.selectedRecords}
      setSelectedRecords={element.configInput.setSelectedRecords}
      emptyGridMessage={element.configInput.emptyGridMessage}
      gridBreakpoint={element.configInput.gridBreakpoint}
      gridOptions={element.configInput.options}
      hidePaginationControl={element.configInput.hidePaginationControl}
    />
  }

  const getFilterCheck = (filterValue: string, data: any): boolean => {
    filterValue = filterValue.toLowerCase()
    return (
      data?.key?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.value?.toString().toLowerCase().indexOf(filterValue) > -1
    )
  }

  let gridItems = metaDataItems

  if (filterText !== '' && metaDataItems) {
    const input = filterText.toLowerCase()
    gridItems = metaDataItems.filter((item: any) => getFilterCheck(input, item))
  } else {
    gridItems = metaDataItems
  }

  const templateBodyModal = <>
    <div className='section'>
      {formElements.map((element, index) => buildCustomInput(element, index, 'formTemplate'))}
      {
        showMetadataMessage ? <div intc-id="MetadataInvalidMessage" className="invalid-feedback">Please select a metadataset</div> : null
      }
    </div>
  </>

  const templateFooterModal = <>
    <ButtonGroup>
      {
        actionsModal.map((action: any, index: number) => (
          <Button
            intc-id={`btn-pcMetadataGridTemplateModal-${action.buttonLabel}`}
            data-wap_ref={`btn-pcMetadataGridTemplateModal-${action.buttonLabel}`}
            aria-label={action.buttonLabel}
            key={index}
            variant={action.buttonVariant}
            onClick={() => onClickActionModal(action.buttonLabel)}
          >
            {action.buttonLabel}
          </Button>
        ))
      }
    </ButtonGroup>
  </>

  const metaDataBodyModal = <>
    <div className='section'>
      {formModalElements.map((element, index) => buildCustomInput(element, index, 'formMetadata'))}
    </div>
  </>

  const metaDataFooterModal = <>
    <ButtonGroup>
      {
        actionsMetaDataModal.map((action: any, index: number) => (
          <Button
            intc-id={`btn-pcMetadataGridModal-${action.buttonLabel}`}
            data-wap_ref={`btn-pcMetadataGridModal-${action.buttonLabel}`}
            aria-label={action.buttonLabel}
            key={index}
            variant={action.buttonVariant}
            onClick={() => onClickActionMetadataModal(action.buttonLabel)}
          >
            {action.buttonLabel}
          </Button>
        ))
      }
    </ButtonGroup>
  </>

  return <>
    <CustomModal
      show={copyTemplateModal.show}
      title={copyTemplateModal.title}
      body={templateBodyModal}
      footer={templateFooterModal}
      onHide={() => setCopyTemplateModal({ ...copyTemplateModal, show: false })} />
    <CustomModal
      show={metaDataModal.show}
      title={metaDataModal.title}
      body={metaDataBodyModal}
      footer={metaDataFooterModal}
      onHide={() => setMetaDataModal({ ...metaDataModal, show: false })} />
    <div className='section'>
      <div className="filter flex-wrap p-0">
        <Button intc-id="btnCopyMetadatafromTemplate" data-wap_ref="btnCopyMetadatafromTemplate" variant="primary" onClick={() => onClickOpenMedataModal()}>
          Copy from templates
        </Button>
        <SearchBox
          intc-id="filterMetadata"
          value={filterText}
          onChange={setFilter}
          placeholder="Filter metadata..."
          aria-label="Type to filter metadata..."
        />
        <Button intc-id="btnAddMetadata" data-wap_ref="btnAddMetadata" variant="primary" onClick={() => setAction({ id: 'add', item: {} })}>
          Add metadata
        </Button>
      </div>
    </div>
    <div className="section">
      <GridPagination data={gridItems} columns={columnsMetadata} loading={false} emptyGrid={emptyGrid} />
    </div>
  </>
}

export default ProductCatalogMetaDataGrid
