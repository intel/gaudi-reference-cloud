import React from 'react'
import { ErrorBoundaryLevel } from '../../utility/Enums'
import SomethingWentWrong from './SomethingWentWrong'
import NetworkError from './NetworkError'
import CannotGetInformation from './CannotGetInformation'
import idcConfig from '../../config/configurator'
import LogoutFromAnotherWindow from './LogoutFromAnotherWindow'
import { Container } from 'react-bootstrap'
import FooterMini from '../../components/footer/FooterMini'
import SingleTopNavBar from '../../components/header/SingleTopNavBar'

export const ErrorBoundaryLevelWrapper = ({ children }) => {
  return (
    <>
      <SingleTopNavBar />
      <Container className="siteContainer-no-toolbar container" role="main">
        <div className="sheet mt-s10">{children}</div>
      </Container>
      <FooterMini />
    </>
  )
}

export const ErrorBoundaryRouteLevelWrapper = ({ children }) => {
  return <div className="section mt-s10">{children}</div>
}

export default class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props)
    this.state = { hasError: false, errorMessage: '', errorCode: '', errorStatus: '', error: '' }
  }

  // eslint-disable-next-line no-unused-vars
  static getDerivedStateFromError(error) {
    // Update state so the next render will show the fallback UI.
    let errorMessage = ''
    let errorCode = ''
    let errorStatus = -1
    const isApiErrorWithErrorMessage = Boolean(error.response && error.response.data && error.response.data.message)
    if (isApiErrorWithErrorMessage) {
      errorMessage = error.response.data.message
      errorCode = error.response.data.code
      errorStatus = error.response.status
    } else {
      errorMessage = error.toString()
    }
    return { hasError: true, errorMessage, errorCode, errorStatus, error }
  }

  // eslint-disable-next-line no-unused-vars, n/handle-callback-err
  componentDidCatch(error, info) {
    // TODO Log error if user wants to send analytics
    // logErrorToMyService(error, info.componentStack);
  }

  clearError = () => {
    this.setState({ hasError: false, error: '' })
  }

  isNetworkError = () => {
    return (
      this.state.errorMessage.toLowerCase().indexOf('network error') !== -1 ||
      this.state.errorMessage.indexOf(`timeout of ${idcConfig.REACT_APP_AXIOS_TIMEOUT}ms exceeded`) !== -1
    )
  }

  isLogoutError = () => {
    return this.state.errorMessage.toLowerCase().indexOf('idc logout window') !== -1
  }

  render() {
    const { hasError } = this.state
    const { children, errorBoundaryLevel } = this.props
    if (hasError) {
      switch (errorBoundaryLevel) {
        case ErrorBoundaryLevel.AppLevel: {
          if (this.isNetworkError()) {
            return (
              <ErrorBoundaryLevelWrapper>
                <NetworkError />
              </ErrorBoundaryLevelWrapper>
            )
          }
          if (this.isLogoutError()) {
            return (
            <ErrorBoundaryLevelWrapper>
              <LogoutFromAnotherWindow />
            </ErrorBoundaryLevelWrapper>
            )
          }
          return (
            <ErrorBoundaryLevelWrapper>
              <SomethingWentWrong error={this.state.errorMessage} />
            </ErrorBoundaryLevelWrapper>
          )
        }
        case ErrorBoundaryLevel.RouteLevel:
          return (
            <ErrorBoundaryRouteLevelWrapper>
              <CannotGetInformation clearError={this.clearError} />
            </ErrorBoundaryRouteLevelWrapper>
          )
        default:
          return (
          <ErrorBoundaryLevelWrapper>
            <SomethingWentWrong />
          </ErrorBoundaryLevelWrapper>
        )
      }
    }

    return children
  }
}
