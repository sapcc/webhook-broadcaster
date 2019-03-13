Concourse webhook broadcaster
=============================

This repo contains a small service that is meant to be run as a sidecar to a concourse deployment.

Its purpose is to process incoming webhooks from a github repository and broadcast them to all concourse
resources that reference the git repository.


```
  ┌─────────────┐
┌─┤  Concourse  ├────────────────┐                                           ┌─────────────────────┐
│ └─────────────┘                │                                         ┌─┤ Github (Enterprise) ├─┐
│ ┌────────────────────────────┐ │                                         │ └─────────────────────┘ │
│ │name: resource1             │ │                                         │                         │
│ │type: git                   │ │                                         │                         │
│ │source:                     │◀┼─┐                                       │                         │
│ │  uri: github.com/some/where│ │ │                                       │                         │
│ └────────────────────────────┘ │ │                                       │                         │
│ ┌────────────────────────────┐ │ │    ┌─────────────────────┐            │                         │
│ │name: resource2             │ │ │    │                     │            │ ┌─────────────────────┐ │
│ │type: git                   │ │ │    │                     │    push    │ │     repository:     │ │
│ │source:                     │◀┼─┼────│ Webhook Broadcaster │◀───────────┼─│     some/where      │ │
│ │  uri: github.com/some/where│ │ │    │                     │   webhook  │ └─────────────────────┘ │
│ └────────────────────────────┘ │ │    │                     │            │                         │
│ ┌────────────────────────────┐ │ │    └─────────────────────┘            │                         │
│ │name: resource3             │ │ │                                       │                         │
│ │type: git                   │ │ │                                       │                         │
│ │source:                     │◀┼─┘                                       │                         │
│ │  uri: github.com/some/where│ │                                         │                         │
│ └────────────────────────────┘ │                                         │                         │
└────────────────────────────────┘                                         └─────────────────────────┘
```


Why?
====

Our concourse deployment was causing considerable load on our github enterprise appliance and we needed to increase the `check_every` properties on all our git resources and rely on [resource webhooks](https://concourse-ci.org/resources.html#resource-webhook-token) to propagte updates in a timely fashion.

We are also using a single repository in 300 resources across 70 pipelines and adding 300 webhooks to this github repository is unfeasable.

How?
====
The webhook broadcaster periodically (`--refresh-interval`) fetches all pipelines via the concourse api and builds up a cache of all resources that have a `webhook_token` configured. When  a push event is received the broadcaster scans its local cache and calls the webhook url of all resources that reference this particular repository using the resources individual `webhook_token`.

Usage
======
1. Start a webhook-broadcaster you need to provide the following configuration
   * `--concourse-url` external url of your concourse deployment
   * `--auth-user` concourse basic auth admin user. 
   * `--auth-password` concourse basic auth admin password
2. Create a github webhook for push events pointing it to `http://webhook-broadcaster.somewhere:8080/github`
3. Make sure resources of type `git` have a `webhook_token` configured

Compatibility
=============
* webhook-broadcaster should work with concourse `>=4.x`. There is a branch https://github.com/sapcc/webhook-broadcaster/tree/concourse-3.x that supports concourse `3.x`.
* The broadcaster only supports github webhooks yet. Adding different types of webhooks, even for resources of different types should be simple (PRs welcome).
