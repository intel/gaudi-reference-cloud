// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
namespace b2c.helper.api
{
    public class ValidateEmailDTO
    {
        public string Email { get; set; }
        public string? ObjectId { get; set; }
    }

    public class ValidateLoginDTO
    {
        public string Email { get; set; }
        public string? ObjectId { get; set; }
        public int? IncorrectAttempts { get; set; }
        public string? NextLoginEnabledTime { get; set; }

        public bool IsCorrectPwd { get; set; }
    }
}
