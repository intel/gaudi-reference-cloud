// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import CloudCreditsView from '../../components/userSummary/CloudCreditsView'
import useCloudCreditsStore from '../../store/userManagement/CloudCreditsStore'
import moment from 'moment'
import { setFormValue } from '../../utility/updateFormHelper/UpdateFormHelper'
import { formatCurrency } from '../../utility/numberFormatHelper/NumberFormatHelper'
import useErrorBoundary from '../../hooks/useErrorBoundary'

const CloudCreditsContainers = (props: any): JSX.Element => {
  const userId = props.userId
  // local state
  const columns = [
    {
      columnName: 'Credit Type',
      targetColumn: 'creditType'
    },
    {
      columnName: 'Obtained on',
      targetColumn: 'createdAt'
    },
    {
      columnName: 'Expiration date',
      targetColumn: 'ExpiryDate'
    },
    {
      columnName: 'Total credit amount',
      targetColumn: 'total',
      className: 'text-end'
    },
    {
      columnName: 'Amount used',
      targetColumn: 'totalUsed',
      className: 'text-end'
    },
    {
      columnName: 'Amount remaining',
      targetColumn: 'totalRemaining',
      className: 'text-end'
    }
  ]

  // local state
  const initialState = {
    form: {
      filterCredits: {
        type: 'text', // options = 'text ,'textArea'
        label: 'Filter credits',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        isRequired: false,
        maxLength: 63,
        validationRules: {
          isRequired: false,
          checkMaxLength: true
        }
      },
      isValidForm: false
    }
  }

  const currentTimeDDateFormat = 'dddd, MMMM D, YYYY [at] hh:mm:ss A'
  const dateFormat = 'M/D/YYYY HH:mm'

  const throwError = useErrorBoundary()

  const [filteredCredits, setFilteredCredits] = useState<any[]>([])
  const [myRemainingCredits, setMyRemainingCredits] = useState('')
  const [myUsedCredits, setMyUsedCredits] = useState('')
  const [state, setState] = useState(initialState)
  const [currentDateTime, setCurrentDateTime] = useState(moment().format(currentTimeDDateFormat))

  // Global State
  const loading = useCloudCreditsStore((state) => state.loading)
  const cloudCredits: any = useCloudCreditsStore((state) => state.cloudCredits)
  const lastUpdated = useCloudCreditsStore((state) => state.lastUpdated)
  const remainingCredits = useCloudCreditsStore((state) => state.remainingCredits)
  const usedCredits = useCloudCreditsStore((state) => state.usedCredits)
  const setCloudCredits = useCloudCreditsStore((state) => state.setCloudCredits)

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await setCloudCredits(userId)
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch(() => {})
  }, [])

  useEffect(() => {
    setCurrentDateTime(moment(lastUpdated).format(currentTimeDDateFormat))
    setMyRemainingCredits(formatCurrency(remainingCredits))
    setMyUsedCredits(formatCurrency(usedCredits))
    setGridInfo()
  }, [cloudCredits])

  // functions
  function setGridInfo(filter = ''): void {
    const gridInfo = []

    for (const index in cloudCredits) {
      const credit = { ...cloudCredits[index] }

      const checkFilter = filter === '' || getFilterCheck(filter, credit)

      if (checkFilter) {
        gridInfo.push({
          creditType: credit.creditType,
          createdAt: {
            showField: true,
            type: 'date',
            value: credit.createdAt,
            toLocalTime: true,
            format: dateFormat
          },
          ExpiryDate: {
            showField: true,
            type: 'date',
            toLocalTime: true,
            value: credit.ExpiryDate,
            format: dateFormat
          },
          total: {
            showField: true,
            type: 'currency',
            value: credit.total
          },
          totalUsed: {
            showField: true,
            type: 'currency',
            value: credit.totalUsed
          },
          totalRemaining: {
            showField: true,
            type: 'currency',
            value: credit.totalRemaining
          }
        })
      }
    }

    setFilteredCredits(gridInfo)
  }

  function getFilterCheck(filterValue: string, data: any): boolean {
    filterValue = filterValue.toLowerCase()

    return (
      data.creditType.toString().toLowerCase().indexOf(filterValue) > -1 ||
      moment(data.createdAt).format(dateFormat).toLowerCase().includes(filterValue) ||
      moment(data.ExpiryDate).format(dateFormat).toLowerCase().includes(filterValue) ||
      data.total.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.totalUsed.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.totalRemaining.toString().toLowerCase().indexOf(filterValue) > -1
    )
  }

  function onChangeInput(event: any, formInputName: string): void {
    const updatedState = {
      ...state
    }

    const updatedForm = setFormValue(formInputName, event.target.value, updatedState.form)
    updatedState.form = updatedForm

    setState(updatedState)
    setGridInfo(event.target.value)
  }

  return (
    <CloudCreditsView
      loading={loading}
      filteredCredits={filteredCredits}
      myRemainingCredits={myRemainingCredits}
      myUsedCredits={myUsedCredits}
      columns={columns}
      state={state}
      onChangeInput={onChangeInput}
      currentDateTime={currentDateTime}
      hasCredits={cloudCredits && cloudCredits.length > 0}
    />
  )
}

export default CloudCreditsContainers
