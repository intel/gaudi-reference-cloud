import CustomInput from '../../../utility/customInput/CustomInput'
import OnSubmitModal from '../../../utility/modals/onSubmitModal/OnSubmitModal'
import { Button, ButtonGroup } from 'react-bootstrap'
import SearchBox from '../../../utility/searchBox/SearchBox'
import InstanceTable from './InstanceTable'
import SelectedAccountCard from '../../../utility/selectedAccountCard/SelectedAccountCard'

const SkuQuotaCreate = (props) => {
  // props variables
  const state = props.state
  const mainSubtitle = state.mainSubtitle
  const form = state.form
  const isValidForm = state.isValidForm
  const products = props.products
  const showLoader = props.showLoader
  const backButtonLabel = props.backButtonLabel
  const selectedCloudAccount = props.selectedCloudAccount
  const selectedProducts = props.selectedProducts
  const setSelectedProducts = props.setSelectedProducts

  // props functions
  const onChangeInput = props.onChangeInput
  const onSearchCloudAccount = props.onSearchCloudAccount
  const onSubmit = props.onSubmit
  const onCancel = props.onCancel

  // functions
  function buildCustomInput(element) {
    return (
      <CustomInput key={element.id} type={element.configInput.type} fieldSize={element.configInput.fieldSize} placeholder={element.configInput.placeholder} isRequired={element.configInput.validationRules.isRequired} label={element.configInput.validationRules.isRequired ? element.configInput.label + ' *' : element.configInput.label} value={element.configInput.value} onChanged={(event) => onChangeInput(event, element.id)} isValid={element.configInput.isValid} isTouched={element.configInput.isTouched} helperMessage={element.configInput.helperMessage} isReadOnly={element.configInput.isReadOnly} options={element.configInput.options} validationMessage={element.configInput.validationMessage} readOnly={element.configInput.readOnly} refreshButton={element.configInput.refreshButton} isMultiple={element.configInput.isMultiple} onChangeSelectValue={element.configInput.onChangeSelectValue} extraButton={element.configInput.extraButton} emptyOptionsMessage={element.configInput.emptyOptionsMessage} />
    )
  }

  // variables
  const formElementsInstanceConfiguration = []
  const formElementsCloudAccount = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.sectionGroup === 'configuration') {
      formElementsInstanceConfiguration.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.sectionGroup === 'cloudAccount') {
      formElementsCloudAccount.push({
        id: key,
        configInput: formItem
      })
    }
  }

  return (
    <>
      <OnSubmitModal showModal={showLoader.isShow} message={showLoader.message}></OnSubmitModal>
      <div className="section">
        <Button intc-id="navigationTop-BackButton" variant="link" className='p-s0' onClick={onCancel}>
          {backButtonLabel}
        </Button>
      </div>

      <div className="section">
        <h2 className='h4'>{mainSubtitle}</h2>
      </div>
      <div className="section">
        <div className="row">
          <div className="col-md-6 d-flex flex-column gap-s6">

            {formElementsCloudAccount.map((element, index) => (
              <div className="d-flex-customInput" key={index}>
                <div className='customInputLabel gap-s6' intc-id="filterText">
                  Cloud Account: *
                </div>
                <div className='col-12'>
                  <SearchBox
                    intc-id="searchCloudAccounts"
                    placeholder={`${element.configInput.placeholder}`}
                    aria-label="Type to search cloud account..."
                    value={ element.configInput.value || ''}
                    onChange={(e) => onChangeInput(e, element.id)}
                    onClickSearchButton={onSearchCloudAccount}
                  />
                  {element.configInput.validationMessage && <div className='invalid-feedback pt-s3' intc-id='cloudAccountSearchError'>{element.configInput.validationMessage}</div>}
                </div>
              </div>
            ))}

            {selectedCloudAccount && <div className="d-md-none">
              <SelectedAccountCard selectedCloudAccount={selectedCloudAccount}/>
            </div>}

            {formElementsInstanceConfiguration.map((element, index) => (
              <div className="col-12" key={index}>{buildCustomInput(element)}</div>
            ))}
          </div>
          {selectedCloudAccount && <div className="col-lg-3 col-md-6 d-md-block d-none">
            <SelectedAccountCard selectedCloudAccount={selectedCloudAccount}/>
          </div>}
        </div>
      </div>

      <div className="section">
        <InstanceTable products={products} setSelectedProducts={setSelectedProducts} selectedProducts={selectedProducts}/>
        <ButtonGroup>
          <Button intc-id="navigationBottom-whitelistAccount" disabled={!(isValidForm && selectedProducts?.length > 0)} variant="primary" onClick={onSubmit}>
            Whitelist Account
          </Button>
          <Button intc-id="navigationBottom-Cancel" variant="link" onClick={onCancel}>
            Cancel
          </Button>
        </ButtonGroup>
      </div>
    </>
  )
}

export default SkuQuotaCreate
