# import

A vanity url http service for `go import`.

Uses [boltdb](htpts://github.com/boltdb/bolt), but otherwise all standard library imports.

Plans include:

 - API for managing repos that this service handles
 - UI for listing and exploring repos, and managing through the API

## Installation

Use `go get jw4.us/import` to get the source code.

Use the Dockerfile or the image `docker.jw4.us/import` to run the default version in a container.


(Acknowledgements and thanks to [rsc](https://rsc.io/go-import-redirector))
