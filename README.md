# Track USCIS Case Processing Times
This is a utility to fetch USCIS case processing times as published at
https://egov.uscis.gov/processing-times/

Processing times, by default, are fetched for all forms and offices.
However, this can be controlled to work with specific combination(s) of forms
and offices using a configuration file.

This project is developed primarily to serve the needs of
[Volunteers of Legal Service](http://www.volsprobono.org/). However, this can
be used by anyone who wishes to track USCIS processing times closely.

## Building
Golang(1.11.4+) must be installed to be able build this project.

This project can be compiled for various operating systems and architectures.
Build instructions for some of the common ones are given below:

For windows:

    For 32-bit architectures:
    GOOS=windows GOARCH=386 go build

    For 64-bit architectures:
    GOOS=windows GOARCH=amd64 go build

For macOS:

    For 32-bit architectures:
    GOOS=darwin GOARCH=386 go build

    For 64-bit architectures:
    GOOS=darwin GOARCH=amd64 go build

For linux:

    For 32-bit architectures:
    GOOS=linux GOARCH=386 go build

    For 64-bit architectures:
    GOOS=linux GOARCH=amd64 go build


## Usage
This project, when built, should produce an executable named `uscis-tracker`.
On windows, the executable will be named `uscis-tracker.exe`

The `uscis-tracker` executable can be invoked with no configuration to fetch
processing time data for all forms and processing centers:

    uscis-tracker
    (or just double-click on the executable)

The output is, by default, dumped into a tab-separated text file in the user's
home directory. Location of the output file is displayed when the executable
finishes processing. However, the location of this file can be controlled by an
argument, `-output`, to the executable.

    uscis-tracker -output=/tmp/processing-times.txt


If data for all forms and processing centers isn't required, a configuration
file with required combinations of forms and processing centers maybe provided
to only include relevant data in the report. This can be provided with the
`-config` argument to the executable.

    uscis-tracker -config=/tmp/config

This configuration file is a simple text file containing combination(s) of
forms and processing centers separated by a comma. Each combination should
appear in a new line. And, the name of the center is NOT case-sensitive.
Sample configuration file `config` is included in this repository.


## Contributing
This is really my first attempt at writing Go code. I'm sure this can be
improved in many different ways going further. However, a few good starting
points, if one is interested in contributing, are:
- Add tests to the current code
- Currently data is dumped into a tab-separated text file, which is excepted to
  be opened with Excel using `tab` as the data separator. However, using
  something like [excelize](https://github.com/360EntSecGroup-Skylar/excelize),
  we can generate a valid Excel spreadsheet directly.
- There is definitely scope for refactoring the current code. If a
  configuration file is provided, it, currently, still fetches all the forms
  and offices but filters out the unnecessary data. This is unnecessary
  processing and adds to execution time.
