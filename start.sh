#!/bin/bash

bin/webhook-ingester -concourse-url $CONCOURSE_URL -auth-user $CONCOURSE_USER -auth-password $CONCOURSE_PASSWORD
