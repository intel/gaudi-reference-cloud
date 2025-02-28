// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Button from 'react-bootstrap/Button'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import ObjectStorageUsersPermissionsSelection from '../objectStorageUsersPermissionsManagement/ObjectStorageUsersPermissionsSelection'
import EmptyView from '../../../utils/emptyView/EmptyView'
import Spinner from '../../../utils/spinner/Spinner'

const ObjectStorageUserEdit = (props) => {
  // props
  const state = props.state
  const onSubmit = props.onSubmit
  const loading = props.loading
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const objectStorages = props.objectStorages
  const emptyBuckets = props.emptyBuckets

  // state
  const mainTitle = state.mainTitle
  const navigationBottom = state.navigationBottom
  const showReservationModal = state.showReservationModal
  const showErrorModal = state.showErrorModal
  const errorMessage = state.errorMessage
  const errorTitleMessage = state.errorTitleMessage
  const errorDescription = state.errorDescription

  // variables
  return (
    <>
      <ReservationSubmit showReservationCreateModal={showReservationModal} />
      <ErrorModal
        showModal={showErrorModal}
        titleMessage={errorTitleMessage}
        description={errorDescription}
        message={errorMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      {loading ? (
        <Spinner />
      ) : objectStorages.length > 0 ? (
        <>
          <div className="section">
            <h2 intc-id="title-objectStorageUserEdit">{mainTitle}</h2>
          </div>
          <div className="section">
            <div className="row">
              <div className="col-xs-12 col-md-8 col-xl-6">
                <h3>Permission management</h3>
                <ObjectStorageUsersPermissionsSelection buckets={objectStorages} isEdit={true} />
                {navigationBottom.map((item, index) => (
                  <Button
                    intc-id={`btn-objectStorageUserEdit-navigationBottom ${item.buttonLabel}`}
                    data-wap_ref={`btn-objectStorageUserEdit-navigationBottom ${item.buttonLabel}`}
                    key={index}
                    variant={item.buttonVariant}
                    className="btn"
                    onClick={item.buttonLabel === 'Save' ? onSubmit : item.buttonFunction}
                  >
                    {item.buttonLabel}
                  </Button>
                ))}
              </div>
            </div>
          </div>
        </>
      ) : (
        <EmptyView title={emptyBuckets.title} subTitle={emptyBuckets.subTitle} action={emptyBuckets.action} />
      )}
    </>
  )
}

export default ObjectStorageUserEdit
