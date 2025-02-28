// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
namespace b2c.helper.api
{
    public class EmailInfo
    {
        public string email { get; set; }
        public bool Primary { get; set; }
        public bool Verified { get; set; }
        public string Visibility
        {
            get; set;
        }
    }
}
