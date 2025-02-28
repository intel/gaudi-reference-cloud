# Intel Tiber AI Cloud Registration

## Getting Started

### HTML Local Development

To test the HTML templates locally follow these steps:

First install npm modules using ```npm install```

Deploy the HTML templates for local test ```npm run build-html-local```

Run your local server ```npm run server-html```

Use one of the different LOCAL policies to test your changes, refresh after each change to reflect the change.

To reflect any change you need to call again ```npm run build-html-local```

If you need to update the policies for local development please run ```npm run build-local``` and upload the files to B2C.

This is the entry URL to test local HTML:

https://auth-dev.tiberaicloud.com/IDCB2CDEV.onmicrosoft.com/oauth2/v2.0/authorize?p=B2C_1A_LOCAL_NEW_SIGNUP&client_id=e4e51a15-c4c5-4102-89a9-918ce8297112&nonce=defaultNonce&redirect_uri=https%3A%2F%2Fjwt.ms%2F&scope=openid&response_type=id_token

### Available Scripts

