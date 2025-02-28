// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

using Azure.Storage.Blobs;
using Azure.Storage.Blobs.Models;
using b2c.helper.api;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Options;
using Microsoft.Identity.Client;
using Newtonsoft.Json;
using System.Net;
using System.Net.Http.Headers;
using System.Text;
using static System.Net.Mime.MediaTypeNames;

namespace Contoso.AADB2C.API.Controllers
{
    [ApiController]
    [Route("api/[controller]")]
    public class IdentityController : ControllerBase
    {
        private readonly HttpClient _httpClient;
        private readonly IConfiguration _configuration;
        private readonly string _blockedDomainsFileName;
        private readonly string _blockedEmailssFileName;
        private readonly AppSettingsModel AppSettings;


        public IdentityController(IOptions<AppSettingsModel> appSettings, HttpClient httpClient, IConfiguration configuration)
        {
            _httpClient = httpClient;
            _configuration = configuration;
            _blockedDomainsFileName = "blocked_domains.txt";
            _blockedEmailssFileName = "blocked_emails.txt";
            this.AppSettings = appSettings.Value;

        }

        [HttpPost("signup")]
        public async Task<IActionResult> SignUpAsync()
        {
            // If no data came in, return bad request
            if (Request.Body == null)
            {
                return BadRequest(new B2CResponseContent("Request content is null.", HttpStatusCode.BadRequest));
            }

            using (var reader = new StreamReader(Request.Body, Encoding.UTF8))
            {
                var input = await reader.ReadToEndAsync();

                if (string.IsNullOrEmpty(input))
                {
                    return Conflict(new B2CResponseContent("Request content is empty", HttpStatusCode.Conflict));
                }

                var inputClaims = JsonConvert.DeserializeObject<InputClaimsModel>(input);
                if (inputClaims == null || string.IsNullOrEmpty(inputClaims.captcha))
                {
                    return Conflict(new B2CResponseContent("Captcha is missing or invalid.", HttpStatusCode.Conflict));
                }

                // Validate captcha blob at Google
                var validationEndpoint = "https://www.google.com/recaptcha/api/siteverify";
                var secret = AppSettings.GoogleCaptchaSecert; // Replace with your actual secret

                if (string.IsNullOrEmpty(secret))
                {
                    throw new InvalidOperationException("Recaptcha secret is not configured.");
                }

                var postBody = $"secret={secret}&response={inputClaims.captcha}";
                var response = await _httpClient.PostAsync(validationEndpoint, new StringContent(postBody, Encoding.UTF8, "application/x-www-form-urlencoded"));
                var content = await response.Content.ReadAsStringAsync();

                if (response.IsSuccessStatusCode)
                {
                    dynamic json = JsonConvert.DeserializeObject(content);

                    // Check if the captcha verification was successful
                    if (json.success == "true")
                    {
                        // Check if the email is blocked
                        if (await IsEmailBlocked(inputClaims.Email))
                        {
                            return Conflict(new B2CResponseContent("Please choose a different email address to continue.", HttpStatusCode.Conflict));
                        }

                        return Ok(new B2CResponseContent("Captcha validated successfully.", HttpStatusCode.OK));
                    }
                    else
                    {
                        return Conflict(new B2CResponseContent("Captcha failed, retry the Captcha.", HttpStatusCode.Conflict));
                    }
                }

                return Conflict(new B2CResponseContent("Captcha API call failed.", HttpStatusCode.Conflict));
            }
        }

        [HttpPost("validateEmail")]
        public async Task<IActionResult> ValidateEmail(ValidateEmailDTO validateEmailDTO)
        {
            try
            {
                var result = await IsEmailBlocked(validateEmailDTO.Email);

                if (result)
                {
                    if (validateEmailDTO.ObjectId != null)
                    {
                        await LockAccount(validateEmailDTO.ObjectId);

                    }

                    return Conflict(new B2CResponseContent("You are not allowed to login and blocked.", HttpStatusCode.Conflict));


                }
                else
                {
                    return Ok(new B2CResponseContent("Valid Email.", HttpStatusCode.OK));

                }

            }
            catch
            {
                return Conflict(new B2CResponseContent("You are not allowed to login and blocked.", HttpStatusCode.Conflict));
            }

        }

        [HttpPost("GetContent")]
        public async Task<IActionResult> GetContent(string filename)
        {
            try
            {
                List<string> blockedDomains = new List<string>();
                List<string> blockedEmails = new List<string>();

                Console.WriteLine($"Checking email: {filename}"); // Log the incoming email

                var blocked_domains = await GetFileContent(_blockedDomainsFileName);
                var blocked_emails = await GetFileContent(_blockedEmailssFileName);

                blockedEmails = blocked_emails.Split('\n').ToList();
                blockedDomains = blocked_domains.Split('\n').ToList();
                return Ok(blockedEmails);


            }
            catch (Exception ex) { return Ok(ex); }

                
      


        }

        [HttpPost("validateLogin")]
        public async Task<IActionResult> validateLogin(ValidateLoginDTO validateLoginDTO)
        {
            try
            {
                var result = await IsEmailBlocked(validateLoginDTO.Email);

                if (result)
                {
                    if (validateLoginDTO.ObjectId != null)
                    {
                        await LockAccount(validateLoginDTO.ObjectId);

                    }

                    return Conflict(new B2CResponseContent("You are not allowed to login and blocked.", HttpStatusCode.Conflict));


                }
                else
                {

                    string utcDateString = validateLoginDTO.NextLoginEnabledTime; // Ensure format matches input string
                    DateTime parsedUtcDate;

                    // Parse the string as UTC
                    if (DateTime.TryParse(utcDateString, out parsedUtcDate))
                    {
                        // Convert the parsed time to UTC (if not already in UTC)
                        parsedUtcDate = DateTime.SpecifyKind(parsedUtcDate, DateTimeKind.Utc);

                        if (DateTime.UtcNow < parsedUtcDate)
                        {
                            return Conflict(new B2CResponseContent("Your account is in locked state due to incorrect password attempts. Please try after sometime.", HttpStatusCode.Conflict));

                        }

                        await ResetPasswordAttempts(validateLoginDTO.ObjectId);

                        validateLoginDTO.IncorrectAttempts = 0;

                    }

                    if (validateLoginDTO.IsCorrectPwd == false)
                    {
                        validateLoginDTO.IncorrectAttempts = validateLoginDTO.IncorrectAttempts == null ? 1 : validateLoginDTO.IncorrectAttempts + 1;

                        if (validateLoginDTO.IncorrectAttempts >= AppSettings.IncorrectAttempts)
                        {
                            string incorrectPasswordAttemptKey = AppSettings.IncorrectPasswordAttempt;
                            string nextLoginEnabledTimeKey = AppSettings.NextLoginEnabledTime;


                            // Dynamically create the payload
                            var payloadObject = new Dictionary<string, object>
            {
                { incorrectPasswordAttemptKey, validateLoginDTO.IncorrectAttempts  }, // Dynamic key for attempts
                { nextLoginEnabledTimeKey, DateTime.UtcNow.AddMinutes(AppSettings.LockDurationInMinutes).ToString() } // Dynamic key for next login time
            };

                            // Serialize the dictionary to JSON
                            string jsonPayload = JsonConvert.SerializeObject(payloadObject);

                            var payload = new StringContent(jsonPayload, Encoding.UTF8, "application/json");

                            await UpdatePasswordAttempts(validateLoginDTO.ObjectId, payload);

                            return Conflict(new B2CResponseContent("Incorrect Password. You account is locked", HttpStatusCode.Conflict));

                        }
                        else
                        {
                            string incorrectPasswordAttemptKey = AppSettings.IncorrectPasswordAttempt;


                            // Dynamically create the payload
                            var payloadObject = new Dictionary<string, object>
            {
                { incorrectPasswordAttemptKey, validateLoginDTO.IncorrectAttempts  }// Dynamic key for attempts
            };

                            // Serialize the dictionary to JSON
                            string jsonPayload = JsonConvert.SerializeObject(payloadObject);

                            var payload = new StringContent(jsonPayload, Encoding.UTF8, "application/json");

                            await UpdatePasswordAttempts(validateLoginDTO.ObjectId, payload);


                            return Conflict(new B2CResponseContent("Incorrect Password. Attempts remaining - " + (5 - validateLoginDTO.IncorrectAttempts), HttpStatusCode.Conflict));
                        }
                    }
                    else
                    {
                        if (validateLoginDTO.IsCorrectPwd && validateLoginDTO.IncorrectAttempts > 0)
                        {

                            await ResetPasswordAttempts(validateLoginDTO.ObjectId);
                        }
                    }

                    return Ok(new B2CResponseContent("Valid Email.", HttpStatusCode.OK));

                }

            }
            catch
            {
                return Conflict(new B2CResponseContent("You are not allowed to login and blocked.", HttpStatusCode.Conflict));
            }

        }

        [HttpPost("validateSocialEmail")]
        public async Task<IActionResult> validateSocialEmail(ValidateEmailDTO validateEmailDTO)
        {
            try
            {
                var result = await IsEmailBlocked(validateEmailDTO.Email);

                if (result)
                {
                    if (validateEmailDTO.ObjectId != null)
                    {
                        await LockAccount(validateEmailDTO.ObjectId);

                    }

                    return Ok(new SociaEmailValidation { isSocialEmailValid = false });
                }
                else
                {
                    return Ok(new SociaEmailValidation { isSocialEmailValid = true });

                }

            }
            catch
            {
                return Ok(new SociaEmailValidation { isSocialEmailValid = false });
            }

        }

        private async Task LockAccount(string objectId)
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

                // Create the JSON payload for disabling the user
                var payload = new StringContent("{\"accountEnabled\": false}", Encoding.UTF8, "application/json");

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


        private async Task ResetPasswordAttempts(string objectId)
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

                string incorrectPasswordAttemptKey = AppSettings.IncorrectPasswordAttempt;
                string nextLoginEnabledTimeKey = AppSettings.NextLoginEnabledTime;


                // Dynamically create the payload
                var payloadObject = new Dictionary<string, object>
            {
                { incorrectPasswordAttemptKey, 0 }, // Dynamic key for attempts
                { nextLoginEnabledTimeKey, "" } // Dynamic key for next login time
            };

                // Serialize the dictionary to JSON
                string jsonPayload = JsonConvert.SerializeObject(payloadObject);

                var payload = new StringContent(jsonPayload, Encoding.UTF8, "application/json");

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

        private async Task UpdatePasswordAttempts(string objectId, StringContent payload)
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

        private async Task<bool> IsEmailBlocked(string email)
        {
            List<string> blockedDomains = new List<string>();
            List<string> blockedEmails = new List<string>();

            Console.WriteLine($"Checking email: {email}"); // Log the incoming email

            var blocked_domains = await GetFileContent(_blockedDomainsFileName);
            var blocked_emails = await GetFileContent(_blockedEmailssFileName);

            blockedEmails = blocked_emails.Split("\r\n").ToList();
            blockedDomains = blocked_domains.Split("\r\n").ToList();

            string[] emailParts = email.Split('@');

            if (emailParts.Length != 2)
            {
                return false; // Invalid email format
            }

            string domain = emailParts[1];

            Console.WriteLine($"Extracted domain: {domain}"); // Log the extracted domain

            // Check if the domain is in the blocked list
            var blockedDomain = blockedDomains.Contains(domain, StringComparer.OrdinalIgnoreCase);
            var blockedEmail = blockedEmails.Contains(email, StringComparer.OrdinalIgnoreCase);

            return blockedDomain || blockedEmail;
        }

        private async Task<string> GetFileContent(string fileName)
        {

            // Azure Blob Storage connection string
            string connectionString = AppSettings.StorageConnectionString;

            // Container and blob details
            string containerName = AppSettings.ContainerName;
            string blobName = fileName;

            // Create a BlobServiceClient
            BlobServiceClient blobServiceClient = new BlobServiceClient(connectionString);

            // Get a reference to the container
            BlobContainerClient containerClient = blobServiceClient.GetBlobContainerClient(containerName);

            // Get a reference to the blob
            BlobClient blobClient = containerClient.GetBlobClient(blobName);

            // Check if the blob exists
            if (await blobClient.ExistsAsync())
            {
                // Download the blob's content
                BlobDownloadInfo download = await blobClient.DownloadAsync();

                // Read the content as a stream
                using (StreamReader reader = new StreamReader(download.Content))
                {
                    string content = await reader.ReadToEndAsync();
                    Console.WriteLine("Blob content:");
                    return content;
                }
            }
            else
            {
                Console.WriteLine($"Blob '{blobName}' does not exist in container '{containerName}'.");

                return string.Empty;
            }
        }
    }
}
