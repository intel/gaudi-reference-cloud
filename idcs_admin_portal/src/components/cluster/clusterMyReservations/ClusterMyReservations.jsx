import React, { useEffect, useState } from 'react'
import GridPagination from '../../../utility/gridPagination/gridPagination'
import TapContent from '../../../utility/TapContent/TapContent'
import { Modal, CloseButton } from 'react-bootstrap'
import CustomInput from '../../../utility/customInput/CustomInput'
import Button from 'react-bootstrap/Button'
import ModalCreatePublicKey from '../../../utility/modals/modalCreatePublicKey/ModalCreatePublicKey'
import Filter from '../../filter/Filter'
import EmptyView from '../../../utility/emptyView/EmptyView'
import ExportDataToCSV from '../../../utility/ExportDataToCSV'
import TabsNavigation from '../../../utility/tabsNavigation/TabsNavagation'
import SearchBox from '../../../utility/searchBox/SearchBox'
import { useCopy } from '../../../hooks/useCopy'
import SearchCloudAccountModal from '../../../utility/modals/searchCloudAccountModal/SearchCloudAccountModal'

function GridPaginationSearch(props) {
  const {
    data,
    columns,
    emptyGrid,
    loading,
    emptyGridByFilter,
    showSearch,
    summary
  } = props

  const emptyGridByFilterObject = {
    title: 'No Results found',
    subTitle: 'The applied filter criteria did not match any items',
    ...emptyGridByFilter,
    action: {
      type: 'function',
      href: () => onClearFilters(),
      label: 'Clear filters'
    }
  }

  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)

  const onChangeSearchInput = (event) => {
    setEmptyGridObject(emptyGridByFilterObject)
    setFilterText(event.target.value)
  }

  const onClearFilters = () => {
    setEmptyGridObject(emptyGrid)
    setFilterText('')
  }

  let gridItems = data
  if (filterText !== '' && data) {
    const input = filterText.toLowerCase()

    gridItems = data.filter((item) => item.name.toString().toLowerCase().includes(input))
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.stateOriginal.value.toString().toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.imi.toString().toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.ipaddress.toString().toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.wekaStorage.toString().toLowerCase().includes(input))
    }
  }

  return (
    <div>
      {
        showSearch ? (
        <div className="section">
          <div className="filter flex-wrap p-0">
            <div>
              {summary}
            </div>
            <SearchBox
              intc-id="filterAccounts"
              value={filterText}
              onChange={onChangeSearchInput}
              placeholder="Filter accounts..."
              aria-label="Type to filter cloud accounts..."
            />
          </div>
        </div>
        ) : null
      }
      <div className="section px-0">
        <GridPagination
          data={gridItems}
          columns={columns}
          emptyGrid={emptyGridObject}
          loading={loading}
        />
      </div>
    </div>
  )
}

const ClusterMyReservations = (props) => {
  // *****
  // props
  // *****

  const columns = props.columns
  const myreservations = props.myreservations
  const clusterDetails = props.clusterDetails
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const taps = props.taps
  const tapDetails = props.tapDetails
  const activeTap = props.activeTap
  const setActiveTap = props.setActiveTap
  const showPublicKeyModal = props.showPublicKeyModal
  const onShowHidePublicKeyModal = props.onShowHidePublicKeyModal
  const afterPubliKeyCreate = props.afterPubliKeyCreate
  const setDetails = props.setDetails
  const nodegroupsInfo = props.nodegroupsInfo
  const upgradeK8sAvailable = props.upgradeK8sAvailable
  const upgradeK8sVersions = props.upgradeK8sVersions
  const certExpirations = props.certExpirationInfo
  const snapShotDetails = props.snapshots
  const storages = props.storages
  const onChangeInput = props.onChangeInput
  const isValidForm = props.isValidForm
  const formElementsLoadBalancerCreate = props.formElementsLoadBalancerCreate
  const formOnChangeTagValue = props.onChangeTagValue
  const formOnClickActionTag = props.onClickActionTag
  const onChangeTagValueMetadata = props.onChangeTagValueMetadata
  const onClickActionTagMetadata = props.onClickActionTagMetadata
  const onChangeAnnotationsValue = props.onChangeAnnotationsValue
  const onClickActionAnnotations = props.onClickActionAnnotations
  const loadBalancers = props.loadBalancers
  const securityDetails = props.securityDetails
  const resfreshLoadBalancers = props.resfreshLoadBalancers
  const emptyGrid = props.emptyGrid
  const filterText = props.filterText
  const setFilter = props.setFilter
  const showLBModal = props.showLBModal
  const showDeleteNodegroupModal = props.showDeleteNodegroupModal
  const closeDeleteModal = props.closeDeleteModal
  const deleteLoadBalancer = props.deleteLoadBalancer
  const successAPICall = props.successAPICall
  const closeSuccessAPIModal = props.closeSuccessAPIModal
  const showCreateLBModal = props.showCreateLBModal
  const closeLBModal = props.closeLBModal
  const addLoadBalancer = props.addLoadBalancer
  const setDeleteLbId = props.setDeleteLbId
  const showDeleteModal = props.showDeleteModal
  const upgradeNodeGroup = props.upgradeNodeGroup
  const showIMIUpgradeModal = props.showIMIUpgradeModal
  const closeIMIModal = props.closeIMIModal
  const setSelectedVersion = props.setSelectedVersion
  const upgradeClusterK8sVersion = props.upgradeClusterK8sVersion
  const setSucessAPICall = props.setSucessAPICall
  const successMessage = props.successMessage
  const setSuccessMessage = props.setSuccessMessage
  const onChangeLoadBalancerInput = props.onChangeLoadBalancerInput
  const loadBalancerBottom = props.loadBalancerBottom
  const getActionItemLabelCluster = props.getActionItemLabelCluster
  const deleteLbName = props.deleteLbName
  const setDeleteLbName = props.setDeleteLbName
  const loading = props.loading
  const backToHome = props.backToHome
  const sshKeys = props.sshKeys
  const isSREAdminUser = props.isSREAdminUser
  const filterIKS = props.filterIKS
  // Search Cloud Account Props
  const cloudAccountError = props.cloudAccountError
  const handleSearchInputChange = props.handleSearchInputChange
  const handleSubmit = props.handleSubmit
  const cloudAccount = props.cloudAccount
  const selectedCloudAccount = props.selectedCloudAccount
  const selectedClusterCloudAccount = props.selectedClusterCloudAccount
  const showSearchCloudAccount = props.showSearchCloudAccount
  const setShowSearchModal = props.setShowSearchModal
  const showLoader = props.showLoader
  const {
    showAuthModal,
    openAuthModal,
    closeAuthModal,
    onAuthSuccess,
    authPassword,
    setAuthPassword,
    authFormFooterEvents = {},
    showUnAuthModal,
    setShowUnAuthModal
  } = props.authModalData || {}
  const csvData = props.csvData

  // *****
  // custom component
  // *****
  const { copyToClipboard } = useCopy()
  function buildCustomInput(element, option) {
    let onClickActionTag
    let onChangeTagValue
    let onChangeInputOption
    switch (option) {
      case 'nodegroupForm':
        onClickActionTag = formOnClickActionTag
        onChangeTagValue = formOnChangeTagValue
        onChangeInputOption = onChangeInput
        break
      case 'loadbalancer':
        onClickActionTag = formOnClickActionTag
        onChangeTagValue = formOnChangeTagValue
        onChangeInputOption = onChangeLoadBalancerInput
        break
      case 'tagsForm':
        onClickActionTag = onClickActionTagMetadata
        onChangeTagValue = onChangeTagValueMetadata
        break
      case 'annotationsForm':
        onClickActionTag = onClickActionAnnotations
        onChangeTagValue = onChangeAnnotationsValue
        break
      default:
        break
    }
    return <CustomInput
        key={element.id}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={element.configInput.validationRules.isRequired ? element.configInput.label + ' *' : element.configInput.label}
        value={element.configInput.value}
        onChanged={(event) => onChangeInputOption(event, element.id)}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        readOnly={element.configInput.readOnly}
        refreshButton={element.configInput.refreshButton}
        isMultiple={element.configInput.isMultiple}
        onChangeSelectValue={element.configInput.onChangeSelectValue}
        extraButton={element.configInput.extraButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
        onClickActionTag={onClickActionTag}
        onChangeTagValue={onChangeTagValue}
        onChangeDropdownMultiple={onChangeDropdownMultiple}
      />
  }

  // *****
  // variables
  // *****
  let instanceDetails = null

  if (clusterDetails !== null) {
    instanceDetails = <div className='section'>
      <div className="d-flex flex-row justify-content-between align-items-center w-100">
        <h3>Cluster name: {clusterDetails.name}</h3>
        <CloseButton onClick={() => setDetails()}/>
      </div>
      <TabsNavigation tabs={taps} activeTab={activeTap} setTabActive={setActiveTap} />
      <TapContent infoToDisplay={getTapContent(activeTap)} />
    </div>
  }

  useEffect(() => {
    setFilteredData(myreservations)
  }, [myreservations])

  const [filteredData, setFilteredData] = useState(myreservations)

  let gridItems = []
  if (filterText !== '' && filteredData) {
    const input = filterText.toLowerCase()
    gridItems = filteredData.filter((item) => {
      const clusterValue = item['cluster-id'].value
      return clusterValue === undefined || clusterValue.toLowerCase().includes(input)
    })
    if (gridItems.length === 0) {
      gridItems = filteredData.filter((item) => item['cluster-name'].toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = filteredData.filter((item) => item.account.value.includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = filteredData.filter((item) => item.clusterstatus.toLowerCase().includes(input))
    }
  } else {
    gridItems = filteredData
  }

  // *****
  // functions
  // *****
  function getTapContent(tapNumber) {
    if (activeTap >= taps.length) return <></>
    let content = <></>
    const tabLabel = taps?.[activeTap]?.label
    switch (tabLabel) {
      case 'Overview':
        tapDetails[activeTap].customContent = getOverview()
        content = tapDetails[activeTap]
        break
      case 'Compute':
        tapDetails[activeTap].customContent = getManagedNodes()
        content = tapDetails[activeTap]
        break
      case 'Upgrade K8s Version':
        tapDetails[activeTap].customContent = getUpgradeK8sVersion()
        content = tapDetails[activeTap]
        break
      case 'Load Balancers':
        tapDetails[activeTap].customContent = getNetwork()
        content = tapDetails[activeTap]
        break
      case 'Certificates Expiration':
        tapDetails[activeTap].customContent = getCertificateDetails()
        content = tapDetails[activeTap]
        break
      case 'Snapshots':
        tapDetails[activeTap].customContent = getSnapshots()
        content = tapDetails[activeTap]
        break
      case 'SSH Keys':
        tapDetails[activeTap].customContent = getSSHKeys()
        content = tapDetails[activeTap]
        break
      case 'Storage':
        tapDetails[activeTap].customContent = getStorage()
        content = tapDetails[activeTap]
        break
      case 'Security':
        tapDetails[activeTap].customContent = getSecurity()
        content = tapDetails[activeTap]
        break
      default:
        tapDetails[activeTap].customContent = getDefaultContent()
        content = tapDetails[activeTap]
        break
    }

    return content
  }

  function getDefaultContent() {
    return (
      <div></div>
    )
  }

  function getNodegroupMessage(nodegroup) {
    let message = 'No Status'
    if (nodegroup) {
      message = nodegroup.status ? nodegroup.status : 'No Status'
    }
    return message
  }
  function getNodeMessage(node) {
    let message = 'No Status'
    if (node) {
      message = node.status ? node.status : 'No Status'
    }
    return message
  }
  function getIlbMessage(ilb) {
    let message = 'No Status'
    if (ilb) {
      message = ilb.status ? ilb.status : 'No Status'
    }
    return message
  }

  function writeToFile(event, data, clusterDetails, keyType) {
    const element = document.createElement('a')
    const file = new Blob([data], { type: 'text/plain' })
    element.href = URL.createObjectURL(file)
    const fileName = clusterDetails.name + '- ' + clusterDetails.account + '-' + keyType + '.yaml'
    element.download = fileName
    document.body.appendChild(element)
    element.click()
  }
  // *****
  // taps
  // *****

  // intel load balancer tap

  function getNetwork() {
    if (loadBalancers.length <= 0) {
      return (
        <EmptyView
          title={'Failed to Fetch Load Balancers'}
          action={{
            type: 'function',
            href: () => { resfreshLoadBalancers() },
            label: 'Refresh'
          }}
        />
      )
    }

    return (
      <div className='w-100'>
        <div className="d-grid d-md-flex justify-content-md-end mb-2">
          <Button variant='primary' type="button" onClick={showLBModal} disabled>Add Load Balancer</Button>
        </div>

        <div>
          <Modal className="md" show={showDeleteNodegroupModal} onHide={closeDeleteModal}>
            <Modal.Header closeButton>
              <Modal.Title className="text-break">Delete Load Balancer</Modal.Title>
            </Modal.Header>
            <Modal.Body className="text-left small">
              <div className="w-100">
                <p>Are you sure you want to delete {deleteLbName}.</p>
              </div>
            </Modal.Body>
            <Modal.Footer>
              <Button variant="danger" onClick={async () => { await deleteLoadBalancer() }}>
                Delete
              </Button>
              <Button variant="outline-primary" onClick={closeDeleteModal}>
                Close
              </Button>
            </Modal.Footer>
          </Modal>
        </div>
        <div>
          <Modal className="modal-lg" show={showCreateLBModal} onHide={closeLBModal} >
            <Modal.Header closeButton>
              <Modal.Title>Add Load Balancer</Modal.Title>
            </Modal.Header>
            <Modal.Body className="m-3">
                  { formElementsLoadBalancerCreate &&
                      formElementsLoadBalancerCreate.map((element, index) => (
                        (<div className="d-flex justify-content-center" key={index}>
                        <div className='col-12 col-lg-10'>{buildCustomInput(element, 'loadbalancer')}</div>
                      </div>)
                      ))
                    }
            </Modal.Body>
            <Modal.Footer>
            {loadBalancerBottom.map((item, index) => (
            < Button
              key={index}
              disabled={item.buttonLabel === 'Create' ? !isValidForm : false}
              variant={item.buttonVariant}
              onClick={item.buttonLabel === 'Create' ? addLoadBalancer : item.buttonFunction} >
              {item.buttonLabel}
            </Button>
            ))}
            </Modal.Footer>
          </Modal>
        </div>
        { loadBalancers.length > 0 && (
        <table className="table table-row-dashed table-row-gray-500 gy-5 gs-5 mb-0">
          <thead>
            <tr className="fw-semibold ">
              <th scope="col">Name</th>
              <th scope="col">Backend Ports</th>
              <th scope="col">Frontend Ports</th>
              <th scope="col">Type</th>
              <th scope="col">State</th>
              <th scope="col">IP</th>
              <th scope="col">Created At</th>
              <th scope="col">Actions</th>
            </tr>
          </thead>
          <tbody>
            {loadBalancers.map((lb, index) => (
              <tr key={index}>
                <td>{lb.lbname}</td>
                <td>{lb.backendports}</td>
                <td>{lb.frontendportd}</td>
                <td>{lb.viptype}</td>
                <td>{getActionItemLabelCluster(lb.status, getIlbMessage(lb))}</td>
                <td>{lb.vip}</td>
                <td>{(new Date(lb.createddate)).toLocaleString(undefined, { timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone, day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' })}</td>
                <td><Button variant='outline-primary' onClick={() => { setDeleteLbName(lb.lbname); setDeleteLbId(lb.vip); showDeleteModal() }} disabled>Delete</Button></td>
              </tr>))}
          </tbody>
        </table>
        )}
      </div>)
  }

  // overview tap

  function getOverview() {
    return (
      (clusterDetails != null) && (selectedClusterCloudAccount != null) &&
        <div className='row w-100'>
          <div className='col-6'>
            <p className='my-2 fw-bold small mb-auto'><span className="lead fs-6 fw-bolder">Name:</span></p>
            <p className="lead fs-6">{clusterDetails.name}</p>
          </div>
          <div className='col-6'>
            <p className='my-2 fw-bold small mb-auto'><span className="lead fs-6 fw-bolder">Cluster ID:</span></p>
            <p><span className="lead fs-6"> {clusterDetails.uuid}</span></p>
          </div>
          <div className='col-6'>
            <p className='my-2 fw-bold small mb-auto'><span className="lead fs-6 fw-bolder">Account:</span></p>
            <p><span className="lead fs-6"> {clusterDetails.account}</span></p>
          </div>
          <div className='col-6'>
            <p className='my-2 fw-bold small mb-auto'><span className="lead fs-6 fw-bolder">Email:</span></p>
            <p><span className="lead fs-6"> {selectedClusterCloudAccount.name}</span></p>
          </div>
          <div className='col-6'>
          </div>
          <div className='col-6'>
            <p className='my-2 fw-bold small mb-auto'><span className="lead fs-6 fw-bolder">Kubernetes Version:</span></p>
            <p><span className="lead fs-6">{clusterDetails.k8sversion}</span></p>
          </div>
        </div>
    )
  }

  // compute tap

  function getManagedNodes() {
    const nodeGroupColumns = [
      {
        columnName: 'Name',
        targetColumn: 'name'
      },
      {
        columnName: 'State',
        targetColumn: 'state'
      },
      {
        columnName: 'State',
        targetColumn: 'stateOriginal',
        hideField: true
      },
      {
        columnName: 'IMI',
        targetColumn: 'imi'
      },
      {
        columnName: 'IP Address',
        targetColumn: 'ipaddress'
      },
      {
        columnName: 'Storage Status',
        targetColumn: 'wekaStorage'
      },
      {
        columnName: 'Created Date',
        targetColumn: 'createddate'
      },
      {
        columnName: 'DNS Name',
        targetColumn: 'dnsname'
      }
    ]

    const getGridItems = (nodes) => {
      const gridItems = []

      if (Array.isArray(nodes)) {
        nodes.forEach((node) => {
          gridItems.push({
            name: node.name,
            state: getActionItemLabelCluster(node.state, getNodeMessage(node)),
            stateOriginal: {
              showField: false,
              value: node.state
            },
            imi: node.imi,
            ipaddress: node.ipaddress,
            wekaStorage: node.wekaStorage?.status,
            createddate: new Date(node.createddate).toLocaleString(undefined, {
              timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
              day: '2-digit',
              month: '2-digit',
              year: 'numeric',
              hour: '2-digit',
              minute: '2-digit'
            }),
            dnsname: node.dnsname
          })
        })
      }

      return gridItems
    }

    return (
      <div className='w-100'>
        <Modal className="md" show={showIMIUpgradeModal} onHide={closeIMIModal}>
          <Modal.Header closeButton>
            <Modal.Title className="text-break">Upgrade IMI for Nodegroup</Modal.Title>
          </Modal.Header>
          <Modal.Body className="text-left small">

            <div className="w-100">
              <p>Do you want to continue?</p>
            </div>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="primary" onClick={() => { upgradeNodeGroup() }}>
              Continue
            </Button>
            <Button variant="outline-primary" onClick={closeIMIModal}>
              Close
            </Button>
          </Modal.Footer>
        </Modal>
        <table className="table table-fixed table-row-dashed table-row-gray-500 gy-5 gs-5 mb-0">
          {nodegroupsInfo.map((data, index) => (
              <tbody key={index}>
                <tr>
                  <th scope="row"><span className="lead fs-6 fw-bolder">Uuid: </span><span className="lead fs-6 fw-semibold"> {data.nodgroupuuid} </span></th>
                  <th scope="row"><span className="lead fs-6 fw-bolder">Name: </span><span className="lead fs-6 fw-semibold"> {data.name} </span></th>
                  <th scope="row">{getActionItemLabelCluster(data.nodegroupstate, getNodegroupMessage(data), 'nodegroup')} </th>
                  <th scope="row"><span className="lead fs-6 fw-bolder">Instance Type: </span><span className="lead fs-6 fw-semibold"> {data.instancetype} </span></th>
                  <th scope="row"><span className="lead fs-6 fw-bolder">Count: </span><span className="lead fs-6 fw-semibold"> {data.count} </span></th>
                  <th scope="row"><span className="lead fs-6 fw-bolder">Status: </span><span className="lead fs-6 fw-semibold"> {getNodegroupMessage(data)} </span></th>
                  {data.imiupgradeavailable && <th scope="row" className='border border-end-0 border-start-0'>
                      <div className='mt-s3'>
                        <Button
                          type="button"
                          variant='outline-primary'
                          className="mx-1 px-3"
                          disabled={isSREAdminUser}
                          onClick={() => {
                            openAuthModal()
                            onAuthSuccess(() => {
                              // showIMIModal(data)
                              console.log('Callback called with Data - ', data)
                            })
                          }}>
                            Upgrade IMI
                        </Button>
                      </div>
                  </th>
                  }
                </tr>
                <tr>
                  <td colSpan="6" className='px-0'>
                    {
                      data.nodegrouptype_name === 'ControlPlane' ? (
                        <div className='px-2'>
                          <table className="table table-row-dashed table-row-gray-500 gy-5 gs-5 mb-0">
                            <thead>
                              <tr className="fw-semibold ">
                                <th scope="col">Name</th>
                                <th scope="col">State</th>
                                <th scope="col">IMI</th>
                                <th scope="col">IP Address</th>
                                <th scope="col">Storage Status</th>
                                <th scope="col">Created Date</th>
                                <th scope="col">DNS Name</th>
                              </tr>
                            </thead>
                            <tbody>
                              {data.nodes.map((node, index) => (
                                <tr key={index}>
                                  <td>{node.name}</td>
                                  <td>{getActionItemLabelCluster(node.state, getNodeMessage(node))}</td>
                                  <td>{node.imi}</td>
                                  <td>{node.ipaddress}</td>
                                  <td>{node.wekaStorage?.status}</td>
                                  <td>{(new Date(node.createddate)).toLocaleString(undefined, { timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone, day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' })}</td>
                                  <td>{node.dnsname}</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      ) : (
                        <GridPaginationSearch
                          summary={
                            data.nodegrouptype_name === 'Worker' ? (
                              <div>
                                <span className="lead fs-6 fw-bolder">Summary: </span><br />
                                <span className="lead fs-6 fw-semibold">
                                  {`${data.nodegroupsummary?.activenodes}/${data.nodes.length} ${'Nodes Active'}`}
                                </span>
                              </div>
                            ) : null
                          }
                          showSearch={data.nodegrouptype_name === 'Worker'}
                          data={getGridItems(data.nodes)}
                          columns={nodeGroupColumns}
                          emptyGrid={{
                              title: 'No Managed Node Groups found',
                              subTitle: 'There are currently no node groups'
                          }}
                          emptyGridByFilter={{
                            title: 'No Managed Node Groups found',
                            subTitle: 'The applied filter criteria did not match any items'
                          }}
                          loading={false}
                        />
                      )
                    }
                  </td>
                </tr>
              </tbody>
          ))}
        </table>
      </div>
    )
  }

  // upgrade tap

  function getUpgradeK8sVersion() {
    if (upgradeK8sAvailable) {
      return (
        <div className='w-100'>
          <Modal className="md" show={successAPICall} onHide={closeSuccessAPIModal}>
            <Modal.Header closeButton>
              <Modal.Title className="text-break">Upgrade Confirmation</Modal.Title>
            </Modal.Header>
            <Modal.Body className="text-left small">
              <div className="w-100">
                <p>{successMessage}</p>
              </div>
            </Modal.Body>
            <Modal.Footer>
              <Button variant="outline-primary" onClick={closeSuccessAPIModal}>
                Close
              </Button>
              <Button variant="primary" onClick={upgradeClusterK8sVersion}>
                Accept
              </Button>
            </Modal.Footer>
          </Modal>
          <div className="m-0">
            <div className="row">
              <div className="col">
                <label className="mb-2">Select the K8s Version:</label>
                <select className="form-select form-select-sm w-50" onChange={(e) => { setSelectedVersion(e.target.value) }} >
                  {upgradeK8sVersions.map((option, index) => (
                    <option key={index} value={option}>{option}</option>
                  ))}
                </select>
              </div>
              <div className="col">
                <div className="section-component">
                  <h4 className="alert-heading">Release Notes:</h4>
                  <p>Please read the following release notes:
                    <a href={'https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG'} className="alert-link">{'https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG'}</a>
                    .</p>
                </div>
              </div>
            </div>
            <Button
              variant={'primary'}
              onClick={(e) => {
                setSucessAPICall(true)
                setSuccessMessage('Do you want to upgrade the cluster?')
              }} disabled>
              Upgrade
            </Button>
          </div>
        </div>)
    } else {
      return (
        <p>You are already running the latest version of Kubernetes</p>
      )
    }
  }

  // certificate tap

  function getCertificateDetails() {
    return (
      <div className='w-100'>
        <table className="table table-fixed table-row-dashed table-row-gray-500 gy-5 gs-5 mb-0">
          {certExpirations.map((data, index) => (
              <tbody key={index}>
                <tr>
                  <td colSpan="6">
                    <table className="table table-row-dashed table-row-gray-500 gy-5 gs-5 mb-0">
                      <thead>
                        <tr className="fw-semibold ">
                          <th scope="col">Certification Name</th>
                          <th scope="col">Expiration Date</th>
                        </tr>
                      </thead>
                      <tbody>
                          <tr key={index}>
                            <td>{data.cpname}</td>
                            <td>{data.certexpirydate}</td>
                          </tr>
                      </tbody>
                    </table>
                  </td>
                </tr>
              </tbody>
          ))}
        </table>
      </div>
    )
  }

  // SshKeys Tap

  function getSSHKeys() {
    return (
      <div className='w-100'>
        <table className="table table-fixed table-row-dashed table-row-gray-500 gy-5 gs-5 mb-0" style={{ minWidth: 'unset' }}>
          { sshKeys && clusterDetails && (
          <tbody>
            <tr className="fw-semibold">
              <td scope="col">SSH Private Key</td>
              <td><Button variant='outline-primary' onClick={(event) => { writeToFile(event, sshKeys.sshprivatekey, clusterDetails, 'private-key') }}>Download</Button></td>
              <td><Button variant='outline-primary' onClick={(event) => { copyToClipboard(sshKeys.sshprivatekey) }}>Copy</Button></td>
            </tr>
            <tr>
              <td style={{ overflowWrap: 'anywhere' }}><p>{sshKeys.sshprivatekey}</p></td>
              <td></td>
            </tr>
            <tr className="fw-semibold">
              <td scope="col">SSH Public Key</td>
              <td><Button variant='outline-primary' onClick={(event) => { writeToFile(event, sshKeys.sshprivatekey, clusterDetails, 'public-key') }}>Download</Button></td>
              <td><Button variant='outline-primary' onClick={(event) => { copyToClipboard(sshKeys.sshpublickey) }}>Copy</Button></td>
            </tr>
            <tr>
              <td style={{ overflowWrap: 'anywhere' }}><p>{sshKeys.sshpublickey}</p></td>
              <td></td>
            </tr>
          </tbody>
          )}
        </table>
      </div>
    )
  }

  // snapshot tap

  function getSnapshots() {
    return (
        <div className='w-100'>
          <table className="table table-fixed table-row-dashed table-row-gray-500 gy-5 gs-5 mb-0">
            {snapShotDetails.map((data, index) => (
                <tbody key={index}>
                  <tr>
                    <td colSpan="6">
                      <table className="table table-row-dashed table-row-gray-500 gy-5 gs-5 mb-0">
                        <thead>
                          <tr className="fw-semibold ">
                            <th scope="col">Created</th>
                            <th scope="col">File Name</th>
                            <th scope="col">Name</th>
                            <th scope="col">State</th>
                            <th scope="col">Type</th>
                          </tr>
                        </thead>
                        <tbody>
                            <tr key={index}>
                              <td>{data.created}</td>
                              <td>{data.filename}</td>
                              <td>{data.name}</td>
                              <td>{data.state}</td>
                              <td>{data.type}</td>
                            </tr>
                        </tbody>
                      </table>
                    </td>
                  </tr>
                </tbody>
            ))}
          </table>
        </div>
    )
  }

  // storage tap

  function getStorage() {
    if (!(Array.isArray(storages) && storages.length > 0)) {
      return null
    }

    const {
      storageprovider,
      state,
      size,
      reason,
      message
    } = storages[0] || {}

    return (
      <div>
        <div className='row my-2'>
          <div className='col-6'>
            <p className='my-2 fw-bold small mb-auto'><span className="lead fs-6 fw-bolder">Storage Provider:</span></p>
            <p className="lead fs-6">{storageprovider}</p>
          </div>
          <div className='col-6'>
            <p className='my-2 fw-bold small mb-auto'><span className="lead fs-6 fw-bolder">State:</span></p>
            <p><span className="lead fs-6"> {state}</span></p>
          </div>
          <div className='col-6'>
            <p className='my-2 fw-bold small mb-auto'><span className="lead fs-6 fw-bolder">Size:</span></p>
            <p><span className="lead fs-6"> {size}</span></p>
          </div>
          <div className='col-6'>
            <p className='my-2 fw-bold small mb-auto'><span className="lead fs-6 fw-bolder">Reason:</span></p>
            <p><span className="lead fs-6"> {reason}</span></p>
          </div>
          <div className='col-6'>
            <p className='my-2 fw-bold small mb-auto'><span className="lead fs-6 fw-bolder">Message:</span></p>
            <p><span className="lead fs-6">{message}</span></p>
          </div>
        </div>
      </div>
    )
  }

  // Security Tap

  function getSecurity() {
    const securityEmptyGrid = {
      title: 'No Security Details lists found',
      subTitle: 'There are currently no security details'
    }

    const securityColumns = [
      {
        columnName: 'Name',
        targetColumn: 'vipname'
      },
      {
        columnName: 'Type',
        targetColumn: 'viptype'
      },
      {
        columnName: 'State',
        targetColumn: 'state'
      },
      {
        columnName: 'Source IPs',
        targetColumn: 'sourceip'
      },
      {
        columnName: 'Protocol',
        targetColumn: 'protocol'
      },
      {
        columnName: 'Internal IP',
        targetColumn: 'destinationip'
      },
      {
        columnName: 'Internal Port',
        targetColumn: 'internalport'
      }
    ]

    const securityDetailsData = securityDetails.map((data) => {
      const sourceip = (Array.isArray(data.sourceip)) ? data.sourceip.join(', ') : ''
      const protocol = (Array.isArray(data.protocol)) ? data.protocol.join(', ') : ''

      return ({
        vipname: data.vipname,
        viptype: data.viptype,
        state: data.state,
        sourceip,
        protocol,
        destinationip: data.destinationip,
        internalport: data.internalport
      })
    })

    return (
      <div className='w-100'>
        <GridPagination
          data={securityDetailsData}
          columns={securityColumns}
          emptyGrid={securityEmptyGrid}
          loading={false}
        />
      </div>
    )
  }

  return (
    <>
      <SearchCloudAccountModal
        showModal={showSearchCloudAccount}
        selectedCloudAccount={selectedCloudAccount}
        cloudAccount={cloudAccount}
        cloudAccountError={cloudAccountError}
        showLoader={showLoader}
        setShowModal={setShowSearchModal}
        handleSearchInputChange={handleSearchInputChange}
        onClickSearchButton={handleSubmit}
      />
      <div className="section">
        <Button variant="link" className='p-s0' onClick={() => backToHome()}>
          ⟵ Back to Home
        </Button>
      </div>
      <ModalCreatePublicKey
        showModalActionConfirmation={showPublicKeyModal}
        closeCreatePublicKeyModal={onShowHidePublicKeyModal}
        afterPubliKeyCreate={afterPubliKeyCreate}
        isModal={true}
      />
      <div className="filter flex-wrap">
        <h2 className="h4">Cluster Details ({myreservations.length || 0})</h2>
        <ExportDataToCSV
          data={csvData.data}
          csvHeaders={csvData.headers}
          filename={'clusters.csv'}
        />
      </div>

      <div className="section">
          {
            myreservations.length > 0
            ? (
              <Filter
                data={myreservations}
                filters={filterIKS}
                onFiltersChanged={(data) => { setFilteredData(data) }}
              />
            )
            : null
          }
      </div>
      <div className='section d-flex flex-column'>
        <div className="filter flex-wrap p-0">
          <Button variant='primary' onClick={() => {
            setShowSearchModal(true)
            }}
          >
            Search Cloud Account
          </Button>
          <div className='d-inline-flex'>
            <SearchBox
              intc-id="searchClusters"
              value={filterText}
              onChange={setFilter}
              placeholder="Search Clusters..."
              aria-label="Type to search clusters..."
            />
          </div>
        </div>
        <GridPagination data={gridItems}
          columns={columns}
          emptyGrid={emptyGrid}
          loading={loading}
          />
      </div>
      {instanceDetails}

      <Modal
        className='modal-lg'
        backdrop="static"
        centered
        show={showAuthModal}
        onHide={() => {
          closeAuthModal()
        }}
      >
        <Modal.Header closeButton>
          <Modal.Title>{'Please Enter Password to Perform Action'}</Modal.Title>
        </Modal.Header>

        <Modal.Body className='m-3'>
          <div>
            <div className='row'>
              <div className='col-12 col-lg-6'>
                <CustomInput
                  key={'password'}
                  type={'password'}
                  fieldSize={'small'}
                  placeholder={'Enter Password'}
                  isRequired={true}
                  label={'Authorization Password'}
                  value={authPassword}
                  onChanged={(event) => { setAuthPassword(event.target.value) }}
                  isValid={true}
                />
              </div>
            </div>
          </div>
        </Modal.Body>

        <Modal.Footer>
          <Button
            variant={'primary'}
            className='btn'
            onClick={authFormFooterEvents.onPrimaryBtnClick}
          >
            Submit
          </Button>

          <Button variant={'link'} className='btn' onClick={authFormFooterEvents.onSecondaryBtnClick}>
            Cancel
          </Button>
        </Modal.Footer>
      </Modal>

      <Modal
        className='modal-lg'
        centered
        show={showUnAuthModal}
        onHide={() => {
          setShowUnAuthModal(false)
        }}
      >
        <Modal.Body className='m-3'>
          <br />
          <div className="text-center">
            <h5>Unauthorized</h5>
            <p>
              You are not allowed to perform this action. Please contact Admin.
            </p>
          </div>
        </Modal.Body>

        <Modal.Footer>
          <Button
            variant={'primary'}
            className='btn'
            onClick={() => {
              setShowUnAuthModal(false)
            }}
          >
            Go Back
          </Button>
        </Modal.Footer>
      </Modal>
    </>
  )
}

export default ClusterMyReservations
