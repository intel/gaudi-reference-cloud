import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'

import useErrorBoundary from '../../hooks/useErrorBoundary'
import UserManagement from '../../components/userManagement/UserManagement'
import CloudAccountService from '../../services/CloudAccountService'
import useUserManagementStore from '../../store/userManagement/UserManagementStore'
import useUserStore from '../../store/userStore/UserStore'
import useToastStore from '../../store/toastStore/ToastStore'

const UserManagementContainer = () => {
  // local state
  const columns = [
    {
      columnName: 'ID',
      targetColumn: 'id'
    },
    {
      columnName: 'Email',
      targetColumn: 'owner'
    },
    {
      columnName: 'Type',
      targetColumn: 'type'
    },
    {
      columnName: 'Created',
      targetColumn: 'created'
    },
    {
      columnName: 'Blocked?',
      targetColumn: 'restricted'
    },
    {
      columnName: 'By',
      targetColumn: 'adminName'
    },
    {
      columnName: 'When',
      targetColumn: 'accessLimitedTimestamp'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'buttons',
        behaviorFunction: null
      }
    }
  ]

  const emptyGrid = {
    title: 'Cloud Account details empty',
    subTitle: 'No details found'
  }

  const emptyGridByFilter = {
    title: 'No details found',
    subTitle: 'The applied filter criteria did not match any items',
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
  const [users, setUsers] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [confirmModalData, setConfirmModalData] = useState(initialConfirmData)
  const [showLoader, setShowLoader] = useState(initialLoaderData)

  // Global State
  const loading = useUserManagementStore((state) => state.loading)
  const userList = useUserManagementStore((state) => state.userList)
  const setUserList = useUserManagementStore((state) => state.setUserList)
  const user = useUserStore((state) => state.user)
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
        await setUserList()
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    setGridInfo()
  }, [userList])

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in userList) {
      const user = { ...userList[index] }

      gridInfo.push({
        id: user.id,
        owner: user.owner,
        type: user.type,
        created: user.created,
        restricted: user.restricted,
        adminName: user.adminName,
        accessLimitedTimestamp: user.accessLimitedTimestamp,
        actions: getActionButton(user)
      })
    }

    setUsers(gridInfo)
  }

  function closeConfirmModal() {
    setConfirmModalData(initialConfirmData)
  }

  function getActionButton(user) {
    const buttonLabel = user.restricted === 'No' ? 'Block' : 'Unblock'

    return {
      showField: true,
      type: 'button',
      value: buttonLabel,
      function: () => {
        setAction(user)
      }
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

  function setAction(user) {
    const label = user.restricted === 'No' ? 'Block' : 'Unblock'

    const data = [
      {
        col: 'Cloud Account ID',
        value: user.id
      },
      {
        col: 'Cloud Account Email',
        value: user.owner
      },
      {
        col: 'Cloud Account Type',
        value: user.type
      },
      {
        col: 'Is Blocked',
        value: user.restricted === 'Yes' ? 'Yes' : 'No'
      }
    ]

    setConfirmModalData({
      title: `${label} ${user.owner} (${user.id})`,
      data,
      isShow: true,
      id: user.id,
      isBlocked: user.restricted,
      onClose: closeConfirmModal
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
      const status = confirmModalData.isBlocked === 'No'
      const payload = {
        restricted: status,
        adminName: user.email
      }

      await CloudAccountService.updateCloudAccount(confirmModalData.id, payload)
      showSuccess(`Cloud account ${status ? 'Blocked' : 'Unblocked'}`)
      await setUserList()
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
    <UserManagement
    loading={loading}
    users={users}
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

export default UserManagementContainer
