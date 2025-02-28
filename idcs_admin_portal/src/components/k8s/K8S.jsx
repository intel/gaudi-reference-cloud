import React from 'react'
import GridPagination from '../../utility/gridPagination/gridPagination'
import Button from 'react-bootstrap/Button'
import SearchBox from '../../utility/searchBox/SearchBox'

function K8S(props) {
  const {
    gridData: { data, columns, emptyGridObject, loading },
    backToHome,
    filterText,
    setFilter
  } = props

  let gridItems = []

  if (filterText !== '' && data) {
    const input = filterText.toLowerCase()
    gridItems = data.filter((item) => {
      const k8sName = item.versionName
      return k8sName === undefined || k8sName.toLowerCase().includes(input)
    })
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.provider.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.cpIMI.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.workerIMI.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.state.toLowerCase().includes(input))
    }
  } else {
    gridItems = data
  }
  return (
    <>
      <div className='section'>
        <Button variant='link' className='p-s0' onClick={() => backToHome()}>
          ⟵ Back to Home
        </Button>
      </div>

      <div className="filter">
          <h2 className='h4'>Intel Kubernetes K8s Versions</h2>
          <div>
              <SearchBox
                intc-id="searchIKSVersions"
                value={filterText}
                onChange={setFilter}
                placeholder="Search versions..."
                aria-label="Type to search versions.."
              />
          </div>
        </div>

      <div className='section'>
        <GridPagination data={gridItems} columns={columns} emptyGrid={emptyGridObject} loading={loading} />
      </div>
    </>
  )
}

export default K8S
