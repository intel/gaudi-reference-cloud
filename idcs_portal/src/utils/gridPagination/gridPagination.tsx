// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import Grid from './Grid'
import CustomInput from '../customInput/CustomInput'
import EmptyView, { type EmptyViewProps } from '../emptyView/EmptyView'
import moment from 'moment'
import { type CustomInputOption } from '../customInput/CustomInput.types'
import { type ColumnDefinition } from './GridPagination.types'

interface GridPaginationProps {
  idField?: string
  emptyGrid?: EmptyViewProps | null
  hidePaginationControl?: boolean
  hideSortControls?: boolean
  fixedFirstColumn?: boolean
  loading?: boolean
  isSelectable?: boolean
  singleSelection?: boolean
  showSelectAll?: boolean
  maxElementsByPage?: 10 | 25 | 50
  data?: any[] | null
  columns: ColumnDefinition[]
  feedback?: React.ReactElement | string | null
  tableClassName?: string
  selectedRecords?: string[]
  setSelectedRecords?: (selectedRecords: any[]) => void
  expandedRecords?: string[]
  setExpandedRecords?: (selectedRecords: any[]) => void
  isRecordExpandable?: (record: any) => boolean
  'intc-id'?: string
}

const GridPagination: React.FC<GridPaginationProps> = ({
  idField,
  emptyGrid = null,
  hidePaginationControl = false,
  hideSortControls = false,
  fixedFirstColumn = false,
  loading = false,
  isSelectable = false,
  singleSelection = false,
  showSelectAll = false,
  maxElementsByPage = 10,
  data = [],
  columns = [],
  feedback = null,
  tableClassName = '',
  selectedRecords,
  setSelectedRecords,
  expandedRecords,
  setExpandedRecords,
  isRecordExpandable,
  ...props
}): JSX.Element => {
  const [pageLimit, setPageLimit] = useState<number>(maxElementsByPage || 10)
  const [localData, setLocalData] = useState(data)
  const [pageData, setPageData] = useState<any[]>([])
  const [currentPage, setCurrentPage] = useState(1)
  const [paginationGroup, setPaginationGroup] = useState<number[]>([])
  const [lastSort, setLastSort] = useState({ orientation: 'default', columnName: 'default' })
  const [sortedColumns, setSortedColumns] = useState({})

  const dataLength = data?.length ?? 0

  function getNumbersOfElementsByPageOfDropdown(): CustomInputOption[] {
    const options = [{ value: '10/Page', name: '10/Page' }]
    if (dataLength > 10) options.push({ value: '25/Page', name: '25/Page' })
    if (dataLength > 25) options.push({ value: '50/Page', name: '50/Page' })
    return options
  }

  function calcNumberOfPages(): number {
    return Math.ceil(dataLength / pageLimit)
  }

  function changePageSize(pageSize: string | undefined): void {
    let localPageLimit = 0
    if (pageSize === '10/Page') localPageLimit = 10
    if (pageSize === '25/Page') localPageLimit = 25
    if (pageSize === '50/Page') localPageLimit = 50
    setPageLimit(localPageLimit)
    setPageData(getPaginatedData(localPageLimit, currentPage))
  }

  function goToNextPage(event: any): void {
    event.preventDefault()
    const newPage = currentPage + 1
    setCurrentPage(newPage)
    setPageData(getPaginatedData(pageLimit, currentPage + 1))
  }

  function goToPreviousPage(event: any): void {
    event.preventDefault()
    const newPage = currentPage - 1
    setCurrentPage(newPage)
  }

  function changePage(event: any): void {
    event.preventDefault()
    const pageNumber = Number(event.target.textContent)
    setCurrentPage(pageNumber)
  }

  function getPaginatedData(dataLimit: number = 10, page: number): any[] {
    const startIndex = page * dataLimit - dataLimit
    const endIndex = startIndex + dataLimit
    return data?.slice(startIndex, endIndex) ?? []
  }
  function getPaginationGroup(): number[] {
    const start = Math.floor((currentPage - 1) / pageLimit) * pageLimit
    return new Array(pageLimit).fill(0).map((_, idx) => start + idx + 1)
  }
  function sortNumber(a: number, b: number, column: ColumnDefinition): number {
    if (column.targetColumn === lastSort.columnName) {
      if (lastSort.orientation === 'up') {
        setLastSort({ orientation: 'down', columnName: column.targetColumn })
        return sortDown(a, b, 'number')
      }
      setLastSort({ orientation: 'up', columnName: column.targetColumn })
      return sortUp(a, b, 'number')
    }
    setLastSort({ orientation: 'up', columnName: column.targetColumn })
    return sortUp(a, b, 'number')
  }

  function sortByColumn(column: ColumnDefinition): void {
    const sortedData = localData ?? []
    sortedData.sort((a, b) => {
      if (typeof a[column.targetColumn] === 'string' && typeof b[column.targetColumn] === 'string') {
        const nameA = a[column.targetColumn].toUpperCase()
        const nameB = b[column.targetColumn].toUpperCase()
        if (column.targetColumn === lastSort.columnName) {
          if (lastSort.orientation === 'up') {
            setLastSort({ orientation: 'down', columnName: column.targetColumn })
            return sortDown(nameA, nameB, 'string')
          } else {
            setLastSort({ orientation: 'up', columnName: column.targetColumn })
            return sortUp(nameA, nameB, 'string')
          }
        }
        setLastSort({ orientation: 'up', columnName: column.targetColumn })
        return sortUp(nameA, nameB, 'string')
      }

      if (
        typeof a[column.targetColumn] === 'object' &&
        (typeof a[column.targetColumn].value === 'string' ||
          (typeof a[column.targetColumn].value === 'object' && a[column.targetColumn].sortValue))
      ) {
        if (a[column.targetColumn].type === 'date') {
          const dateA = moment(a[column.targetColumn].value).utc().unix()
          const dateB = moment(b[column.targetColumn].value).utc().unix()
          return sortNumber(dateA, dateB, column)
        }

        const nameA = a[column.targetColumn].sortValue?.toUpperCase() ?? a[column.targetColumn].value.toUpperCase()
        const nameB = b[column.targetColumn].sortValue?.toUpperCase() ?? b[column.targetColumn].value.toUpperCase()
        if (column.targetColumn === lastSort.columnName) {
          if (lastSort.orientation === 'up') {
            setLastSort({ orientation: 'down', columnName: column.targetColumn })
            return sortDown(nameA, nameB, 'string')
          }
          setLastSort({ orientation: 'up', columnName: column.targetColumn })
          return sortUp(nameA, nameB, 'string')
        }
        setLastSort({ orientation: 'up', columnName: column.targetColumn })
        return sortUp(nameA, nameB, 'string')
      }

      if (typeof a[column.targetColumn] === 'object' && typeof a[column.targetColumn].value === 'number') {
        return sortNumber(a[column.targetColumn].value, b[column.targetColumn].value, column)
      }

      return sortNumber(a[column.targetColumn], b[column.targetColumn], column)
    })
    setLocalData(sortedData)
    setPageData(getPaginatedData(pageLimit, currentPage))
  }
  function sortUp(a: any, b: any, type: string): number {
    if (type === 'string') {
      if (a < b) return -1
      if (a > b) return 1
      return 0
    }
    return a - b
  }
  function sortDown(a: any, b: any, type: string): number {
    if (type === 'string') {
      if (a > b) return -1
      if (a < b) return 1
      return 0
    }
    return b - a
  }

  const onExpandSelected = (id: string): void => {
    if (!expandedRecords || !setExpandedRecords || !idField) {
      return
    }
    const isExpanded = expandedRecords.some((x) => x === id)
    if (isExpanded) {
      expandedRecords = expandedRecords.filter((x) => x !== id)
    } else {
      expandedRecords.push(id)
    }
    setExpandedRecords([...expandedRecords])
  }

  const onChangeSelected = (id: string): void => {
    if (!selectedRecords || !setSelectedRecords || !idField) {
      return
    }
    if (singleSelection) {
      setSelectedRecords([id])
      return
    }
    const isSelected = selectedRecords.some((x) => x === id)
    if (isSelected) {
      selectedRecords = selectedRecords.filter((x) => x !== id)
    } else {
      selectedRecords.push(id)
    }
    setSelectedRecords([...selectedRecords])
  }

  const onChangeAllSelected = (): void => {
    if (!selectedRecords || !setSelectedRecords || !idField) {
      return
    }
    const shouldDeselectAll = selectedRecords.length === data?.length
    if (shouldDeselectAll) {
      setSelectedRecords([])
    } else {
      setSelectedRecords(data?.map((x) => x[idField]) ?? [])
    }
  }

  useEffect(() => {
    if (currentPage > 1) {
      const gridMaxCapacity = currentPage * pageLimit
      const isCurrentPageEmpty = gridMaxCapacity - dataLength >= pageLimit
      if (isCurrentPageEmpty) {
        const newPage = currentPage - 1
        setCurrentPage(newPage)
        return
      }
    }
    if (hidePaginationControl && pageLimit !== dataLength) {
      setPageLimit(dataLength)
      return
    }
    setPageData(getPaginatedData(pageLimit, currentPage))
    setLocalData(data)
    setPaginationGroup(getPaginationGroup())
  }, [dataLength, data, currentPage, pageLimit])

  let content = null

  if (!loading && dataLength === 0) {
    if (emptyGrid) {
      return <EmptyView title={emptyGrid.title} subTitle={emptyGrid.subTitle} action={emptyGrid.action} />
    }
  } else {
    const canShowCheckboxes = selectedRecords !== undefined && setSelectedRecords !== undefined && idField !== undefined
    content = (
      <>
        <Grid
          intc-id={props['intc-id']}
          idField={idField ?? ''}
          records={pageData}
          fixedFirstColumn={fixedFirstColumn}
          hideSortControls={hideSortControls}
          isSelectable={isSelectable && canShowCheckboxes}
          singleSelection={singleSelection}
          showSelectAll={showSelectAll && canShowCheckboxes}
          columns={columns}
          loading={loading}
          sortByColumn={sortByColumn}
          selectedRecords={selectedRecords}
          expandedRecords={expandedRecords}
          onChangeSelected={onChangeSelected}
          onChangeAllSelected={onChangeAllSelected}
          onExpandSelected={onExpandSelected}
          isRecordExpandable={isRecordExpandable}
          tableClassName={tableClassName}
          areAllChecked={selectedRecords !== undefined && selectedRecords.length === data?.length}
          sortedColumns={sortedColumns}
          setSortedColumn={setSortedColumns}
        />
        {!hidePaginationControl && !loading ? (
          <div className="d-flex flex-row align-self-stretch justify-content-between flex-wrap">
            <div className="d-flex flex-row flex-wrap gap-s4">
              {canShowCheckboxes && (
                <>
                  <span className="feedback">{`${selectedRecords?.length ?? 0} of ${data?.length ?? 0} rows selected`}</span>
                  <span>{'|'}</span>
                </>
              )}
              <span className="feedback">{feedback}</span>
            </div>
            <div className="d-flex flex-column flex-sm-row gap-s3 align-items-center">
              <div className="pagination-dropdown">
                <CustomInput
                  label="Rows per page"
                  hiddenLabel
                  type="dropdown"
                  value={`${pageLimit}/Page`}
                  onChanged={(event) => {
                    changePageSize(event.target.value?.toString())
                  }}
                  isValid
                  isTouched
                  options={getNumbersOfElementsByPageOfDropdown()}
                />
              </div>
              <nav aria-label="Page navigation">
                <ul className="pagination justify-content-end flex-wrap mb-0">
                  <li className={currentPage === 1 ? 'page-item disabled' : 'page-item'}>
                    <a
                      className="page-link"
                      href="#"
                      aria-disabled="true"
                      aria-label="Go to previous page"
                      onClick={goToPreviousPage}
                      intc-id="goToPreviousPageButton"
                    >
                      Previous
                    </a>
                  </li>
                  {paginationGroup.map((item, index) =>
                    item <= calcNumberOfPages() ? (
                      <li
                        key={index}
                        className={item === currentPage ? 'page-item currentPaginationIndex' : 'page-item'}
                      >
                        <a
                          className="page-link"
                          aria-label={item === currentPage ? `Page ${item} is already selected` : `Go to page ${item}`}
                          intc-id={`goPage${index + 1}Button`}
                          onClick={changePage}
                          href="#"
                        >
                          {item}
                        </a>
                      </li>
                    ) : null
                  )}
                  <li
                    className={
                      currentPage < calcNumberOfPages() && calcNumberOfPages() > 0 ? 'page-item' : 'page-item disabled'
                    }
                  >
                    <a
                      className="page-link"
                      aria-label="Go to next page"
                      intc-id="goNextPageButton"
                      onClick={goToNextPage}
                      href="#"
                    >
                      Next
                    </a>
                  </li>
                </ul>
              </nav>
            </div>
          </div>
        ) : null}
      </>
    )
  }

  return (
    <div intc-id={props['intc-id']} className="section p-0">
      {content}
    </div>
  )
}

export default GridPagination
