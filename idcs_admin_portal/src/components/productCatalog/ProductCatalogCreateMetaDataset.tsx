// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { Button, ButtonGroup } from 'react-bootstrap'
import SearchBox from '../../utility/searchBox/SearchBox'
import CustomModal from '../../utility/modals/customModal/CustomModal'
import CustomInput from '../../utility/customInput/CustomInput'
import GridPagination from '../../utility/gridPagination/gridPagination'
import LineDivider from '../../utility/lineDivider/LineDivider'

const ProductCatalogCreateMetaDataset = (props: any): JSX.Element => {
  const addFormMetaSetState = props.addFormMetaSetState
  const metadatasetForm = addFormMetaSetState.form
  const addFormMetaDataManualState = props.addFormMetaDataManualState
  const addMetaDataStateModal = addFormMetaDataManualState.actionsModal
  const metaDataForm = addFormMetaDataManualState.form
  const metadataSetactionsModal = addFormMetaSetState.actionsModal
  const title = addFormMetaSetState.title
  const filterMetadatasetText = props.filterMetadatasetText
  const setFilterMetaDataset = props.setFilterMetaDataset
  const copyTemplateModal = props.copyTemplateModal
  const onChangeInput = props.onChangeInput
  const onClickActionModals = props.onClickActionModals
  const showMetadataMessage = props.showMetadataMessage
  const metadataSetColumns = props.metadataSetColumns
  const metaDataSetItems = props.metaDataSetItems
  const metaDataSetEmptyGrid = props.metaDataSetEmptyGrid
  const metadataColumns = props.metadataColumns
  const metaDataItems = props.metaDataItems
  const metaDataEmptyGrid = props.metaDataEmptyGrid
  const addManualModal = props.addManualModal
  const addFormMetaSetManualState = props.addFormMetaSetManualState
  const addManualMetadataModal = props.addManualMetadataModal
  const addForm = addFormMetaSetManualState.form
  const addMetadataSetactionsModal = addFormMetaSetManualState.actionsModal
  const filterMetadataText = props.filterMetadataText
  const setFilterMetaData = props.setFilterMetaData
  const generalFormOptions = props.generalFormOptions
  const setSelectedMetaDataSet = props.setSelectedMetaDataSet
  const selectedMetaDataSet = props.selectedMetaDataSet
  const discardChangesModal = props.discardChangesModal
  const discardOptions = props.discardOptions
  const deleteManualMetadataModal = props.deleteManualMetadataModal
  const deleteMetaOptions = props.deleteMetaOptions
  const metadatasetFormElements = []
  const metadataFormElements = []
  const metadataFormManualElements = []

  for (const key in metadatasetForm) {
    const formItem = {
      ...metadatasetForm[key]
    }
    metadatasetFormElements.push({
      id: key,
      configInput: formItem
    })
  }

  for (const key in addForm) {
    const formItem = {
      ...addForm[key]
    }
    metadataFormElements.push({
      id: key,
      configInput: formItem
    })
  }

  for (const key in metaDataForm) {
    const formItem = {
      ...metaDataForm[key]
    }
    metadataFormManualElements.push({
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
      onChanged={(event: any) => onChangeInput(event, element.id, formType)}
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

  const getFilterMetaDatasetCheck = (filterValue: string, data: any): boolean => {
    filterValue = filterValue.toLowerCase()
    return (
      data?.name?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.context?.toString().toLowerCase().indexOf(filterValue) > -1
    )
  }

  const getFilterMetaDataCheck = (filterValue: string, data: any): boolean => {
    filterValue = filterValue.toLowerCase()
    return (
      data?.key?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.context?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.value?.toString().toLowerCase().indexOf(filterValue) > -1
    )
  }

  const copyTemplateBodyModal = <>
    <div className='section'>
      {metadatasetFormElements.map((element, index) => buildCustomInput(element, index, 'formTemplate'))}
      {
        showMetadataMessage ? <div intc-id="MetadataInvalidMessage" className="invalid-feedback">Metadataset is required</div> : null
      }
    </div>
  </>

  const copyTemplateFooterModal = <>
    <ButtonGroup>
      {
        metadataSetactionsModal.map((action: any, index: number) => (
          <Button
            intc-id={`btn-pcCopyModal-${action.buttonLabel}`}
            data-wap_ref={`btn-pcCopyModal-${action.buttonLabel}`}
            aria-label={action.buttonLabel}
            key={index}
            variant={action.buttonVariant}
            onClick={() => onClickActionModals(action.buttonAction)}
          >
            {action.buttonLabel}
          </Button>
        ))
      }
    </ButtonGroup>
  </>

  const addMetaDatasetManualBodyModal = <>
    <div className='section'>
      {metadataFormElements.map((element, index) => buildCustomInput(element, index, 'addmetadasetForm'))}
    </div>
  </>

  const addMetaDatasetManualFooterModal = <>
    <ButtonGroup>
      {
        addMetadataSetactionsModal.map((action: any, index: number) => (
          <Button
            intc-id={`btn-pcAddMetadataSetModal-${action.buttonLabel}`}
            data-wap_ref={`btn-pcAddMetadataSetModal-${action.buttonLabel}`}
            aria-label={action.buttonLabel}
            key={index}
            variant={action.buttonVariant}
            onClick={() => onClickActionModals(action.buttonAction)}
          >
            {action.buttonLabel}
          </Button>
        ))
      }
    </ButtonGroup>
  </>

  const metaDataBodyModal = <>
    <div className='section'>
      {metadataFormManualElements.map((element, index) => buildCustomInput(element, index, 'addmetadaForm'))}
    </div>
  </>

  const metaDataFooterModal = <>
    <ButtonGroup>
      {
        addMetaDataStateModal.map((action: any, index: number) => (
          <Button
            intc-id={`btn-pcAddMetadataModal-${action.buttonLabel}`}
            data-wap_ref={`btn-pcAddMetadataModal-${action.buttonLabel}`}
            aria-label={action.buttonLabel}
            key={index}
            variant={action.buttonVariant}
            onClick={() => onClickActionModals(action.buttonAction)}
          >
            {action.buttonLabel}
          </Button>
        ))
      }
    </ButtonGroup>
  </>

  const discardBodyModal = <>
    <div className='section'>
      <div className="text-left small">
        {discardChangesModal.question}<br />
      </div>
    </div>
  </>

  const discardFooterModal = <>
    <ButtonGroup>
      {
        discardOptions.map((action: any, index: number) => (
          <Button
            intc-id={`btn-pcDiscardModal-${action.buttonLabel}`}
            data-wap_ref={`btn-pcDiscardModal-${action.buttonLabel}`}
            aria-label={action.buttonLabel}
            key={index}
            variant={action.buttonVariant}
            onClick={() => onClickActionModals(action.buttonAction)}
          >
            {action.buttonLabel}
          </Button>
        ))
      }
    </ButtonGroup>
  </>

  const metaDataDeleteBodyModal = <>
    <div className='section'>
      <div className="text-left small">
        {deleteManualMetadataModal.question}<br />
      </div>
    </div>
  </>

  const metaDataDeleteFooterModal = <>
    <ButtonGroup>
      {
        deleteMetaOptions.map((action: any, index: number) => (
          <Button
            intc-id={`btn-pcDeleteMetadataModal-${action.buttonLabel}`}
            data-wap_ref={`btn-pcDeleteMetadataModal-${action.buttonLabel}`}
            aria-label={action.buttonLabel}
            key={index}
            variant={action.buttonVariant}
            onClick={() => onClickActionModals(action.buttonAction)}
          >
            {action.buttonLabel}
          </Button>
        ))
      }
    </ButtonGroup>
  </>

  let gridMetaDatasetItems = metaDataSetItems

  if (filterMetadatasetText !== '' && metaDataSetItems) {
    const input = filterMetadatasetText.toLowerCase()
    gridMetaDatasetItems = metaDataSetItems.filter((item: any) => getFilterMetaDatasetCheck(input, item))
  } else {
    gridMetaDatasetItems = metaDataSetItems
  }

  let gridMetaDataItems = metaDataItems

  if (filterMetadataText !== '' && metaDataItems) {
    const input = filterMetadataText.toLowerCase()
    gridMetaDataItems = metaDataItems.filter((item: any) => getFilterMetaDataCheck(input, item))
  } else {
    gridMetaDataItems = metaDataItems
  }

  // *****
  // Functions
  // *****
  return <>
    <CustomModal size='xl' show={copyTemplateModal.show} title={copyTemplateModal.title} body={copyTemplateBodyModal} footer={copyTemplateFooterModal} onHide={() => onClickActionModals('close')} />
    <CustomModal size='xl' show={addManualModal.show} title={addManualModal.title} body={addMetaDatasetManualBodyModal} footer={addMetaDatasetManualFooterModal} onHide={() => onClickActionModals('close')} />
    <CustomModal size='xl' show={addManualMetadataModal.show} title={addManualMetadataModal.title} body={metaDataBodyModal} footer={metaDataFooterModal} onHide={() => onClickActionModals('close')} />
    <CustomModal size='md' show={discardChangesModal.show} title={discardChangesModal.title} body={discardBodyModal} footer={discardFooterModal} onHide={() => onClickActionModals('cancelForm')} />
    <CustomModal size='md' show={deleteManualMetadataModal.show} title={deleteManualMetadataModal.title} body={metaDataDeleteBodyModal} footer={metaDataDeleteFooterModal} onHide={() => onClickActionModals('close')} />
    <div className="section">
      <Button variant="link" className="p-s0" onClick={() => onClickActionModals('cancelForm')}>
        ⟵ Back to product
      </Button>
    </div>
    <div className="section">
      <h2 intc-id="maintitle">{title}</h2>
    </div>
    <div className='section'>
      <div className="filter flex-wrap p-0">
        <h3 intc-id="subTitle">Product metadata set</h3>
        <SearchBox
          intc-id="filterMetadataset"
          value={filterMetadatasetText}
          onChange={setFilterMetaDataset}
          placeholder="Filter metadatasets..."
          aria-label="Type to filter metadatasets..."
        />
        <>
          {
            metaDataSetItems.length === 0
              ? <Button intc-id="btnCopyMetadataset" variant="outline-primary" onClick={() => onClickActionModals('openMetaDataSetFromTemplate')}>
                Copy from template
              </Button> : null
          }
        </>
      </div>
    </div>
    <div className="section">
      <GridPagination data={gridMetaDatasetItems} columns={metadataSetColumns} loading={false} emptyGrid={metaDataSetEmptyGrid} isSelectable={true} idField={'id'} selectedRecords={selectedMetaDataSet} setSelectedRecords={setSelectedMetaDataSet} singleSelection={true} />
    </div>
    <LineDivider horizontal />
    <div className='section'>
      <div className="filter flex-wrap p-0">
        <h3 intc-id="subTitle2">Product metadata</h3>
        <SearchBox
          intc-id="filterMetadata"
          value={filterMetadataText}
          onChange={setFilterMetaData}
          placeholder="Filter metadata..."
          aria-label="Type to filter metadata..."
        />
        <Button intc-id="btnAddMetadata" variant="outline-primary" onClick={() => onClickActionModals('openMetaDataManual')}>
          Add metadata
        </Button>
      </div>
    </div>
    <div className="section">
      <GridPagination data={gridMetaDataItems} columns={metadataColumns} loading={false} emptyGrid={metaDataEmptyGrid} />
    </div>
    <div className='section'>
      <ButtonGroup className='ms-auto'>
        {
          generalFormOptions.map((action: any, index: number) => (
            <Button
              intc-id={`btn-changes-${action.buttonLabel}`}
              data-wap_ref={`btn-changes-${action.buttonLabel}`}
              aria-label={action.buttonLabel}
              key={index}
              variant={action.buttonVariant}
              onClick={() => onClickActionModals(action.buttonAction)}
            >
              {action.buttonLabel}
            </Button>
          ))
        }
      </ButtonGroup>
    </div>
  </>
}

export default ProductCatalogCreateMetaDataset
