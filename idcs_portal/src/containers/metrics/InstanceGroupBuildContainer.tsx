// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useErrorBoundary from '../../hooks/useErrorBoundary'

import { UpdateFormHelper, setFormValue, setSelectOptions } from '../../utils/updateFormHelper/UpdateFormHelper'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import InstanceGroupBuild from '../../components/metrics/InstanceGroupBuild'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

const InstanceGroupBuildContainer = (props: any): JSX.Element => {
  const throwError = useErrorBoundary()

  const instances = props?.instances

  const initialState = {
    form: {
      instanceGroups: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Instance Groups:',
        maxWidth: '25rem',
        placeholder: 'Please select instance group',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        options: [],
        validationRules: {
          isRequired: false
        }
      }
    }
  }

  const allowedInstanceStatus = ['Ready']
  const allowedInstanceCategories = ['VirtualMachine']

  if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_METRICS_BARE_METAL)) {
    allowedInstanceCategories.push('BareMetalHost')
  }

  // Local State

  const [selectedResource, setSelectedResource] = useState('')
  const [state, setState] = useState(initialState)
  const [selectedResourceCategory, setSelectedResourceCategory] = useState('')
  const [instancesFilteredValues, setInstancesFilteredValues] = useState<any[]>([])

  // Global State
  const instanceGroupInstances = useCloudAccountStore((state) => state.instanceGroupInstances)
  const setInstanceGroupInstances = useCloudAccountStore((state) => state.setInstanceGroupInstances)

  // Hooks

  useEffect(() => {
    setForm()
  }, [])

  useEffect(() => {
    if (selectedResource) {
      void getInstancesFromGroup()
    }
  }, [selectedResource])

  useEffect(() => {
    getInstanceValues()
  }, [instanceGroupInstances])

  // functions

  const setForm = (): void => {
    const stateUpdated = { ...state }

    if (instances.length > 0) {
      stateUpdated.form = setSelectOptions('instanceGroups', instances, stateUpdated.form)
      const selectedValue = instances[0]?.value

      stateUpdated.form = setFormValue('instanceGroups', selectedValue, stateUpdated.form)
      setSelectedResource(selectedValue)
      setSelectedResourceCategory(instances[0]?.instanceCategory)
    }

    setState(stateUpdated)
  }

  const onChange = (event: any, inputName: string): void => {
    const updatedState = { ...state }
    const value: string = event.target.value

    const updatedForm = UpdateFormHelper(value, inputName, updatedState.form)
    setInstancesFilteredValues([])

    setSelectedResource(value)

    const selectedInstance = instances.filter((x: any) => x.value === value)[0]
    setSelectedResourceCategory(selectedInstance.instanceCategory)

    updatedState.form = updatedForm
    setState(updatedState)
  }

  const getInstancesFromGroup = async (): Promise<void> => {
    try {
      await setInstanceGroupInstances(selectedResource)
    } catch (error) {
      throwError(error)
    }
  }

  const getInstanceValues = (): void => {
    if (instanceGroupInstances.length > 0) {
      const instancesValues = instanceGroupInstances
        .filter(
          (x) =>
            allowedInstanceCategories.includes(x.instanceTypeDetails?.instanceCategory as string) &&
            allowedInstanceStatus.includes(x.status)
        )
        .map((instance) => {
          return {
            name: instance.name + ` (${instance.instanceType})`,
            value: instance.resourceId,
            instanceName: instance.name,
            instanceCategory: instance.instanceTypeDetails?.instanceCategory
          }
        })

      setInstancesFilteredValues(instancesValues)
    }
  }

  return (
    <InstanceGroupBuild
      onChange={onChange}
      state={state}
      selectedResourceCategory={selectedResourceCategory}
      instancesFilteredValues={instancesFilteredValues}
      selectedResource={selectedResource}
    />
  )
}

export default InstanceGroupBuildContainer
