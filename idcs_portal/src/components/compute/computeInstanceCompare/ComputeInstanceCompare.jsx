// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Wrapper from '../../../utils/Wrapper'

const ComputeInstanceCompare = (props) => {
  // props
  const columns = [
    {
      columnName: 'Family',
      targetColumn: 'familyDisplayName'
    },
    {
      columnName: 'Name',
      targetColumn: 'displayName'
    },
    {
      columnName: 'Service',
      targetColumn: 'service'
    },
    {
      columnName: 'ID',
      targetColumn: 'name',
      className: 'text-uppercase'
    },
    {
      columnName: 'Cores',
      targetColumn: 'cpuCores'
    },
    {
      columnName: 'Sockets',
      targetColumn: 'cpuSockets'
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

  const colsLenght = columns.length

  const columnHtml = columns.map((column, index) => {
    return (
      <th
        key={index}
        className="border-top-0 grid-header"
        intc-id={column.columnName.replaceAll(' ', '') + 'Column' + index}
      >
        {column.columnName}
      </th>
    )
  })

  function compare(a, b) {
    if (a.recommendedUseCase < b.recommendedUseCase) {
      return -1
    }
    if (a.recommendedUseCase > b.recommendedUseCase) {
      return 1
    }
    return 0
  }

  const sortedProducts = [...props.products].sort(compare)

  let rowsHtml = ''
  let groupRow = ''
  if (sortedProducts.length > 0) {
    rowsHtml = sortedProducts.map((row, index) => {
      let rowGr = ''

      if (groupRow === '' || groupRow !== row.recommendedUseCase) {
        groupRow = row.recommendedUseCase
        rowGr = (
          <tr key={index + 'group'} role="row">
            <td colSpan={colsLenght + 1} role="cell">
              {groupRow}
            </td>
          </tr>
        )
      }

      const rowTd = columns.map((column, colIndex) => {
        return (
          <td key={index + colIndex} intc-id={'fillTable'} className="mt-1" role="cell">
            <span intc-id={index + column.columnName + 'Span'} className={column?.className}>
              {row[column.targetColumn]}
            </span>
          </td>
        )
      })

      return (
        <Wrapper key={index}>
          {rowGr}
          <tr role="row">
            <td intc-id="fillTable" className="align-middle" role="cell">
              <input
                className="form-check-input ml-2"
                type="radio"
                role="radio"
                name="invalidCheck"
                intc-id={index + 'RadioTable'}
                onClick={(e) => selectInstance(row)}
                aria-label={row}
              />
            </td>
            {rowTd}
          </tr>
        </Wrapper>
      )
    })
  } else {
    rowsHtml = (
      <tr>
        <td className="text-center" colSpan={colsLenght + 1}>
          No instance availible
        </td>
      </tr>
    )
  }

  function selectInstance(instance) {
    props.afterInstanceSelected(instance)
  }

  return (
    <div className="main-grid">
      <div className="section-component">
        <div className="my-3 mx-2">
          <Wrapper>
            <div className="table-responsive mh-400">
              <table className="table" intc-id="instanceCompareTable" role="table">
                <thead>
                  <tr>
                    <th className="border-top-0 grid-header" intc-id="RadioColumn"></th>
                    {columnHtml}
                  </tr>
                </thead>
                <tbody role="rowgroup">{rowsHtml}</tbody>
              </table>
            </div>
          </Wrapper>
        </div>
      </div>
    </div>
  )
}

export default ComputeInstanceCompare
