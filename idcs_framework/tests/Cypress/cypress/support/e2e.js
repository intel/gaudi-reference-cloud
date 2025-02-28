// ***********************************************************
// This example support/index.js is processed and
// loaded automatically before your test files.
//
// This is a great place to put global configuration and
// behavior that modifies Cypress.
//
// You can change the location of this file or turn off
// automatically serving support files with the
// 'supportFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/configuration
// ***********************************************************

// Import commands.js using ES2015 syntax:
import './azureLogin'
import './commands'
require('cypress-xpath')
import "cypress-ntlm-auth/dist/commands";
import 'cypress-mailslurp';
// Alternatively you can use CommonJS syntax:
// require('./commands')

// new function that takes 3 arguments
const createCustomErrorMessage = (error, steps, runnableObj) => {
    // Let's generate the list of steps from an array of strings
    let lastSteps = "Last logged steps:\n"
    steps.map((step, index) => {
      lastSteps += `${index + 1}. ${step}\n`
    })
  
      // I decided to keep the following as an array
    //   for easier customization. But basically in the end
    //   I'll be building the text from the array by combining those
    //   and adding new line at the end
    const messageArr = [
      `Context: ${runnableObj.parent.title}`, // describe('...')
      `Test: ${runnableObj.title}`, // it('...')
      `----------`,
      `${error.message}`, // actual Cypress error message
      `\n${lastSteps}`, // additional empty line to get some space
                        //   and the list of steps generated earlier
    ]
  
    // Return the new custom error message
    return messageArr.join('\n')
  }


  // When the test fails, run this function
Cypress.on('fail', (err, runnable) => {

    // let's store the error message by creating it using our
    //   custom function we just made earlier. We need to pass
    //   "err" and "runnable" that we get from Cypress test fails
    //   but we have to remember to also pass our steps. In case
    //   no steps were provided we have to provide either empty string
    //   or some form of a message to help understand what's going on.
    const customErrorMessage = createCustomErrorMessage(
      err,
      Cypress.env('step') || ['no steps provided...'],
      runnable,
    )
    
    // Our custom error will now be defaulted back to the
    //   original default Cypress Error
    const customError = err
  
    // BUT we will change the message we're presenting to our custom one
    customError.message = customErrorMessage
  
    // aaaand let's throw that error nicely
    throw customError
  })

