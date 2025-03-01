# Sleuth

Sleuth is a pipeline of tools for the analysis of News footage. It is designed to find multimedia content across various news sources, "providers", extract metadata, download the content, and ultimately create a dataset of news footage focused around murder and bodies found for further forensic analysis.

# Sleuth CLI

The sleuth CLI is responsible for searching various news providers for multimedia content to be further processed.

## Building

To build, install golang on your system and run `./build/build.sh`. This script will output the sleuth CLI into the root of the project.

After it's built, you can execute like this:

```
./sleuth search -q "murder cases"
```

## Running without building first

```shell
# from root directory
go run cmd/sleuth/main.go search -q "body found"
```
