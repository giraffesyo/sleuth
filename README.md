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

## Starting the database

From the root directory, run

```shell
docker compose up -d
```

Connecting to the mongo database can be done via tools like [MongoDB Compass](https://www.mongodb.com/products/tools/compass) or using `mongosh`

i.e. `mongosh mongodb://localhost:9000/?directConnection=true`

## Running without building first

```shell
# from root directory
export MONGODB_URI="mongodb://localhost:9000/?directConnection=true"
go run cmd/sleuth/main.go search -q "body found"
```

## Audio Transcription

### Requirements:

1. Python 3.11 (`brew install python@3.11`)
2. ffmpeg (`brew install ffmpeg`)

### Installation:

1. Create a virtual environment

```bash
python3.11 -m venv venv
```

2. Activate the virtual environment (MacOS/Linux)

```bash
source venv/bin/activate
```

3. Upgrade pip version (optional)

```bash
pip install --upgrade pip
```

4. Install the dependencies

```bash
pip install -r requirements.txt
```
