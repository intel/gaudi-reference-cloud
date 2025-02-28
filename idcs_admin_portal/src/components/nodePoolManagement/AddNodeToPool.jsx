import CustomInput from '../../utility/customInput/CustomInput'
import { Button, ButtonGroup, Modal } from 'react-bootstrap'

const AddNodeToPool = (props) => {
  // props Variables
  const addNodeModal = props.addNodeModal
  const nodeListFormItem = props.nodeListFormItem
  const selectedNode = props.selectedNode
  const onSelectNodeFromList = props.onSelectNodeFromList
  const cancelNodeAddition = props.cancelNodeAddition
  const addNodeFn = props.addNodeFn

  // Props Functions

  function buildCustomInput(element) {
    return (
      <CustomInput
        key={element.id}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired ? element.configInput.label + ' *' : element.configInput.label
        }
        value={element.configInput.value}
        onChanged={(event) => onSelectNodeFromList(event)}
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
      />
    )
  }

  return (
    <>
      <Modal show={addNodeModal.show} backdrop="static" keyboard={false} intc-id="OnSubmitAddNodeModal">
        <Modal.Header>{addNodeModal.title}</Modal.Header>
        <Modal.Body>
          <div className="modal-body row justify-content-center">
            <div className="col-12 row">
              {addNodeModal.isLoader ? (
                <div className="spinner-border text-primary center"></div>
              ) : (
                buildCustomInput({ id: 'nodeList', configInput: nodeListFormItem })
              )}
            </div>
          </div>
        </Modal.Body>
        <Modal.Footer>
          <ButtonGroup>
            <Button onClick={cancelNodeAddition} variant="outline-primary" intc-id={'btn-nodepool-cancelAddNode'}>
              Cancel
            </Button>
            <Button onClick={addNodeFn} variant="primary" intc-id={'btn-nodepool-addNode'} data-wap_ref='btn-nodepool-addNode' disabled={!selectedNode}>
              Add node
            </Button>
          </ButtonGroup>
        </Modal.Footer>
      </Modal>
    </>
  )
}

export default AddNodeToPool
