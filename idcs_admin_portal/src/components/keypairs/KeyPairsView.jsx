import { Alert, Button } from 'react-bootstrap'
import DeleteConfirmationModal from '../../utility/DeleteConfirmationModal'
import GridPagination from '../../utility/gridPagination/gridPagination'
import Wrapper from '../../utility/wrapper/Wrapper'
import { Link } from 'react-router-dom'

const KeyPairsView = (props) => {
  // props variables
  const columns = props.columns
  const myPublicKeys = props.myPublicKeys
  const displayConfirmationModal = props.displayConfirmationModal
  const deleteMessage = props.deleteMessage
  const deletedKPMessage = props.deletedKPMessage
  const deleteShow = props.deleteShow

  // props functions
  const deleteKeyPairsHandler = props.deleteKeyPairsHandler
  const hideConfirmationModal = props.hideConfirmationModal
  const setDeleteShow = props.setDeleteShow

  let infoView = (
    <div className="p-3">
      <div className="jumbotron jumbotron-fluid bg-white">
        <div className="text-center">
          <h1>No keys available</h1>
          <p>Your account currently has no keys.</p>
          <span className="mr-2">
            <Link to="/security/publickeys/import">
              <Button intc-id="UploadKeyOnNoKeysAvailableButton" variant="primary">
                Upload key
              </Button>
            </Link>
          </span>
        </div>
      </div>
    </div>
  )

  if (myPublicKeys.length > 0) {
    infoView = (
      <div intc-id="sshPublicKeysTable">
        <GridPagination
        data={myPublicKeys}
        columns={columns}
      />
      </div>
    )
  }

  return (
    <Wrapper>
      <br />
      {deletedKPMessage && deleteShow && (
        <Alert
          variant="success"
          className="col-6 center"
          onClose={() => setDeleteShow(false)}
          dismissible
        >
          {deletedKPMessage}
        </Alert>
      )}

      <DeleteConfirmationModal
        showModal={displayConfirmationModal}
        confirmModal={deleteKeyPairsHandler}
        hideModal={hideConfirmationModal}
        message={deleteMessage}
      />

      <div className="ps-2 pb-3">
        {myPublicKeys.length <= 0
          ? <div className="d-flex justify-content-between main">
          <span className="col-6">
            <strong className="h3" intc-id="myPublicKeysTitle">
              {' '}
              My keys
            </strong>
          </span>
          <span className="col-6 text-end">
            <Link to="/security/publickeys/import">
              <Button variant="primary" intc-id="UploadKeyButton">
                Upload key
              </Button>
            </Link>
          </span>
        </div>
          : null}
      </div>

      {props.loading
        ? (
        <div className="col-12 row">
          <div className="spinner-border text-primary center"></div>
        </div>
          )
        : (
        <>
        {myPublicKeys.length > 0
          ? <h1 className="ps-3">Account keys</h1>
          : null}
        <div style={{ clear: 'both' }} className={'p-3 bg-white'}>
        {myPublicKeys.length > 0
          ? <div className="row">
          <h3 className="col-6">My Keys</h3>
          <span className="col-6 text-end">
            <Link to="/security/publickeys/import">
              <Button variant="primary" intc-id="UploadKeyButtonSecondary">
                Upload key
              </Button>
            </Link>
          </span>
        </div>

          : null}
          {infoView}
        </div>
        </>
          )}
    </Wrapper>
  )
}

export default KeyPairsView
