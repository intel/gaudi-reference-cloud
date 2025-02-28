import React from 'react'
import CodeLine from '../../../utility/CodeLine'
import idcConfig from '../../../config/configurator'

const HowToCreateSSHKeyLinux = (props) => {
  return (
    <>
      <ol className='pt-3' style={{ paddingInlineStart: '1rem' }}>
        <li className="mb-2">Launch a Terminal on your local system.</li>
        <li className="mb-2">Copy & paste the following to your terminal to generate SSH Keys.
          <CodeLine codeline={'ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa'}/>
        </li>
        <li className="mb-2">If you are prompted to overwrite, select <strong className="text-danger">No</strong>.</li>
        <li className="mb-2">Copy & paste the following to your terminal to open your SSH key:
          <CodeLine codeline={'vi ~/.ssh/id_rsa.pub'}/>
        </li>
        <li>Upload the generated file</li>
      </ol>
      <p className="mb-0">For more information go to <a rel="noreferrer" href={idcConfig.REACT_APP_SHH_KEYS} className="text-decoration-none" target="_blank">SSH key documentation</a>.</p>
    </>
  )
}

export default HowToCreateSSHKeyLinux
