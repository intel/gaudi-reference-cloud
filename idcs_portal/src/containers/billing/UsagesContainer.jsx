// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import moment from 'moment'

import UsagesView from '../../components/billing/usages/UsagesView'
import useUsagesReport from '../../store/billingStore/UsagesStore'
import { setFormValue } from '../../utils/updateFormHelper/UpdateFormHelper'
import useErrorBoundary from '../../hooks/useErrorBoundary'

const UsagesContainers = () => {
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

  const initialFilters = {
    filterUsages: null
  }

  const currentTimeDDateFormat = 'dddd, MMMM D, YYYY [at] hh:mm:ss A'

  // local state
  const [filteredUsages, setFilteredUsages] = useState([])
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
    const fetch = async () => {
      try {
        await setUsage()
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    setCurrentDateTime(moment(lastUpdated).format(currentTimeDDateFormat))
    setGridInfo()
  }, [usageDetails, filters])

  // functions
  function setGridInfo() {
    const gridInfo = []
    let usageProductFamily = ''

    const getUsagesDescription = (usages) => {
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

  function getFilterCheck(data) {
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

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }
    if (formInputName === 'clearFilters') {
      setState(initialState)
      setFilters(initialState)
    } else {
      const value = event.target.value
      const updatedForm = setFormValue(formInputName, value, updatedState.form)
      updatedState.form = updatedForm
      setState(updatedState)
      const updateFilter = { ...filters }
      updateFilter[formInputName] = value
      setFilters(updateFilter)
    }
  }

  return (
    <UsagesView
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

export default UsagesContainers
