import React from 'react'
import { Nav, NavDropdown } from 'react-bootstrap'
import { Link } from 'react-router-dom'

const Breadcrumb = (props) => {
  let isDashboardDisabled = null
  let isMonitoringDisabled = null
  let isBillingPageClass = null

  if (props.activePage === 'dashboard') isDashboardDisabled = true
  if (props.activePage === 'monitoring') isMonitoringDisabled = true
  if (props.activePage === 'billing') isBillingPageClass = 'text-muted'

  return (
        <Nav activeKey="dropdrop">
            <Nav.Item>
                <Nav.Link disabled={isDashboardDisabled} as={Link} to="/">Active</Nav.Link>
            </Nav.Item>

            <NavDropdown title={<span className={isBillingPageClass}>Billing</span>} >
              <NavDropdown.Item as={Link} to="/billing/cloudcredits">Cloud Credits</NavDropdown.Item>
              <NavDropdown.Item>Internal Billing</NavDropdown.Item>
              <NavDropdown.Item>Customer Billing</NavDropdown.Item>
            </NavDropdown>

            <Nav.Item>
                <Nav.Link disabled={isMonitoringDisabled} as={Link} to="/monitoring">Monitoring</Nav.Link>
            </Nav.Item>
        </Nav>
  )
}

export default Breadcrumb
