# Examples

In addition to our tutorial, we've put together a number of good example configs to get you started.  

| Example                                                    | Description                                             |
|------------------------------------------------------------|---------------------------------------------------------|
| [Weblog](examples/weblog/weblog.yml)                       | Quientissential log.  Example used throughout the tutorial.  
| [CSV](examples/csv/csv.yml)                                | Generates CSV Data.  Showcases Gogen can be used for use cases aside from time series data. 
| [UNIX](examples/weblog/nix.yml)                            | This example showcases custom generators.  Generates data like running `df`, `ps`, etc on a UNIX box the way Splunk's UNIX TA collects data.  Variables correlate and make heavy use of the LUA Generators feature. 
| [Users](examples/generators/users.yml)                     | This example also showcases LUA Generators.  Simulates a group of users going through a set of actions.  Show cases complex interplay that's only possible by writing code.
| [Splunktel Demo](https://github.com/coccyx/splunktel_demo) | This example uses Gogen as a Splunk app and has a five relevant configs for Gogen stored in gogen_assets.  Shows to how to make a composable, managable Gogen config.

If you have relevant examples, send a PR and we'll add them to this list!
