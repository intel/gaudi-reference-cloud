// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useState, useEffect } from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { clearAxiosMock } from '../../mocks/utils'
import { mockStandardUser } from '../../mocks/authentication/authHelper'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import { Button } from 'react-bootstrap'
import userEvent from '@testing-library/user-event'

const TestComponent = ({ initialCountOfItems, emptyGrid }) => {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)

  const resetData = () => {
    setLoading(true)
    const newData = []
    for (let i = 1; i <= initialCountOfItems; i++) {
      const newItem = {
        title: `Row Number ${i}`
      }
      newData.push(newItem)
    }
    setData(newData)
    setLoading(false)
  }

  useEffect(() => {
    resetData()
  }, [])

  const removeRow = () => {
    if (data.length > 0) {
      data.pop()
      setData([...data])
    }
  }

  const addRow = () => {
    const newData = data
    const newItem = {
      title: `Row Number ${data.length + 1}`
    }
    newData.push(newItem)
    setData([...newData])
  }

  const columns = [
    {
      columnName: 'Title',
      targetColumn: 'title'
    }
  ]

  return (
    <AuthWrapper>
      <Button onClick={addRow} aria-label="addRow">
        Add row
      </Button>
      <Button onClick={removeRow} aria-label="removeRow">
        Remove row
      </Button>
      <GridPagination data={data} columns={columns} emptyGrid={emptyGrid} loading={loading} />
    </AuthWrapper>
  )
}

describe('Grid Pagination', () => {
  beforeEach(() => {
    clearAxiosMock()
  })

  beforeEach(() => {
    mockStandardUser()
  })

  it('Show only first 10 rows by default', async () => {
    render(<TestComponent initialCountOfItems={12} />)
    expect(await screen.findByText('Row Number 1')).toBeVisible()
    expect(await screen.findByText('Row Number 2')).toBeVisible()
    expect(await screen.findByText('Row Number 3')).toBeVisible()
    expect(await screen.findByText('Row Number 4')).toBeVisible()
    expect(await screen.findByText('Row Number 5')).toBeVisible()
    expect(await screen.findByText('Row Number 6')).toBeVisible()
    expect(await screen.findByText('Row Number 7')).toBeVisible()
    expect(await screen.findByText('Row Number 8')).toBeVisible()
    expect(await screen.findByText('Row Number 9')).toBeVisible()
    expect(await screen.findByText('Row Number 10')).toBeVisible()
  })

  it('Show rows 6 and 7 when go to second page', async () => {
    render(<TestComponent initialCountOfItems={12} />)
    const gotoPage2Button = await screen.findByLabelText('Go to page 2')
    await userEvent.click(gotoPage2Button)
    expect(await screen.findByText('Row Number 11')).toBeVisible()
    expect(await screen.findByText('Row Number 12')).toBeVisible()
  })

  it('After remove and refresh keep selected page', async () => {
    render(<TestComponent initialCountOfItems={12} />)
    const gotoPage2Button = await screen.findByLabelText('Go to page 2')
    await userEvent.click(gotoPage2Button)
    expect(await screen.findByText('Row Number 11')).toBeVisible()
    expect(await screen.findByText('Row Number 12')).toBeVisible()
    const removeButton = screen.getByLabelText('removeRow')
    await userEvent.click(removeButton)
    await waitFor(() => {
      expect(screen.queryByText('Row Number 12')).toBeNull()
    })
    expect(screen.queryByText('Row Number 11')).toBeVisible()
  })

  it('Go to previous page when selected page is empty, and show all prevous page records', async () => {
    render(<TestComponent initialCountOfItems={11} />)
    const gotoPage2Button = await screen.findByLabelText('Go to page 2')
    await userEvent.click(gotoPage2Button)
    expect(await screen.findByText('Row Number 11')).toBeVisible()
    const removeButton = await screen.findByLabelText('removeRow')
    await userEvent.click(removeButton)
    await waitFor(() => {
      expect(screen.queryByText('Row Number 11')).toBeNull()
    })
    expect(await screen.findByText('Row Number 1')).toBeVisible()
    expect(await screen.findByText('Row Number 2')).toBeVisible()
    expect(await screen.findByText('Row Number 3')).toBeVisible()
    expect(await screen.findByText('Row Number 4')).toBeVisible()
    expect(await screen.findByText('Row Number 5')).toBeVisible()
    expect(await screen.findByText('Row Number 6')).toBeVisible()
    expect(await screen.findByText('Row Number 7')).toBeVisible()
    expect(await screen.findByText('Row Number 8')).toBeVisible()
    expect(await screen.findByText('Row Number 9')).toBeVisible()
    expect(await screen.findByText('Row Number 10')).toBeVisible()
  })

  it('Show Empty object when grid is empty', async () => {
    render(
      <TestComponent
        initialCountOfItems={0}
        emptyGrid={{
          title: 'Empty title',
          subTitle: 'Empty subtitle',
          action: {
            label: 'Empty action button',
            href: () => {}
          }
        }}
      />
    )
    expect(await screen.findByText('Empty title')).toBeVisible()
    expect(await screen.findByText('Empty subtitle')).toBeVisible()
    expect(await screen.findByText('Empty action button')).toBeEnabled()
  })

  it('Add 25 per page option when increasing when rows are 10 and add more', async () => {
    render(<TestComponent initialCountOfItems={10} />)
    expect(await screen.findByText('10/Page')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.queryByText('25/Page')).toBeNull()
    })
    const addButton = screen.getByLabelText('addRow')
    userEvent.click(addButton)
    expect(await screen.findByText('25/Page')).toBeInTheDocument()
    expect(await screen.findByLabelText('Rows per page')).toBeEnabled()
  })

  it('Add 50 per page option when increasing when rows are 25 and add more', async () => {
    render(<TestComponent initialCountOfItems={25} />)
    expect(await screen.findByText('10/Page')).toBeInTheDocument()
    expect(await screen.findByText('25/Page')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.queryByText('50/Page')).toBeNull()
    })
    const addButton = screen.getByLabelText('addRow')
    userEvent.click(addButton)
    expect(await screen.findByText('50/Page')).toBeInTheDocument()
  })
})
