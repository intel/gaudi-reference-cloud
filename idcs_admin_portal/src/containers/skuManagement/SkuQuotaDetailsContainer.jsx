import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import SkuQuotaDetails from '../../components/skuManagement/skuQuotaDetails/SkuQuotaDetails'
import useSkuQuotaStore from '../../store/skuManagement/SkuQuotaStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import SkuQuotaService from '../../services/SkuQuotaService'
import useToastStore from '../../store/toastStore/ToastStore'
import useVendorStore from '../../store/vendorStore/VendorStore'

const SkuQuotaDetailsContainer = (props) => {
  const viewMode = props?.viewMode
  const userId = props?.userId ?? ''
  // local state
  const columns = [
    {
      columnName: 'Service Type',
      targetColumn: 'serviceType'
    },
    {
      columnName: 'Product Family',
      targetColumn: 'family'
    },
    {
      columnName: 'Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Instance Type',
      targetColumn: 'instanceType'
    },
    {
      columnName: 'Description',
      targetColumn: 'instanceName'
    },
    {
      columnName: 'Cloud Account',
      targetColumn: 'cloudAccountId'
    },
    {
      columnName: 'Assignment Date',
      targetColumn: 'creationTimestamp'
    },
    {
      columnName: 'Assignment Admin',
      targetColumn: 'creator'
    }
  ]

  columns.forEach(col => {
    col.width = '12rem'
  })

  if (!viewMode) {
    columns.push({
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'buttons',
        behaviorFunction: null
      }
    })
  }

  const emptyGrid = {
    title: 'Cloud Account instance whitelist details empty',
    subTitle: 'No details found',
    action: {
      type: 'redirect',
      href: '/camanagement/create',
      label: 'Create'
    }
  }

  const emptyGridByFilter = {
    title: 'No details found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => {
        setFilter('', true)
        setSelectedServiceType('')
      },
      label: 'Clear filters'
    }
  }

  // Initial state for confirm modal
  const initialConfirmData = {
    title: 'Remove',
    family: '',
    instanceName: '',
    instanceType: '',
    cloudAccount: '',
    onClose: closeConfirmModal,
    isShow: false,
    action: 'Remove'
  }

  // Initial state for loader
  const initialLoaderData = {
    isShow: false,
    message: ''
  }

  // Local State
  const [adminSkuQuotas, setAdminSkuQuotas] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [confirmModalData, setConfirmModalData] = useState(initialConfirmData)
  const [showLoader, setShowLoader] = useState(initialLoaderData)
  const [serviceTypes, setServiceTypes] = useState([{ name: 'All', value: '' }])
  const [selectedServiceType, setSelectedServiceType] = useState('')
  const [idcVendor, setIDCVendor] = useState(null)
  // Global State
  const vendors = useVendorStore((state) => state.vendors)
  const getIDCVendor = useVendorStore((state) => state.getIDCVendor)
  const loading = useSkuQuotaStore((state) => state.loading)
  const skuQuotas = useSkuQuotaStore((state) => state.skuQuotas)
  const setSkuQuotas = useSkuQuotaStore((state) => state.setSkuQuotas)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // Navigation
  const navigate = useNavigate()

  // Error Boundry
  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        if (!vendors) await getIDCVendor()
        await setSkuQuotas(userId)
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [userId])

  useEffect(() => {
    if (vendors) {
      const idcVendor = vendors.find((x) => x.name === 'idc')
      if (idcVendor) setIDCVendor(idcVendor)
    }
  }, [vendors])

  useEffect(() => {
    setServiceTypeOptions()
  }, [idcVendor])

  useEffect(() => {
    setGridInfo()
  }, [skuQuotas])

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in skuQuotas) {
      const quota = { ...skuQuotas[index] }
      const gridData = {
        serviceType: getServiceType(quota.familyId),
        family: quota.family,
        name: quota.name,
        instanceType: quota.instanceType,
        instanceName: quota.instanceName,
        cloudAccountId: quota.cloudAccountId,
        creationTimestamp: quota.creationTimestamp,
        creator: quota.creator
      }
      if (!viewMode) gridData.actions = getActionButton(quota)
      gridInfo.push(gridData)
    }

    setAdminSkuQuotas(gridInfo)
  }

  function closeConfirmModal() {
    setConfirmModalData(initialConfirmData)
  }

  function getServiceType(familyId) {
    if (!idcVendor || !familyId) return ''
    const services = idcVendor?.families
    const matchedService = services.find(service => service.id === familyId)
    return matchedService?.description || ''
  }

  function setServiceTypeOptions() {
    const services = [{ name: 'All', value: '' }]
    if (idcVendor) {
      idcVendor?.families.map(service => services.push({
          name: service.description.split(':')[0],
          value: service.description
        }))
    }
    setServiceTypes(services)
  }

  function getActionButton(quota) {
    const buttonLabel = 'Remove'

    return {
      showField: true,
      type: 'button',
      value: buttonLabel,
      function: () => {
        setAction(quota)
      }
    }
  }

  function setServiceTypeFilter(event) {
    setEmptyGridObject(emptyGridByFilter)
    setSelectedServiceType(event.target.value)
  }

  function setFilter(event, clear) {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  function setAction(quota) {
    setConfirmModalData({
      title: 'Remove Cloud Account Assignment',
      family: quota.family,
      instanceName: quota.instanceName,
      instanceType: quota.instanceType,
      cloudAccount: quota.cloudAccountId,
      vendorId: quota.vendorId,
      resourceId: quota.resourceId,
      isShow: true,
      onClose: closeConfirmModal,
      action: 'Remove'
    })
  }

  function backToHome() {
    navigate('/')
  }

  function onSubmit() {
    closeConfirmModal()
    setShowLoader({ isShow: true, message: 'Working on your request' })
    submitForm()
  }

  async function submitForm() {
    try {
      await SkuQuotaService.deleteAcl(confirmModalData.cloudAccount, confirmModalData.resourceId, confirmModalData.vendorId)
      showSuccess('Cloud account removed.')
      await setSkuQuotas(userId)
    } catch (error) {
      let message = ''
      if (error.response) {
        if (error.response.data.message !== '') {
          message = error.response.data.message
        } else {
          message = error.message
        }
      } else {
        message = error.message
      }
      showError(message)
    }
    setShowLoader({ isShow: false })
  }

  return (
    <SkuQuotaDetails
      viewMode={viewMode}
      loading={loading}
      adminSkuQuotas={adminSkuQuotas}
      columns={columns}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      confirmModalData={confirmModalData}
      showLoader={showLoader}
      serviceTypes={serviceTypes}
      selectedServiceType={selectedServiceType}
      setFilter={setFilter}
      onSubmit={onSubmit}
      backToHome={backToHome}
      setServiceTypeFilter={setServiceTypeFilter}
    />
  )
}

export default SkuQuotaDetailsContainer
