// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { type ColumnDefinition } from '../GridPagination.types'
import { BsCaretUpFill, BsCaretDownFill } from 'react-icons/bs'
import { Button } from 'react-bootstrap'

interface ColumnHeaderProps {
  index: number
  enableSort: boolean
  sortByColumn: (column: ColumnDefinition) => void
  column: ColumnDefinition
  sortedColumns: any
  setSortedColumn: (sortedColumns: any) => void
}

const ColumnHeader: React.FC<ColumnHeaderProps> = ({
  index,
  enableSort,
  sortByColumn,
  column,
  sortedColumns,
  setSortedColumn
}): JSX.Element => {
  const className = 'border-top-0 grid-header' + (column.className ? ' ' + column.className : '')
  const isActionsColumn = column.columnName === 'Actions'
  const sort = sortedColumns[column.columnName]

  const changeSortedColumn = (): void => {
    if (sort === undefined || sort === 'none') {
      sortedColumns[column.columnName] = 'asc'
    } else if (sort === 'asc') {
      sortedColumns[column.columnName] = 'desc'
    } else if (sort === 'desc') {
      sortedColumns[column.columnName] = 'none'
    }
    setSortedColumn({ ...sortedColumns })
  }

  return (
    <th
      key={index}
      className={`${className} ${isActionsColumn ? 'text-center' : ''}`}
      style={{ width: column.width }}
      intc-id={`${column.columnName.replaceAll(' ', '')}Column${index}`}
    >
      <span className="fw-semibold">
        {column.columnName}
        {(Object.prototype.hasOwnProperty.call(column, 'isSort') && !column.isSort) ||
        isActionsColumn ||
        !enableSort ? null : (
          <Button
            intc-id={column.columnName + 'sortTableButton'}
            className="sort-button ms-s4"
            variant="icon-simple"
            aria-label={`Sort ${column.columnName} column`}
            size="sm"
            onClick={() => {
              sortByColumn(column)
              changeSortedColumn()
            }}
          >
            {sort === undefined || sort === 'none' ? (
              <div className="d-flex flex-column gap-s0">
                <BsCaretUpFill />
                <BsCaretDownFill />
              </div>
            ) : null}
            {sort === 'asc' && (
              <div className="d-flex flex-column">
                <BsCaretUpFill />
              </div>
            )}
            {sort === 'desc' && (
              <div className="d-flex flex-column">
                <BsCaretDownFill />
              </div>
            )}
          </Button>
        )}
      </span>
    </th>
  )
}

export default ColumnHeader
