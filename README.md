# AWS Mail Parser

![AWSMailParser](https://jelius.dev/assets/AWSMailParser.png)

A lightweight Go service that polls SQS Events and fetches mails from S3 Bucket, it then saves it to your Mail server running on EC2 instance.  

## Features

- Polls events from SQS.
- Fetches mails from S3 bucket.
- Parse and save the mail to your mail server's mail directory.

## Installation

Clone the repository:

```bash
git clone https://github.com/jelius-sama/AWSMailParser.git
cd AWSMailParser
nvim . # Modify the `vars` package with your AWS config
./build.sh
```

## Usage

Run the mail parser:

```bash
./bin/AWSMailParser-1.x.x-linux-amd64
```

## How It Works

1. Polls SQS Events to check for new mails.
2. When a new SQS event is received it fetches mail from S3.
3. After fetching the mail from S3 it then parses it.
4. When it is done with parsing it saves the mail to your mail server's maildir.

## Requirements

* Go 1.24.5+ (Only to build otherwise binaries in the release section are static and can run without go being installed)
* Dovecot or similar mail server already up and running

## License

MIT License

