// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation 
using Microsoft.Extensions.Options;
using SendGrid;
using SendGrid.Helpers.Mail;
using System.Net.Mail;

namespace b2c.helper.api
{
    public class SendGridEmailSenderOptions
    {
        public string ApiKey { get; set; }
        public string SenderEmail { get; set; }
        public string SenderName { get; set; }
    }

    public class SendGridEmailSender
    {
        public SendGridEmailSender(IOptions<SendGridEmailSenderOptions> options)
        {
            this.Options = options.Value;
        }

        public SendGridEmailSenderOptions Options { get; set; }

        public async Task<Response> SendEmailAsync(string toEmail, string templateId, object templateData)
        {
            return await Execute(toEmail, templateId, templateData);
        }

        private async Task<Response> Execute(string toEmail, string templateId, object templateData)
        {
            var mailClient = new SendGridClient(Options.ApiKey);
            var mailMessage = new SendGridMessage();

            mailMessage.SetFrom(Options.SenderEmail, Options.SenderName);
            mailMessage.AddTo(new EmailAddress(toEmail));
            mailMessage.SetTemplateId(templateId);
            mailMessage.SetTemplateData(templateData);

            var response = await mailClient.SendEmailAsync(mailMessage);

            return response;
        }
    }
}
