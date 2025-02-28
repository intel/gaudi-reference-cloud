const auth = require("../fixtures/IDC2.0/login");

Cypress.Commands.add("PrepareSession", () => {
  if (Cypress.env("testTags").includes("IntelAll")) {
    cy.loginWithSessionIntel("Intel");
  } else if (Cypress.env("testTags").includes("PremiumAll")) {
    cy.loginWithSessionPremium("Premium");
  } else if (Cypress.env("testTags").includes("StandardAll")) {
    cy.loginWithSessionStandard("Standard");
  } else if (Cypress.env("testTags").includes("ependingAll")) {
    cy.loginWithSessionEnterprisePending("EnterprisePending");
  } else if (Cypress.env("testTags").includes("EnterpriseAll")) {
    cy.loginWithSessionEnterprise("Enterprise");
  }
});

Cypress.Commands.add("GetSession", () => {
  if (Cypress.env("testTags").includes("IntelAll")) {
    Cypress.env('accountType', "Intel");
    cy.loginWithSessionIntel("Intel");
  } else if (Cypress.env("testTags").includes("PremiumAll")) {
    Cypress.env('accountType', "Premium");
    cy.loginWithSessionPremium("Premium");
  } else if (Cypress.env("testTags").includes("StandardAll")) {
    Cypress.env('accountType', "Standard");
    cy.loginWithSessionStandard("Standard");
  } else if (Cypress.env("testTags").includes("ependingAll")) {
    Cypress.env('accountType', "EnterprisePending");
    cy.loginWithSessionEnterprisePending("EnterprisePending");
  } else if (Cypress.env("testTags").includes("EnterpriseAll")) {
    Cypress.env('accountType', "Enterprise");
    cy.loginWithSessionStandard("Enterprise");
  }
  cy.visit(Cypress.env("baseUrl"));
  cy.wait(2000);
});

Cypress.Commands.add("GetAdminSession", () => {
  cy.clearCookies();
  cy.clearAllLocalStorage();
  cy.loginWithSessionAdmin("admin");
  cy.wait(2000);
});

Cypress.Commands.add("TestClean", () => {
  cy.visit("./index.html"); // Go to index.html to avoid enter infinite loop of logging after tests finish
  cy.wait(100);
});

Cypress.Commands.add("loginWithSessionIntel", (sesionName) => {
  cy.session(
    sesionName,
    () => {
      cy.loginAD(Cypress.env("iuser"), Cypress.env("iupass"));
      cy.on("uncaught:exception", (e) => {
        console.error(e);
        return false;
      });
    },
    {
      cacheAcrossSpecs: true,
    }
  );
});

Cypress.Commands.add("loginWithSessionIntel2", (sesionName) => {
  cy.session(
    sesionName,
    () => {
      cy.loginAD(Cypress.env("iuser2"), Cypress.env("iupass2"));
      cy.on("uncaught:exception", (e) => {
        console.error(e);
        return false;
      });
    },
    {
      cacheAcrossSpecs: true,
    }
  );
});

Cypress.Commands.add("loginWithSessionPremium", (sesionName) => {
  cy.session(
    sesionName,
    () => {
      cy.loginADExternal(Cypress.env("puser"), Cypress.env("pupass"));
      cy.on("uncaught:exception", (e) => {
        console.error(e);
        return false;
      });
    },
    {
      cacheAcrossSpecs: true,
    }
  );
});

Cypress.Commands.add("loginWithSessionStandard", (sesionName) => {
  cy.session(
    sesionName,
    () => {
      cy.loginADExternal(Cypress.env("suser"), Cypress.env("supass"));
      cy.on("uncaught:exception", (e) => {
        console.error(e);
        return false;
      });
    },
    {
      cacheAcrossSpecs: true,
    }
  );
});

Cypress.Commands.add("loginWithSessionEnterprisePending", (sesionName) => {
  cy.session(
    sesionName,
    () => {
      cy.loginADExternal(Cypress.env("epuser"), Cypress.env("eppass"));
      cy.on("uncaught:exception", (e) => {
        console.error(e);
        return false;
      });
    },
    {
      cacheAcrossSpecs: true,
    }
  );
});

Cypress.Commands.add("loginWithSessionEnterprise", (sesionName) => {
  cy.session(
    sesionName,
    () => {
      cy.loginADExternal(Cypress.env("euser"), Cypress.env("eupass"));
      cy.on("uncaught:exception", (e) => {
        console.error(e);
        return false;
      });
    },
    {
      cacheAcrossSpecs: true,
    }
  );
});

Cypress.Commands.add("loginAD", (user, pass) => {
  var authURL =
    auth.azureProd.authority + "client_id=" + auth.azureProd.clientId + "&scope=" + auth.azureProd.scope + "&redirect_uri=" + auth.azureProd.redirect + "&" + auth.azureProd.other;
  cy.visit("./index.html"); // Workaround for Cypress issue with origins needed for Azure B2C SSO
  cy.origin(
    "https://consumer.intel.com",
    {
      args: {
        user,
        authURL,
      },
    },
    ({ user, authURL }) => {
      cy.visit(authURL);
      cy.get('input[type="email"]').should("be.visible").type(user);
      cy.wait(2000);
      cy.get('.btngroupitm.textleftalign').contains("Employee Sign In")
        .should("be.visible")
        .click({ force: true });
      cy.wait(100);
      Cypress.on("uncaught:exception", (_err, _runnable, _promise) => {
        return false;
      });
    }
  );
  cy.origin(
    "https://login.microsoftonline.com",
    {
      args: {
        pass,
      },
    },
    ({ pass }) => {
      cy.get('input[type="password"]').should("be.visible").type(pass);
      cy.wait(50);
      cy.get('input[type="submit"]')
        .contains("Sign in")
        .should("be.visible")
        .click({ force: true });
      cy.wait(12000);
      Cypress.on("uncaught:exception", (_err, _runnable) => {
        return false
      });
    }
  );
});

Cypress.Commands.add("loginADExternal", (user, pass) => {
  var authURL = auth.azureProd.authority + "client_id=" + auth.azureProd.clientId + "&scope=" + auth.azureProd.scope + "&redirect_uri=" + auth.azureProd.redirect + "&" + auth.azureProd.other;
  cy.visit("./index.html"); // Workaround for Cypress issue with origins needed for Azure B2C SSO
  cy.origin(
    "https://consumer.intel.com",
    {
      args: {
        user,
        pass,
        authURL,
      },
    },
    ({ user, pass, authURL }) => {
      cy.visit(authURL);
      cy.wait(500);
      cy.get('input[type="email"]').should("be.visible").type(user);
      cy.wait(2000);
      cy.get('button[type="submit"]')
        .should("be.visible")
        .click({ force: true });
      cy.wait(100);
      cy.get('input[type="password"]').should("be.visible").type(pass);
      cy.wait(50);
      cy.get("#continue")
        .contains("Sign In")
        .should("be.visible")
        .click({ force: true });
      cy.wait(12000);
      cy.on("uncaught:exception", (_err, _runnable, _promise) => {
        return false;
      });
    }
  );
});

Cypress.Commands.add("signout", () => {
  cy.get('[id="dropdown-header-user-toggle"]').click({ multiple: true });
  cy.get('[intc-id="signOutHeaderButton"]').click({ force: true });
});

Cypress.Commands.add("loginWithSessionAdmin", (sesionName) => {
  cy.session(
    sesionName,
    () => {
      cy.loginAdmin();
      cy.on("uncaught:exception", (e) => {
        console.error(e);
        return false;
      });
    },
    {
      cacheAcrossSpecs: false,
    }
  );
});

Cypress.Commands.add("loginAdmin", () => {
  //Login via Azure AD & Set Local Storage
  const options = {
    method: "POST",
    url: auth.adminConsole.tokenEndPoint,
    form: true,
    failOnStatusCode: false,
    body: {
      grant_type: "password",
      client_id: auth.adminConsole.clientId,
      scope: auth.adminConsole.scope,
      client_secret: auth.adminConsole.clientSecret,
      client_info: 1, // returns an extra token that MSAL needs
      //redirectUri: auth.adminConsole.redirect,
      username: Cypress.env("adminuser"),
      password: Cypress.env("adminpass"),
    },
  };
  cy.request(options).then((response) => {
    const token = response.body.access_token;
    Cypress.env('token', token);
    console.log("token");
    response.body.type = "azure";
    const authValue = JSON.stringify(response.body);
    window.localStorage.setItem("auth", authValue);
  });
});
Cypress.Commands.add("saveLocalStorageCacheFirstTimeLogin", () => {
  //cy.log('save')
  localStorage.setItem("completedSteps", "9");
  localStorage.setItem("firstRunStatus", "Completed");
  cy.reload();
  Object.keys(localStorage).forEach((key) => {
    LOCAL_STORAGE_MEMORY[key] = localStorage[key];
  });

  const token = localStorage.getItem("authToken").slice(1, -1);
  Cypress.env("token", token);
});
Cypress.Commands.add("saveLocalStorageCache", () => {
  cy.log("save");
  Object.keys(localStorage).forEach((key) => {
    LOCAL_STORAGE_MEMORY[key] = localStorage[key];
  });
});
Cypress.Commands.add("restoreLocalStorageCache", () => {
  cy.log("restore");
  Object.keys(LOCAL_STORAGE_MEMORY).forEach((key) => {
    localStorage.setItem(key, LOCAL_STORAGE_MEMORY[key]);
  });
});
Cypress.Commands.add("clearLocalStorageCache", () => {
  localStorage.clear();
  LOCAL_STORAGE_MEMORY = {};
});