// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { Button } from 'react-bootstrap'
import SearchBox from '../../../utility/searchBox/SearchBox'
import GridPagination from '../../../utility/gridPagination/gridPagination'
import CustomInput from '../../../utility/customInput/CustomInput'
import LineDivider from '../../../utility/lineDivider/LineDivider'
import { BsDownload } from 'react-icons/bs'
import './KubeScoreDetail.scss'

const KubeScoreDetail = (props: any): JSX.Element => {
  // *****
  // Variables
  // *****
  const onCancel = props.onCancel
  const isPageReady = props.isPageReady
  const downloadSbom = props.downloadSbom
  const summaryItems = props.summaryItems
  const summaryColumns = props.summaryColumns
  const summaryEmptyGrid = props.summaryEmptyGrid
  const summaryFilterText = props.summaryFilterText
  const setSummaryFilter = props.setSummaryFilter
  const summaryLoading = props.summaryLoading
  const vulnerabilitiesItems = props.vulnerabilitiesItems
  const vulnerabilitiesEmptyGrid = props.vulnerabilitiesEmptyGrid
  const vulnerabilitiesColumns = props.vulnerabilitiesColumns
  const vulnerabilitiesLoading = props.vulnerabilitiesLoading
  const setVulnerabiltyFilter = props.setVulnerabiltyFilter
  const vulnerabilitiesFilterText = props.vulnerabilitiesFilterText
  const state = props.state
  const form = state.form
  const onChangeInput = props.onChangeInput
  const isKubernetesSelected = props.isKubernetesSelected
  const insightsVersions = props.insightsVersions
  const formSbomElements = []
  const formVulnerabilityElements = []
  const formSummaryElements = []
  const formComponentElements = []

  function getFilterCheckSummary(filterValue: string, data: any): boolean {
    filterValue = filterValue.toLowerCase()
    return (
      data?.componentName?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.componentVersion?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.releaseId?.toString().toLowerCase().indexOf(filterValue) > -1
    )
  }

  function getFilterCheckVulnerabilities(filterValue: string, data: any): boolean {
    filterValue = filterValue.toLowerCase()
    return (
      data?.id?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.description?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.componentName?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.componentVersion?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.severity?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.affectedPackage?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.fixedVersion?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.publishedAt?.toString().toLowerCase().indexOf(filterValue) > -1
    )
  }

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.groupBy === 'sbom') {
      formSbomElements.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.groupBy === 'vulnerability') {
      formVulnerabilityElements.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.groupBy === 'summary') {
      formSummaryElements.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.groupBy === 'component') {
      formComponentElements.push({
        id: key,
        configInput: formItem
      })
    }
  }

  let gridSummaryItems = []

  if (summaryFilterText !== '' && summaryItems) {
    const input = summaryFilterText.toLowerCase()
    gridSummaryItems = summaryItems.filter((item: any) => getFilterCheckSummary(input, item))
  } else {
    gridSummaryItems = summaryItems
  }

  let gridVulnerabilityItems = []

  if (vulnerabilitiesFilterText !== '' && vulnerabilitiesItems) {
    const input = vulnerabilitiesFilterText.toLowerCase()
    gridVulnerabilityItems = vulnerabilitiesItems.filter((item: any) => getFilterCheckVulnerabilities(input, item))
  } else {
    gridVulnerabilityItems = vulnerabilitiesItems
  }

  // *****
  // Functions
  // *****
  const buildCustomInput = (element: any, key: number): JSX.Element => {
    return (
      <CustomInput
        key={key}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={element.configInput.label}
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
        hiddenLabel={element.configInput.hiddenLabel}
      />
    )
  }

  const content = (
    <>
      <div className="section">
        <h2 className="h4">Kube Score</h2>
      </div>
      <div className="filter w-100 mb-s8">
        <Button
          variant="primary"
          disabled={!isKubernetesSelected}
          onClick={downloadSbom}
        >
          <BsDownload />
          <span>Download Sbom</span>
        </Button>
        <div className="d-flex flex-row flex-wrap flex-sm-nowrap align-items-center gap-s6 text-nowrap">
          {formSbomElements.map((element, index) => buildCustomInput(element, index))}
        </div>
      </div>
      {
        isKubernetesSelected
          ? <div className='section'>
              <div className='d-flex gap-s4'><span className="fw-semibold">End of life date:</span>{!summaryLoading ? insightsVersions?.eolTimeStamp : ''}</div>
              <div className='d-flex gap-s4'><span className="fw-semibold">End of support date:</span>{!summaryLoading ? insightsVersions?.eosTimeStamp : ''}</div>
          </div>
          : null
      }
      <div className="section p-0 gap-s8">
        <div className="section">
          <div className="filter flex-wrap p-0">
            <h2 className="h4">Summary</h2>
            <SearchBox
              intc-id="filterSummaryItems"
              className="ms-sm-auto"
              value={summaryFilterText}
              onChange={setSummaryFilter}
              placeholder="Filter items..."
              aria-label="Type to filter cloud items..."
            />
          </div>
          <GridPagination
            data={gridSummaryItems}
            columns={summaryColumns}
            emptyGrid={summaryEmptyGrid}
            loading={summaryLoading}
          />
        </div>
        <LineDivider horizontal />
        <div className="section">
          <div className="filter flex-wrap p-0">
            <h2 className="h4">Vulnerabilities</h2>
            <SearchBox
              intc-id="filterSummaryItems"
              className="ms-sm-auto"
              value={vulnerabilitiesFilterText}
              onChange={setVulnerabiltyFilter}
              placeholder="Filter items..."
              aria-label="Type to filter cloud items..."
            />
          </div>
          <GridPagination
            data={gridVulnerabilityItems}
            columns={vulnerabilitiesColumns}
            emptyGrid={vulnerabilitiesEmptyGrid}
            loading={vulnerabilitiesLoading}
          />
        </div>
      </div>
    </>
  )

  return (
    <>
      <div className="section">
        <Button variant="link" className="p-s0" onClick={() => onCancel()}>
          ⟵ Back to Home
        </Button>
      </div>
      {isPageReady ? (
        <>{content}</>
      ) : (
        <div className="section">
          <div className="col-12 row mt-s2">
            <div className="spinner-border text-primary center"></div>
          </div>
        </div>
      )}
    </>
  )
}

export default KubeScoreDetail
