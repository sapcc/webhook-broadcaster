#!/bin/bash

bin/webhook-broadcaster -concourse-url $CONCOURSE_URL -auth-user $CONCOURSE_USER -auth-password $CONCOURSE_PASSWORD --refresh-interval 1m
