// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
using b2c.helper.api;
using Microsoft.Extensions.Configuration;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.

builder.Services.AddControllers();
// Learn more about configuring Swagger/OpenAPI at https://aka.ms/aspnetcore/swashbuckle
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddHttpClient();
builder.Services.AddSwaggerGen();
builder.Services.Configure<AppSettingsModel>(builder.Configuration.GetSection("AppSettings"));
builder.Services.AddTransient<SendGridEmailSender>();
builder.Services.Configure<SendGridEmailSenderOptions>(options =>
{
    options.ApiKey = builder.Configuration["SendGrid:ApiKey"];
    options.SenderEmail = builder.Configuration["SendGrid:SenderEmail"];
    options.SenderName = builder.Configuration["SendGrid:SenderName"];
});

var app = builder.Build();

// Configure the HTTP request pipeline.

app.UseSwagger();
app.UseSwaggerUI();


app.UseHttpsRedirection();

app.UseAuthorization();

app.MapControllers();

app.Run();
