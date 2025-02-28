// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import moment from 'moment'
import './Grid.scss'
import Spinner from '../spinner/Spinner'
import ColumnHeader from './headerRender/ColumnHeader'
import GridRow from './GridRow'
import { type ColumnDefinition } from './GridPagination.types'

interface GridProps {
  'intc-id'?: string
  idField: string
  hideSortControls: boolean
  fixedFirstColumn: boolean
  isSelectable: boolean
  singleSelection: boolean
  showSelectAll: boolean
  loading: boolean
  records: any[]
  columns: ColumnDefinition[]
  sortByColumn: (column: ColumnDefinition) => void
  tableClassName?: string
  selectedRecords?: string[]
  expandedRecords?: string[]
  isRecordExpandable?: (record: any) => boolean
  onChangeSelected: (id: string) => void
  onChangeAllSelected: () => void
  onExpandSelected: (id: string) => void
  areAllChecked: boolean
  sortedColumns: any
  setSortedColumn: (sortedColumns: any) => void
}

const Grid: React.FC<GridProps> = ({
  idField,
  hideSortControls = false,
  fixedFirstColumn = false,
  isSelectable = false,
  singleSelection = false,
  showSelectAll = false,
  loading = false,
  records = [],
  columns = [],
  sortByColumn,
  tableClassName = '',
  selectedRecords = [],
  expandedRecords = [],
  isRecordExpandable = undefined,
  onChangeSelected = () => {},
  onChangeAllSelected = () => {},
  onExpandSelected = () => {},
  areAllChecked = false,
  sortedColumns,
  setSortedColumn,
  ...props
}): JSX.Element => {
  const formatCellDate = (value: any, format: string | undefined, toLocalTime: boolean | undefined): string => {
    const dateFormatted = toLocalTime ? moment(value).format(format) : moment(value).utc().format(format)
    if (dateFormatted === 'Invalid date') {
      return ''
    }
    return dateFormatted
  }

  const onSelect = (record: any, isExpandableRow: boolean): void => {
    if (isSelectable && !isExpandableRow) {
      onChangeSelected(record[idField])
    } else if (isExpandableRow) {
      onExpandSelected(record[idField])
    }
  }

  const gridColumns = []

  const rows = []

  let index = 0
  const hasExpandableRows = isRecordExpandable ? records?.some((x) => isRecordExpandable(x) && x.expandableData) : false

  for (const item of columns) {
    index++

    const column = { ...item }

    if (!column?.hideField) {
      gridColumns.push(
        <ColumnHeader
          key={index}
          index={index}
          column={column}
          enableSort={!hideSortControls}
          sortByColumn={sortByColumn}
          sortedColumns={sortedColumns}
          setSortedColumn={setSortedColumn}
        />
      )
    }
  }

  index = 0

  for (const record of records) {
    index++
    const isExpanded = expandedRecords.some((x) => x === record[idField]) ?? false
    const isSelected = selectedRecords.some((x) => x === record[idField]) ?? false
    rows.push(
      <GridRow
        intc-id={props['intc-id']}
        index={index}
        key={index}
        id={record[idField] ?? ''}
        record={record}
        columns={columns}
        isSelectable={isSelectable}
        singleSelection={singleSelection}
        isExpanded={isExpanded}
        isSelected={isSelected}
        isExpandedRow={false}
        hasExpandableRows={hasExpandableRows}
        isRecordExpandable={isRecordExpandable}
        onSelect={onSelect}
        formatCellDate={formatCellDate}
      />
    )
    if (record.expandableData !== undefined && isExpanded) {
      let expandIndex = 0
      for (const expandableRecord of record.expandableData) {
        rows.push(
          <GridRow
            index={index}
            key={`${index}-${expandIndex}`}
            id={expandableRecord[idField] ?? ''}
            record={expandableRecord}
            columns={columns}
            isSelectable={false}
            singleSelection={false}
            isExpanded={false}
            isSelected={false}
            hasExpandableRows={false}
            isExpandedRow
            onSelect={onSelect}
            formatCellDate={formatCellDate}
          />
        )
        expandIndex++
      }
    }
  }

  let contentView = <Spinner />

  if (!loading) {
    contentView = (
      <div className="tableContainer">
        <table
          className={`table ${tableClassName} ${fixedFirstColumn ? 'fixedColumn' : ''} ${isSelectable || hasExpandableRows ? 'selectable' : ''} mb-0`}
        >
          <thead>
            <tr>
              {isSelectable || hasExpandableRows ? (
                <th scope="col" className="border-top-0 table-checkbox">
                  <>
                    {!singleSelection && showSelectAll && onChangeAllSelected && (
                      <input
                        type="checkbox"
                        className="form-check-input"
                        name="selectAll"
                        checked={areAllChecked}
                        onChange={() => {
                          onChangeAllSelected()
                        }}
                        intc-id={'checkBoxSelectAllTable'}
                      />
                    )}
                  </>
                </th>
              ) : (
                ''
              )}
              {gridColumns.map((column) => {
                return column
              })}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => {
              return row
            })}
          </tbody>
        </table>
      </div>
    )
  }

  return <>{contentView}</>
}

export default Grid
