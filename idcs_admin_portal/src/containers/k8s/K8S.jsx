import React, { useEffect, useState } from 'react'
import useK8Store from '../../store/k8sStore/K8sStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import K8S from '../../components/k8s/K8S'
import { useNavigate } from 'react-router'

function K8SContainer() {
  const emptyGrid = {
    title: 'No K8S lists found',
    subTitle: 'There are currently no k8s'
  }

  const emptyGridByFilter = {
    title: 'No K8S found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  // K8S States
  const [gridItems, setGridItems] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)

  // IMI Store
  const loading = useK8Store((state) => state.loading)
  const stopLoading = useK8Store((state) => state.stopLoading)
  const k8sData = useK8Store((state) => state.k8sData)
  const getK8SData = useK8Store((state) => state.getK8sData)

  // Error Boundary
  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()

  useEffect(() => {
    fetchK8SData()
  }, [])

  const fetchK8SData = async (isBackground) => {
    try {
      await getK8SData(isBackground)
    } catch (error) {
      throwError(error)
    } finally {
      stopLoading()
    }
  }

  useEffect(() => {
    setGridInfo()
  }, [k8sData])

  const setGridInfo = () => {
    const items = []
    k8sData.forEach((k8s) => {
      items.push({
        id: k8s.provider === 'rke2' ? k8s.provider + '-' + k8s.name.split('+')[0] : k8s.provider + '-' + k8s.name,
        versionName: k8s.name,
        releaseName: k8s.releasename,
        provider: k8s.provider,
        cpIMI: k8s.cpimi,
        workerIMI: k8s.workimi,
        state: k8s.state,
        actions: {
          showField: true,
          type: 'HyperLink',
          value: 'View',
          function: () => { console.log('TODO Add Handler') }
        }
      })
    })

    setGridItems(items)
  }

  const gridColumns = [
    {
      columnName: 'ID',
      targetColumn: 'id'
    },
    {
      columnName: 'K8s Version Name',
      targetColumn: 'versionName'
    },
    {
      columnName: 'Release Name',
      targetColumn: 'releaseName'
    },
    {
      columnName: 'Provider',
      targetColumn: 'provider'
    },
    {
      columnName: 'Control Plane IMI',
      targetColumn: 'cpIMI'
    },
    {
      columnName: 'Worker IMI',
      targetColumn: 'workerIMI'
    },
    {
      columnName: 'Status',
      targetColumn: 'state'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'hyperLink',
        behaviorFunction: null
      }
    }
  ]

  function backToHome() {
    navigate('/')
  }

  function setFilter(event, clear) {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  return (
    <K8S
      gridData={{
        data: gridItems,
        columns: gridColumns,
        emptyGridObject,
        loading
      }}
      backToHome={backToHome}
      filterText={filterText}
      setFilter={setFilter}
    />
  )
}

export default K8SContainer
