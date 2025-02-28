// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useCloudAccountStore from '../../../../store/cloudAccountStore/CloudAccountStore'
import useErrorBoundary from '../../../../hooks/useErrorBoundary'
import Wrapper from '../../../Wrapper'
import CustomInput from '../../../customInput/CustomInput'
import { UpdateFormHelper, isValidForm } from '../../../updateFormHelper/UpdateFormHelper'
import { Nav } from 'react-bootstrap'

export const buildCustomInput = (element, onChangeInput) => {
  return (
    <Wrapper key={element.id}>
      <CustomInput
        type={element.type}
        fieldSize={element.fieldSize}
        placeholder={element.placeholder}
        isRequired={element.validationRules.isRequired}
        label={element.validationRules.isRequired ? element.label + ' *' : element.label}
        value={element.value}
        onChanged={(event) => onChangeInput(event, element.id)}
        isValid={element.isValid}
        isTouched={element.isTouched}
        isReadOnly={element.isReadOnly}
        options={element.options}
        readOnly={element.readOnly}
      />
    </Wrapper>
  )
}

const SelectSingleInstanceView = ({ formState, onChangeInput }) => {
  return (
    <>
      <h3>{formState.title}</h3>
      <div className="row">
        <div className="col-xs-12 col-md-10 col-xl-8">{buildCustomInput(formState.form.instance, onChangeInput)}</div>
      </div>
    </>
  )
}

const SelectInstanceGroupInstanceView = ({ formState, onChangeInput }) => {
  return (
    <>
      <h3>{formState.title}</h3>
      <div className="row">
        <div className="col-xs-12 col-md-10 col-xl-8">
          {buildCustomInput(formState.form.instanceGroups, onChangeInput)}
        </div>
      </div>
      <div className="row">
        <div className="col-xs-12 col-md-10 col-xl-8">
          {buildCustomInput(formState.form.instanceGroupInstances, onChangeInput)}
        </div>
      </div>
    </>
  )
}

const HotToConnectSelectInstance = ({ setSelectedInstance }) => {
  const [activeTab, setActiveTab] = useState(0)
  const {
    instances,
    setInstances,
    instanceGroups,
    setInstanceGroups,
    setInstanceGroupInstances,
    instanceGroupInstances
  } = useCloudAccountStore((state) => state)
  const throwError = useErrorBoundary()

  const initialState = {
    title: activeTab === 0 ? 'Select a instance to continue' : 'Select a instance group to continue',
    form: {
      instance: {
        id: 'instance',
        type: 'dropdown',
        label: 'Instance:',
        placeholder: 'Select a instance',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: ''
      },
      instanceGroups: {
        id: 'instanceGroups',
        type: 'dropdown',
        label: 'Instance Group:',
        placeholder: 'Select a instance group',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: ''
      },
      instanceGroupInstances: {
        id: 'instanceGroupInstances',
        type: 'dropdown',
        label: 'Instance:',
        placeholder: 'Select a instance',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: ''
      }
    }
  }

  const [formState, setFormState] = useState(initialState)

  const goToInstructions = () => {
    switch (activeTab) {
      case 0: {
        const instance = instances.find((x) => x.name === formState.form.instance.value)
        setSelectedInstance(instance)
        break
      }
      case 1: {
        const instance = instanceGroupInstances.find((x) => x.name === formState.form.instanceGroupInstances.value)
        setSelectedInstance(instance)
        break
      }
      default:
        break
    }
  }

  const onChangeInput = (event, formInputName) => {
    const updatedState = {
      ...formState
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    if (formInputName === 'instanceGroups') {
      setInstanceGroupInstances(updatedState.form.instanceGroups.value)
    }

    setFormState(updatedState)
  }

  const getTabView = () => {
    switch (activeTab) {
      case 0:
        return <SelectSingleInstanceView formState={formState} onChangeInput={onChangeInput} />
      case 1:
        return <SelectInstanceGroupInstanceView formState={formState} onChangeInput={onChangeInput} />
      default:
        return null
    }
  }

  const getTapClass = (key) => {
    let classValue = 'tap-inactive'

    if (activeTab === key) {
      classValue = 'tap-active'
    }

    return classValue
  }

  const getInstanceGroupOptions = () => {
    const dropDownOptions = []
    const readyInstanceGroups = instanceGroups.filter((x) => x.readyCount === x.instanceCount)
    readyInstanceGroups.forEach((instance) => {
      const dropdownOption = {
        name: instance.name,
        value: instance.name
      }
      dropDownOptions.push(dropdownOption)
    })
    formState.form.instanceGroups.options = dropDownOptions
    setFormState({ ...formState })
  }

  const getInstanceOptions = (isInstanceGroup) => {
    const dropDownOptions = []
    const instanceList = isInstanceGroup ? instanceGroupInstances : instances
    const readyInstances = instanceList.filter((x) => x.status === 'Ready')
    readyInstances.forEach((instance) => {
      const dropdownOption = {
        imageSrc: '',
        name: `${instance.name} -IP: ${instance.interfaces[0].addresses[0]} -Type: ${instance.instanceType}`,
        value: instance.name
      }
      dropDownOptions.push(dropdownOption)
    })
    if (isInstanceGroup) {
      formState.form.instanceGroupInstances.options = dropDownOptions
    } else {
      formState.form.instance.options = dropDownOptions
    }
    setFormState({ ...formState })
  }

  useEffect(() => {
    const fetch = async () => {
      try {
        if (instances.length === 0) {
          await setInstances(true)
        }
        if (instanceGroups.length === 0) {
          await setInstanceGroups(true)
        }
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    getInstanceOptions(false)
  }, [instances])

  useEffect(() => {
    getInstanceGroupOptions()
  }, [instanceGroups])

  useEffect(() => {
    getInstanceOptions(true)
  }, [instanceGroupInstances])

  useEffect(() => {
    switch (activeTab) {
      case 0:
        if (formState.form.instance.value) {
          goToInstructions()
        }
        break
      case 1:
        if (formState.form.instanceGroupInstances.value) {
          goToInstructions()
        }
        break
      default:
        break
    }
  }, [formState])

  return (
    <div className="section">
      <div className="bd-highlight">
        <span className="h6">First, choose your instance:</span>
      </div>
      <>
        <Nav variant="tabs" className="tabs-secondary" activeKey={activeTab}>
          <Nav.Link className={getTapClass(0)} onClick={() => setActiveTab(0)} aria-current="page">
            Single instance
          </Nav.Link>
          <Nav.Link className={getTapClass(0)} onClick={() => setActiveTab(0)} aria-current="page">
            Instance group
          </Nav.Link>
        </Nav>
        {getTabView()}
      </>
    </div>
  )
}

export default HotToConnectSelectInstance
