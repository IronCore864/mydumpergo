# Mydumper in Go

A wrapper for [mydumper](https://github.com/maxbube/mydumper) written in Golang with limited available storage in mind.

## Usage

Example:

```bash
BUCKET=mybucket-123123123 ./mydumpergo
```

## Configuration

Using ENV variables. For more details, see file [config/config.go](config/config.go).

Explanation:

- AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY: used to access AWS.
- REGION: the S3 bucket region, defaults to "eu-central-1".
- BUCKET: the name of the S3 bucket for uploading. Mandatory.
- CHUNK_FILE_SIZE_MB: Split tables into chunks of this output file size. This value is in MB. Defaults to 1 for local testing purpose.
- MAX_FILE_COUNT: maximum files allowed before pausing and uploading to S3. Defaults to 10 for local testing purpose.
- HOST: mysql host, defaults to "localhost".
- PORT: mysql port, defaults to 3306.
- USERNAME: mysql username, defaults to "root" for local testing purpose.
- PASSWORD: mysql password, defaults to empty string "" for local testing purpose.
- OUTPUTDIR: defaults to "output", used as the `-o` parameter for mydumper.

**NOTE:**

`CHUNK_FILE_SIZE_MB` * `MAX_FILE_COUNT` decides maximum storage usage. For example, if `CHUNK_FILE_SIZE_MB=50` and `MAX_FILE_COUNT=20`, total max disk usage is 1GB. This is an MVP implementation, so counting file number is used (which is simpler) rather than calculating output dir size.

## Main Logic

- start the mydumper process
- start a goroutine that checks disk usage, which loops:
    - if disk usage is higher than pre-configured threshold, send SIGTSTP to the mydumper process to pause it, then:
        - for each existing file in the output folder, upload the file (using the s3manager package's Uploader, which provides concurrent upload by taking advantage of S3's Multipart API)
        - after all existing files are uploaded, delete them to release disk usage
    - send SIGCONT to the mydumper process to continue
- after the mydumper process ends, clean up the output folder.

## TODOs

To make it production-ready, there are a few improvements to be done:

- Change counting files to counting disk usage. Counting file is only meant as a simple implementation for local testing purpose where you don't have large tables and output files.
- Add robust error handling, like retry for S3 upload, only delete successfully uploaded files, etc.
- Add unit test and functional test.
- Running in K8s: docker [multistage build](https://docs.docker.com/develop/develop-images/multistage-build/), adding mydumper into the same image; run as a K8s [job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) wth persistent volume, using the mountPath as the output dir.
