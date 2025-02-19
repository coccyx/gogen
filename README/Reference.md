# Gogen Reference

This document will walk through all the objects in the configuration and Lua API and attempt to document them fully.

## Environment Variables

| Environment Variable   | Expected Value   | Description                                                                  |
|------------------------|------------------|------------------------------------------------------------------------------|
| GOGEN\_EXPORT          | 1 or unset       | Specifies whether Gogen is running to export configs.  If set to export, internal defaults aren't added and a few more miscellaneous things.
| GOGEN\_INFO            | 1 or unset       | Specifies whether to turn on info level logging
| GOGEN\_DEBUG           | 1 or unset       | Specifies whether to turn on debug level logging
| GOGEN\_GENERATORS      | integer          | Number of generator threads for Gogen to run
| GOGEN\_OUTPUTTERS      | integer          | Number of output threads for Gogen to run
| GOGEN\_OUTPUTTEMPLATE  | name of template | Name of an output template to use.
| GOGEN\_OUT             | outputter        | Name of outputter to use, stdout, file, http, etc
| GOGEN\_FILENAME        | filename         | When outputter is file, output to this filename
| GOGEN\_URL             | URL              | When outputter is http, URL to send data to
| GOGEN\_HEC_TOKEN       | Token            | When outputter is HTTP, use the token for authentication to Splunk's HTTP Event Collector
| GOGEN\_SAMPLES\_DIR    | directory        | Directory to find sample configs
| GOGEN\_CONFIG          | file             | Path to a full configs
| GOGEN\_CONFIG\_DIR     | directory        | Directory to use as a base config directory, contains sample, templates, generator directories.

## Configuration

The configuration 

| Section    | Description                                                                                                        |
|------------|--------------------------------------------------------------------------------------------------------------------|
| global     | Defines global parameters, such as output                                                                          |
| generators | Defines custom generators, written in Lua, which can greatly extend gogen's capabilities                           |
| raters     | Defines raters, which allow you to rate the count of events or value of tokens based on time or custom Lua scripts |
| samples    | Define sample configurations, which is the core data structure in Gogen                                            |
| mix        | Defines mix configurations, which allow you to reuse existing sample configurations in new configurations          |
| templates  | Defines output templates, which allow you to format the output of Gogen using Go's templating language             |

### Global

Global options:

| Setting          | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| debug            | Turns on debug internal logging output                                                         | bool        |
| verbose          | Turns on verbose internal logging output                                                       | bool        |
| generatorWorkers | Sets the number of threads generating output tasks                                             | int         |
| outputWorkers    | Sets the number of threads outputting output tasks                                             | int         |
| rotInterval      | Interval in seconds to output internal statistics to stderr                                    | int         |
| output           | Set the output plugin to use                                                                   | string      |
| samplesDir       | Sets the directory to look for Sample YAML, CSV or .Samples files                              | string list |
| cacheIntervals   | Sets the number of intervals to reuse generated events                                         | int         |


### Output

Output options:

| Setting          | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| fileName         | For file output, sets the file name to output to                                               | string      |
| maxBytes         | For file output, sets the max bytes before rolling a new file                                  | int64       |
| backupFiles      | For file output, sets the number of files to keep before discarding older files                | int         |
| bufferBytes      | For HTTP, S2S and other outputs, sets the number of bytes to buffer before flushing            | int         |
| outputter        | Sets the output module to use, currently supports devnull, file, http and stdout               | string      |
| outputTemplate   | Set the output template to format output, builtins include csv, json, splunkhec, modinput      | string      |
| endpoints        | For http, or potentially others, lists endpoints to send data to.                              | string list |
| headers          | For http, sets headers                                                                         | string obj  |
| protocol         | For network, set to `tcp` or `udp`                                                             | string      |
| timeout          | For network based outputs, a time in seconds, default `10s`                                    | string      |

### Sample

Samples are represented as a list of sample objects, which can consist of the following configuration options:

| Setting          | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| name             | Name of the sample                                                                             | string      |
| description      | Description of the sample                                                                      | string      |
| notes            | Notes about the sample                                                                         | string      |
| disabled         | Sets whether the sample should be disabled and not generate events                             | bool        |
| generator        | Sets the sample to use a custom generator as defined in the generators stanza                  | string      |
| rater            | Sets the sample to use a custom rater as defined in the raters stanza                          | string      |
| interval         | Sets the interval, in seconds, between generations of this sample                              | int         |
| delay            | Sets the delay, in seconds, before starting generation for the first time                      | int         |
| count            | Sets the number of events to generate each sample                                              | int         |
| earliest         | Sets the beginning of the time window to generate an event in for this interval (ex: -1m)      | string      |
| latest           | Sets the end of the time window to generate an event in for this interval (ex: now)            | string      |
| begin            | Sets the timestamp to begin generation at (ex: -1h)                                            | string      |
| end              | Sets the timestamp to end generation at (ex: now). If unspecified, generates in real time      | string      |
| endIntervals     | Sets generation to run for `endInterval` intervals                                             | int         |
| randomizeCount   | Percentage of randomness of `count` events.  Ex: 0.2 will randomly increase count +/- 20%      | float       |
| randomizeEvents  | Randomize the events from the sample when picking events.  By default will pick top X events   | bool        |
| tokens           | List of tokens (see below)                                                                     | token       |
| lines            | List of line objects.  Arbitrary key/value pairs to be used for generation.                    | list string obj
| field            | Sets the default field to replace in (default '_raw')                                          | string      |
| fromSample       | Bring in lines from another named sample                                                       | string      |
| singlePass       | Allows disabling SinglePass optimization, if for example you have chained replacements         | bool        |

### Token

Tokens are the core unit of the replacement engine, and they contain the following configuration options:

| Setting          | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| name             | Name of the token                                                                              | string      |
| format           | Format of the replacement, either `regex` or `template`. Default `template`                    | string      |
| token            | Replacement text to find.  Required for `regex`, for `template` defaults to `$name$`           | string      |
| type             | Sets the type of replacement.  See token types below.                                          | string      |
| replacement      | Value to use for the replacement.  Depends on the token type (see below)                       | string      |
| group            | Token group. All items from the same group will pick the same index across multiple tokens     | int         |
| sample           | For choice types, pulls the items from another sample                                          | string      |
| field            | Field to replace into, defaults to `_raw`                                                      | string      |
| srcField         | Field to replace from, used in `fieldChoice`                                                   | string      |
| precision        | For `float` `random` or `rated` tokens, how many decision points to generate                   | int         |
| lower            | Lower value for a `random` or `rated` token                                                    | int         |
| upper            | Upper value for a `random` or `rated` token                                                    | int         |
| length           | Length of a `random` `string` or `hex` replacement                                             | int         |
| weightedChoice   | Used for `weightedChoice` type, a list of objects containing `choice` and `weight` (`int`)     | list of obj |
| fieldChoice      | Used for `fieldChoice` type, a list of objects containing fields and values                    | list string obj |
| choice           | Used for `choice` type, a list of strings to use for replacements                              | list of string |
| script           | LUA script to use for replacement.                                                             | string      |
| init             | Initialize keys and values in the Lua engine                                                   | object      |
| rater            | Use the specified rater to rate this token (see below)                                         | string      |
| disabled         | Disables this sample.                                                                          | bool        |

Token types:

| Type             | Description                                                                                    |
|------------------|------------------------------------------------------------------------------------------------|
| timestamp        | Timestamp in strftime format                                                                   |
| gotimestamp      | Timestamp in [go timestamp format](https://golang.org/pkg/time/#pkg-constants).  This is signifcantly more performant than strftime. |
| epochtimestamp   | Timestamp in seconds since the epoch.                                                          | 
| static           | Replaces with a static string                                                                  |
| random           | Replaces the token with random values.  Valid replacement values: `int`, `float`, `string`, `hex`, `guid`, `ipv4`, or `ipv6` |
| rated            | Replaces the token with a rated value. Valid replacement values: `int`, `float`                |
| choice           | Replaces from the `choice` stanza, which is a list                                                 |
| weightedChoice   | Replaces from the `weightedChoice` stanza, which is a list of objects containing weight (`int`) and choice |
| fieldChoice      | Replaces from the `fieldChoice` stanza, which is an object containing values. Selects field based on `field` stanza. |
| script           | Replaces using a lua script, which is defined inline.                                          |

### Mix

Mixes allow grabbing other full configurations, overriding a few parameters, and creating a new config.

| Setting          | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| sample           | Name of the sample.  Path to a config file or a public config (ex: coccyx/weblog)              | string      |
| interval         | Overrides `interval` of sample.                                                                | int         |
| count            | Overrides `count` of sample.                                                                   | int         |
| begin            | Overrides `begin` of sample.                                                                   | string      |
| end              | Overrides `end` of sample.                                                                     | string      |
| endIntervals     | Overrides `endIntervals` of sample.                                                            | int         |
| realtime         | Sets sample to realtime. This exists because if end is set, this will override to realtime.    | bool        |

### Raters

Raters will dynamically determine value based on the time of day or a custom script.

| Setting          | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| name             | Name of the rater                                                                              | string      |
| type             | Type of the rater. Either `config` or `script`                                                 | string      |
| script           | For `script` rater, specifies a Lua script to use to rate.                                     | string      |
| options          | Options to pass to the config rater.  See [here](https://github.com/coccyx/gogen/blob/master/tests/rater/fullraterconfig.yml) for example.  | object |
| init             | Initialize Lua variables for `script` rater.                                                   | object      |

### Template

Templates let the user change how data is output using Go's template language.  See [here](https://github.com/coccyx/gogen/blob/master/examples/tutorial/tutorial4.yml) for an example.

| Setting          | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| name             | Name of the template                                                                           | string      |
| header           | Header for the template                                                                        | string      |
| row              | Row for the template                                                                           | string      |
| footer           | Footer for the template                                                                        | string      |

### Generators

Generators let the user define custom logic in Lua for how events should be generated. Generators contain a configuration component as well as an API the user can access to send events and access configuration data.


| Setting          | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| name             | Name of the generator                                                                          | string      |
| init             | Object of key value pairs to initialize in the `state` variable in in the Lua engine           | object      |
| options          | Object of key value pairs to be placed in the `options` variable in the Lua engine.  Unlike `init`, can pass complex structures. | object |
| script           | Script to execute.                                                                             | string      |
| fileName         | File on disk containing the Lua script.                                                        | string      |
| singleThreaded   | Execute SingleThreaded or not.  Scripts may be wary of stomping on state in multithreaded mode.  | bool      |

## Generator API

The Lua environment provides a rich set of APIs to access running state inside of Gogen.  This documents the global variables as well as all the functions and their parameters.

### Globals

| Variable         | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| state            | State from the `init` setting in config                                                        | table       |
| options          | Options passed to the generator from config, as userdata                                       | userdata    |
| lines            | Lines configured in config, as a table of tables                                               | table       |
| count            | Count to generate for this interval                                                            | number      |
| earliest         | Earliest time to generate for, in epoch time                                                   | number      |
| latest           | Latest time to generate for, in epoch time                                                     | number      |
| now              | Current time, in epoch time                                                                    | number      |


### sleep

sleep uses Go's time.sleep to sleep the generator thread.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| duration         | Duration to sleep in seconds                                                                   | int64       |

### debug

debug outputs debug output to stderr.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| message          | Message to output                                                                              | string      |

### Info

info outputs info output to stderr.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| message          | Message to output                                                                              | string      |

### replaceTokens

replaceTokens calls Gogen's replacement engine, and replaces all tokens in a particular line.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| event            | Takes a Lua table of key/value pairs to replace tokens in. Generally retrieved from `getLine`  | table       |
| choices          | Dictionary of choices, passed back in as a return from a previous `replaceTokens` call         | userdata    |
| replaceFirst     | Replace tokens set via `setTokens` first if `true`                                             | bool        |

*Returns:*

| Return           | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| event            | Event with all tokens replaced                                                                 | table       |
| choices          | Dictionary of choices                                                                          | userdata    |

### send

send sends events to the outputter.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| events           | Takes a Lua table of tables of key/value pairs to output                                       | table       |

### sendEvent

sendEvent sends a single event to the outputter.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| events           | Takes a Lua table              key/value pairs to output                                       | table       |

### round

round rounds a Lua Number to the specified precision.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| number           | Number to round                                                                                | number      |
| precision        | Signifcant digits to round to                                                                  | number      |

*Returns:*

| Return           | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| number           | Rounded number                                                                                 | number      |

### setToken

setToken sets a token for Gogen's replacement engine.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| name             | Token name, represented by `$name$` in the event                                               | string      |
| value            | Token value                                                                                    | string      |
| field            | Field to replace in, defaults to `_raw`                                                        | string      |

### removeToken

removeToken removes a token set via `setToken`.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| name             | Token name to remove                                                                           | string      |

### getLine

getLine returns a line from the sample config.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| line             | Line number to return                                                                          | number      |

*Returns:*

| Return           | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| line             | Table of key/value parameters from the sample config                                           | table       |

### getLines

getLines returns all lines from the sample config.

*Returns:*

| Return           | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| lines            | Table of tables of key/value parameters from the sample config                                 | table       |

### getChoice

getChoice returns all choices from the sample config for a given token.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| token            | Token to retrieve                                                                              | string      |

*Returns:*

| Return           | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| choices          | Table of choices of key/value parameters from the sample config                                | table       |

### getChoiceItem

getChoice returns a choice item from the sample config.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| token            | Token to retrieve                                                                              | string      |
| index            | Retrieve choice at line `index`                                                                | number      |

*Returns:*

| Return           | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| choice           | Choice at `index` from the sample config                                                       | string      |

### getFieldChoice

getFieldChoice returns all fieldChoices from the sample config for a given token.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| token            | Token to retrieve                                                                              | string      |
| field            | Field to retrieve                                                                              | string      |

*Returns:*

| Return           | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| choices          | Table of choices of key/value parameters from the sample config                                | table       |

### getFieldChoiceItem

getFieldChoiceItem returns a choice from the sample config.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| token            | Token to retrieve                                                                              | string      |
| field            | Field to retrieve                                                                              | string      |
| index            | Retrieve choice at line `index`                                                                | number      |

*Returns:*

| Return           | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| choice           | Choice of `field` at `index` from the sample config                                            | string      |

### getWeightedChoiceItem

getWeightedChoiceItem returns a choice from the sample config.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| token            | Token to retrieve                                                                              | string      |
| index            | Retrieve choice at line `index`                                                                | number      |

*Returns:*

| Return           | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| choice           | Choice `index` from the sample config                                                          | string      |

### getGroupIdx

getGroupIdx returns the index for a particular choice.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| choices          | Choices, returned from a previous `replaceTokens` call                                         | userdata    |
| group            | Index of group to return                                                                       | number      |

*Returns:*

| Return           | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| choice           | Index of the row chosen for a particular group                                                 | numbe       |


### setTime

setTime sets the time for a particular interval.

| Parameter        | Description                                                                                    | Type        |
|------------------|------------------------------------------------------------------------------------------------|-------------|
| time             | Epoch Time to set time to                                                                      | number      |
