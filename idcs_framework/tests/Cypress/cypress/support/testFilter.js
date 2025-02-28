/// <reference types="Cypress" />

const TestFilter = (definedTags, runTest) => {
  console.log('testFilter is called')
  const envTags = Cypress.env('testTags')
    if (typeof envTags === 'string') {
      const tags = Cypress.env('testTags').split(',');
      console.log('Tags are being verified')
      const isFound = definedTags.some((definedTag) => tags.includes(definedTag));
 
      if (isFound) {
        console.log('Tags matched')
        runTest();
      }
    } else if (Array.isArray(envTags)) {
      const isFound = definedTags.some((definedTag) => envTags.includes(definedTag));
      if (isFound) {
        runTest();
      }
    }
};
 
  export default TestFilter;
