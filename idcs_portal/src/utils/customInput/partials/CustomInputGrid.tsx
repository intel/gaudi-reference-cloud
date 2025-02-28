// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import './_CustomInputs.scss'
import { getCustomInputId, type CustomInputProps } from '../CustomInput.types'
import CustomInputLabel from './CustomInputLabel'
import GridPagination from '../../gridPagination/gridPagination'

const CustomInputGrid: React.FC<CustomInputProps> = ({
  idField,
  label = '',
  labelButton,
  maxWidth,
  columns = [],
  gridOptions = [],
  gridBreakpoint,
  hiddenLabel,
  singleSelection,
  selectedRecords,
  setSelectedRecords,
  emptyGridMessage
}): JSX.Element => {
  let regularColumns: any[] = []
  const regularOptions: any[] = []

  let breakpointColumns: any[] = []
  const breakpointOptions: any[] = []

  regularColumns = columns.filter((x) => !x.showOnBreakpoint)
  breakpointColumns = columns.filter((x) => x.showOnBreakpoint ?? x.targetColumn === idField)
  gridOptions.forEach((gridOptions) => {
    const newBreakpointOption: any = {}
    breakpointColumns.forEach((column) => {
      newBreakpointOption[column.targetColumn] = gridOptions[column.targetColumn]
    })
    breakpointOptions.push(newBreakpointOption)

    const newRegularOption: any = {}
    regularColumns.forEach((column) => {
      newRegularOption[column.targetColumn] = gridOptions[column.targetColumn]
    })
    regularOptions.push(newRegularOption)
  })

  const labelId = getCustomInputId(label)

  return (
    <div className="d-flex-customInput" style={maxWidth ? { maxWidth } : undefined}>
      <CustomInputLabel label={label} labelButton={labelButton} hiddenLabel={hiddenLabel} />
      <div className={gridBreakpoint ? `d-block d-${gridBreakpoint}-none` : 'd-none'}>
        <GridPagination
          intc-id={`${labelId}-breakpoint-grid`}
          idField={idField}
          hidePaginationControl
          hideSortControls
          isSelectable
          singleSelection={singleSelection}
          data={breakpointOptions}
          columns={breakpointColumns}
          selectedRecords={selectedRecords}
          setSelectedRecords={setSelectedRecords}
          emptyGrid={emptyGridMessage}
          tableClassName="tableFixed"
        ></GridPagination>
      </div>
      <div className={gridBreakpoint ? `d-none d-${gridBreakpoint}-block` : 'd-block'}>
        <GridPagination
          intc-id={`${labelId}-grid`}
          idField={idField}
          hidePaginationControl
          hideSortControls
          isSelectable
          singleSelection={singleSelection}
          data={regularOptions}
          columns={regularColumns}
          selectedRecords={selectedRecords}
          setSelectedRecords={setSelectedRecords}
          emptyGrid={emptyGridMessage}
          tableClassName="tableFixed"
        ></GridPagination>
      </div>
    </div>
  )
}

export default CustomInputGrid
