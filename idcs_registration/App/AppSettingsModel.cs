// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
namespace b2c.helper.api
{
    public class AppSettingsModel
    {
        public string B2CTenant { get; set; }
        public string B2CPolicy { get; set; }
        public string B2CClientId { get; set; }
        public string B2CRedirectUri { get; set; }
        public string B2CSignUpUrl { get; set; }
        public int LinkExpiresAfterMinutes { get; set; }
        public string CertificatePassword { get; set; }
        public string CertificateThumbPrint { get; set; }
        public string B2CtenantId { get; set; }
        public string MgmtAppClientId { get; set; }
        public string MgmtAppClientSecret { get; set; }

        public string SendgridWelcomTemplateId { get; set; }
        public string SendgridVerifyTemplateId { get; set; }
        public string SendgridPasswordResetTemplateId { get; set; }
        public string APIAppURL { get; set; }

        public string GoogleCaptchaSecert { get; set; }
        public string PRLinkFlag { get; set; }
        public string SignupLinkFlag { get; set; }

        public string IncorrectPasswordAttempt { get; set; }
        public string NextLoginEnabledTime { get; set; }

        public int LockDurationInMinutes { get; set; }
        public int IncorrectAttempts { get; set; }
        public string StorageConnectionString { get; set; }

        public string ContainerName { get; set; }

    }
}
