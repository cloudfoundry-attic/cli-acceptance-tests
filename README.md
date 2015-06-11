GATS
====

We got some CATS strapped with GATS.
![cache-cats](http://41.media.tumblr.com/8b3d8038a788d661b59ed7b35a7f73a2/tumblr_njuqwy3abv1r9khx4o4_1280.jpg)

What???
=======

GATS (a.k.a. the Go-CLI Acceptance Test Suite) are a collection of integration tests for the Go-CLI. In general, these should all be passed on recent CF Releases, but not on the stable version of the CLI. We may be using these at any given time to test new features that haven't been released yet.

Where is this going?
--------------------

Our goal is to delete these tests overtime as we push these tests upstream into [CATS](https://github.com/cloudfoundry/cf-acceptance-tests). All of the tests in this repo should be written in the same style, with the same testhelpers.
