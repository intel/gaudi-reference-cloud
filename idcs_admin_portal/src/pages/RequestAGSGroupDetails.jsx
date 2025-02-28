import { Button } from 'react-bootstrap'
import { useNavigate } from 'react-router-dom'

const RequestAGSGroupDetails = () => {
  // Navigation
  const navigate = useNavigate()

  const backToHome = () => {
    navigate('/')
  }
  return (
    <>
      <div className="section">
        <Button variant="link" className="p-s0" onClick={() => backToHome()}>
          ⟵ Back to Home
        </Button>
      </div>
      <div className="section">
        <h2 className="h4">Pre-Production AGS Entitlement</h2>
        <ul>
          <li>
            {' '}
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Administrator%20-%20Test"
              target="_blank"
              rel="noreferrer"
            >
              IDC Administrator - Test
            </a>
          </li>
          <li>
            {' '}
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20IKS%20-%20Test"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin IKS - Test
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20SRE%20-%20Test"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin SRE - Test
            </a>
          </li>
          <li>
            {' '}
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20Compute%20-%20Test"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin Compute - Test
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20Product%20-%20Test"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin Product - Test
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20Banners"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin Banner
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20Storage%20Quota%20-%20Test"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin Storage Quota- Test
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20Quota-%20Test"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin Quota- Test
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=ITAC%20Node%20Pools%20Admin%20-%20Test"
              target="_blank"
              rel="noreferrer"
            >
              ITAC Node Pools Admin - Test
            </a>
          </li>
        </ul>
      </div>
      <div className="section">
        <h2 className="h4">Production AGS Entitlement</h2>
        <ul>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Administrator%20-%20prod"
              target="_blank"
              rel="noreferrer"
            >
              IDC Administrator - Prod
            </a>{' '}
          </li>
          <li>
            {' '}
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20IKS%20-%20prod"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin IKS - Prod
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20SRE%20-%20prod"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin SRE - Prod
            </a>
          </li>
          <li>
            {' '}
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20Compute%20-%20prod"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin Compute - Prod
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20Product%20-%20prod"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin Product - Prod
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20Banners"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin Banner
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20Storage%20Quota%20-%20Prod"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin Storage Quota- Prod
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=IDC%20Admin%20Quota-%20Prod"
              target="_blank"
              rel="noreferrer"
            >
              IDC Admin Quota- Prod
            </a>
          </li>
          <li>
            <a
              href="https://ags.intel.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=ITAC%20Node%20Pools%20Admin%20-%20Prod"
              target="_blank"
              rel="noreferrer"
            >
              ITAC Node Pools Admin - Prod
            </a>
          </li>
        </ul>
      </div>
      <div className="section">
        <table className="table-bordered w-100">
          <thead>
            <tr>
              <td></td>
              <th scope="col">Home page</th>
              <th scope="col">Cloud Credits</th>
              <th scope="col">Cloud Account Instance Whitelist</th>
              <th scope="col">Terminate Instances</th>
              <th scope="col">Developer Tools</th>
              <th scope="col">Cloud Account Management</th>
              <th scope="col">IKS</th>
              <th scope="col">Banners</th>
              <th scope="col">Node Pools</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <th scope="row">IDC Administrator - (Test/Prod)</th>
              <td data-label="Home page only">X</td>
              <td data-label="Cloud Credits"></td>
              <td data-label="Cloud Account Instance Whitelist"></td>
              <td data-label="Terminate Instances"></td>
              <td data-label="Developer Tools"></td>
              <td data-label="Cloud Account Management"></td>
              <td data-label="IKS"></td>
              <td data-label="Banner Management"></td>
              <td data-label="Node Pools"></td>
            </tr>
            <tr>
              <th scope="row">IDC Admin SRE - (Test/Prod)</th>
              <td data-label="Home page only"></td>
              <td data-label="Cloud Credits">X</td>
              <td data-label="Cloud Account Instance Whitelist"></td>
              <td data-label="Terminate Instances">X</td>
              <td data-label="Developer Tools"></td>
              <td data-label="Cloud Account Management">X</td>
              <td data-label="IKS"></td>
              <td data-label="Banner Management"></td>
              <td data-label="Node Pools"></td>
            </tr>
            <tr>
              <th scope="row">IDC Admin IKS - (Test/Prod)</th>
              <td data-label="Home page only"></td>
              <td data-label="Cloud Credits">X</td>
              <td data-label="Cloud Account Instance Whitelist"></td>
              <td data-label="Terminate Instances"></td>
              <td data-label="Developer Tools">X</td>
              <td data-label="Cloud Account Management"></td>
              <td data-label="IKS">X</td>
              <td data-label="Banner Management"></td>
              <td data-label="Node Pools"></td>
            </tr>
            <tr>
              <th scope="row">IDC Admin Compute - (Test/Prod)</th>
              <td data-label="Home page only"></td>
              <td data-label="Cloud Credits">X</td>
              <td data-label="Cloud Account Instance Whitelist"></td>
              <td data-label="Terminate Instances">X</td>
              <td data-label="Developer Tools">X</td>
              <td data-label="Cloud Account Management">X</td>
              <td data-label="IKS"></td>
              <td data-label="Banner Management"></td>
              <td data-label="Node Pools"></td>
            </tr>
            <tr>
              <th scope="row">IDC Admin Product - (Test/Prod)</th>
              <td data-label="Home page only"></td>
              <td data-label="Cloud Credits">X</td>
              <td data-label="Cloud Account Instance Whitelist">X</td>
              <td data-label="Terminate Instances"></td>
              <td data-label="Developer Tools"></td>
              <td data-label="Cloud Account Management"></td>
              <td data-label="IKS"></td>
              <td data-label="Banner Management"></td>
              <td data-label="Node Pools"></td>
            </tr>
            <tr>
              <th scope="row">IDC Admin Banners</th>
              <td data-label="Home page only"></td>
              <td data-label="Cloud Credits"></td>
              <td data-label="Cloud Account Instance Whitelist"></td>
              <td data-label="Terminate Instances"></td>
              <td data-label="Developer Tools"></td>
              <td data-label="Cloud Account Management"></td>
              <td data-label="IKS"></td>
              <td data-label="Banner Management">X</td>
              <td data-label="Node Pools"></td>
            </tr>
            <tr>
              <th scope="row">ITAC Node Pools</th>
              <td data-label="Home page only"></td>
              <td data-label="Cloud Credits"></td>
              <td data-label="Cloud Account Instance Whitelist"></td>
              <td data-label="Terminate Instances"></td>
              <td data-label="Developer Tools"></td>
              <td data-label="Cloud Account Management"></td>
              <td data-label="IKS"></td>
              <td data-label="Banner Management"></td>
              <td data-label="Node Pools">X</td>
            </tr>
          </tbody>
        </table>
      </div>
    </>
  )
}

export default RequestAGSGroupDetails
