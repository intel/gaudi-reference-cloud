import { NavLink } from 'react-router-dom'
import AddNodeToPoolContainer from '../../containers/nodePoolManagement/AddNodeToPoolContainer'
import GridPagination from '../../utility/gridPagination/gridPagination'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'
import SearchBox from '../../utility/searchBox/SearchBox'
import { Button } from 'react-bootstrap'

const PoolList = (props) => {
  // props Variables
  const columns = props.columns
  const pools = props.pools
  const emptyGrid = props.emptyGrid
  const filterText = props.filterText
  const loading = props.loading
  const showRequestModal = props.showRequestModal
  const selectedPool = props.selectedPool
  const cancelAddNode = props.cancelAddNode
  const addNodeToPoolFn = props.addNodeToPoolFn

  // Props Functions
  const setFilter = props.setFilter
  const backToHome = props.backToHome

  let gridItems = []

  if (filterText !== '') {
    const input = filterText.toLowerCase()

    if (gridItems.length === 0) {
      gridItems = pools.filter((item) => item.poolId.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = pools.filter((item) => item.poolName.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = pools.filter((item) => item.poolAccountManagerAgsRole.toLowerCase().includes(input))
    }
  } else {
    gridItems = pools
  }

  return (
    <>
      {selectedPool && (
        <AddNodeToPoolContainer
          selectedPool={selectedPool}
          cancelAddNode={cancelAddNode}
          addNodeToPoolFn={addNodeToPoolFn}
        />
      )}
      <OnSubmitModal showModal={showRequestModal} message="Working on your request" />
      <div className="section">
        <Button variant="link" className="p-s0" onClick={() => backToHome()}>
          ⟵ Back to Home
        </Button>
      </div>
      <div className={`filter flex-wrap ${!pools || pools.length === 0 ? 'd-none' : ''}`}>
        <NavLink className="btn btn-primary" to="/npm/pools/create" aria-label='Create new pool' intc-id={'btn-navigate-PoolCreate'}>
          Create new pool
        </NavLink>
        <SearchBox
          intc-id="searchPools"
          value={filterText}
          onChange={setFilter}
          placeholder="Search pools..."
          aria-label="Type to search pools.."
        />
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default PoolList
