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
* Add md5sum for gogen exe's and download new gogen from modinput if md5sum does not match
* Add timemultiple
* Consider finding a way to break up config package and refactor using better interface design
* Unit test coverage 90%
