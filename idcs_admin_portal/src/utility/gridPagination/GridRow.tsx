// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Dropdown from 'react-bootstrap/Dropdown'
import Button from 'react-bootstrap/Button'
import { BsDashSquare, BsPlusSquare } from 'react-icons/bs'
import { type CellDefitition, type ColumnDefinition } from './GridPagination.types'
import { type CustomInputOption } from '../customInput/CustomInput.types'
import { formatCurrency } from '../numberFormatHelper/NumberFormatHelper'

interface GridRowProps {
  'intc-id'?: string
  index: number
  id: string
  record: any
  columns: ColumnDefinition[]
  isSelectable: boolean
  singleSelection: boolean
  isExpanded: boolean
  isSelected: boolean
  isExpandedRow: boolean
  hasExpandableRows: boolean
  isRecordExpandable?: (record: any) => boolean
  onSelect: (record: any, isExpandableRow: boolean) => void
  formatCellDate: (value: any, format: string | undefined, toLocalTime: boolean | undefined) => string
}

const GridRow: React.FC<GridRowProps> = ({
  index,
  id,
  record,
  columns,
  isSelectable,
  singleSelection,
  isExpanded,
  isSelected,
  isExpandedRow,
  hasExpandableRows,
  isRecordExpandable,
  onSelect,
  formatCellDate,
  ...props
}): JSX.Element => {
  const cells = []
  const isExpandableRow = (isRecordExpandable?.(record) && record?.expandableData) ?? false
  if (isSelectable && !isExpandableRow) {
    cells.push(
      <td
        key={index}
        intc-id={`checkBoxTableRow${index}`}
        className="table-checkbox"
        onClick={() => {
          onSelect(record, isExpandableRow)
        }}
      >
        <>
          <input
            type={singleSelection ? 'radio' : 'checkbox'}
            className="form-check-input"
            name={`checkBoxTable-${props['intc-id']}-${id}`}
            checked={isSelected}
            onChange={() => {
              onSelect(record, isExpandableRow)
            }}
            intc-id={`checkBoxTable-${props['intc-id']}-${id}`}
          />
        </>
      </td>
    )
  } else if (isExpandableRow) {
    cells.push(
      <td
        key={index}
        intc-id={`expandableTableRow${index}`}
        className="table-expand-icon"
        onClick={() => {
          onSelect(record, isExpandableRow)
        }}
      >
        <Button
          intc-id={`btn-expand-row-${id}`}
          variant="icon-simple"
          aria-label={`Toggle expand row with id ${id}`}
          className="p-0"
        >
          {!isExpanded && (
            <BsPlusSquare intc-id={`expandableTable${index}`} aria-label={`Toggle expand row with id ${id}`} />
          )}
          {isExpanded && (
            <BsDashSquare intc-id={`expandableTable${index}`} aria-label={`Toggle expand row with id ${id}`} />
          )}
        </Button>
      </td>
    )
  } else if (isExpandedRow || hasExpandableRows) {
    cells.push(
      <td key={index} intc-id={`expandedTableRow${index}`} className="table-expand-icon">
        {' '}
      </td>
    )
  }

  let cellIndex = 0
  for (const cell in record) {
    const cellObject: CellDefitition = { ...record[cell] }

    // Validate if the cell needs to add a function action
    if (cellObject.showField) {
      if (cellObject.type.toLowerCase() === 'dropdown') {
        cells.push(
          <td colSpan={cellObject.colSpan}>
            <Dropdown className="dropdown-grid dropdown-table" intc-id={'actionDropdownTable'}>
              <Dropdown.Toggle
                size="sm"
                className="m-2"
                variant="outline-primary"
                id="dropdown-basic"
                intc-id={'actionDropdownToggleButtonTable'}
              >
                {cellObject.value ?? 'Actions'}
              </Dropdown.Toggle>
              <Dropdown.Menu>
                {cellObject.options?.map((option: CustomInputOption, index: number) => {
                  return (
                    <Dropdown.Item
                      key={index}
                      eventKey="{item.name}"
                      onClick={() => {
                        if (option.onChanged) {
                          option.onChanged({ target: { value: option.value } })
                        }
                      }}
                      data-key={index}
                      intc-id={`${option.name}DropdownOptionTableOption`}
                    >
                      {option.name}
                    </Dropdown.Item>
                  )
                })}
              </Dropdown.Menu>
            </Dropdown>
          </td>
        )
      } else if (cellObject.type.toLowerCase() === 'hyperlink') {
        cells.push(
          <td colSpan={cellObject.colSpan}>
            <span className="d-flex flex-row align-items-center">
              {cellObject?.noHyperLinkValue ? <span>{cellObject?.noHyperLinkValue}</span> : null}
              {cellObject?.href
                  ? <a
                  intc-id={`${cellObject.value?.toString().replace(' ', '')}HyperLinkTable`}
                  className="btn btn-link hyperlink"
                  href={cellObject.href}
                  rel={cellObject.href.startsWith('/') ? undefined : 'noreferrer'}
                  target={cellObject.href.startsWith('/') ? undefined : '_blank'}
                >
                  {cellObject.value}
                </a> : <Button
                intc-id={`${cellObject.value?.toString().replace(' ', '')}HyperLinkTable`}
                variant="link"
                className="hyperlink"
                onClick={() => cellObject.function()}
              >
                {cellObject.value}
                {cellObject.icon ? cellObject.icon : null}
              </Button>}
            </span>
          </td>
        )
      } else if (cellObject.type.toLowerCase() === 'checkbox') {
        cells.push(
          <td colSpan={cellObject.colSpan}>
            <input
              className="form-check-input"
              type="checkbox"
              id="invalidCheck"
              intc-id={`${cellObject.value?.toString().replace(' ', '')}CheckboxTable`}
              onChange={(e) => cellObject.function(e.target.value)}
            />
          </td>
        )
      } else if (cellObject.type.toLowerCase() === 'buttons') {
        cells.push(
          <td colSpan={cellObject.colSpan} className={columns[cellIndex].columnName === 'Actions' ? 'text-center' : ''}>
            {cellObject?.selectableValues?.map((item, index) => {
              if (item.type === 'icon') {
                return (
                  <div key={index} className="btn border-0 shadow-none cursor-default">
                    {item.name}
                  </div>
                )
              } else {
                return (
                  <Button
                    key={index}
                    variant="link"
                    intc-id={`ButtonTable ${item.label}`}
                    aria-label={item.label}
                    onClick={() => cellObject.function(item, cellObject.value)}
                  >
                    {item.name}
                  </Button>
                )
              }
            })}
          </td>
        )
      } else if (cellObject.type.toLowerCase() === 'button') {
        cells.push(
          <td colSpan={cellObject.colSpan} className='text-center'>
            <Button
              intc-id={`${cellObject.value?.toString().replace(' ', '')}ButtonTable`}
              variant="link"
              className="shadow-none"
              aria-label={cellObject.value?.toString()}
              onClick={() => cellObject.function()}
            >
              {cellObject.value?.toString() ?? ''}
            </Button>
          </td>
        )
      } else if (cellObject.type.toLowerCase() === 'currency') {
        cells.push(
          <td
            intc-id={'fillTable'}
            colSpan={cellObject.colSpan}
            onClickCapture={() => {
              onSelect(record, isExpandableRow)
            }}
            className={columns[cellIndex]?.className ?? ''}
          >
            <span>{formatCurrency(cellObject.value)}</span>
          </td>
        )
      } else if (cellObject.type.toLowerCase() === 'date') {
        cells.push(
          <td
            intc-id={'fillTable'}
            colSpan={cellObject.colSpan}
            onClickCapture={() => {
              onSelect(record, isExpandableRow)
            }}
            className={columns[cellIndex]?.className ?? ''}
          >
            <span>{formatCellDate(cellObject.value, cellObject.format ?? '', cellObject.toLocalTime)}</span>
          </td>
        )
      } else if (cellObject.type.toLowerCase() === 'hyperlink-date') {
        cells.push(
          <td colSpan={cellObject.colSpan}>
            <div className="d-flex align-items-center gap-s4">
              {cellObject?.noHyperLinkValue ? (
                <span>{`${formatCellDate(
                  cellObject.noHyperLinkValue,
                  cellObject.format,
                  cellObject.toLocalTime
                )} `}</span>
              ) : null}
              <Button
                variant="link"
                className={cellObject.type === 'hyperlink-date' ? 'text-decoration-underline' : ''}
                intc-id={`${cellObject.value?.toString().replace(' ', '')}HyperLinkTable`}
                onClick={() => cellObject.function()}
              >
                {cellObject.icon ? cellObject.icon : null}
                {cellObject.value}
              </Button>
            </div>
          </td>
        )
      } else if (cellObject.type.toLowerCase() === 'function') {
        cells.push(
          <td
            intc-id={'fillTable'}
            colSpan={cellObject.colSpan}
            className={columns[cellIndex]?.className ?? ''}
            onClickCapture={() => {
              if (cellObject.canSelectRow) onSelect(record, isExpandableRow)
            }}
          >
            <span>
              <div className="d-flex flex-row align-items-center gap-s4">{cellObject.function(cellObject.value)}</div>
            </span>
          </td>
        )
      } else if (cellObject.type.toLowerCase() === 'text') {
        cells.push(
          <td
            intc-id={'fillTable'}
            colSpan={cellObject.colSpan}
            onClickCapture={() => {
              onSelect(record, isExpandableRow)
            }}
            className={columns[cellIndex]?.className ?? ''}
          >
            <span>{cellObject.value}</span>
          </td>
        )
      }
    } else {
      if (cell !== 'expandableData' && !columns[cellIndex]?.hideField) {
        cells.push(
          <td
            intc-id={'fillTable'}
            onClickCapture={() => {
              onSelect(record, isExpandableRow)
            }}
            className={columns[cellIndex]?.className ?? ''}
          >
            <span>{record[cell]}</span>
          </td>
        )
      }
    }
    cellIndex = cellIndex + 1
  }
  return (
    <tr key={index} intc-id={`RowTable${index}`} className={isExpanded || isSelected ? 'active' : ''}>
      {cells.map((cell, index) => {
        return <React.Fragment key={index}>{cell}</React.Fragment>
      })}
    </tr>
  )
}

export default GridRow
