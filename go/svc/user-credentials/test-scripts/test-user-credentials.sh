#/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

set -e

forwardPids=
cloudaccount=

trap '{
    atexit
}' EXIT

atexit() {
    [ -n "$cloudaccount" ] && grpcurl -insecure -d '{"id":"'$cloudaccount'"}' -H "Authorization: Bearer $ADMIN_TOKEN" localhost:9443 proto.CloudAccountService/Delete >/dev/null 2>&1
    if [ -n "$forwardPids" ]; then
        kill $forwardPids 2>/dev/null || true
    fi
}

fatal() {
    echo FAIL: $* 1>& 2
    exit 1
}

forwardPort() {
    svc=$1
    lport=$2
    rport=$3

    kubectl port-forward -n idcs-system svc/$svc $lport:$rport >/dev/null 2>&1 &
    forwardPids="$forwardPids $!"

    try=0
    while ! nc -vz localhost $lport > /dev/null 2>&1 ; do
        sleep 0.1
        try=$((try + 1))
        if [ $try -gt 20 ]; then
            fatal "timed out waiting for port-forward %svc"
        fi
    done
}

k8s_exists() {
    kubectl get -n idcs-system $* >/dev/null 2>&1 || fatal "k8s object $* missing"
}

# Check for some prequisites:
#   make deploy-all-in-kind in sigle-cluster mode
k8s_exists pod -l app.kubernetes.io/instance=us-dev-1-grpc-proxy-external
k8s_exists pod -l app.kubernetes.io/instance=grpc-proxy-external


forwardPort us-dev-1-grpc-proxy-external 8443 8443
forwardPort grpc-proxy-external 9443 8443
forwardPort oidc 8888 80


ADMIN_TOKEN=$(curl -s "http://localhost:8888/token?email=admin@intel.com&enterpriseId=309b6e1b-e280-41ec-8f0d-0ee8a9314161&tid=d4557930-4982-47b7-b0cf-52034e48c98d&idp=intel.com&groups=IDC.Admin")

#delete cloud account if it already exists.
existing=$(grpcurl -insecure -d '{"name":"abc@a.com"}' -H "Authorization: Bearer $ADMIN_TOKEN" localhost:9443 proto.CloudAccountService/GetByName 2>/dev/null | jq -r .id || true)
[ -n "$existing" ] && grpcurl -insecure -d '{"id":"'$existing'"}' -H "Authorization: Bearer $ADMIN_TOKEN" localhost:9443 proto.CloudAccountService/Delete > /dev/null 2>&1

existing=$(grpcurl -insecure -d '{"name":"def@intel.com"}' -H "Authorization: Bearer $ADMIN_TOKEN" localhost:9443 proto.CloudAccountService/GetByName 2>/dev/null | jq -r .id || true)
[ -n "$existing" ] && grpcurl -insecure -d '{"id":"'$existing'"}' -H "Authorization: Bearer $ADMIN_TOKEN" localhost:9443 proto.CloudAccountService/Delete > /dev/null 2>&1

echo "Creating standard cloudaccount"

cloudaccount=$(grpcurl -insecure -d '{"tid":"0566686b-f46a-4571-938b-f844c05a584a","oid":"7eb42708-a1ae-4df8-bec3-e3b01d7553a7","name":"abc@a.com","owner":"abc@a.com","type":"ACCOUNT_TYPE_STANDARD","billingAccountCreated":true,"enrolled":true,"paidServicesAllowed":true,"personId":"123456", "countryCode":"GB"}'  -H "Authorization: Bearer $ADMIN_TOKEN" localhost:9443 proto.CloudAccountService/Create | jq -r .id)

if [ -z "$cloudaccount" ]; then
    fatal "unable to determine cloudaccount"
fi

echo "Creating cloudaccount for Intel user."
# Country code is not populated for intel users.
cloudaccount_intel=$(grpcurl -insecure -d '{"tid":"0566686b-f46a-4571-938b-f844c05a584a","oid":"7eb42708-a1ae-4df8-bec3-e3b01d7553a1","name":"def@intel.com","owner":"def@intel.com","type":"ACCOUNT_TYPE_INTEL","billingAccountCreated":true,"enrolled":true,"paidServicesAllowed":true,"personId":"234567"}'  -H "Authorization: Bearer $ADMIN_TOKEN" localhost:9443 proto.CloudAccountService/Create | jq -r .id)

if [ -z "$cloudaccount_intel" ]; then
    fatal "unable to determine cloudaccount of type intel"
fi

USER_TOKEN=$(curl -s "http://localhost:8888/token?email=abc@a.com&enterpriseId=7eb42708-a1ae-4df8-bec3-e3b01d7553a7&idp=0566686b-f46a-4571-938b-f844c05a584a&groups=DevCloud%20Console%20Standard")
USER_TOKEN_INTEL=$(curl -s "http://localhost:8888/token?email=def@intel.com&enterpriseId=7eb42708-a1ae-4df8-bec3-e3b01d7553a1&idp=0566686b-f46a-4571-938b-f844c05a584a&groups=DevCloud%20Console%20Standard")


appclientname=${appclientname:-std-client-$(uuidgen)}

clientId=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'","appClientName":"'$appclientname'"}'  -H "Authorization: Bearer $USER_TOKEN" localhost:9443 proto.UserCredentialsService/CreateUserCredentials | jq -r .clientId)
[[ "$clientId" =~ "Message: upstream connect error" ]] && fatal $clientId
echo "PASS: creating client credential, clientId: $clientId"

out=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'"}'  -H "Authorization: Bearer $USER_TOKEN" localhost:9443 proto.UserCredentialsService/GetUserCredentials 2>&1 || true)
[[ "$out" =~ "Message: upstream connect error" ]] && fatal $out
echo "PASS: get client credential, out: $out"


out=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'","clientId":"'$clientId'"}'  -H "Authorization: Bearer $USER_TOKEN" localhost:9443 proto.UserCredentialsService/RemoveUserCredentials 2>&1 || true)
[[ "$out" =~ "Message: upstream connect error" ]] && fatal $out
echo "PASS: delete client credential, out: $out"



appmemberclientname=${appmemberclientname:-std-member-client-$(uuidgen)}
appmemberclientname2=${appmemberclientname2:-std-member-client-2-$(uuidgen)}

memberemail=${memberemail:-member-$(uuidgen| tr -d '-'|cut -c 1-5)@member.com}
MEMBER_USER_TOKEN=$(curl -s "http://localhost:8888/token?email=$memberemail&enterpriseId=7eb42708-a1ae-4df8-bec3-e3b01d7553a7&idp=0566686b-f46a-4571-938b-f844c05a584a&groups=DevCloud%20Console%20Standard")



memberclientId=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'","appClientName":"'$appmemberclientname'"}'  -H "Authorization: Bearer $MEMBER_USER_TOKEN" localhost:9443 proto.UserCredentialsService/CreateUserCredentials | jq -r .clientId)
[[ "$memberclientId" =~ "Message: upstream connect error" ]] && fatal $memberclientId
echo "PASS: creating member client credential, memberclientId: $memberclientId"

memberclientId=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'","appClientName":"'$appmemberclientname2'"}'  -H "Authorization: Bearer $MEMBER_USER_TOKEN" localhost:9443 proto.UserCredentialsService/CreateUserCredentials | jq -r .clientId)
[[ "$memberclientId" =~ "Message: upstream connect error" ]] && fatal $memberclientId
echo "PASS: creating member client credential, memberclientId: $memberclientId"

out=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'"}'  -H "Authorization: Bearer $MEMBER_USER_TOKEN" localhost:9443 proto.UserCredentialsService/GetUserCredentials 2>&1 || true)
[[ "$out" =~ "Message: upstream connect error" ]] && fatal $out
echo "PASS: get member client credential, out: $out"

out=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'","memberEmail":"'$memberemail'"}'  -H "Authorization: Bearer $USER_TOKEN" localhost:9443 proto.UserCredentialsService/RemoveMemberUserCredentials 2>&1 || true)
[[ "$out" =~ "Message: upstream connect error" ]] && fatal $out
echo "PASS: remove all member client credential, out: $out"

out=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'"}'  -H "Authorization: Bearer $MEMBER_USER_TOKEN" localhost:9443 proto.UserCredentialsService/GetUserCredentials 2>&1 || true)
[[ "$out" =~ "Message: upstream connect error" ]] && fatal $out
echo "PASS: get member client credential, out: $out"

#intelappclientname=${intelappclientname:-intel-client-$(uuidgen)}

#clientId=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'","appClientName":"'$intelappclientname'"}'  -H "Authorization: Bearer $USER_TOKEN_INTEL" localhost:9443 proto.UserCredentialsService/CreateUserCredentials | jq -r .clientId)
#[[ "$clientId" =~ "Message: upstream connect error" ]] && fatal $clientId
#echo "PASS: creating client credential, clientId: $clientId"

#out=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'"}'  -H "Authorization: Bearer $USER_TOKEN_INTEL" localhost:9443 proto.UserCredentialsService/GetUserCredentials 2>&1 || true)
#[[ "$out" =~ "Message: upstream connect error" ]] && fatal $out
#echo "PASS: get client credential, out: $out"

#out=$(grpcurl -insecure -d '{"cloudaccountId":"'$cloudaccount'","clientId":"'$clientId'"}'  -H "Authorization: Bearer $USER_TOKEN_INTEL" localhost:9443 proto.UserCredentialsService/RemoveUserCredentials 2>&1 || true)
#[[ "$out" =~ "Message: upstream connect error" ]] && fatal $out
#echo "PASS: delete client credential, out: $out"
