// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React, { useState, useEffect } from 'react'
import ProductCatalogView from '../../components/productCatalog/ProductCatalogView'
import { useNavigate } from 'react-router'
import useFamilyStore from '../../store/familyStore/FamilyStore'
import useProductV2Store from '../../store/productV2Store/ProductV2Store'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { UpdateFormHelper, setSelectOptions } from '../../utility/updateFormHelper/UpdateFormHelper'
import { BsPencilFill } from 'react-icons/bs'
const ProductCatalogContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****

  const families = useFamilyStore((state) => state.families)
  const getFamilies = useFamilyStore((state) => state.getFamilies)
  const productsByFamily = useProductV2Store((state) => state.productsByFamily)
  const getProductByFamily = useProductV2Store((state) => state.getProductByFamily)
  const getProducts = useProductV2Store((state) => state.getProducts)

  // *****
  // local state
  // *****
  const navigate = useNavigate()
  const throwError = useErrorBoundary()

  const initialState = {
    mainTitle: 'Products',
    form: {
      family: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: '',
        placeholder: 'Please select family',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: <>
          Choose the desired family
        </>
      }
    }
  }

  const produtColumns = [
    {
      columnName: 'Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Region',
      targetColumn: 'regionName'
    },
    {
      columnName: 'Service',
      targetColumn: 'serviceName'
    },
    {
      columnName: 'Usage',
      targetColumn: 'usage'
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

  const emptyGrid = {
    title: 'No product found',
    subTitle: 'Please select a family'
  }

  const EmptyGridByFilter = {
    title: 'No product found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => { setFilter('', true) },
      label: 'Clear filters'
    }
  }

  const actionsOptions = [
    {
      id: 'edit',
      name: <>
        <BsPencilFill /> Edit{' '}
      </>
    }
  ]

  const [state, setState] = useState(initialState)
  const [selectedFamily, setSelectedFamily] = useState('')
  const [productItems, setProductItems] = useState<any[]>([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [isPageReady, setIsPageReady] = useState(false)
  // *****
  // Hooks
  // *****
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      const promiseArray = []
      promiseArray.push(getFamilies())
      promiseArray.push(getProducts())
      await Promise.all(promiseArray)
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    loadFamiliesForm()
  }, [families])

  useEffect(() => {
    if (selectedFamily) {
      getProductByFamily(selectedFamily)
    }
  }, [selectedFamily])

  useEffect(() => {
    if (selectedFamily) {
      loadProductGrid()
    }
  }, [productsByFamily])

  // *****
  // functions
  // *****
  const loadProductGrid = (): void => {
    const productGrid: any = []
    for (const index in productsByFamily) {
      const productItem = { ...productsByFamily[Number(index)] }
      productGrid.push({
        name: productItem.name,
        regionName: productItem.regionName,
        serviceName: productItem.serviceName,
        usage: productItem.usage,
        actions: {
          showField: true,
          type: 'Buttons',
          value: productItem,
          selectableValues: actionsOptions,
          function: setAction
        }
      })
    }
    setProductItems(productGrid)
  }

  const loadFamiliesForm = (): void => {
    const familiesItems: any = []
    families?.forEach((family) => {
      familiesItems.push({
        name: family.name,
        value: family.name
      })
    })
    const stateUpdate = { ...state }
    const form = setSelectOptions('family', familiesItems, stateUpdate.form)
    stateUpdate.form = form
    setIsPageReady(true)
    setState(stateUpdate)
  }

  const setFilter = (event: any, clear: boolean): void => {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(EmptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  const setAction = (action: any, item: any): void => {
    switch (action.edit) {
      default:
        navigate(`/products/d/${item.name}`)
        break
    }
  }

  const onChangeInput = (event: any, formInputName: string): void => {
    const value = event.target.value
    const updatedState = {
      ...state
    }

    let updatedForm = updatedState.form

    updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

    updatedState.form = updatedForm

    if (formInputName === 'family') {
      setSelectedFamily(value)
    }

    setState(updatedState)
  }

  const onCancel = (): void => {
    navigate('/')
  }

  return <ProductCatalogView
    state={state}
    onChangeInput={onChangeInput}
    onCancel={onCancel}
    produtColumns={produtColumns}
    productItems={productItems}
    emptyGrid={emptyGridObject}
    filterText={filterText}
    setFilter={setFilter}
    isPageReady={isPageReady} />
}

export default ProductCatalogContainer
