import { Button } from 'react-bootstrap'
import { useNavigate } from 'react-router-dom'
import { BsCopy } from 'react-icons/bs'
import { useCopy } from '../../../hooks/useCopy'
const ApiKeysView = ({ token, expirationDate, refreshKey, copyKey }) => {
  const { copyToClipboard } = useCopy()

  // Navigation
  const navigate = useNavigate()

  function backToHome() {
    navigate('/')
  }

  const spinner = (
    <div className="col-12 row mt-s2">
      <div className="spinner-border text-primary center"></div>
    </div>
  )

  const tokenContent = <div className="section code-line rounded-3 mt-s4">
      <div className="row mt-0 align-items-center gap-s4 ps-s4">
        <pre className="col-12">
          <span>{token}</span>
        </pre>
        <div className="col-12 mt-0">
          <Button variant="secondary" onClick={() => copyToClipboard(token)}>
            <BsCopy />
            Copy
          </Button>
        </div>
      </div>
    </div>

  return (
    <>
      <div className="section">
        <Button variant="link" className='p-s0' onClick={backToHome}>
          ⟵ Back to Home
        </Button>
      </div>
      <div className="filter">
        <div className="flex-fill bd-highlight">
          <span className='h4 pe-s4'>Key expiration date:</span>
          <span intc-id="api-key-expiration-date" className="lead">
            {expirationDate}
          </span>
        </div>
        <Button
          intc-id="api-key-refresh-key"
          variant='primary'
          onClick={refreshKey}
        >
          Refresh Key
        </Button>
      </div>
        {!token ? spinner : tokenContent}
    </>
  )
}

export default ApiKeysView
