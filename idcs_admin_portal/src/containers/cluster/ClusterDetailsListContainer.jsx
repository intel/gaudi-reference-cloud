import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import ClusterMyReservations from '../../components/cluster/clusterMyReservations/ClusterMyReservations'
import IKSService from '../../services/IKSService'
import { UpdateFormHelper, isValidForm } from '../../utility/updateFormHelper/UpdateFormHelper'
import useClusterStore from '../../store/clusterStore/ClusterStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import CloudAccountService from '../../services/CloudAccountService'
import Popover from 'react-bootstrap/Popover'
import OverlayTrigger from 'react-bootstrap/OverlayTrigger'
import {
  BiMinusCircle,
  BiCheckCircle,
  BiXCircle,
  BiPlayCircle
} from 'react-icons/bi'
import useToastStore from '../../store/toastStore/ToastStore'
import useUserStore from '../../store/userStore/UserStore'

const getCustomMessage = (messageType) => {
  let message = null

  switch (messageType) {
    case 'kubeletArgs':
      message = <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'} >

            </div>
      break
    case 'kubeletApiArgs':
      message = <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'} >

            </div>
      break
    case 'kubeletControllerArgs':
      message = <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'} >

            </div>
      break
    case 'kubeletSchedulerArgs':
      message = <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'} >

            </div>
      break
    case 'kubeletProxyArgs':
      message = <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'} >

            </div>
      break
    default:
      break
  }

  return message
}

const ClusterDetailsListContainer = (props) => {
  const isSREAdminUser = useUserStore(state => state.isSREAdminUser)

  const emptyGrid = {
    title: 'No clusters found',
    subTitle: 'Your account currently has no clusters'
  }

  const emptyGridByFilter = {
    title: 'No clusters found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  // *****
  // cluster table structure
  // *****

  const columns = [
    {
      columnName: 'Cluster Id',
      targetColumn: 'cluster-id',
      columnConfig: {
        behaviorType: 'hyperLink',
        behaviorFunction: 'setDetails'
      }
    },
    {
      columnName: 'Cluster Name',
      targetColumn: 'cluster-name'
    },
    {
      columnName: 'Cloud Account',
      targetColumn: 'account',
      columnConfig: {
        behaviorType: 'hyperLink',
        behaviorFunction: 'getCloudAccountDetailsById'
      }
    },
    {
      columnName: 'K8s Version',
      targetColumn: 'k8sversion'
    },
    {
      columnName: 'Status',
      targetColumn: 'clusterStatus'
    },
    {
      columnName: 'Created at',
      targetColumn: 'creationTimestamp'
    }
  ]

    // *****
  // CSV file structure
  // *****

   const csvHeaders = [
    { label: 'Cluster Id', key: 'cluster-id' },
    { label: 'Cluster Name', key: 'cluster-name' },
    { label: 'Cloud Account', key: 'account' },
    { label: 'K8s Version', key: 'k8sversion' },
    { label: 'Status', key: 'clusterStatus' },
    { label: 'Created at', key: 'creationTimestamp' }
  ]

  // *****
  // taps structure
  // *****

  const taps = [
    {
      label: 'Overview'
    },
    {
      label: 'Compute'
    },
    {
      label: 'Load Balancers'
    },
    {
      label: 'SSH Keys'
    }
  ]

  const tapDetails = [
    {
      tapTitle: 'Cluster Information',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'Manage Node Groups',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'Manage Load Balancers',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'SSH Keys Information',
      tapConfig: { type: 'custom' },
      customContent: null
    }
  ]

  const filterIKS = [
    {
      label: 'Status',
      key: 'clusterstatus'
    },
    {
      label: 'K8s',
      key: 'k8sversion'
    }
  ]

  if (!isSREAdminUser()) {
    columns.push({
      columnName: 'IMI Upgrade Available',
      targetColumn: 'cpupgradeavailable'
    },
    {
      columnName: 'K8s Upgrade Available',
      targetColumn: 'k8supgradeavailable'
    })

    csvHeaders.push({ label: 'IMI Upgrade Available', key: 'cpupgradeavailable' },
    { label: 'K8s Upgrade Available', key: 'k8supgradeavailable' })

    taps.push({
      label: 'Upgrade K8s Version'
    },
    {
      label: 'Certificates Expiration'
    },
    {
      label: 'Snapshots'
    },
    {
      label: 'Storage'
    },
    {
      label: 'Security'
    })

    tapDetails.push({
      tapTitle: 'Upgrade K8s Version',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'Certificates Expiration Details',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'Snapshot Details',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'Storage Details',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'Security Details',
      tapConfig: { type: 'custom' },
      customContent: null
    })

    filterIKS.push({
      label: 'IMI Upgrade Available',
      key: 'cpupgradeavailable'
    },
    {
      label: 'K8s Upgrade Available',
      key: 'k8supgradeavailable'
    })
  }

  // *****
  // nodegroup form structure
  // *****

  const initialFormCreateNodegroup = {
    nodegroupName: {
      sectionGroup: 'tags',
      type: 'text', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Name',
      placeholder: 'Node Group Name',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 63,
      validationRules: {
        isRequired: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: getCustomMessage('kubeletArgs')
    },
    nodegroupDescription: {
      sectionGroup: 'tags',
      type: 'text', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Description',
      placeholder: 'Node Group Description',
      value: '', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 63,
      validationRules: {
        isRequired: false
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: getCustomMessage('kubeletArgs')
    },
    nodegroupCount: {
      sectionGroup: 'tags',
      type: 'integer', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Count',
      placeholder: 'Node Group Count',
      value: '1', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 3,
      validationRules: {
        isRequired: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: getCustomMessage('kubeletArgs')
    },
    nodegroupInstanceType: {
      sectionGroup: 'tags',
      type: 'select', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Node Group Instance Type',
      placeholder: 'Please select instance type',
      value: '',
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      validationRules: {
        isRequired: true
      },
      options: [
      ],
      validationMessage: '',
      helperMessage: ''
    },
    nodegroupVnets: {
      sectionGroup: 'tags',
      type: 'select', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Node Group vnet',
      placeholder: 'Please select vnet',
      value: '',
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      validationRules: {
        isRequired: true
      },
      options: [
      ],
      validationMessage: '',
      helperMessage: ''
    },
    nodegroupPublicKeys: {
      sectionGroup: 'tags',
      type: 'select', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Select keys ',
      placeholder: 'Please select',
      value: [], // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      isMultiple: true,
      validationRules: {
        isRequired: true
      },
      options: [], // Only required for select inputs, contains the seletable options
      validationMessage: '', // Errror message to display to the user
      extraButton: {
        status: true,
        label: '+ Upload Key',
        buttonFunction: () => onShowHidePublicKeyModal(true)
      },
      refreshButton: {
        status: true,
        label: 'Refresh Keys'
      },
      emptyOptionsMessage: 'No keys found. Please create a key to continue.'
    },
    nodegroupTags: {
      sectionGroup: 'tags',
      type: 'dictionary',
      options: [],
      fieldSize: 'small',
      label: 'Tags',
      validationRules: {
        isRequired: false
      },
      isValid: true
    }
  }

  // *****
  // loadbalancer form structure
  // *****

  const initialFormCreateLoadbalancer = {
    loadbalancerName: {
      sectionGroup: 'lb',
      type: 'text', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Name',
      placeholder: 'Load Balancer Name',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 63,
      validationRules: {
        isRequired: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: getCustomMessage('kubeletArgs')
    },
    loadbalancerDescription: {
      sectionGroup: 'lb',
      type: 'text', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Description',
      placeholder: 'Load Balancer Description',
      value: '', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 63,
      validationRules: {
        isRequired: false
      },
      validationMessage: '' // Errror message to display to the user
    },
    loadbalancerPort: {
      sectionGroup: 'lb',
      type: 'integer', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Port',
      placeholder: 'Load Balancer Port',
      value: '80', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 3,
      validationRules: {
        isRequired: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: ''
    },
    loadbalancerType: {
      sectionGroup: 'lb',
      type: 'select', // options = 'text ,'textArea'
      fieldSize: 'small', // options = 'small', 'medium', 'large'
      label: 'Load Balancer Type',
      placeholder: 'Please select load balancer type',
      value: '',
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      validationRules: {
        isRequired: true
      },
      options: [{ key: 1, name: 'private', value: 'private' }, { key: 2, name: 'public', value: 'public' }],
      validationMessage: ''
    }
  }

  const loadBalancerBottom = [{
    buttonLabel: 'Create',
    buttonVariant: 'primary'
  },
  {
    buttonLabel: 'Cancel',
    buttonVariant: 'link',
    buttonFunction: () => closeLBModal()
  }]

  // *****
  // metadata tab structure
  // *****

  const initialFormMetadataTab = {
    clusterTags: {
      sectionGroup: 'tags',
      type: 'dictionary',
      options: [],
      fieldSize: 'medium',
      label: '',
      validationRules: {
        isRequired: false
      }
    },
    clusterAnnotations: {
      sectionGroup: 'annotations',
      type: 'dictionary',
      options: [],
      fieldSize: 'medium',
      label: '',
      validationRules: {
        isRequired: false
      }
    }
  }
  const navigationMetadataBottom = [{
    buttonLabel: 'Save',
    buttonVariant: 'primary'
  }]

  // *****
  // variables
  // *****

  // filtered forms

  const formElementsNodegroupCreate = []
  const formElementsTagsTab = []
  const formElementsAnnotationsTab = []
  const formElementsLoadBalancerCreate = []
  const formElementsUpgradeCluster = []

  // cloudAccounts
  const [cloudAccount, setCloudAccount] = useState(null)
  const [showLoader, setShowLoader] = useState('')
  const [selectedCloudAccount, setSelectedCloudAccount] = useState(null)
  const [selectedClusterCloudAccount, setSelectedClusterCloudAccount] = useState(null)
  const [cloudAccountError, setCloudAccountError] = useState(false)

  // alerts
  const [successMessage, setSuccessMessage] = useState('')
  const [successAPICall, setSucessAPICall] = useState(false)
  const clusters = useClusterStore((state) => state.clustersData)
  const cluster = useClusterStore((state) => state.clusterData)
  const sshKeys = useClusterStore((state) => state.sshKeys)
  const loadBalancers = useClusterStore((state) => state.loadBalancerDetails)
  const securityDetails = useClusterStore((state) => state.securityDetails)
  const loading = useClusterStore((state) => state.loading)
  const clusterNodegroups = useClusterStore((state) => state.clusterNodegroups)
  const certExpiration = useClusterStore((state) => state.certExpiration)
  const snapshotList = useClusterStore((state) => state.snapshot)
  const storagesList = useClusterStore((state) => state.storages)
  const refreshIlb = useClusterStore((state) => state.refreshIlb)
  const setClusters = useClusterStore((state) => state.setClustersData)
  const setCluster = useClusterStore((state) => state.setClusterData)
  const setSSHKeys = useClusterStore((state) => state.setSSHKeys)
  const setLoadbalancerDetails = useClusterStore((state) => state.setLoadBalancerDetails)
  const setSecurityDetails = useClusterStore((state) => state.setSecurityDetails)
  const resetSecurityDetails = useClusterStore((state) => state.resetSecurityDetails)
  const resetLoadBalancerDetails = useClusterStore((state) => state.resetLoadBalancerDetails)
  const setRefreshRate = useClusterStore((state) => state.setRefreshRate)
  const setIlbRefresh = useClusterStore((state) => state.setIlbRefresh)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // cluster information
  const [myreservations, setMyreservations] = useState([])
  const [csvGridInfo, setCsvGridInfo] = useState([])
  const [clusterDetails, setClusterDetails] = useState(null)
  const [clusterNodegroupsDetails, setClusterNodegroupsDetails] = useState([])
  const [certExpirationDetails, setCertExpirationDetails] = useState([])
  const [upgradeK8sVersions, setUpgradeK8sVersions] = useState([])
  const [upgradeK8sAvailable, setUpgradeK8sAvailable] = useState(false)
  const [filterText, setFilterText] = useState('')
  const [snapshots, setsnapshots] = useState([])
  const [storages, setStorages] = useState([])
  const [deleteLbName, setDeleteLbName] = useState('')
  const [deleteLbId, setDeleteLbId] = useState('')
  const [nodeGroupID, setNodegroupID] = useState('')
  // modals
  const [showPublicKeyModal, setShowPublicKeyModal] = useState(false)
  const [showCreateLBModal, setShowCreateLBModal] = useState(false)
  const [showIMIUpgradeModal, setShowIMIUpgradeModal] = useState(false)
  const [showDeleteNodegroupModal, setShowDeleteNodegroupModal] = useState(false)
  const [showPutCluster, setShowPutCluster] = useState(false)
  const [showSearchCloudAccount, setShowSearchCloudAccount] = useState(false)
  // taps
  const [activeTap, setActiveTap] = useState(0)
  // form
  const [isValid, setIsValid] = useState(false)
  const [loadbalancerForm, setLoadbalancerForm] = useState(initialFormCreateLoadbalancer)
  const [nodegroupForm, setNodegroupForm] = useState(initialFormCreateNodegroup)
  const [formMetadataTab, setFormMetadataTab] = useState(initialFormMetadataTab)

  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)

  // Auth Modal States
  const [showAuthModal, setShowAuthModal] = useState(false)
  const [showUnAuthModal, setShowUnAuthModal] = useState(false)
  const [authPassword, setAuthPassword] = useState('')
  const [authSuccessCallback, setAuthSuccessCallback] = useState(() => {})

  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()

  // *****
  // use effect
  // *****

  const fetchCluster = async (uuid) => {
    try {
      await setCluster(uuid)
      if (clusterNodegroups) {
        setClusterNodegroupsDetails(clusterNodegroups)
      }
      if (certExpiration) {
        setCertExpirationDetails(certExpiration)
      }
      if (snapshotList) {
        setsnapshots(snapshotList)
      }
      if (storagesList) {
        setStorages(storagesList)
      }
    } catch (error) {
      throwError(error)
    }
  }
  const fetchSSHKeys = async (uuid) => {
    try {
      await setSSHKeys(uuid)
    } catch (error) {
      throwError(error)
    }
  }
  const fetchLoadBalancers = async (uuid) => {
    try {
      await setLoadbalancerDetails(uuid)
    } catch (error) {
      resetLoadBalancerDetails()
    }
  }
  const fetchSecurity = async (uuid) => {
    try {
      await setSecurityDetails(uuid)
    } catch (error) {
      resetSecurityDetails()
    }
  }

  const resfreshLoadBalancers = () => {
    if (clusterDetails) {
      fetchLoadBalancers(clusterDetails.uuid)
    }
  }

  useEffect(() => {
    if (clusterDetails) {
      fetchCluster(clusterDetails.uuid)
      fetchSSHKeys(clusterDetails.uuid)
      getCloudAccountDetailsById(clusterDetails.account, true)
    }
    if (clusterDetails) {
      fetchLoadBalancers(clusterDetails.uuid)
      if (!isSREAdminUser()) {
        fetchSecurity(clusterDetails.uuid)
        setUpgradeK8sVersions(clusterDetails.k8supgradeversions)
        setUpgradeK8sAvailable(clusterDetails.k8supgradeavailable)
      }
    }
    setGridInfo()
  }, [clusterDetails])

  useEffect(() => {
    if (clusterDetails && cluster && clusterNodegroups) {
      fetchCluster(clusterDetails.uuid)
      setIlbRefresh()
      if (refreshIlb) {
        fetchLoadBalancers(clusterDetails.uuid)
      }
    }
    setGridInfo()
  }, [clusters])

  const fetchClusters = async (isBackground) => {
    try {
      await setClusters(isBackground)
    } catch (error) {
      throwError(error)
    }
  }
  useEffect(() => {
    fetchClusters()
    setRefreshRate(true)
    return () => {
      setRefreshRate(false)
    }
  }, [])

  const debounceClusterRefresh = () => {
    setTimeout(() => {
      fetchClusters(true)
    }, 1000)
    setTimeout(() => {
      fetchClusters(true)
    }, 3000)
    setTimeout(() => {
      fetchClusters(true)
    }, 5000)
  }
  // *****
  // functions
  // *****

  const getActionItemLabelCluster = (text, statusStep = null, option = null) => {
    let message = null
    switch (text) {
      case 'Start':
        message = (
          <>
            {' '}
            <BiPlayCircle className="mb-1" /> {text}{' '}
          </>
        )
        break
      case 'Deleting':
        message = (
          <>
            {' '}
            <BiMinusCircle className="mb-1" /> {text}{' '}
          </>
        )
        break
      case 'Delete':
        message = (
          <>
            {' '}
            <BiXCircle className="mb-1" /> {text}{' '}
          </>
        )
        break
      case 'Pending':
        message = (
          <div className="d-flex flex-row bd-highlight">
          {option && (
            <>
            <div>
              {'State: '}
            </div>
            <div className="p-0 bd-highlight" intc-id='Tooltip Message'>
              <OverlayTrigger trigger="focus" placement="right" overlay={popover(statusStep)}>
                <button className='btn btn-link status-text'>
                  {text}
                </button>
              </OverlayTrigger>
              <div className="spinner-border spinner-border-sm" role="status">
                <span className="visually-hidden">Loading...</span>
              </div>
            </div>
            </>
          )}
          {!option && (
            <>
            <div className="p-0 bd-highlight">
              <div className="spinner-border spinner-border-sm" role="status">
                <span className="visually-hidden">Loading...</span>
              </div>
            </div>
            <div className="p-0 bd-highlight" intc-id='Tooltip Message'>
              <OverlayTrigger trigger="focus" placement="right" overlay={popover(statusStep)}>
                <button className='btn btn-link status-text'>
                  {text}
                </button>
              </OverlayTrigger>
            </div>
            </>
          )}
          </div>
        )
        break
      case 'Updating':
        message = (
          <div className="d-flex flex-row bd-highlight">
          {option && (
            <>
            <div>
              {'State: '}
            </div>
            <div className="p-0 bd-highlight" intc-id='Tooltip Message'>
              <OverlayTrigger trigger="focus" placement="right" overlay={popover(statusStep)}>
                <button className='btn btn-link status-text'>
                  {text}
                </button>
              </OverlayTrigger>
              <div className="spinner-border spinner-border-sm" role="status">
                <span className="visually-hidden">Loading...</span>
              </div>
            </div>
            </>
          )}
          {!option && (
            <>
            <div className="p-0 bd-highlight">
              <div className="spinner-border spinner-border-sm" role="status">
                <span className="visually-hidden">Loading...</span>
              </div>
            </div>
            <div className="p-0 bd-highlight" intc-id='Tooltip Message'>
              <OverlayTrigger trigger="focus" placement="right" overlay={popover(statusStep)}>
                <button className='btn btn-link status-text'>
                  {text}
                </button>
              </OverlayTrigger>
            </div>
            </>
          )}
          </div>
        )
        break
      case 'DeletePending':
        message = (
          <>
            {' '}
            <BiMinusCircle className="mb-1" /> {text}{' '}
          </>
        )
        break
      case 'Active':
        message = (
          <>
            {' '}
            <BiCheckCircle className="mb-1" /> {text}{' '}
          </>
        )
        break
      default:
        message = <> {text} </>
        break
    }

    return message
  }

  const popover = (message) => {
    return <Popover id="popover-basic">
      <Popover.Header as="h3">Provisioning status</Popover.Header>
      <Popover.Body>
        {message}
      </Popover.Body>
    </Popover>
  }

  function onShowHidePublicKeyModal(status = false) {
    setShowPublicKeyModal(status)
  }

  function showPutModal () {
    setShowPutCluster(true)
  }

  function closePutModal () {
    setShowPutCluster(false)
  }

  function setFilter(event, clear) {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
    setClusterDetails(null)
  }

  function onChangeDropdownMultiple(values) {
    const updatedNodegroupForm = {
      ...nodegroupForm
    }

    const updatedFormElement = { ...updatedNodegroupForm.nodegroupPublicKeys }

    updatedFormElement.value = values

    updatedFormElement.isTouched = false
    updatedFormElement.isValid = false

    if (values.length > 0) {
      updatedFormElement.isTouched = true
      updatedFormElement.isValid = true
    }

    updatedNodegroupForm.nodegroupPublicKeys = updatedFormElement

    const isValid = isValidForm(updatedNodegroupForm)
    setNodegroupForm(updatedNodegroupForm)
    setIsValid(isValid)
  }

  function onClickActionTag(index, actionType) {
    const updatedOptions = nodegroupForm.nodegroupTags.options

    if (actionType.toUpperCase() === 'ADD') {
      updatedOptions.push({ key: '', value: '' })
    } else if (actionType.toUpperCase() === 'DELETE') {
      updatedOptions.splice(index, 1)
    }

    const updatedForm = {
      ...nodegroupForm,
      nodegroupTags: {
        ...nodegroupForm.nodegroupTags,
        options: updatedOptions
      }
    }

    setNodegroupForm(updatedForm)
  }

  function parseStringToDictionary(inputStringList) {
    try {
      const list = inputStringList.split(' ')
      const dictionary = []
      for (let i = 0; i < list.length; i++) {
        if (list[i] !== '') {
          const line = list[i]
          const [key, value] = line.split(':')
          dictionary.push({ key: key.trim(), value: value.trim() })
        }
      }
      return dictionary
    } catch { }
    return inputStringList
  }

  function onChangeTagValue(event, field, index) {
    const { value } = event.target
    const updatedOptions = [...nodegroupForm.nodegroupTags.options]

    let updatedForm = {
      ...nodegroupForm
    }

    const parseValue = parseStringToDictionary(value)

    if (typeof parseValue === 'object') {
      updatedForm = {
        ...nodegroupForm,
        nodegroupTags: {
          ...nodegroupForm.nodegroupTags,
          options: parseValue
        }
      }
    } else {
      const tagItem = updatedOptions[index]
      tagItem[field] = value.replace(/\s/g, '')

      updatedForm = {
        ...nodegroupForm,
        nodegroupTags: {
          ...nodegroupForm.nodegroupTags,
          options: updatedOptions
        }
      }
    }
    setNodegroupForm(updatedForm)
  }

  function handleClickActionTag(index, actionType) {
    const updatedOptions = formMetadataTab.clusterTags.options

    if (actionType.toUpperCase() === 'ADD') {
      updatedOptions.push({ key: '', value: '' })
    } else if (actionType.toUpperCase() === 'DELETE') {
      updatedOptions.splice(index, 1)
    }

    const updatedForm = {
      ...formMetadataTab,
      clusterTags: {
        ...formMetadataTab.clusterTags,
        options: updatedOptions
      }
    }

    setFormMetadataTab(updatedForm)
  }

  function handleClickActionAnnotations(index, actionType) {
    const updatedOptions = formMetadataTab.clusterAnnotations.options

    if (actionType.toUpperCase() === 'ADD') {
      updatedOptions.push({ key: '', value: '' })
    } else if (actionType.toUpperCase() === 'DELETE') {
      updatedOptions.splice(index, 1)
    }

    const updatedForm = {
      ...formMetadataTab,
      clusterAnnotations: {
        ...formMetadataTab.clusterAnnotations,
        options: updatedOptions
      }
    }

    setFormMetadataTab(updatedForm)
  }

  function showIMIModal(data) {
    setNodegroupID(data.name)
    setShowIMIUpgradeModal(true)
  }

  function closeIMIModal() {
    setShowIMIUpgradeModal(false)
  }

  function showDeleteModal() {
    setShowDeleteNodegroupModal(true)
  }

  function closeDeleteModal() {
    setShowDeleteNodegroupModal(false)
  }

  function closeSuccessAPIModal() {
    setSucessAPICall(false)
  }
  function showLBModal() {
    setShowCreateLBModal(true)
  }

  function closeLBModal() {
    setShowCreateLBModal(false)
  }

  function handleTagChange(event, field, index) {
    const { value } = event.target
    const updatedOptions = [...formMetadataTab.clusterTags.options]
    let updatedForm = {
      ...formMetadataTab
    }

    const parseValue = parseStringToDictionary(value)

    if (typeof parseValue === 'object') {
      updatedForm = {
        ...formMetadataTab,
        clusterTags: {
          ...formMetadataTab.clusterTags,
          options: parseValue
        }
      }
    } else {
      const tagItem = updatedOptions[index]
      tagItem[field] = value.replace(/\s/g, '')

      updatedForm = {
        ...formMetadataTab,
        clusterTags: {
          ...formMetadataTab.clusterTags,
          options: updatedOptions
        }
      }
    }

    setFormMetadataTab(updatedForm)
  }

  function handleAnnotationsChange(event, field, index) {
    const { value } = event.target
    const updatedOptions = formMetadataTab.clusterAnnotations.options
    let updatedForm = {
      ...formMetadataTab
    }

    const parseValue = parseStringToDictionary(value)

    if (typeof parseValue === 'object') {
      updatedForm = {
        ...formMetadataTab,
        clusterAnnotations: {
          ...formMetadataTab.clusterAnnotations,
          options: parseValue
        }
      }
    } else {
      const tagItem = updatedOptions[index]
      tagItem[field] = value.replace(/\s/g, '')

      updatedForm = {
        ...formMetadataTab,
        clusterAnnotations: {
          ...formMetadataTab.clusterAnnotations,
          options: updatedOptions
        }
      }
    }

    setFormMetadataTab(updatedForm)
  }
  async function upgradeNodeGroup() {
    try {
      await IKSService.upgradeNodeGroup(null, nodeGroupID, clusterDetails.uuid)
      debounceClusterRefresh()
      setShowIMIUpgradeModal(false)
      showSuccess('IMI upgraded succesfully.')
    } catch (error) {
      setShowIMIUpgradeModal(false)
      throwError(error)
    }
  }
  function setGridInfo() {
    const gridInfo = []
    const gridInfoCSV = []

    for (const item in clusters) {
      const clusterItem = { ...clusters[item] }
      clusterItem.value = clusterItem.state

      if (
        (props.page === 'Cluster' && clusterItem.clustertype === 'supercompute') ||
        (props.page === 'Super Compute Cluster' && clusterItem.clustertype !== 'supercompute')
      ) {
        continue
      }
      const gridInfoData = {
        'cluster-id': (clusterItem.state !== 'Deleting' && clusterItem.state !== 'DeletePending' && clusterItem.state !== 'Deleted')
          ? {
              showField: true,
              type: 'HyperLink',
              value: clusterItem.uuid,
              function: () => { setDetails(clusterItem) }
            }
          : clusterItem.uuid,
        'cluster-name': clusterItem.name,
        account: {
          showField: true,
          type: 'HyperLink',
          value: clusterItem.account,
          function: () => { getCloudAccountDetailsById(clusterItem.account, false) }
        },
        k8sversion: clusterItem.k8sversion,
        clusterstatus: clusterItem.state,
        creationTimestamp: {
          showField: true,
          type: 'date',
          toLocalTime: true,
          value: clusterItem.createddate,
          format: 'MM/DD/YYYY h:mm a'
        }
      }

      const gridInfoCSVData = {
        'cluster-id': clusterItem.uuid.toString(),
        'cluster-name': clusterItem.name.toString(),
        account: clusterItem.account.toString(),
        k8sversion: clusterItem.k8sversion.toString(),
        clusterStatus: clusterItem.state.toString(),
        creationTimestamp: clusterItem.createddate.toString()
      }
      if (!isSREAdminUser()) {
        gridInfoData.cpupgradeavailable = clusterItem.cpupgradeversions.toString()
        gridInfoData.k8supgradeavailable = clusterItem.k8supgradeversions.toString()

        gridInfoCSVData.cpupgradeavailable = clusterItem.cpupgradeversions.toString()
        gridInfoCSVData.k8supgradeavailable = clusterItem.k8supgradeversions.toString()
      }

      gridInfo.push(gridInfoData)
      gridInfoCSV.push(gridInfoCSVData)
    }

    setCsvGridInfo(gridInfoCSV)
    setMyreservations(gridInfo)
  }

  function onChangeInput(event, formInputName) {
    const formCopy = {
      ...nodegroupForm
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, formCopy)
    const isValidoptions = isValidForm(updatedForm)

    setNodegroupForm(updatedForm)

    setIsValid(isValidoptions)
  }
  function onChangeLoadBalancerInput(event, formInputName) {
    const formCopy = {
      ...loadbalancerForm
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, formCopy)
    const isValidoptions = isValidForm(updatedForm)

    setLoadbalancerForm(updatedForm)

    setIsValid(isValidoptions)
  }

  async function setDetails(clusterDetail = null) {
    try {
      setClusterDetails(null)
      if (clusterDetail) {
        await setCluster(clusterDetail.uuid)
        setClusterDetails(clusterDetail)
      }
    } catch (error) {
      throwError(error)
    }
  }

  const openAuthModal = () => {
    setShowAuthModal(true)
  }

  const closeAuthModal = () => {
    // Close Auth Modal
    setShowAuthModal(false)
    // Reset Auth Password
    setAuthPassword('')
  }

  const onAuthSuccess = (callback) => {
    setAuthSuccessCallback(() => callback)
  }

  const setUserAuthorization = () => {
    const payload = {
      iksadminkey: btoa(authPassword)
    }

    IKSService.authenticateIMIS(payload)
      .then(({ data }) => {
        if (data?.isAuthenticatedUser) {
          authSuccessCallback()
        } else {
          setShowUnAuthModal(true)
        }
        closeAuthModal()
      })
      .catch((error) => {
        closeAuthModal()
        throwError(error.response)
      })
  }

  const authFormFooterEvents = {
    onPrimaryBtnClick: () => {
      setUserAuthorization()
    },
    onSecondaryBtnClick: () => {
      closeAuthModal()
    }
  }

  function backToHome() {
    navigate('/')
  }
  const onClearSearchInput = () => {
    // Making cloud account null.
    setCloudAccount('')
    // Clearing / Making empty instance group state.
    setSelectedCloudAccount('')
    // Clearing Error Message
    setCloudAccountError(false)
  }
  const handleSearchInputChange = (e) => {
    // Update the state with the numeric value
    setCloudAccount(e.target.value)
    // Clearing Error Message
    setCloudAccountError(false)
  }
  // Function to handle form submission
  const handleSubmit = async (e) => {
    setCloudAccountError('')
    setSelectedCloudAccount('')
    if (cloudAccount !== '') {
      setShowLoader({ isShow: true, message: 'Searching for Details...' })
      try {
        let data = null
        // Calling the specific service based on ID to fetch the cloud account details.
        if (cloudAccount.includes('@') || /[a-zA-Z]/.test(cloudAccount)) {
          data = await CloudAccountService.getCloudAccountDetailsByName(cloudAccount)
        } else {
          data = await CloudAccountService.getCloudAccountDetailsById(cloudAccount)
        }
        setCloudAccount(cloudAccount)
        // Picks the selected searched data.
        setSelectedCloudAccount(data?.data)
        // Making error state with false.
        setCloudAccountError(false)
      } catch (e) {
        const code = e?.response?.data?.code
        const errorMsg = e?.response?.data?.message
        const message = code && [3, 5].includes(code) ? errorMsg.charAt(0).toUpperCase() + errorMsg.slice(1) : 'Cloud Account ID is not found'
        // Assigning the error message.
        setCloudAccountError(message)
        // Clearing selected search data.
        setSelectedCloudAccount(false)
      }
    } else {
    // Assigning the error message.
      setCloudAccountError('Cloud Account is required')
    }
    setShowLoader('')
  }

  const getCloudAccountDetailsById = async (cloudAccount, clusterClick) => {
    if (cloudAccount !== '') {
      setShowSearchModal(!clusterClick)
      setShowLoader({ isShow: true, message: 'Searching for Details...' })
      try {
        const data = await CloudAccountService.getCloudAccountDetailsById(cloudAccount)

        // Picks the selected searched data.
        if (clusterClick) {
          setSelectedClusterCloudAccount(data?.data)
        } else {
          setSelectedCloudAccount(data?.data)
        }
        // Making error state with false.
        setCloudAccountError(false)
      } catch (e) {
        const code = e.response.data?.code
        const errorMsg = e.response.data?.message
        const message = code && [3, 5].includes(code) ? errorMsg.charAt(0).toUpperCase() + errorMsg.slice(1) : 'Cloud Account ID is not found'
        // Assigning the error message.
        setCloudAccountError(message)
        // Clearing selected search data.
        setSelectedCloudAccount(false)
      }
      setShowLoader('')
    }
  }

  function setShowSearchModal(status) {
    if (!status) onClearSearchInput()
    setShowSearchCloudAccount(status)
  }

  async function onSubmit() {
    try {
      await IKSService.putClusterMetadata(formMetadataTab, clusterDetails)
      closePutModal()
      showSuccess('Metadata saved')
    } catch (error) {
      throwError(error)
    }
  }
  async function addLoadBalancer() {
    try {
      closeLBModal()
      showError('We are still working on this functionality.')
    } catch (error) {
      closeLBModal()
      throwError(error)
    }
  }

  async function deleteLoadBalancer() {
    try {
      closeDeleteModal()
      showError('We are still working on this functionality.')
    } catch (error) {
      closeDeleteModal()
      throwError(error)
    }
  }
  async function upgradeClusterK8sVersion() {
    try {
      closeSuccessAPIModal()
      showError('We are still working on this functionality.')
    } catch (error) {
      throwError(error)
    }
  }

  for (const key in nodegroupForm) {
    const formItem = {
      ...nodegroupForm[key]
    }

    if (formItem.sectionGroup === 'tags') {
      formElementsNodegroupCreate.push({
        id: key,
        configInput: formItem
      })
    }
  }
  for (const key in loadbalancerForm) {
    const formItem = {
      ...loadbalancerForm[key]
    }
    formElementsLoadBalancerCreate.push({
      id: key,
      configInput: formItem
    })
  }

  for (const key in formMetadataTab) {
    const formItem = {
      ...formMetadataTab[key]
    }

    if (formItem.sectionGroup === 'tags') {
      formElementsTagsTab.push({
        id: key,
        configInput: formItem
      })
    }
    if (formItem.sectionGroup === 'annotations') {
      formElementsAnnotationsTab.push({
        id: key,
        configInput: formItem
      })
    }
  }

  return (
    <ClusterMyReservations
        columns={columns}
        myreservations={myreservations}
        clusterDetails={clusterDetails}
        onChangeDropdownMultiple={onChangeDropdownMultiple}
        taps={taps}
        tapDetails={tapDetails}
        activeTap={activeTap}
        snapshots = {snapshots}
        storages={storages}
        setActiveTap={setActiveTap}
        showPublicKeyModal={showPublicKeyModal}
        onShowHidePublicKeyModal={onShowHidePublicKeyModal}
        setDetails={setDetails}
        nodegroupsInfo={clusterNodegroupsDetails}
        certExpirationInfo={certExpirationDetails}
        upgradeK8sAvailable={upgradeK8sAvailable}
        upgradeK8sVersions={upgradeK8sVersions}
        onChangeInput={onChangeInput}
        isValidForm={isValid}
        formElementsTagsTab={formElementsTagsTab}
        formElementsAnnotationsTab={formElementsAnnotationsTab}
        formElementsTags={formElementsNodegroupCreate}
        formElementsLoadBalancerCreate={formElementsLoadBalancerCreate}
        onChangeTagValue={onChangeTagValue}
        onClickActionTag={onClickActionTag}
        onChangeTagValueMetadata={handleTagChange}
        onClickActionTagMetadata={handleClickActionTag}
        onChangeAnnotationsValue={handleAnnotationsChange}
        onClickActionAnnotations={handleClickActionAnnotations}
        navigationMetadataBottom={navigationMetadataBottom}
        showPutModal={showPutModal}
        showPutCluster={showPutCluster}
        closePutModal={closePutModal}
        onSubmit={onSubmit}
        loadBalancers={loadBalancers}
        securityDetails={securityDetails}
        resfreshLoadBalancers={resfreshLoadBalancers}
        emptyGrid={emptyGridObject}
        filterText={filterText}
        setFilter={setFilter}
        showLBModal={showLBModal}
        showDeleteNodegroupModal={showDeleteNodegroupModal}
        closeDeleteModal={closeDeleteModal}
        deleteLoadBalancer={deleteLoadBalancer}
        successAPICall={successAPICall}
        closeSuccessAPIModal={closeSuccessAPIModal}
        showCreateLBModal={showCreateLBModal}
        closeLBModal={closeLBModal}
        addLoadBalancer={addLoadBalancer}
        setDeleteLbId={setDeleteLbId}
        deleteLbId={deleteLbId}
        showDeleteModal={showDeleteModal}
        upgradeClusterK8sVersion={upgradeClusterK8sVersion}
        setSucessAPICall={setSucessAPICall}
        successMessage={successMessage}
        setSuccessMessage={setSuccessMessage}
        onChangeLoadBalancerInput={onChangeLoadBalancerInput}
        loadBalancerBottom={loadBalancerBottom}
        getActionItemLabelCluster={getActionItemLabelCluster}
        deleteLbName={deleteLbName}
        setDeleteLbName={setDeleteLbName}
        backToHome={backToHome}
        loading={loading}
        formElementsUpgradeCluster={formElementsUpgradeCluster}
        handleSearchInputChange={handleSearchInputChange}
        handleSubmit={handleSubmit}
        selectedCloudAccount={selectedCloudAccount}
        cloudAccountError={cloudAccountError}
        cloudAccount={cloudAccount}
        showLoader={showLoader}
        showIMIUpgradeModal={showIMIUpgradeModal}
        closeIMIModal={closeIMIModal}
        upgradeNodeGroup={upgradeNodeGroup}
        showIMIModal={showIMIModal}
        sshKeys={sshKeys}
        selectedClusterCloudAccount={selectedClusterCloudAccount}
        authModalData={{
          showAuthModal,
          openAuthModal,
          closeAuthModal,
          onAuthSuccess,
          authPassword,
          setAuthPassword,
          authFormFooterEvents,
          showUnAuthModal,
          setShowUnAuthModal
        }}
        csvData={{
          data: csvGridInfo,
          headers: csvHeaders
        }}
        showSearchCloudAccount={showSearchCloudAccount}
        setShowSearchModal={setShowSearchModal}
        isSREAdminUser={isSREAdminUser()}
        filterIKS={filterIKS}
    />
  )
}

export default ClusterDetailsListContainer
