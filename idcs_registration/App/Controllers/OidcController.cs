// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Options;
using Microsoft.IdentityModel.Tokens;
using Newtonsoft.Json;
using System.Security.Cryptography.X509Certificates;

namespace b2c.helper.api.Controllers
{
    [ApiExplorerSettings(IgnoreApi = true)]
    [Route("[controller]")]
    public class OidcController : Controller
    {
        private static Lazy<X509SigningCredentials> SigningCredentials;
        private readonly Microsoft.AspNetCore.Hosting.IHostingEnvironment HostingEnvironment;
        private readonly AppSettingsModel AppSettings;
        bool validOnly = false;

        // Sample: Inject an instance of an AppSettingsModel class into the constructor of the consuming class, 
        // and let dependency injection handle the rest
        public OidcController(IOptions<AppSettingsModel> appSettings, Microsoft.AspNetCore.Hosting.IHostingEnvironment hostingEnvironment)
        {
            this.HostingEnvironment = hostingEnvironment;
            this.AppSettings = appSettings.Value;


            //// Sample: Load the certificate with a private key (must be pfx file)
            //SigningCredentials = new Lazy<X509SigningCredentials>(() =>
            //{
            //    var bytes = System.IO.File.ReadAllBytes(hostingEnvironment.ContentRootPath + "\\cert.pfx");
            //    var cert = new X509Certificate2(bytes, appSettings.Value.CertificatePassword);
            //    return new X509SigningCredentials(cert);
            //});

            using (X509Store certStore = new X509Store(StoreName.My, StoreLocation.CurrentUser))
            {
                certStore.Open(OpenFlags.ReadOnly);

                X509Certificate2Collection certCollection = certStore.Certificates.Find(
                    X509FindType.FindByThumbprint,
                    // Replace below with your certificate's thumbprint
                    AppSettings.CertificateThumbPrint,
                    validOnly);
                // Get the first cert with the thumbprint
                X509Certificate2 cert = certCollection.OfType<X509Certificate2>().FirstOrDefault();

                if (cert is null)
                    throw new Exception($"Certificate with thumbprint {AppSettings.CertificateThumbPrint} was not found");

                // Use certificate
                Console.WriteLine(cert.FriendlyName);
                SigningCredentials = new Lazy<X509SigningCredentials>(() => new X509SigningCredentials(cert));
            }
        }

        [Route(".well-known/openid-configuration", Name = "OIDCMetadata")]
        public ActionResult Metadata()
        {
            return Content(JsonConvert.SerializeObject(new OidcModel
            {
                // Sample: The issuer name is the application root path
                Issuer = AppSettings.APIAppURL,

                // Sample: Include the absolute URL to JWKs endpoint
                JwksUri = AppSettings.APIAppURL + "Oidc/.well-known/keys", //Url.Link("JWKS", null),

                // Sample: Include the supported signing algorithms
                IdTokenSigningAlgValuesSupported = new[] { OidcController.SigningCredentials.Value.Algorithm },
            }), "application/json");
        }

        [Route(".well-known/keys", Name = "JWKS")]
        public ActionResult JwksDocument()
        {
            return Content(JsonConvert.SerializeObject(new JwksModel
            {
                Keys = new[] { JwksKeyModel.FromSigningCredentials(OidcController.SigningCredentials.Value) }
            }), "application/json");
        }
    }
}
