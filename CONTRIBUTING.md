# Contributing

First of all, thanks for taking the time to contribute!

# How to contribute

## Reporting Bugs

Bugs are tracked as [Gitlab issues](https://gitlab.com/itk.fr/lorhammer/issues). Following these guidelines helps maintainers and the community understand your report, reproduce the behavior, and find related reports.

Explain the problem and include additional details to help maintainers reproduce the problem:

* Look at existing issues, if you find an already existing issue for your probem, don't hesitate to :+1: and add new details if possible.
* Use a clear and descriptive title for the issue to identify the problem.
* Detail your environment (windows, linux, mac...)
* Add the command line and/or the scenario file to reproduce the problem
* You can delete some private address or deployment strategy but don't forget to tell us what is your network server type
* Explain the behavior you expect to see instead and why.

## Code contribution

Merge-request (or pull-request) are made at [Gitlab merge-request](https://gitlab.com/itk.fr/lorhammer/merge_requests).

Before implementing a new feature, please open an issue to discuss the impacts and the design.

The continuous-integration will check :

* That the code is well formatted : `make lint` return no error
* That all tests are green, please add unit tests with your code : `make test` to run them
* That integration tests are also ok : launch some instances with scenarios located at `resources/scenarios/ci` to check performance regressions

Obviously, a merge-request will be accepted only if pipeline is green.

## Other

If you have any question, please create an issue, we will answer as soon as possible.