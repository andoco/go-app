#!/usr/bin/env bash

aws --endpoint-url "http://localhost:4566" sqs send-message \
    --queue-url "http://localhost:4566/000000000000/test-queue" \
    --message-body '{"message":"hello"}' \
    --message-attributes '{"msgType":{"DataType":"String","StringValue":"foo"}}'
