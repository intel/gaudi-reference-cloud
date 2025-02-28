import React, { useEffect, useState } from 'react'
import useFamilyStore, { type Family } from '../../store/familyStore/FamilyStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { useNavigate } from 'react-router-dom'
import ProductFamiliesView from '../../components/productCatalog/ProductFamiliesView'
import { BsPencilFill, BsTrash3 } from 'react-icons/bs'
import useToastStore from '../../store/toastStore/ToastStore'
import PublicService from '../../services/PublicService'

const ProductFamiliesContainer = (): JSX.Element => {
  const columns = [
    {
      columnName: 'Family Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Description',
      targetColumn: 'description'
    },
    {
      columnName: 'Vendor Name',
      targetColumn: 'vendor'
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
    title: 'Family view',
    subTitle: 'No families found'
  }

  const emptyGridByFilter = {
    title: 'No families found',
    subTitle: 'The applied filter criteria did not match any families',
    action: {
      type: 'function',
      href: () => { setFilter('', true) },
      label: 'Clear filters'
    }
  }

  const deleteModalInitial = {
    show: false,
    title: 'Delete',
    itemToDelete: '',
    question: 'Are you sure you want to delete family {familyName}?',
    feedback: 'Delete the family cannot be undone',
    buttonLabel: 'Delete'
  }

  const actionsOptions = [
    {
      id: 'edit',
      name: (
        <>
          <BsPencilFill /> Edit{' '}
        </>
      )
    },

    {
      id: 'delete',
      name: (
        <>
          <BsTrash3 /> Delete{' '}
        </>
      )
    }
  ]

  // Local State
  const [rows, setRows] = useState<Family[]>([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [deleteModal, setDeleteModal] = useState(deleteModalInitial)

    // Global State
  const families = useFamilyStore((state) => state.families)
  const getFamilies = useFamilyStore((state) => state.getFamilies)
  const loading = useFamilyStore((state) => state.loading)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // Navigation
  const navigate = useNavigate()

  // Error Boundry
  const throwError = useErrorBoundary()

  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await getFamilies()
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    setGridInfo()
  }, [families])

  // *****
  // functions
  // *****
  function setGridInfo(): void {
    const gridInfo: any[] = []

    families?.forEach((family) => {
      gridInfo.push({
        name: family.name,
        description: family.description,
        vendor: family.vendor,
        actions: {
          showField: true,
          type: 'Buttons',
          value: family,
          selectableValues: actionsOptions,
          function: setAction
        }
      })
    })

    setRows(gridInfo)
  }

  function onRedirect(location: string, familyId = null): void {
    navigate(`/products/${location}`, { state: { familyId } })
  }

  function setFilter(event: any, clear: boolean): void {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  function setAction(action: any, family: Family): void {
    const type = action.id
    if (type === 'edit') {
      navigate({
        pathname: `/products/families/d/${family.name}/edit`
      })
    } else if (type === 'delete') {
      setDeleteModal({
        ...deleteModalInitial,
        show: true,
        question: deleteModalInitial.question.replace('{familyName}', family.name),
        itemToDelete: family.name
      })
    }
  }

  const onCloseDeleteModal = async (action: boolean): Promise<void> => {
    if (action) {
      await deleteService(deleteModal.itemToDelete)
    } else {
      setDeleteModal(deleteModalInitial)
    }
  }

  const deleteService = async (familyName: string): Promise<void> => {
    try {
      await PublicService.deleteCatalogFamily(familyName)
      await getFamilies()
      showSuccess('Family deleted successfully', false)
      setDeleteModal(deleteModalInitial)
    } catch (error: any) {
      const message = String(error.message)
      if (error.response) {
        const errData = error.response.data
        const errMessage = errData.message
        showError(errMessage, false)
      } else {
        showError(message, false)
      }
      setDeleteModal(deleteModalInitial)
    }
  }

  function backToHome(): void {
    navigate('/')
  }

  return (
    <ProductFamiliesView
      columns={columns}
      rows={rows}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      loading={loading}
      onRedirect={onRedirect}
      setFilter={setFilter}
      backToHome={backToHome}
      deleteModal={deleteModal}
      onCloseDeleteModal={onCloseDeleteModal}/>
  )
}

export default ProductFamiliesContainer
