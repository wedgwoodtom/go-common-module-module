
# How to Build
See the Makefile for options

* make itest - to run integration tests
* make build - to build


* Useful Commands
    * go mod init <modulename>.  (Creates go.mod)
    * go get -u ./….  (creates go.mod and go.sum)
    * go mod vendor   (pull code to vendor dir)
    * Update a dependency
        * go get -u <repo url>
        * go mod vendor
        
        