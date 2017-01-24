# Gogen [![CircleCI](https://img.shields.io/circleci/project/github/RedSparr0w/node-csgo-parser.svg)](https://github.com/coccyx/gogen) [![Coverage Status](https://coveralls.io/repos/github/coccyx/gogen/badge.svg?branch=master)](https://coveralls.io/github/coccyx/gogen?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/coccyx/gogen)](https://goreportcard.com/report/github.com/coccyx/gogen) 

Gogen is an open source data generator for generating demo and test data.  Gogen can be used for any type of data, 
and it has first class support for time series data.  Primarily, it's been used and tested to generate log and metric data 
for testing time series systems.  It's very performant and can be used for functional and performance testing as well 
as generating compelling data sets to be used for demo purposes.

## Features

* Generates complex time series data via token substitutions in original raw events
* Flexible output methods and formats: raw, JSON, CSV, or custom template output to stdout, file, HTTP and more
* Support for arbitrary key/value datasets, tokens can replace in any field
* Many token types: static, randomly generated, different types of choices from lists, or custom scripts
* Three generation modes: random substitution of tokens from a sample file, replaying a sample in time series order, or custom generation scripts
* Extensible via custom Lua scripts
* Easy configuration via YAML or JSON files
* Easy sharing of configurations via a centralized service
* Simple getting started experience as one statically linked binary, compiled on multiple platforms

## Getting started

Getting started with Gogen is easy.  Grab the latest binary for [OS X](https://api.gogen.io/osx/gogen), [Windows](https://api.gogen.io/windows/gogen.exe), 
or [Linux](https://api.gogen.io/linux/gogen) and place it somewhere in your `$PATH` with something like `wget -O /usr/local/bin/gogen https://api.gogen.io/osx/gogen && chmod 755 /usr/local/bin/gogen`. 
First lets see what our options are.

    gogen --help

Gogen has a centralized service which makes sharing Gogen configs very simple.  Lets see if we can generate a weblog, for example.

    gogen search weblog

Seems a nice user named Coccyx has published a weblog configuration for us.  Lets see what that will generate, and see what the configuration looks like.

    gogen info coccyx/weblog

As we see from the info, the actual configuration is stored in a GitHub gist.  Feel free to click through the link in the info to see the full configuration. 
Let's generate a few weblog entries.

    gogen -c coccyx/weblog

If you looked at the configuration, you can see it by default will generate 10 events over one interval and end at one interval.  Let's change a few 
of the options to generate events over a longer time window.

    gogen -c coccyx/weblog gen -c 1 -ei 10 -i 60

We get the same number of events, but this time we're generating the events over 10 minutes, at an interval of once per minute with 1 event per interval. 
Now let's get a view into the structure of events as it flows through the system and also look at outputting things differently, to a file.

    gogen -c coccyx/weblog -ot json -o file -f weblog.json gen -c 1 -ei 10 -i 60

In weblog.json, you will find the same events output a file as JSON.  You can see there are 5 fields: _raw, index, host, source, sourcetype.  This mirrors 
how Splunk defines it's metadata, but these can be any arbitary key/value pairs.  They are all treated as strings.  Lastly, lets see what other configs are available.

    gogen list

## Documentation


* Learn how to [configure and run Gogen](README/Tutorial.md)
* Look through [many more examples](README/Examples.md)
* [Sharing configurations](README/Sharing.md) and participating in the community
* Can't find something?  [Reference documentation](README/Reference.md)