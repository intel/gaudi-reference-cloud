// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React, { useEffect, useState } from 'react'
import useStorageManagementStore from '../../store/storageManagementStore/StorageManagementStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import StorageDetailView from '../../components/storageManagement/storageDetailView/StorageDetailView'
import { useNavigate } from 'react-router'

const StorageUsagesContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const fileSystemUsages = useStorageManagementStore((state) => state.fileSystemUsages)
  const bucketUsages = useStorageManagementStore((state) => state.bucketUsages)
  const loading = useStorageManagementStore((state) => state.loading)
  const getStorageUsages = useStorageManagementStore((state) => state.getStorageUsages)

  // *****
  // local state
  // *****
  const columns = [
    {
      columnName: 'Account ID',
      targetColumn: 'cloudAccountId'
    },
    {
      columnName: 'Account Owner',
      targetColumn: 'email'
    },
    {
      columnName: 'File Systems',
      targetColumn: 'numFilesystems'
    },
    {
      columnName: 'Provisioned',
      targetColumn: 'totalProvisioned'
    },
    {
      columnName: 'Region',
      targetColumn: 'region'
    },
    {
      columnName: 'CSI Volumes',
      targetColumn: 'hasIksVolumes'
    },
    {
      columnName: 'Cluster Scheduled',
      targetColumn: 'clusterScheduled'
    }
  ]

  const emptyGrid = {
    title: 'No file storage found',
    subTitle: 'No file storage items'
  }

  const emptyGridByFilter = {
    title: 'No file storage found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => { setFilter('', true) },
      label: 'Clear filters'
    }
  }

  const bucketColumns = [
    {
      columnName: 'Account ID',
      targetColumn: 'cloudAccountId'
    },
    {
      columnName: 'Account Owner',
      targetColumn: 'email'
    },
    {
      columnName: 'Buckets',
      targetColumn: 'buckets'
    },
    {
      columnName: 'Bucket Size',
      targetColumn: 'bucketSize'
    },
    {
      columnName: 'Use Capacity',
      targetColumn: 'usedCapacity'
    },
    {
      columnName: 'Region',
      targetColumn: 'region'
    },
    {
      columnName: 'Cluster Scheduled',
      targetColumn: 'clusterScheduled'
    }
  ]

  const bucketEmptyGrid = {
    title: 'No file storage found',
    subTitle: 'No file storage items'
  }

  const bucketEmptyGridByFilter = {
    title: 'No file storage found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => { setBucketFilter('', true) },
      label: 'Clear filters'
    }
  }

  const moduleName = 'Usages'

  const throwError = useErrorBoundary()
  const navigate = useNavigate()
  const [fileUsages, setFileUsages] = useState<any[]>([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [bucketEmptyGridObject, setBucketEmptyGridObject] = useState(bucketEmptyGrid)
  const [bucketFilterText, setBucketFilterText] = useState('')
  const [bucketItems, setButcketItems] = useState<any[]>([])

  // *****
  // Hooks
  // *****
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await getStorageUsages(!!fileSystemUsages)
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    setGridFileSystemUsages()
  }, [fileSystemUsages])

  useEffect(() => {
    setGridBucketUsages()
  }, [bucketUsages])

  // *****
  // functions
  // *****
  const setGridFileSystemUsages = (): void => {
    const gridInfo: any[] = []
    for (const index in fileSystemUsages) {
      const fileSystem = { ...fileSystemUsages[Number(index)] }
      gridInfo.push({
        cloudAccountId: fileSystem.cloudAccountId,
        email: fileSystem.email,
        numFilesystems: fileSystem.numFilesystems,
        totalProvisioned: fileSystem.totalProvisioned,
        region: fileSystem.region,
        hasIksVolumes: fileSystem.hasIksVolumes,
        clusterScheduled: fileSystem.clusterScheduled
      })
    }
    setFileUsages(gridInfo)
  }

  const setGridBucketUsages = (): void => {
    const gridInfo: any[] = []
    for (const index in bucketUsages) {
      const bucketUsage = { ...bucketUsages[Number(index)] }
      gridInfo.push({
        cloudAccountId: bucketUsage.cloudAccountId,
        email: bucketUsage.email,
        buckets: bucketUsage.buckets,
        bucketSize: bucketUsage.bucketSize,
        usedCapacity: bucketUsage.usedCapacity,
        region: bucketUsage.region,
        clusterScheduled: bucketUsage.clusterScheduled
      })
    }
    setButcketItems(gridInfo)
  }

  const setFilter = (event: any, clear: boolean): void => {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  const setBucketFilter = (event: any, clear: boolean): void => {
    if (clear) {
      setBucketEmptyGridObject(bucketEmptyGrid)
      setBucketFilterText('')
    } else {
      setBucketEmptyGridObject(bucketEmptyGridByFilter)
      setBucketFilterText(event.target.value)
    }
  }

  const onCancel = (): void => {
    navigate('/')
  }

  return (
    <>
      <StorageDetailView
        storageUsages={fileUsages}
        columns={columns}
        emptyGrid={emptyGridObject}
        loading={loading}
        filterText={filterText}
        setFilter={setFilter}
        bucketUsages={bucketItems}
        bucketColumns={bucketColumns}
        bucketEmptyGrid={bucketEmptyGridObject}
        bucketFilterText={bucketFilterText}
        setBucketFilter={setBucketFilter}
        onCancel={onCancel}
        moduleName={moduleName}
      />
    </>
  )
}
export default StorageUsagesContainer
