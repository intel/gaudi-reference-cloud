# Intel Tiber AI Cloud Admin Console

This repository contains the React application for the Intel Tiber AI Cloud Admin Console.

## Recommended IDE Setup

[VSCode](https://code.visualstudio.com/) + [Prettier](https://marketplace.visualstudio.com/items?itemName=esbenp.prettier-vscode)

To read this document as in Github, using vsCode just use ctrl+shift+v.

## Getting Started

### Available Scripts

In the project directory, you can run:

#### `npm install`

Install all pre-requiered packages

#### `npm run start`

Runs the app in the development mode.\
Open [http://localhost:3000](http://localhost:3000) to view it in your browser.

## Coding Standards

In you are new developer, or want to refresh the coding standards for this project please take a look at this section.

### Design system IDC 2.0

For new IDC 2.0 related US we use the [Intel Tiber AI Cloud Design system](../idcs_design_system/README.md) this is based on bootstrap 5.3.

For IDC 2.0 following rules should be follow:

* No css or sass in allowed, all should be managed through the design system. Exceptions are non shared component specific styles that should be in a sass file in the same folder of the component file with the same name.

### Coding rules set

ESlint is configured with two plugings that enforces its rules as the default sytanx of the project:

* [eslint-plugin-react](https://www.npmjs.com/package/eslint-plugin-react)
* [eslint-config-standard-with-typescript](https://github.com/standard/eslint-config-standard-with-typescript)

To review syntax problems just use athe command `npm run lint` if you want to fix the issues automatically us `npm run lint -- --fix` note that not all the issues can be fixed automatically

You cannot commit your code unless the lints are passing, because the project has a pre-commit that review if the lints are passing.

### Application State Management

For application state management we recommend to use [Zustand](https://github.com/pmndrs/zustand). Use the [immer middleware](https://docs.pmnd.rs/zustand/integrations/immer-middleware) in case you need to update complex objects or inmmutable state. The prefer way is to use TypeScript for our stores that way we can use interfaces to declare the shape of our state.

### Error handling

There are two scenarios that we faced when talking about error handling:

1. Managed exceptions: These are exceptions that we want to manage and show a non generic message to the user, also sometimes we want the user to do something specific. For these use cases use a normal try/catch and do the logic in the catch block, just remember to use the error modals if you want to display some errors, ask the UX Designer to know more on this.

2. Unmanaged exceptions: These are the exceptions that we don't want to manage and let the application control and manage the error. For this use case we use an ErrorBounday. The ErrorBounday Works very well with synchornous code, but when you want your asyncrhonous code or handler exceptions be catch by the ErrorBoundary you need to use the useErrorBoundary hook. This is an example when calling an async function inside an useEffect or in an eventHandler:

```js
/**
 * Calling a normal async function
 */
import useErrorBoundary from "../../hooks/useErrorBoundary";
...
const throwError = useErrorBoundary();
...
useEffect(() => {
        const fetchProducts = async () => {
            try {
                if (products === null) {
                    await setProducts();
                }
            } catch (error) {
                throwError(error);
            }
        };

        fetchProducts();
    }, [products]);

```

```js
/**
 * Calling several async functions using Promise.all
 */

import useErrorBoundary from "../../hooks/useErrorBoundary";
...
const throwError = useErrorBoundary();
...
const fetchData = async() => {
    const promises = [
        machineOs === null ? setMachineOs : new Promise(),
        publicKeys === null ? setPublickeys : new Promise(),
        products === null ? setProducts : new Promise()
    ];
    try {
        await Promise.all(promises);
    } catch (error) {
        throwError(error);
    }
};
...

useEffect(() => {
    fetchData();
}, [machineOs, publicKeys, products]);

```

```js
/**
 * Unmanaged error inside event handler
 */

import useErrorBoundary from "../../hooks/useErrorBoundary";
...
const throwError = useErrorBoundary();
...
  const handler = () => {
    try {
      functionThatThrowsError();
    } catch (error) {
      throwError(error);
    }
  }
...

return (
    <>
    ...
     <Button variant="primary" onClick={handler}>
        My Button
    </Button>
    ...
    </>
)

```
### HTML and SCSS code smells
The following are practices to avoid in HTML and SCSS. Correct practices help achieve a crisper look and feel.
#### HTML smells
1. Inline CSS: Applying styles directly within HTML tags.
2. Inline Event Handlers: Using inline JavaScript event handlers within HTML tags.
3. Overuse of divs: Excessive usage of div elements without semantic meaning.
4. Deprecated Attributes: Using outdated or non-standard HTML attributes.
5. Missing Alt Attributes: Omitting alt attributes on img tags for accessibility.
6. Adding `<br>` elements in HTML for layout purposes is not recommended. The `<br>` element is intended for inserting line breaks within text content. It is not meant for creating complex layout structures. Instead, it is better to use appropriate CSS techniques for layout, such as using margin, padding, flexbox, or grid layouts.


#### SCSS smells
1. Selector Overuse: Having overly complex or nested selectors.
2. `!important` Overuse: Relying too heavily on the !important declaration.
3. Inline Styles: Applying styles directly within HTML tags.
4. Magic Numbers: Hardcoding specific pixel values instead of using variables or calculations.
5. Unused Styles: Having CSS rules that are no longer used in the codebase.

#### Bootstrap best practices
1. When assigning color to a property, use one of the bootstrap color variables instead of using a `hex` code or `color()` mix.
2. When using `col-x` classes, use always the `col-md-x` breakpoint as default.
3. Avoid nesting divs with `row` as class. In bootstrap, row classes add extra margin and padding to elements, possibly adding unnecessary whitespace to the layout.
3. As a general rule, Columns within a `row` are meant only to contain a horizontal set of components, instead of adding multiple `col-12` divs to simulate several rows, create multiple row divs.
4. Learn about [flex](https://getbootstrap.com/docs/5.1/utilities/flex/) and make it your friend.

## Azure Authentication

This application uses Azure B2B in order to authenticate the user. Below is the information for both environments.

### Azure B2C Local, Development and Staging information

- ApplicationId: [IDC Admin Console Development](https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationMenuBlade/~/Overview/appId/5cf0ff86-a5ea-47ab-a3fc-29934041d044/isMSAApp~/false)

### Azure B2C production information

- ApplicationId: [IDC Admin Console Production](https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationMenuBlade/~/Overview/appId/8d696d04-e138-4c61-bd52-a0a25f9e1d05/isMSAApp~/false)

## Build and environment configuration

### Change configuration variables

The configuration variables are located in a filed called configMap.json, the variables are loaded every time the user load the page and are stored in memory, therefore do not put any sensitive data on configuration variables. Depends on the deployment environment the process to update have small differences, the environments that uses docker images and helm deployment will build the configMap on the helm deployment and mount it as a volume so update something will required a redeploy, on static deployment the configMap is just a static file that can be create and updated onDemand.

Following files must be updated when you add a new config variable:

* [public/configMap.json](./public/configMap.json) : This is the configMap that is taken for local development, when you do npm run start.
* [src/config/configurator.js](./src/config/configurator.js) : When you add a new config variable remember to update the typeDef on this file and put a meaningful description, it will allow to do intelligence to that variable on the project.
* [configmapreference/configMap.example.json](./configmapreference/configMap.example.json): This is the configMap example for reference, please keep it updated.
* [charts/values.yaml](./charts/values.yaml): When you add a new config variable you must add here as well, it contains the default values and is used to create the configMap for docker images.

### Build and release to any environment

To do a Production ready Migration, please follow up the next steps:

1. Run the `npm run build` command to generate a Production build, that would create a Production ready build in the ./build folder.
2. Prepare a configmap.json file for your environment and paste in the ./build folder.
3. Push the contents of the ./build server to your prefered server.