// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import CloudCreditsView from '../../components/billing/cloudCredits/CloudCreditsView'
import useCloudCreditsStore from '../../store/billingStore/CloudCreditsStore'
import moment from 'moment'
import { setFormValue } from '../../utils/updateFormHelper/UpdateFormHelper'
import { formatCurrency } from '../../utils/numberFormatHelper/NumberFormatHelper'
import useErrorBoundary from '../../hooks/useErrorBoundary'

const CloudCreditsContainers = () => {
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

  const [filteredCredits, setFilteredCredits] = useState([])
  const [myRemainingCredits, setMyRemainingCredits] = useState([])
  const [myUsedCredits, setMyUsedCredits] = useState([])
  const [state, setState] = useState(initialState)
  const [currentDateTime, setCurrentDateTime] = useState(moment().format(currentTimeDDateFormat))

  // Global State
  const loading = useCloudCreditsStore((state) => state.loading)
  const cloudCredits = useCloudCreditsStore((state) => state.cloudCredits)
  const lastUpdated = useCloudCreditsStore((state) => state.lastUpdated)
  const remainingCredits = useCloudCreditsStore((state) => state.remainingCredits)
  const usedCredits = useCloudCreditsStore((state) => state.usedCredits)
  const setCloudCredits = useCloudCreditsStore((state) => state.setCloudCredits)

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await setCloudCredits()
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    setCurrentDateTime(moment(lastUpdated).format(currentTimeDDateFormat))
    setMyRemainingCredits(formatCurrency(remainingCredits))
    setMyUsedCredits(formatCurrency(usedCredits))
    setGridInfo()
  }, [cloudCredits])

  // functions
  function setGridInfo(filter = '') {
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

  function getFilterCheck(filterValue, data) {
    filterValue = filterValue.toLowerCase()

    return (
      data.creditType.toString().toLowerCase().indexOf(filterValue) > -1 ||
      moment(data.createdAt).format(dateFormat).toLowerCase().indexOf(filterValue) > -1 ||
      moment(data.ExpiryDate).format(dateFormat).toLowerCase().indexOf(filterValue) > -1 ||
      data.total.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.totalUsed.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.totalRemaining.toString().toLowerCase().indexOf(filterValue) > -1
    )
  }

  function onChangeInput(event, formInputName) {
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
