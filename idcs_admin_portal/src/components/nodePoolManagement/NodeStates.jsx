import GridPagination from '../../utility/gridPagination/gridPagination'
import SearchBox from '../../utility/searchBox/SearchBox'
import { Button } from 'react-bootstrap'

const NodeStates = (props) => {
  // props Variables
  const columns = props.columns
  const nodeStatesList = props.nodeStatesList
  const emptyGrid = props.emptyGrid
  const filterText = props.filterText
  const loading = props.loading
  const nodeDetails = props.nodeDetails
  const isPageReady = props.isPageReady

  // Props Functions
  const setFilter = props.setFilter
  const backToHome = props.backToHome

  function getFilteredData() {
    let filteredData = []

    if (filterText !== '') {
      for (const index in nodeStatesList) {
        const quota = { ...nodeStatesList[index] }
        if (getFilterCheck(filterText, quota)) {
          filteredData.push(quota)
        }
      }
    } else {
      filteredData = [...nodeStatesList]
    }

    return filteredData
  }

  const gridItems = getFilteredData()

  function getFilterCheck(filterValue, data) {
    filterValue = filterValue.toLowerCase()

    return data.instanceType.toString().toLowerCase().indexOf(filterValue) > -1
  }

  return !isPageReady ? (
    <div className="section">
      <div className="row align-self-center">
        <div className="spinner-border text-primary center"></div>
      </div>
    </div>
  ) : (
    <>
      <div className="section">
        <Button variant="link" className="p-s0" onClick={() => backToHome()}>
          ⟵ Back to nodes
        </Button>
      </div>
      <div className="section">
        <h1 className="h2">statistic for {nodeDetails.nodeName} </h1>
      </div>
      <div className="section">
        <div className="row mb-2">
          <div className="col-12 col-md-2 fw-bold">Node Name:</div>
          <div className="col-12 col-md-4">{nodeDetails.nodeName}</div>
        </div>
        <div className="row mb-2">
          <div className="col-12 col-md-2 fw-bold">Region:</div>
          <div className="col-12 col-md-4">{nodeDetails.region}</div>
        </div>
        <div className="row mb-2">
          <div className="col-12 col-md-2 fw-bold">Availability zone:</div>
          <div className="col-12 col-md-4">{nodeDetails.availabilityZone}</div>
        </div>
        <div className="row mb-2">
          <div className="col-12 col-md-2 fw-bold">Resources Used:</div>
          <div className="col-12 col-md-4">{nodeDetails.percentageResourcesUsed}%</div>
        </div>
        <div className="row mb-2">
          <div className="col-12 col-md-2 fw-bold">Instance Types:</div>
          <div className="col-12 col-md-10">{nodeDetails?.instanceTypes.join(', ')}</div>
        </div>
        <div className="row mb-2">
          <div className="col-12 col-md-2 fw-bold">Compute Node Pools:</div>
          <div className="col-12 col-md-10">{nodeDetails?.poolIds.join(', ')}</div>
        </div>
        <hr />
      </div>
      <div className="filter">
        <SearchBox
          intc-id="searchNodeStates"
          value={filterText}
          onChange={setFilter}
          placeholder="Search instance types..."
          aria-label="Type to search instance types..."
        />
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default NodeStates
