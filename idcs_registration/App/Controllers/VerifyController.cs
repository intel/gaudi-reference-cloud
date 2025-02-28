// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Options;
using Microsoft.Identity.Client;
using Microsoft.IdentityModel.Tokens;
using System.IdentityModel.Tokens.Jwt;
using System.Net;
using System.Net.Http.Headers;
using System.Security.Cryptography.X509Certificates;
using System.Text;
using System.Text.Json;

namespace b2c.helper.api.Controllers
{
    [Route("api/[controller]")]
    [ApiController]
    public class VerifyController : ControllerBase
    {
        private static Lazy<X509SigningCredentials> SigningCredentials;
        private readonly AppSettingsModel AppSettings;
        private readonly Microsoft.AspNetCore.Hosting.IHostingEnvironment HostingEnvironment;
        private readonly SendGridEmailSender _emailService;
        bool validOnly = false;
        public VerifyController(IOptions<AppSettingsModel> appSettings, Microsoft.AspNetCore.Hosting.IHostingEnvironment hostingEnvironment, SendGridEmailSender emailService)
        {
            this.AppSettings = appSettings.Value;
            this.HostingEnvironment = hostingEnvironment;
            _emailService = emailService;

            // Sample: Load the certificate with a private key (must be pfx file)
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

        [HttpPost]
        [Route("activate")]
        public async Task ActivateUser(UserDTO userDTO)
        {

            string tenantId = AppSettings.B2CtenantId;
            string clientId = AppSettings.MgmtAppClientId;
            string clientSecret = AppSettings.MgmtAppClientSecret;
            string scope = "https://graph.microsoft.com/.default";  // e.g., "https://graph.microsoft.com/.default"
            string authority = $"https://login.microsoftonline.com/{tenantId}";

            // Create a confidential client application
            var app = ConfidentialClientApplicationBuilder.Create(clientId)
                    .WithClientSecret(clientSecret)
                    .WithAuthority(new Uri(authority))
                    .Build();

            // Define the scope for the token request
            var scopes = new[] { scope };

            try
            {
                // Request the token
                var authResult = await app.AcquireTokenForClient(scopes).ExecuteAsync();

                // Use the access token to authenticate requests to the API
                Console.WriteLine("Access Token: " + authResult.AccessToken);

                using (HttpClient client = new HttpClient())
                {
                    // Set the Authorization header with the Bearer token
                    client.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", authResult.AccessToken);

                    // Create the JSON payload for disabling the user
                    var payload = new StringContent("{\"accountEnabled\": true}", Encoding.UTF8, "application/json");

                    // Make the PATCH request to Microsoft Graph API
                    HttpResponseMessage response = await client.PatchAsync($"https://graph.microsoft.com/v1.0/users/{userDTO.ObjectId}", payload);

                    if (response.IsSuccessStatusCode)
                    {
                        Console.WriteLine("User accountEnabled status updated successfully.");
                        // EmailContent emailContent = new EmailContent
                        // {
                        //     Email = userDTO.Email,
                        //     Name = userDTO.Name
                        // };

                        //var responseWelcome = await _emailService.SendEmailAsync(userDTO.Email, AppSettings.SendgridWelcomTemplateId, emailContent);
                    }
                    else
                    {
                        Console.WriteLine($"Error: {response.StatusCode}");
                    }
                }

            }
            catch (MsalServiceException ex)
            {
                Console.WriteLine($"Error acquiring token: {ex.Message}");
            }

        }

        [HttpPost]
        [Route("email")]
        public async Task<ActionResult> Verify(UserDTO user)
        {
            try
            {
                string token = BuildIdToken(user);
                string link = BuildUrl(token, "b2c_1a_development_new_signupverify");

                await SetFlag(AppSettings.SignupLinkFlag, user.ObjectId);

                EmailContent emailContent = new EmailContent
                {
                    Email = user.Email,
                    Link = link,
                    Name = user.Name,
                };

                //var responseWelcome = await _emailService.SendEmailAsync(user.Email, "d-5f3dfd8b49444897883b1ea9b019d151", emailContent);

                var response = await _emailService.SendEmailAsync(user.Email, AppSettings.SendgridVerifyTemplateId, emailContent);

                VerifyResult verifyResult = new VerifyResult()
                {
                    IsSucess = true,
                    Message = "A verification link is sent to your email address. Please click the link to activate your account."

                };

                return StatusCode((int)HttpStatusCode.OK, verifyResult);
            }
            catch (Exception ex)
            {
                return StatusCode((int)HttpStatusCode.BadRequest, ex.Message);
            }
        }


        [HttpPost]
        [Route("githubEmail")]
        public async Task<ActionResult> GitHubEmail(GitHubAuth gitHubAuth)
        {
            try
            {
                var client = new HttpClient();
                var request = new HttpRequestMessage(HttpMethod.Get, "https://api.github.com/user/emails");
                request.Headers.Add("Authorization", "Bearer " + gitHubAuth.Token);
                request.Headers.Add("User-Agent", "auth.api");
                var response = await client.SendAsync(request);
                response.EnsureSuccessStatusCode();
                var respose = await response.Content.ReadAsStringAsync();

                var emailInfoList = JsonSerializer.Deserialize<List<EmailInfo>>(respose);

                var email = emailInfoList.Select(s => s.email).FirstOrDefault();

                GitHubAuthResult result = new GitHubAuthResult
                {
                    Email = email
                };

                return StatusCode((int)HttpStatusCode.OK, result);
            }
            catch (Exception ex)
            {
                return StatusCode((int)HttpStatusCode.BadRequest, ex.Message);
            }
        }

        [HttpPost]
        [Route("passwordreset")]
        public async Task<ActionResult> Password(UserDTO user)
        {
            try
            {
                string token = BuildIdToken(user);
                string link = BuildUrl(token, "b2c_1a_development_new_passwordresetverify");

                await SetFlag(AppSettings.PRLinkFlag, user.ObjectId);


                EmailContent emailContent = new EmailContent
                {
                    Email = user.Email,
                    Link = link,
                    Name = user.Name,
                };

                var response = await _emailService.SendEmailAsync(user.Email, AppSettings.SendgridPasswordResetTemplateId, emailContent);

                VerifyResult verifyResult = new VerifyResult()
                {
                    IsSucess = true,
                    Message = "A password reset link is sent to your email address. "

                };

                return StatusCode((int)HttpStatusCode.OK, verifyResult);
            }
            catch (Exception ex)
            {
                return StatusCode((int)HttpStatusCode.BadRequest, ex.Message);
            }
        }

        private async Task SetFlag(string attributeName, string objectId)
        {
            string tenantId = AppSettings.B2CtenantId;
            string clientId = AppSettings.MgmtAppClientId;
            string clientSecret = AppSettings.MgmtAppClientSecret;
            string scope = "https://graph.microsoft.com/.default";  // e.g., "https://graph.microsoft.com/.default"
            string authority = $"https://login.microsoftonline.com/{tenantId}";

            // Create a confidential client application
            var app = ConfidentialClientApplicationBuilder.Create(clientId)
                    .WithClientSecret(clientSecret)
                    .WithAuthority(new Uri(authority))
                    .Build();

            // Define the scope for the token request
            var scopes = new[] { scope };

            // Request the token
            var authResult = await app.AcquireTokenForClient(scopes).ExecuteAsync();

            // Use the access token to authenticate requests to the API
            Console.WriteLine("Access Token: " + authResult.AccessToken);

            using (HttpClient client = new HttpClient())
            {
                // Set the Authorization header with the Bearer token
                client.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", authResult.AccessToken);

                bool value = true; // or true, depending on your logic
                var payload = new StringContent(
                    $"{{\"{attributeName}\": {value.ToString().ToLower()}}}",
                    Encoding.UTF8,
                    "application/json"
                );
                // Make the PATCH request to Microsoft Graph API
                HttpResponseMessage response = await client.PatchAsync($"https://graph.microsoft.com/v1.0/users/{objectId}", payload);

                if (response.IsSuccessStatusCode)
                {
                    Console.WriteLine("User accountEnabled status updated successfully.");
                }
                else
                {
                    Console.WriteLine($"Error: {response.StatusCode}");
                }
            }
        }

        private string BuildIdToken(UserDTO user)
        {
            string issuer = AppSettings.APIAppURL;

            // All parameters send to Azure AD B2C needs to be sent as claims
            IList<System.Security.Claims.Claim> claims = new List<System.Security.Claims.Claim>();
            claims.Add(new System.Security.Claims.Claim("objectId", user.ObjectId, System.Security.Claims.ClaimValueTypes.String, issuer));
            claims.Add(new System.Security.Claims.Claim("id", user.ObjectId, System.Security.Claims.ClaimValueTypes.String, issuer));

            // Create the token
            JwtSecurityToken token = new JwtSecurityToken(
                    issuer,
                    this.AppSettings.B2CClientId,
                    claims,
                    DateTime.Now,
                    DateTime.Now.AddMinutes(this.AppSettings.LinkExpiresAfterMinutes),
                    VerifyController.SigningCredentials.Value);

            // Get the representation of the signed token
            JwtSecurityTokenHandler jwtHandler = new JwtSecurityTokenHandler();

            return jwtHandler.WriteToken(token);
        }

        private string BuildUrl(string token, string policy)
        {
            string nonce = Guid.NewGuid().ToString("n");

            return string.Format(this.AppSettings.B2CSignUpUrl,
                    this.AppSettings.B2CTenant,
                    policy,
                    this.AppSettings.B2CClientId,
                    Uri.EscapeDataString(this.AppSettings.B2CRedirectUri),
                    nonce) + "&prompt=login&id_token_hint=" + token;
        }

    }
}
