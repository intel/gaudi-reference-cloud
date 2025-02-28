# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
import json
import os
import requests
import sys
import time
import uuid

apiToken = os.environ['OAUTH_TOKEN']

apiBaseUrl = 'https://internal-placeholder.com/v1'
createCloudAcctApiUrl = apiBaseUrl + '/cloudaccounts'
getMeteringApiUrl = apiBaseUrl + '/meteringrecords/search'
getUsagesApiUrl = apiBaseUrl + '/billing/usages'
username = ''

def create_cloud_account(account_type):
    print('creating a cloud account')
    acct_type = ''
    username = ''
    if not username:
        if account_type == 'premium':
            username = 'premium-user' + str(uuid.uuid4()) + '@example.com'
            acctType = 'ACCOUNT_TYPE_PREMIUM'
        elif account_type == 'standard':
            username = 'standard-user' + str(uuid.uuid4()) + '@example.com'
            acctType = 'ACCOUNT_TYPE_STANDARD'
        elif account_type == 'intel':
            username = 'intel-user' + str(uuid.uuid4()) + '@intel.com'
            acctType = 'ACCOUNT_TYPE_INTEL'
    
    cloudAcctApiHeaders = {'Authorization': 'Bearer ' + apiToken,
                           'Content-Type': 'application/json'}
    createCloudAcctParam = {
        'name': username,
        'owner': username,
        'tid': str(uuid.uuid4()),
        'oid': str(uuid.uuid4()),
        'type': acctType
    }
    responseForCreateCloudAcct = requests.post(createCloudAcctApiUrl, params = createCloudAcctParam, headers = cloudAcctApiHeaders)
    if responseForCreateCloudAcct.status_code == 200:
        print("created cloud acct")
    else:
        print("failed to create cloud acct")
        response_code = responseForCreateCloudAcct.status_code
        response_text = responseForCreateCloudAcct.text
        print(json.dumps(responseForCreateCloudAcct.json(), indent=4))
            

if len( sys.argv ) > 1:
    if len( sys.argv ) != 2:
        print('only one input argument expected')
        exit(1)
    
    print('running with cli')
    arg = sys.argv[1]
    if arg != 'withCli':
        print('invalid argument, use withCli')
else:    
    print('running without cli parameters')

create_cloud_account('standard')