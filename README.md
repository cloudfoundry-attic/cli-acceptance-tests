CLI Acceptance Tests
====

Formerly known as GATS.

These are the [Cloud Foundry CLI](https://github.com/cloudfoundry/cli) acceptance tests. 

They are run to evaluate the readiness of the next CLI release. 

They run seperately to the rest of the CLI test because they require a full CF stack
to test against. 

We may be using these at any given time to test new features that haven't been released yet.

Some tests can be pushed upstream into [cf-acceptance-tests](https://github.com/cloudfoundry/cf-acceptance-tests). 
All of the tests in this repo should be written in the same style as cf-acceptance-tests, with the same `testhelpers`.
