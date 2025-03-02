# Todo

* Starter Task #1:
    * Move templates to only having one template insteead of header, row and footer templates
    * Update unit tests and tutorial 4 example (hint for each row use {{ range }} operator in go templates)

* Starter Task #2:
    * Measurably increase unit test percentage
    * Look at report from https://coveralls.io/github/coccyx/gogen

* Starter Task #3:
    * Configuration Linter
    * Better error messages, better troubleshooting

* Starter Task #4:
    * Implement checkpointing state
        * Create channels back to each imer thread
        * Outputters should acknowledge output and that should increment state counters
        * Each timer thread should write current state after ack
        * This can also be used for performance counters

* Start Task #5:
    * Implement KBPS Rater

* Having a control REST interface
    * Manipulate running state through REST

* Configuration UI
* Add timemultiple
* Consider finding a way to break up config package and refactor using better interface design
* Unit test coverage 90%

## New Config Generator and Sharing System
* Update config sharing system to login via a hosted website to Github to validate username
* Share configs via a private S3 bucket rather than GitHub gists
* Create chat based UI for generating configs from LLM
* Make sharing via the web as easy as clicking a button
* Remove old GitHub code from Gogen