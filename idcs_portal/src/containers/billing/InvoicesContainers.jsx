// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import moment from 'moment'
import InvoicesView from '../../components/billing/invoices/InvoicesView'
import useInvoicesStore from '../../store/billingStore/InvoicesStore'
import { setFormValue } from '../../utils/updateFormHelper/UpdateFormHelper'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import InvoicesService from '../../services/InvoicesService'
import makePDF from '../../utils/ariaHTMLtoPDF/ariaHTMLtoPDF'
import { checkRoles } from '../../utils/accessControlWrapper/AccessControlWrapper'
import { AppRolesEnum } from '../../utils/Enums'

const InvoicesContainers = () => {
  // local state
  const columns = [
    {
      columnName: 'Invoice ID',
      targetColumn: 'id'
    },
    {
      columnName: 'Billing period',
      targetColumn: 'period'
    },
    {
      columnName: 'Period start',
      targetColumn: 'startDate'
    },
    {
      columnName: 'Period end',
      targetColumn: 'endDate'
    },
    {
      columnName: 'Due date',
      targetColumn: 'dueDate'
    },
    {
      columnName: 'Status',
      targetColumn: 'status'
    },
    {
      columnName: 'Total amount',
      targetColumn: 'total',
      className: 'text-end'
    },
    {
      columnName: 'Amount paid',
      targetColumn: 'totalPaid',
      className: 'text-end'
    },
    {
      columnName: 'Amount due',
      targetColumn: 'totalDue',
      className: 'text-end'
    }
  ]

  const getDownloadValue = (invoice) => {
    return <span>{invoice}</span>
  }

  // local state
  const initialState = {
    form: {
      filterInvoices: {
        type: 'text', // options = 'text ,'textArea'
        label: 'Filter Invoices:',
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
      filterStartDate: {
        type: 'date', // options = 'text ,'textArea'
        label: 'Start Date',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        isRequired: false,
        validationRules: {
          isRequired: false
        },
        formClass: 'rounded-0',
        startDate: 'filterStartDate',
        endDate: 'filterEndDate',
        maxDate: 'filterEndDate',
        selectsStart: true
      },
      filterEndDate: {
        type: 'date', // options = 'text ,'textArea'
        label: 'End Date',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        isRequired: false,
        validationRules: {
          isRequired: false
        },
        showYearDropdown: true,
        showMonthDropdown: true,
        startDate: 'filterStartDate',
        endDate: 'filterEndDate',
        minDate: 'filterStartDate',
        selectsEnd: true
      },
      isValidForm: false
    }
  }

  const initialFilters = {
    filterInvoices: null,
    filterStartDate: null,
    filterEndDate: null
  }

  const currentTimeDDateFormat = 'dddd, MMMM D, YYYY [at] hh:mm:ss A'
  const monthFormat = 'MMMM, YYYY'
  const dateFormat = 'MMMM D, YYYY'
  const dateFormatfilter = 'MM/DD/YYYY'

  const [filteredInvoices, setFilteredInvoices] = useState([])
  const [filters, setFilters] = useState(initialFilters)
  const [state, setState] = useState(initialState)
  const [currentDateTime, setCurrentDateTime] = useState(moment().format(currentTimeDDateFormat))

  // Global State
  const loading = useInvoicesStore((state) => state.loading)
  const invoices = useInvoicesStore((state) => state.invoices)
  const setInvoices = useInvoicesStore((state) => state.setInvoices)
  const lastUpdated = useInvoicesStore((state) => state.lastUpdated)

  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await setInvoices()
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    setCurrentDateTime(moment(lastUpdated).format(currentTimeDDateFormat))
    setGridInfo()
  }, [invoices, filters])

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in invoices) {
      const invoice = { ...invoices[index] }
      if (getFilterCheck(invoice)) {
        gridInfo.push({
          id: checkRoles([AppRolesEnum.Premium, AppRolesEnum.Enterprise])
            ? {
                showField: true,
                type: 'hyperlink',
                value: getDownloadValue(invoice.id),
                function: () => {
                  setDownload(invoice.id)
                }
              }
            : invoice.id,
          period: {
            showField: true,
            type: 'date',
            value: invoice.period,
            format: monthFormat
          },
          startDate: {
            showField: true,
            type: 'date',
            value: invoice.startDate,
            format: dateFormat
          },
          endDate: {
            showField: true,
            type: 'date',
            value: invoice.endDate,
            format: dateFormat
          },
          dueDate: {
            showField: true,
            type: 'date',
            value: invoice.dueDate,
            format: dateFormat
          },
          status: invoice.status,
          total: {
            showField: true,
            type: 'currency',
            value: invoice.total
          },
          totalPaid: {
            showField: true,
            type: 'currency',
            value: invoice.totalPaid
          },
          totalDue: {
            showField: true,
            type: 'currency',
            value: invoice.totalDue
          }
        })
      }
    }

    setFilteredInvoices(gridInfo)
  }

  async function setDownload(id) {
    try {
      const invoice = invoices.find((item) => item.id === id)
      const { data } = await InvoicesService.getInvoicePDF(invoice.id)
      makePDF(atob(data?.statement), invoice.id)
    } catch (error) {
      throwError(error)
    }
  }

  function getFilterCheck(data) {
    let returnType = true

    if (filters.filterInvoices) {
      const filterInvoices = filters.filterInvoices.toLowerCase()

      returnType =
        data.id.toString().toLowerCase().indexOf(filterInvoices) > -1 ||
        data.period.toString().toLowerCase().indexOf(filterInvoices) > -1 ||
        moment(data.dueDate).format(dateFormat).toLowerCase().indexOf(filterInvoices) > -1 ||
        data.status.toString().toLowerCase().indexOf(filterInvoices) > -1 ||
        data.total.toString().toLowerCase().indexOf(filterInvoices) > -1 ||
        data.totalPaid.toString().toLowerCase().indexOf(filterInvoices) > -1 ||
        data.totalDue.toString().toLowerCase().indexOf(filterInvoices) > -1
    }
    if (returnType && filters.filterStartDate) {
      returnType =
        moment(data.startDate).utc().format(dateFormatfilter) >=
        moment(filters.filterStartDate).utc().format(dateFormatfilter)
    }

    if (returnType && filters.filterEndDate) {
      returnType =
        moment(data.endDate).utc().format(dateFormatfilter) <=
        moment(filters.filterEndDate).utc().format(dateFormatfilter)
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
    <InvoicesView
      loading={loading}
      filteredInvoices={filteredInvoices}
      state={state}
      columns={columns}
      currentDateTime={currentDateTime}
      onChangeInput={onChangeInput}
      hasInvoices={invoices && invoices.length > 0}
    />
  )
}

export default InvoicesContainers
