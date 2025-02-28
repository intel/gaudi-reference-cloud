// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor, within } from '@testing-library/react'
import { act } from 'react'
import { Router } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import userEvent from '@testing-library/user-event'

import { mockStandardUser } from '../../mocks/authentication/authHelper'
import { clearAxiosMock, expectValueForInputElement, clearInputElement } from '../../mocks/utils'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockFailedCreatePublicKeys } from '../../mocks/publicKeys/publicKeys'

import ImportKeysConstainer from '../../../containers/keypairs/ImportKeysConstainer'
import idcConfig from '../../../config/configurator'
import ToastContainer from '../../../utils/toast/ToastContainer'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <ImportKeysConstainer />
        <ToastContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('SSH Public Keys container - Create form', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
    mockStandardUser()
    global.navigator.clipboard = { writeText: jest.fn() }
  })

  it('Render component if loads correctly', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByLabelText('Upload key')).toBeInTheDocument()

    const cancelLink = screen.getByTestId('btn-ssh-cancelPublicKey')
    expect(cancelLink).toBeInTheDocument()
    expect(cancelLink).toHaveTextContent('Cancel')
    expect(cancelLink).toHaveAttribute('href', '/security/publickeys/')
  })

  it('Page should have a warning message as default', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    const alertMessage =
      'Never share your private keys with anyone. Never create a SSH Private key without a passphrase'
    expect(await screen.findByText(alertMessage)).toBeInTheDocument()
  })

  it('Page should have a accrodion to display how to create key information', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const accordionDiv = screen.getByTestId('howToCreateSSHKeyAccordion')
    expect(accordionDiv).toBeInTheDocument()

    const accHeader = accordionDiv.getElementsByClassName('accordion-header')
    expect(accHeader[0]).toBeInTheDocument()

    const headerButton = accHeader[0].getElementsByTagName('button')
    expect(headerButton[0]).toHaveTextContent('How to create a SSH key')

    const contentDiv = accHeader[0].nextElementSibling
    expect(contentDiv).toHaveClass('accordion-collapse collapse')

    await userEvent.click(headerButton[0])

    const contentDivWin = accHeader[0].nextElementSibling

    await waitFor(() => {
      expect(contentDivWin).toHaveClass('accordion-collapse collapse show')
    })

    expect(await within(contentDivWin).findByText('Select your OS:')).toBeInTheDocument()
    expect(await within(contentDivWin).findByText('SSH key documentation')).toHaveAttribute(
      'href',
      idcConfig.REACT_APP_SHH_KEYS
    )

    const windowsRadioButton = within(contentDivWin).getByTestId('WindowsRadioButton')
    expect(windowsRadioButton).toBeInTheDocument()

    expect(windowsRadioButton).toHaveAttribute('checked')
    expect(
      await within(contentDivWin).findByText('Launch a new PowerShell window on your local system.')
    ).toBeInTheDocument()

    const preCodeWin = contentDivWin.getElementsByClassName('code-line')
    expect(preCodeWin.length).toEqual(3)

    expect(preCodeWin[0]).toHaveTextContent('mkdir $env:UserProfile\\.ssh')
    expect(preCodeWin[1]).toHaveTextContent('ssh-keygen -t rsa -b 4096 -f $env:UserProfile\\.ssh\\id_rsa')
    expect(preCodeWin[2]).toHaveTextContent('cat $env:UserProfile\\.ssh\\id_rsa.pub')

    const copyButton1Win = preCodeWin[0].getElementsByTagName('button')[0]
    expect(copyButton1Win).toHaveTextContent('Copy')

    const copyButton2Win = preCodeWin[1].getElementsByTagName('button')[0]
    expect(copyButton2Win).toHaveTextContent('Copy')

    const copyButton3Win = preCodeWin[2].getElementsByTagName('button')[0]
    expect(copyButton3Win).toHaveTextContent('Copy')

    await userEvent.click(copyButton1Win)
    expect(navigator.clipboard.writeText).toHaveBeenCalledTimes(1)
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('mkdir $env:UserProfile\\.ssh')

    await userEvent.click(copyButton2Win)
    expect(navigator.clipboard.writeText).toHaveBeenCalledTimes(2)
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      'ssh-keygen -t rsa -b 4096 -f $env:UserProfile\\.ssh\\id_rsa'
    )

    await userEvent.click(copyButton3Win)
    expect(navigator.clipboard.writeText).toHaveBeenCalledTimes(3)
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('cat $env:UserProfile\\.ssh\\id_rsa.pub')

    // Linux
    const contentDivLinux = accHeader[0].nextElementSibling
    const linuxRadioButton = within(contentDivLinux).getByTestId('LinuxRadioButton')
    expect(linuxRadioButton).toBeInTheDocument()

    await waitFor(() => {
      expect(contentDivWin).toHaveClass('accordion-collapse collapse show')
    })
    await userEvent.click(linuxRadioButton)
    expect(await within(contentDivLinux).findByText('Launch a Terminal on your local system.')).toBeInTheDocument()

    const preCodeLinux = contentDivLinux.getElementsByClassName('code-line')
    expect(preCodeLinux.length).toEqual(2)

    expect(preCodeLinux[0]).toHaveTextContent('ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa')
    expect(preCodeLinux[1]).toHaveTextContent('cat ~/.ssh/id_rsa.pub')

    const copyButton1Linux = preCodeLinux[0].getElementsByTagName('button')[0]
    expect(copyButton1Linux).toHaveTextContent('Copy')

    const copyButton2 = preCodeLinux[1].getElementsByTagName('button')[0]
    expect(copyButton2).toHaveTextContent('Copy')

    await userEvent.click(copyButton1Linux)
    expect(navigator.clipboard.writeText).toHaveBeenCalledTimes(4)
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa')

    await userEvent.click(copyButton2)
    expect(navigator.clipboard.writeText).toHaveBeenCalledTimes(5)
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('cat ~/.ssh/id_rsa.pub')
  })

  it('Key Name input should display error while entering characters except alphabets (Lower), numbers and Hyphen(-).', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const keyNameInput = screen.getByTestId('KeyNameInput')
    await expectValueForInputElement(keyNameInput, 'testA123', 'testA123')
    const keyNameInvalidMessage = screen.getByTestId('KeyNameInvalidMessage')
    await waitFor(() => {
      expect(keyNameInvalidMessage).toHaveTextContent(
        'Only lower case alphanumeric and hypen(-) allowed for Key Name: *.'
      )
    })
  })

  it('Key Name input should display error while entering characters Hyphen(-) on first place.', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const keyNameInput = screen.getByTestId('KeyNameInput')
    await expectValueForInputElement(keyNameInput, '-test123', '-test123')
    const keyNameInvalidMessage = screen.getByTestId('KeyNameInvalidMessage')
    await waitFor(() => {
      expect(keyNameInvalidMessage).toHaveTextContent(
        'Only lower case alphanumeric and hypen(-) allowed for Key Name: *.'
      )
    })
  })

  it('Key Name input should display error while entering characters Hyphen(-) on last place.', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const keyNameInput = screen.getByTestId('KeyNameInput')
    await expectValueForInputElement(keyNameInput, 'test123-', 'test123-')
    const keyNameInvalidMessage = screen.getByTestId('KeyNameInvalidMessage')
    await waitFor(() => {
      expect(keyNameInvalidMessage).toHaveTextContent(
        'Only lower case alphanumeric and hypen(-) allowed for Key Name: *.'
      )
    })
  })

  it('Key Name input should display error if left blank after entering the value.', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const keyNameInput = screen.getByTestId('KeyNameInput')
    await expectValueForInputElement(keyNameInput, 'test123', 'test123')
    await clearInputElement(keyNameInput)
    const keyNameInvalidMessage = screen.getByTestId('KeyNameInvalidMessage')
    await waitFor(() => {
      expect(keyNameInvalidMessage).toHaveTextContent('Key Name is required')
    })
  })

  it('First Name input should only characters upto 30 characters.', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const keyNameInput = screen.getByTestId('KeyNameInput')
    const inputValue = 'test123test123test123testt112345testtest' // 40 characters
    const expectValue = 'test123test123test123testt1123' // 30 Characters
    await expectValueForInputElement(keyNameInput, inputValue, expectValue)
    await waitFor(() => {
      expect(keyNameInput).toHaveAttribute('maxlength', '30')
    })
  })

  it('Key Contents input should display error if left blank after entering the value.', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const keyContentsInput = screen.getByTestId('PasteyourkeycontentsTextArea')
    await expectValueForInputElement(keyContentsInput, 'test123', 'test123')
    await clearInputElement(keyContentsInput)
    const keyContentsInvalidMessage = screen.getByTestId('PasteyourkeycontentsInvalidMessage')
    await waitFor(() => {
      expect(keyContentsInvalidMessage).toHaveTextContent('key contents is required')
    })
  })

  it('Upload Button should be enabled after all the correct values', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const createPublicKeyButton = screen.getByTestId('btn-ssh-createpublickey')
    await waitFor(() => {
      expect(createPublicKeyButton).toBeVisible()
    })

    await expectValueForInputElement(screen.getByTestId('KeyNameInput'), 'test1', 'test1')
    await expectValueForInputElement(screen.getByTestId('PasteyourkeycontentsTextArea'), 'test content', 'test content')

    const createPublicKeyButtonNew = screen.getByTestId('btn-ssh-createpublickey')
    await waitFor(() => {
      expect(createPublicKeyButtonNew).toBeEnabled()
    })
  })

  it('Clicking on upload button with wrong ssh key should return an error', async () => {
    const errorText = 'SshPublicKey should have at least algorithm and publickey'
    mockFailedCreatePublicKeys(errorText)
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    await expectValueForInputElement(screen.getByTestId('KeyNameInput'), 'test1', 'test1')
    await expectValueForInputElement(screen.getByTestId('PasteyourkeycontentsTextArea'), 'test content', 'test content')

    const createPublicKeyButton = screen.getByTestId('btn-ssh-createpublickey')
    await userEvent.click(createPublicKeyButton)

    expect(await screen.findByText(errorText)).toBeVisible()
  })

  it('Clicking on upload button with duplicate should return an error', async () => {
    const errorText = 'Duplicate keypair name'
    mockFailedCreatePublicKeys(errorText)
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    await expectValueForInputElement(screen.getByTestId('KeyNameInput'), 'test1', 'test1')
    await expectValueForInputElement(screen.getByTestId('PasteyourkeycontentsTextArea'), 'test content', 'test content')

    const createPublicKeyButton = screen.getByTestId('btn-ssh-createpublickey')
    await userEvent.click(createPublicKeyButton)

    expect(await screen.findByText(errorText)).toBeVisible()
  })
})
