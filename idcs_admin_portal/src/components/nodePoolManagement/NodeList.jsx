import GridPagination from '../../utility/gridPagination/gridPagination'
import SearchBox from '../../utility/searchBox/SearchBox'
import { Button } from 'react-bootstrap'

const NodeList = (props) => {
  // props Variables
  const columns = props.columns
  const nodes = props.nodes
  const emptyGrid = props.emptyGrid
  const filterText = props.filterText
  const loading = props.loading
  const poolId = props.poolId

  // Props Functions
  const setFilter = props.setFilter
  const backToHome = props.backToHome

  function getFilteredData() {
    let filteredData = []

    if (filterText !== '') {
      for (const index in nodes) {
        const quota = { ...nodes[index] }
        if (getFilterCheck(filterText, quota)) {
          filteredData.push(quota)
        }
      }
    } else {
      filteredData = [...nodes]
    }

    return filteredData
  }

  const gridItems = getFilteredData()

  function getFilterCheck(filterValue, data) {
    filterValue = filterValue.toLowerCase()

    return (
      data.name.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.region.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.clusterId.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.az.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.instanceTypes.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.pools?.value.toString().toLowerCase().indexOf(filterValue) > -1
    )
  }

  return (
    <>
      <div className="section">
        <Button variant="link" className="p-s0" onClick={() => backToHome()}>
          ⟵ Back to {poolId ? 'pool ' + poolId : 'Home'}
        </Button>
      </div>
      <div className="filter">
        <h2 className="h4">Nodes {poolId ? 'for Pool: ' + poolId : ''}</h2>
        <SearchBox
          intc-id="searchNodes"
          value={filterText}
          onChange={setFilter}
          placeholder="Search nodes..."
          aria-label="Type to search nodes..."
        />
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default NodeList
