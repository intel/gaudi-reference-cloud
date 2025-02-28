// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import GridPagination from '../../../utility/gridPagination/gridPagination'
import type Product from '../../../store/models/Product/Product'
interface InstanceTableProps {
  products: Product[]
  selectedProducts?: string[]
  setSelectedProducts?: (selectedProducts: any[]) => void
}

const InstanceTable: React.FC<InstanceTableProps> = ({
  products,
  selectedProducts,
  setSelectedProducts
}): JSX.Element => {
  // local state
  const columns = [
    {
      columnName: '',
      targetColumn: 'id',
      hideField: true
    },
    {
      columnName: 'Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Instance Type',
      targetColumn: 'instanceType'
    },
    {
      columnName: 'Description',
      targetColumn: 'displayName'
    },
    {
      columnName: 'Nodes',
      targetColumn: 'nodesCount'
    },
    {
      columnName: 'Cores',
      targetColumn: 'cpuCores'
    },
    {
      columnName: 'Memory',
      targetColumn: 'memorySize'
    },
    {
      columnName: 'Disk',
      targetColumn: 'diskSize'
    }
  ]

  const emptyGrid = {
    title: 'Product table empty',
    subTitle: 'No products found'
  }

  // States
  const [gridData, setGridData] = useState<any[]>([])

  // Hooks
  useEffect(() => {
    setGridInfo()
  }, [products])

  // Functions
  const compare = (a: Product, b: Product): number => {
    if (a.recommendedUseCase < b.recommendedUseCase) {
      return -1
    }
    if (a.recommendedUseCase > b.recommendedUseCase) {
      return 1
    }
    return 0
  }

  const setGridInfo = (): void => {
    const sortedProducts: Product[] = [...products].sort(compare)
    const gridInfo = []

    for (const productItem of sortedProducts) {
      gridInfo.push({
        id: productItem.id,
        name: productItem.name,
        instanceType: productItem.instanceType,
        displayName: productItem.displayName,
        nodesCount: productItem.nodesCount,
        cpuCores: productItem.cpuCores,
        memorySize: productItem.memorySize,
        diskSize: productItem.diskSize
      })
    }
    setGridData(gridInfo)
  }

  return (
    <>
      <GridPagination columns={columns} data={gridData} loading={false} emptyGrid={emptyGrid} isSelectable singleSelection setSelectedRecords={setSelectedProducts} selectedRecords={selectedProducts} idField='id' fixedFirstColumn/>
    </>
  )
}

export default InstanceTable
