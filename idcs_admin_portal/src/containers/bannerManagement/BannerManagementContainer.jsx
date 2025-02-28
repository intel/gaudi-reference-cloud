import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'

import useErrorBoundary from '../../hooks/useErrorBoundary'
import BannerDetailsList from '../../components/bannerManagement/BannerDetailsList'
import useBannerStore from '../../store/bannerStore/BannerStore'
import moment from 'moment'
import useToastStore from '../../store/toastStore/ToastStore'

const dateFormat = 'MM/DD/YYYY hh:mm a'

const BannerManagementContainer = () => {
  // local state
  const columns = [
    {
      columnName: 'ID',
      targetColumn: 'id'
    },
    {
      columnName: 'Type',
      targetColumn: 'type'
    },
    {
      columnName: 'Message Title',
      targetColumn: 'title',
      width: '9rem',
      isSort: false
    },
    {
      columnName: 'Status',
      targetColumn: 'status'
    },
    {
      columnName: 'Message Description',
      targetColumn: 'message',
      width: '12.5rem',
      isSort: false
    },
    {
      columnName: 'User Types',
      targetColumn: 'userTypes',
      width: '7.5rem',
      isSort: false
    },
    {
      columnName: 'Routes',
      targetColumn: 'routes',
      isSort: false
    },
    {
      columnName: 'Regions',
      targetColumn: 'regions',
      width: '5rem',
      isSort: false
    },
    {
      columnName: 'Expiration Time',
      width: '12rem',
      targetColumn: 'expirationDatetime'
    },
    {
      columnName: 'Link',
      targetColumn: 'link',
      width: '6rem'
    },
    {
      columnName: 'Is Maintenance',
      targetColumn: 'isMaintenance'
    },
    {
      columnName: 'Actions',
      width: '7rem',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'buttons',
        behaviorFunction: null
      }
    }
  ]

  const emptyGrid = {
    title: 'Alert details empty',
    subTitle: 'No details found'
  }

  const emptyGridByFilter = {
    title: 'No banners found',
    subTitle: 'The applied filter criteria did not match any banners',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  // Initial state for confirm modal
  const initialConfirmData = {
    isShow: false,
    title: '',
    data: [],
    id: null,
    isBlocked: null,
    onClose: closeConfirmModal
  }

  // Initial state for loader
  const initialLoaderData = {
    isShow: false,
    message: ''
  }

  // Local State
  const [banners, setBanners] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [confirmModalData, setConfirmModalData] = useState(initialConfirmData)
  const [showLoader, setShowLoader] = useState(initialLoaderData)

  // Global State
  const loading = useBannerStore((state) => state.loading)
  const bannerList = useBannerStore((state) => state.bannerList)
  const setBannerList = useBannerStore((state) => state.setBannerList)
  const removeBanner = useBannerStore((state) => state.removeBanner)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // Navigation
  const navigate = useNavigate()

  // Error Boundary
  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await setBannerList()
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    setGridInfo()
  }, [bannerList])

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in bannerList) {
      const banner = { ...bannerList[index] }
      banner.routes = banner.routes.join(', ')

      gridInfo.push({
        id: banner.id,
        type: banner.type,
        title: banner.title,
        status: banner.status,
        message: banner.message,
        userTypes: banner.userTypes.join(', '),
        routes: banner.routes,
        regions: banner.regions.join(', ') || 'None',
        expirationDatetime: banner.expirationDatetime ? moment(banner.expirationDatetime).format(dateFormat) : 'None',
        link: banner.link
          ? {
              showField: true,
              type: 'hyperlink',
              value: banner.link.label,
              href: banner.link.href
            }
          : '',
        isMaintenance: banner.isMaintenance,
        actions: getActionButton(banner)
      })
    }

    setBanners(gridInfo)
  }

  function closeConfirmModal() {
    setConfirmModalData(initialConfirmData)
  }

  function getActionButton(banner) {
    return {
      showField: true,
      type: 'buttons',
      function: setAction,
      value: banner,
      selectableValues: [
        {
          value: banner,
          label: 'Duplicate button',
          name: 'Duplicate'
        },
        {
          value: banner,
          label: 'Update button',
          name: 'Update'
        },
        {
          value: banner,
          label: 'Delete button',
          name: 'Delete'
        }
      ]
    }
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

  function setAction(item, banner) {
    const type = item.name

    if (type === 'Update') {
      navigate('/bannermanagement/update', { state: { banner } })
    } else if (type === 'Duplicate') {
      navigate('/bannermanagement/create', { state: { banner } })
    } else {
      const data = [
        {
          col: 'Banner ID',
          value: banner.id
        },
        {
          col: 'Banner Type',
          value: banner.type
        },
        {
          col: 'Banner Status',
          value: banner.status
        },
        {
          col: 'Banner Message',
          value: banner.message
        },
        {
          col: 'User Types',
          value: banner.userTypes.join(', ')
        },
        {
          col: 'Banner Routes',
          value: banner.routes
        },
        {
          col: 'Banner Regions',
          value: banner.regions.join(', ')
        },
        {
          col: 'Banner Expire Time',
          value: banner.expirationDatetime || 'None'
        },
        {
          col: 'Banner Link',
          value: banner.link ? (
            <a aria-label={banner.link.label} href={banner.link.href} target="_blank" rel="noreferrer">
              {banner.link.label}
            </a>
          ) : (
            'N/A'
          )
        },
        {
          col: 'Is Maintenance',
          value: banner?.isMaintenance ?? ''
        }
      ]

      setConfirmModalData({
        title: `Delete banner: ${banner.title} (${banner.id})`,
        data,
        isShow: true,
        id: banner.id,
        isBlocked: false,
        onClose: closeConfirmModal
      })
    }
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
      await removeBanner(confirmModalData.id)
      showSuccess(`Banner ${confirmModalData.id} has been deleted.`)
      await setBannerList()
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
    <BannerDetailsList
      loading={loading}
      banners={banners}
      columns={columns}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      confirmModalData={confirmModalData}
      showLoader={showLoader}
      setFilter={setFilter}
      onSubmit={onSubmit}
      backToHome={backToHome}
    />
  )
}

export default BannerManagementContainer
