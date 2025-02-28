# Contributing

## Guidelines

If you want to contribute to this project please follow these guidelines.

* All components should build with accesibility and responsiveness in mind.
* All code must be written in typescript and scss.
* Themes should be upgraded only by designers.
* Respect project structure and patterns. Each component should have:
    - Well documented interfaces with jsdocs.
    - Test file to ensure component functionality work as expected.
    - Easy to read and mantain component code.
    - Storybook as a medium of showcase and living documentation for  the users of the design system.
* Before making any change please review it with designers and lead developer.
* Once PR is sent is subject to designers and lead developer approval.
* For versioning please validate if change is breaking change, backward compatibility change or a patch and adjust using [Semantic Versioning Guidelines](https://semver.org/).

## Environment setup

In order to keep code style and use the autoformat and code styling rules as intended please use the following environment:

[VSCode](https://code.visualstudio.com/) + [Prettier](https://marketplace.visualstudio.com/items?itemName=esbenp.prettier-vscode) + [eslint](https://marketplace.visualstudio.com/items?itemName=dbaeumer.vscode-eslint).


## Run the project

To run the project just run `npm run storybook`.