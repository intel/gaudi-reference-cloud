import React, { useEffect, useState } from 'react'
import CustomInput from '../../utility/customInput/CustomInput'
import Wrapper from '../../utility/wrapper/Wrapper'
import { Button } from 'react-bootstrap'
import { AiOutlineReload } from 'react-icons/ai'

function SelectedFilters(props) {
  const { selectedFilters, events } = props

  const onClearEach = (selectedFilter) => {
    events.onClear(selectedFilter)
  }

  const onClearAll = () => {
    events.onClearAll()
  }

  const selectedFiltersArr = []
  for (const filter in selectedFilters) {
    selectedFilters[filter].forEach((value) => {
      selectedFiltersArr.push({
        key: filter,
        value
      })
    })
  }

  if (selectedFiltersArr.length <= 0) {
    return null
  }

  return (
    <Wrapper>
      <div className='m-2 d-flex'>
        {selectedFiltersArr.map((selectedFilter, index) => {
          return (
            <button
              key={index}
              type='button'
              className='px-2 py-1 me-2 btn border rounded'
              aria-label='Close'
              onClick={() => onClearEach(selectedFilter)}
            >
              {selectedFilter.value} &nbsp;
              <span className=' btn-sm btn-close' />
            </button>
          )
        })}

        <button
          type='button'
          className='px-2 py-1 ms-2 btn border rounded'
          aria-label='Clear All'
          onClick={() => onClearAll()}
        >
          Clear All Filters &nbsp;
          <span className=' btn-sm btn-close' />
        </button>
      </div>
    </Wrapper>
  )
}

function Filter(props) {
  const { filters = [], data = [], onFiltersChanged } = props

  const [activeFilter, setActiveFilter] = useState(null)
  const [filtersData, setFiltersData] = useState({})
  const [selectedFilters, setSelectedFilters] = useState(null)
  const [filterValues, setFilterValues] = useState([])

  useEffect(() => {
    getFilters()
  }, [])

  const getFilters = () => {
    const filterObj = {}
    const selectedFiltersObj = {}

    const filterKeys = filters.map((filter) => filter.key)
    data.forEach((obj) => {
      filterKeys.forEach((key) => {
        if (Object.hasOwn(obj, key)) {
          if (!Object.hasOwn(filterObj, key)) {
            filterObj[key] = {}
            selectedFiltersObj[key] = []
          }

          const filterKey = filterObj[key]
          let subKey = obj[key]
          if (typeof subKey === 'undefined' || subKey === '') {
            subKey = 'N/A'
          }
          filterKey[subKey] = (filterKey[subKey] ?? 0) + 1
        }
      })
    })

    setFiltersData({ ...filterObj })
    setSelectedFilters({ ...selectedFiltersObj })
  }

  useEffect(() => {
    getFilteredData()
  }, [selectedFilters])

  const getFilteredData = () => {
    let filteredData = data

    for (const filterKey in selectedFilters) {
      if (Object.hasOwn(selectedFilters, filterKey)) {
        const filterValues = selectedFilters[filterKey]

        filteredData = filteredData.filter((obj) => {
          if (filterValues.length > 0) {
            let subKey = obj[filterKey]
            if (subKey === '') {
              subKey = 'N/A'
            }
            return filterValues.includes(subKey)
          }

          return true
        })
      }
    }

    onFiltersChanged(filteredData)
  }

  const onFilterTabClick = (filter) => {
    const filterVal = []

    for (const subFilterKey in filtersData[filter.key]) {
      filterVal.push({
        key: subFilterKey,
        count: filtersData[filter.key][subFilterKey]
      })
    }

    setFilterValues([...filterVal])
    setActiveFilter((prev) => {
      if (prev?.key === filter.key) {
        return null
      }

      return filter
    })
  }

  const onSelectFilter = (filterValue) => {
    let selectedValues = [...selectedFilters[activeFilter.key]]

    if (selectedValues.includes(filterValue.key)) {
      selectedValues = selectedValues.filter((val) => val !== filterValue.key)
    } else {
      selectedValues.push(filterValue.key)
    }

    const selectedFiltersCopy = { ...selectedFilters }
    selectedFiltersCopy[activeFilter.key] = selectedValues
    setSelectedFilters(selectedFiltersCopy)
  }

  const onRemoveFilter = (selectedFilter) => {
    let updatedValues = [...selectedFilters[selectedFilter.key]]
    updatedValues = updatedValues.filter((val) => val !== selectedFilter.value)

    const selectedFiltersCopy = { ...selectedFilters }
    selectedFiltersCopy[selectedFilter.key] = updatedValues
    setSelectedFilters(selectedFiltersCopy)
  }

  const onClearFilters = () => {
    const selectedFiltersCopy = { ...selectedFilters }

    for (const filterKey in selectedFiltersCopy) {
      selectedFiltersCopy[filterKey] = []
    }

    setSelectedFilters(selectedFiltersCopy)
  }

  const selectedFiltersEvents = {
    onClear: (selectedFilter) => {
      onRemoveFilter(selectedFilter)
    },
    onClearAll: () => {
      onClearFilters()
    }
  }

  return (
    <>
      <div className='d-flex align-items-center'>
        <h4>Filters</h4>

        <Button variant="link"
          onClick={() => { getFilters() }}
        >
          <AiOutlineReload />
        </Button>
      </div>

      <div className='d-flex flex-column gap-s6 w-100'>
        <ul className='nav nav-tabs d-flex flex-row'>
          {filters.map((filter, index) => (
            <li className='nav-item' key={index}>
              <a
                className={`nav-link small ${activeFilter?.key === filter.key ? 'border' : ''}`}
                onClick={() => onFilterTabClick(filter)}
              >
                {filter.label}
              </a>
            </li>
          ))}
        </ul>

        {activeFilter !== null
          ? (
              <div>
                {filterValues.map((filterValue, index) => {
                  let value = false
                  if (selectedFilters) {
                    value = selectedFilters[activeFilter.key].includes(filterValue.key)
                  }

                  return (
                    <CustomInput
                      key={index}
                      type={'checkbox'}
                      fieldSize={'small'}
                      options={[{
                        name: `${filterValue.key} (${filterValue.count})`,
                        value: '0'
                      }]}
                      label={''}
                      value={value}
                      onChanged={() => onSelectFilter(filterValue)}
                    />
                  )
                })}
              </div>
            )
          : null
        }

        <SelectedFilters selectedFilters={selectedFilters} events={selectedFiltersEvents} />
      </div>

    </>
  )
}

export default Filter
