import React from 'react'
import CustomInput from '../../utility/customInput/CustomInput'
import Button from 'react-bootstrap/Button'
import { ButtonGroup } from 'react-bootstrap'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'
import AddNodeToPoolContainer from '../../containers/nodePoolManagement/AddNodeToPoolContainer'
import { NavLink } from 'react-router-dom'
const PoolEdit = (props) => {
  const poolId = props.poolId
  const nodeCount = props.nodeCount
  const state = props.state
  const title = state.title
  const form = state.form
  const navigationTop = state.navigationTop
  const navigationBottom = state.navigationBottom
  const isValidForm = state.isValidForm
  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  const showModal = props.showModal
  const addNodeToPool = props.addNodeToPool
  const showAddNode = props.showAddNode
  const cancelAddNode = props.cancelAddNode
  const addNodeToPoolFn = props.addNodeToPoolFn

  const formElements = []
  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    formElements.push({
      id: key,
      configInput: formItem
    })
  }

  return (
    <>
      {showAddNode && (
        <AddNodeToPoolContainer selectedPool={poolId} cancelAddNode={cancelAddNode} addNodeToPoolFn={addNodeToPoolFn} />
      )}
      <div className="section">
        {navigationTop.map((item, index) => (
          <div className="m-lg-0" key={index}>
            {
              <Button
                intc-id={`navigationTop${item.label}`}
                variant={item.buttonVariant}
                className="p-s0"
                onClick={item.function}
              >
                {item.label}
              </Button>
            }
          </div>
        ))}
      </div>
      <OnSubmitModal showModal={showModal} message="Working on your request" />

      <div className="section">
        <div className="d-flex flex-row flex-wrap justify-content-between align-items-center w-100 gap-s6">
          <h2 className="h4">{title}</h2>
          {poolId && (
            <ButtonGroup>
              <NavLink
                className="btn btn-outline-primary"
                to={`/npm/pools/accounts/${poolId}`}
                intc-id={'btn-editnodepool-navigate-cloud-accounts'}
                aria-label='View cloud accounts nodes'
              >
                View Cloud Accounts
              </NavLink>
              <NavLink
                className="btn btn-outline-primary"
                variant="primary"
                to={`/npm/nodes/${poolId}`}
                aria-label='View nodes'
                intc-id={'btn-editnodepool-navigate-view-nodes'}
              >
                View Nodes
              </NavLink>
              <Button
                className="btn btn-outline-primary"
                variant="outline-primary"
                intc-id={'btn-editnodepool-add-node'}
                aria-label='Add node'
                onClick={() => addNodeToPool()}
              >
                Add node
              </Button>
            </ButtonGroup>
          )}
        </div>
      </div>
      <div className="section">
        {poolId && (
          <>
            <div className="row">
              <div className="col-12 col-md-2">Compute Node Pool ID:</div>
              <div className="col-12 col-md-2">{poolId}</div>
            </div>
            <div className="row">
              <div className="col-12 col-md-2"># of Node:</div>
              <NavLink
                className="btn btn-link col-12 col-md-2 justify-content-start"
                variant="primary"
                to={`/npm/nodes/${poolId}`}
                aria-label='View nodes'
                intc-id={'btn-editnodepool-navigate-view-nodes-from-count'}
              >
                {nodeCount}
              </NavLink>
            </div>
          </>
        )}
        {formElements.map((element, index) => (
          <CustomInput
            key={index}
            customContent={element.configInput.customContent}
            hideContent={false}
            type={element.configInput.type}
            fieldSize={element.configInput.fieldSize}
            placeholder={element.configInput.placeholder}
            isRequired={element.configInput.validationRules.isRequired}
            label={
              element.configInput.validationRules.isRequired
                ? element.configInput.label + ' *'
                : element.configInput.label
            }
            value={element.configInput.value}
            onChanged={(event) => onChangeInput(event, element.id)}
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
            hidden={element.configInput.hidden}
          />
        ))}

        <ButtonGroup>
          {navigationBottom.map((item, index) => (
            <Button
              key={index}
              varian="primary"
              intc-id={`navigationBottom-editnodepool-${item.label}`}
              data-wap_ref={`navigationBottom-editnodepool-${item.label}`}
              disabled={item.label === 'Save' ? !isValidForm : false}
              variant={item.buttonVariant}
              onClick={item.label === 'Save' ? onSubmit : item.function}
            >
              {item.label}
            </Button>
          ))}
        </ButtonGroup>
      </div>
    </>
  )
}

export default PoolEdit
