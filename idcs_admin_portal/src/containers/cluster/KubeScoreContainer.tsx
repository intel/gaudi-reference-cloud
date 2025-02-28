/* eslint-disable */
// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React, { useState, useEffect } from 'react'
import KubeScoreDetail from '../../components/cluster/kubeScore/KubeScoreDetail'
import { useNavigate } from 'react-router'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useClusterStore from '../../store/clusterStore/ClusterStore'
import { UpdateFormHelper, setSelectOptions } from '../../utility/updateFormHelper/UpdateFormHelper'
const KubeScoreContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const throwError = useErrorBoundary()
  const insightsVersions = useClusterStore((state) => state.insightsVersions)
  const setInsightsVersions = useClusterStore((state) => state.setInsightsVersions)
  const insightsComponents = useClusterStore((state) => state.insightsComponents)
  const setInsightsComponents = useClusterStore((state) => state.setInsightsComponents)
  const versionLoading = useClusterStore((state) => state.versionLoading)
  const insightsSboms = useClusterStore((state) => state.insightsSboms)
  const setInsightsSboms = useClusterStore((state) => state.setInsightsSboms)
  const sbomLoading = useClusterStore((state) => state.sbomLoading)
  const insightsVulnerabilities = useClusterStore((state) => state.insightsVulnerabilities)
  const setInsightsVulnerabilities = useClusterStore((state) => state.setInsightsVulnerabilities)
  const vulnerabilitiesLoading = useClusterStore((state) => state.vulnerabilitiesLoading)
  const summary = useClusterStore((state) => state.summary)
  const setSummary = useClusterStore((state) => state.setSummary)
  const summaryLoading = useClusterStore((state) => state.summaryLoading)

  // *****
  // local state
  // *****
  const stateInitial = {
    form: {
      component: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Component',
        hiddenLabel: true,
        placeholder: 'Please select component',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: <>Please select a component</>,
        groupBy: 'sbom'
      },
      version: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Release Version',
        hiddenLabel: true,
        placeholder: 'Please select version',
        value: '',
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        options: [],
        validationMessage: '',
        helperMessage: <>Please select a version</>,
        groupBy: 'sbom'
      }
    }
  }

  const vulnerabilitiesColumns = [
    {
      columnName: 'Id',
      targetColumn: 'id',
      className: 'text-nowrap'
    },
    {
      columnName: 'Description',
      targetColumn: 'description',
      width: '20rem'
    },
    {
      columnName: 'Severity',
      targetColumn: 'severity'
    },
    {
      columnName: 'Component',
      width: '15rem',
      targetColumn: 'componentName'
    },
    {
      columnName: 'Component version',
      targetColumn: 'componentVersion',
      width: '8rem'
    },
    {
      columnName: 'Package',
      targetColumn: 'affectedPackage',
      className: 'text-break cellWidth20rem'
    },
    {
      columnName: 'Affected versions',
      targetColumn: 'affectedVersions',
      width: '9rem'
    },
    {
      columnName: 'Fixed versions',
      targetColumn: 'fixedVersion',
      width: '8.5rem'
    },
    {
      columnName: 'Published',
      targetColumn: 'publishedAt'
    }
  ]

  const summaryColumns = [
    {
      columnName: 'Component',
      targetColumn: 'componentName'
    },
    {
      columnName: 'Version',
      targetColumn: 'componentVersion'
    },
    {
      columnName: 'Release Id',
      targetColumn: 'releaseId'
    },
    {
      columnName: 'Scan date',
      targetColumn: 'scanTimestamp'
    },
    {
      columnName: 'Critical',
      targetColumn: 'critical'
    },
    {
      columnName: 'High',
      targetColumn: 'high'
    },
    {
      columnName: 'Medium',
      targetColumn: 'medium'
    },
    {
      columnName: 'Low',
      targetColumn: 'low'
    }
  ]

  const summaryEmptyGridInitial = {
    title: 'No Summary items found',
    subTitle: '',
    action: {
      type: 'function',
      href: () => {
        setSummaryFilter('', true)
      },
      label: 'Clear filters'
    }
  }

  const vulnerabilitiesEmptyGridInitial = {
    title: 'No Vulnerabilities items found',
    subTitle: '',
    action: {
      type: 'function',
      href: () => {
        setVulnerabiltyFilter('', true)
      },
      label: 'Clear filters'
    }
  }

  const navigate = useNavigate()

  const [summaryEmptyGrid, setSummaryEmptyGrid] = useState(summaryEmptyGridInitial)
  const [summaryFilterText, setSummaryFilterText] = useState('')
  const [summaryItems, setSummaryItems] = useState<any[]>([])

  const [vulnerabilitiesEmptyGrid, setVulnerabilitiesEmptyGrid] = useState(vulnerabilitiesEmptyGridInitial)
  const [vulnerabilitiesFilterText, setVulnerabilitiesFilterText] = useState('')
  const [vulnerabilitiesItems, setVulnerabilitiesItems] = useState<any[]>([])
  const [state, setState] = useState(stateInitial)

  const [componentSelected, setComponentSelected] = useState('')
  const [versionSelected, setVersionSelected] = useState('')
  const [isKubernetesSelected, setIsKubernetesSelected] = useState(false)
  const [isPageReady, setIsPageReady] = useState(false)
  // *****
  // Hooks
  // *****
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      await setInsightsComponents()
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    loadComponents()
  }, [insightsComponents])

  useEffect(() => {
    loadVersions()
  }, [componentSelected])

  useEffect(() => {
    if (versionSelected && componentSelected) {
      getInfoReports()
    }
  }, [componentSelected, versionSelected])

  useEffect(() => {
    createSbomFile()
  }, [insightsSboms])

  useEffect(() => {
    loadVulnerabiltyItems()
  }, [insightsVulnerabilities])

  useEffect(() => {
    loadSummaryItems()
  }, [summary])

  // *****
  // Functions
  // *****
  const loadComponents = (): void => {
    const stateUpdate = { ...state }
    const options: any[] = []
    for (const index in insightsComponents) {
      const component = { ...insightsComponents[index] }
      options.push({
        name: component.name,
        value: component.name
      })
    }
    const form = setSelectOptions('component', options, stateUpdate.form)
    stateUpdate.form = form
    setState(stateUpdate)
    setIsPageReady(true)
  }

  const loadVersions = (): void => {
    const stateUpdate = { ...state }
    const options: any[] = []
    const component = insightsComponents?.find((item) => item.name === componentSelected)
    if (component) {
      const versions = [...component.versions]
      versions.forEach((version: string) => {
        options.push({
          name: version,
          value: version
        })
      })
      const form = setSelectOptions('version', options, stateUpdate.form)
      stateUpdate.form = form
      setState(stateUpdate)
    }
  }

  const getInfoReports = (): void => {
    const fetch = async (): Promise<void> => {
      const promiseArray = []
      promiseArray.push(setInsightsVulnerabilities(componentSelected, versionSelected))
      promiseArray.push(setSummary(componentSelected, versionSelected))
      promiseArray.push(setInsightsVersions(componentSelected, versionSelected))
      await Promise.all(promiseArray)
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }

  const onChangeInput = (event: any, formInputName: string): void => {
    const value = event.target.value
    const updatedState = {
      ...state
    }

    let updatedForm = updatedState.form

    updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

    if (formInputName === 'version') {
      setIsKubernetesSelected(updatedForm.component.value  === 'KUBERNETES')
      setVersionSelected(value)
    }

    if (formInputName === 'component') {      
      setComponentSelected(value)
      setVersionSelected('')
      setIsKubernetesSelected(false)
      updatedForm = setSelectOptions('version', [], updatedForm)
      updatedForm = UpdateFormHelper('', 'version', updatedForm)
    }

    updatedState.form = updatedForm
    setSummaryItems([])
    setVulnerabilitiesItems([])
    setState(updatedState)
  }

  const loadVulnerabiltyItems = (): void => {
    const gridInfo: any[] = []
    for (const index in insightsVulnerabilities) {
      const item = { ...insightsVulnerabilities[index] }
      gridInfo.push({
        id: item.id,
        description: item.description,
        severity: item.severity,
        componentName: item.componentName,
        componentVersion: item.componentVersion,
        affectedPackage: item.affectedPackage,
        affectedVersions: item.affectedVersions,
        fixedVersion: item.fixedVersion,
        publishedAt: item.publishedAt
      })
    }
    setVulnerabilitiesItems(gridInfo)
  }

  const loadSummaryItems = (): void => {
    const gridInfo: any[] = []
    for (const index in summary) {
      const item = { ...summary[index] }
      gridInfo.push({
        componentName: item.componentName,
        componentVersion: item.componentVersion,
        releaseId: item.releaseId,
        scanTimestamp: item.scanTimestamp,
        critical: item.critical,
        high: item.high,
        medium: item.medium,
        low: item.low
      })
    }
    setSummaryItems(gridInfo)
  }

  const setSummaryFilter = (event: any, clear: boolean): void => {
    if (clear) {
      setSummaryEmptyGrid({ ...summaryEmptyGrid, subTitle: '' })
      setSummaryFilterText('')
    } else {
      setSummaryEmptyGrid({ ...summaryEmptyGrid, subTitle: 'The applied filter criteria did not match any items' })
      setSummaryFilterText(event.target.value)
    }
  }

  const setVulnerabiltyFilter = (event: any, clear: boolean): void => {
    if (clear) {
      setVulnerabilitiesEmptyGrid({ ...vulnerabilitiesEmptyGrid, subTitle: '' })
      setVulnerabilitiesFilterText('')
    } else {
      setVulnerabilitiesEmptyGrid({
        ...vulnerabilitiesEmptyGrid,
        subTitle: 'The applied filter criteria did not match any items'
      })
      setVulnerabilitiesFilterText(event.target.value)
    }
  }

  const downloadSbom = (): void => {
    if (versionSelected && componentSelected) {
      const fetch = async (): Promise<void> => {
        await setInsightsSboms(componentSelected, versionSelected)
      }
      fetch().catch((error) => {
        throwError(error)
      })
    }
  }

  const createSbomFile = (): void => {
    if (insightsSboms.length > 0) {
      const element = document.createElement('a')
      const file = new Blob([insightsSboms[0].sbom], { type: 'text/plain' })
      element.href = URL.createObjectURL(file)
      const fileName = `file-sbom-${state.form.version.value}.yaml`
      element.download = fileName
      document.body.appendChild(element)
      element.click()
    }
  }

  const onCancel = (): void => {
    navigate('/')
  }

  return (
    <KubeScoreDetail
      onCancel={onCancel}
      state={state}
      loading={versionLoading}
      onChangeInput={onChangeInput}
      downloadSbom={downloadSbom}
      sbomLoading={sbomLoading}
      summaryEmptyGrid={summaryEmptyGrid}
      summaryFilterText={summaryFilterText}
      setSummaryFilter={setSummaryFilter}
      summaryItems={summaryItems}
      summaryLoading={summaryLoading}
      summaryColumns={summaryColumns}
      vulnerabilitiesFilterText={vulnerabilitiesFilterText}
      setVulnerabiltyFilter={setVulnerabiltyFilter}
      vulnerabilitiesItems={vulnerabilitiesItems}
      vulnerabilitiesColumns={vulnerabilitiesColumns}
      vulnerabilitiesEmptyGrid={vulnerabilitiesEmptyGrid}
      vulnerabilitiesLoading={vulnerabilitiesLoading}
      isKubernetesSelected={isKubernetesSelected}
      insightsVersions={insightsVersions}
      isPageReady={isPageReady}
    />
  )
}

export default KubeScoreContainer
