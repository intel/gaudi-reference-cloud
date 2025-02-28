#  awslocal is required
AWS_BIN=awslocal


${AWS_BIN} sns create-topic --name idc-staging-notifications-topic
${AWS_BIN} sqs create-queue --queue-name idc-staging-notifications-queue-DLQ
${AWS_BIN} sqs create-queue --queue-name idc-staging-notifications-queue --attributes '{"RedrivePolicy": "{\"deadLetterTargetArn\":\"arn:aws:sqs:us-west-2:000000000000:idc-staging-notifications-queue-DLQ\",\"maxReceiveCount\":\"3\"}"}'
${AWS_BIN} sns subscribe --topic-arn arn:aws:sns:us-west-2:000000000000:idc-staging-notifications-topic --protocol sqs --notification-endpoint arn:aws:sqs:us-west-2:000000000000:idc-staging-notifications-queue
