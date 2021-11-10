# meesqls

Meesqls is a CLI application to record the round-trip time of arbitrary SQL statements against one or more Oracle databases on a specified interval. 

## Build-time Requirements
  - Go 1.14
  - C compiler with `CGO_ENABLED=1` - so cross-compilation is hard

## Run-time Requirements
  - Oracle Client libraries - see [ODPI-C](https://oracle.github.io/odpi/doc/installation.html) 

Although Oracle Client libraries are NOT required for compiling, they *are*
needed at run time.  Download the free Basic or Basic Light package from
<https://www.oracle.com/database/technologies/instant-client/downloads.html>.

## Installation

Download the executable files from the releases page or 

`go install github.com/djaustin/meesqls`

## Configuration
Meesqls will look for a configuration file named `meesqls.yml` in one of the following locations: 
* /etc/meesqls
* $HOME/.meesqls
* The working directory of the executable when it is run

A commented reference configuration file is included in this repository.
