// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import moment from 'moment'
import CloudUsageView from '../../components/userSummary/CloudUsageView'
import useUsagesReport from '../../store/userManagement/UsagesStore'
import { setFormValue } from '../../utility/updateFormHelper/UpdateFormHelper'
import useErrorBoundary from '../../hooks/useErrorBoundary'

const CreditUsageContainer = (props: any): JSX.Element => {
  const userId = props.userId
  // local Variables
  const columns = [
    {
      columnName: 'Description',
      targetColumn: 'description',
      isSort: false
    },
    {
      columnName: 'Region',
      targetColumn: 'region',
      isSort: false
    },
    {
      columnName: 'Usage',
      targetColumn: 'usageQuantity',
      className: 'text-end',
      isSort: false
    },
    {
      columnName: 'Estimated amount',
      targetColumn: 'amount',
      className: 'text-end',
      isSort: false
    }
  ]

  const initialState = {
    form: {
      filterUsages: {
        type: 'text', // options = 'text ,'textArea'
        label: '',
        placeholder: 'Type to filter...',
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
      }
    }
  }

  const initialFilters: any = {
    filterUsages: null
  }

  const currentTimeDDateFormat = 'dddd, MMMM D, YYYY [at] hh:mm:ss A'

  // local state
  const [filteredUsages, setFilteredUsages] = useState<any[]>([])
  const [filters, setFilters] = useState(initialFilters)
  const [state, setState] = useState(initialState)
  const [currentDateTime, setCurrentDateTime] = useState('')

  // Global State
  const loading = useUsagesReport((state) => state.loading)
  const usageDetails = useUsagesReport((state) => state.usages)
  const setUsage = useUsagesReport((state) => state.setUsage)
  const totalAmount = useUsagesReport((state) => state.totalAmount)
  const lastUpdated = useUsagesReport((state) => state.lastUpdated)
  const period = useUsagesReport((state) => state.period)

  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await setUsage(userId)
      } catch (error) {
        throwError(error)
      }
    }
    if (userId) fetch().catch(() => {})
  }, [userId])

  useEffect(() => {
    setCurrentDateTime(moment(lastUpdated).format(currentTimeDDateFormat))
    setGridInfo()
  }, [usageDetails, filters])

  // functions
  function setGridInfo(): void {
    const gridInfo = []
    let usageProductFamily = ''

    const getUsagesDescription = (usages: any): JSX.Element => {
      return (
        <span className="ms-4">
          ${usages.rate} {usages.usageUnitName} - {usages.productType}
        </span>
      )
    }

    for (const index in usageDetails) {
      const usages = { ...usageDetails[index] }

      if (getFilterCheck(usages)) {
        if (usageProductFamily !== usages.productFamily) {
          const categoryObj = {
            description: usages.productFamily,
            region: '',
            usageQuantity: '',
            amount: ''
          }
          gridInfo.push(categoryObj)
          usageProductFamily = usages.productFamily
        }

        const usageObj = {
          description: getUsagesDescription(usages),
          region: usages.regionName,
          usageQuantity: `${usages.usageQuantity} ${usages.usageQuantityUnitName}`,
          amount: {
            showField: true,
            type: 'currency',
            value: usages.amount
          }
        }

        gridInfo.push(usageObj)
      }
    }
    setFilteredUsages(gridInfo)
  }

  function getFilterCheck(data: any): any {
    let returnType = true

    if (filters.filterUsages) {
      const filterUsages = filters.filterUsages.toLowerCase()
      returnType =
        data.productFamily.toString().toLowerCase().indexOf(filterUsages) > -1 ||
        data.regionName.toString().toLowerCase().indexOf(filterUsages) > -1 ||
        data.rate.toString().toLowerCase().indexOf(filterUsages) > -1 ||
        data.usageUnitName.toString().toLowerCase().indexOf(filterUsages) > -1 ||
        data.productType.toString().toLowerCase().indexOf(filterUsages) > -1 ||
        data.usageQuantity.toString().toLowerCase().indexOf(filterUsages) > -1 ||
        data.usageQuantityUnitName.toString().toLowerCase().indexOf(filterUsages) > -1 ||
        data.amount.toString().toLowerCase().indexOf(filterUsages) > -1
    }

    return returnType
  }

  function onChangeInput(event: any, formInputName: string): void {
    const updatedState = {
      ...state
    }
    if (formInputName === 'clearFilters') {
      setState(initialState)
      setFilters(initialFilters)
    } else {
      const value = event.target.value
      const updatedForm = setFormValue(formInputName, value, updatedState.form)
      updatedState.form = updatedForm
      setState(updatedState)
      const updateFilter: any = { ...filters }
      updateFilter[formInputName] = value
      setFilters(updateFilter)
    }
  }

  return (
    <CloudUsageView
      loading={loading}
      filteredUsages={filteredUsages}
      state={state}
      columns={columns}
      currentDateTime={currentDateTime}
      onChangeInput={onChangeInput}
      totalAmount={totalAmount}
      period={period}
      hasUsages={usageDetails && usageDetails.length > 0}
    />
  )
}

export default CreditUsageContainer
