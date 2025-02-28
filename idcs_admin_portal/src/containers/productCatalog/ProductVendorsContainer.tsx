// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import ProductVendors from '../../components/productCatalog/ProductVendors'
import { useNavigate } from 'react-router'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { BsPencilFill, BsTrash3 } from 'react-icons/bs'
import useToastStore from '../../store/toastStore/ToastStore'
import useVendorStore from '../../store/vendorStore/VendorStore'
import PublicService from '../../services/PublicService'

const ProductVendorsContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const vendors = useVendorStore((state) => state.vendors)
  const loading = useVendorStore((state) => state.loading)
  const getVendor = useVendorStore((state) => state.getVendor)
  const throwError = useErrorBoundary()
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // *****
  // local state
  // *****
  const productColumns = [
    {
      columnName: 'Vendor Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Description',
      targetColumn: 'description'
    },
    {
      columnName: 'Organization Name',
      targetColumn: 'organizationName'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'Buttons',
        behaviorFunction: null
      }
    }
  ]

  const deleteModalInitial = {
    show: false,
    title: 'Delete',
    itemToDelete: '',
    question: 'Are you sure you want to delete vendor {vendorName}?',
    feedback: 'Delete the vendor cannot be undone',
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

  const emptyGrid = {
    title: 'No vendors found',
    subTitle: 'No vendors items'
  }

  const emptyGridByFilter = {
    title: 'No vendors found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => {
        setFilter('', true)
      },
      label: 'Clear filters'
    }
  }

  const navigate = useNavigate()
  const [productItems, setServiceItems] = useState<any[]>([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [deleteModal, setDeleteModal] = useState(deleteModalInitial)

  // *****
  // Hooks
  // *****
  useEffect(() => {
    setGridVendors()
  }, [vendors])

  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await getVendor()
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])
  // *****
  // functions
  // *****
  const setGridVendors = (): void => {
    const gridInfo: any[] = []
    for (const index in vendors) {
      const vendorItem = { ...vendors[Number(index)] }
      gridInfo.push({
        name: vendorItem.name,
        description: vendorItem.description,
        organizationName: vendorItem.organizationName,
        actions: {
          showField: true,
          type: 'Buttons',
          value: vendorItem,
          selectableValues: actionsOptions,
          function: setAction
        }
      })
    }
    setServiceItems(gridInfo)
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      case 'delete': {
        setDeleteModal({
          ...deleteModalInitial,
          show: true,
          question: deleteModalInitial.question.replace('{vendorName}', item.name),
          itemToDelete: item
        })
        break
      }
      default:
        navigate({
          pathname: `/products/vendors/d/${item.name}/edit`
        })
        break
    }
  }

  const onCancel = (): void => {
    navigate('/')
  }

  const setFilter = (event: any, clear: boolean): void => {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  const onCloseDeleteModal = async (action: boolean): Promise<void> => {
    if (action) {
      await deleteService(deleteModal.itemToDelete)
    } else {
      setDeleteModal(deleteModalInitial)
    }
  }

  const deleteService = async (item: any): Promise<void> => {
    try {
      await PublicService.deleteVendorCatalog(item.name)
      await getVendor()
      showSuccess('Vendor deleted successfully', false)
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

  return (
    <ProductVendors
      onCancel={onCancel}
      productColumns={productColumns}
      productItems={productItems}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      setFilter={setFilter}
      deleteModal={deleteModal}
      loading={loading}
      onCloseDeleteModal={onCloseDeleteModal}
    />
  )
}

export default ProductVendorsContainer
