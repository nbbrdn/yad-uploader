# YaD Uploader

A command-line tool in Go for synchronizing files between a local directory and a remote WebDAV server. This tool compares the files in a specified local folder with those in a remote WebDAV directory and uploads any missing local files to the remote server. Optionally, it supports file filtering using a filename mask.

## Features
* Connects to a specified WebDAV server directory to retrieve a list of files.
* Compares files in a local directory and uploads any that are missing on the server.
* Supports file filtering by name or extension using regular expressions.
* Logs all synchronization operations to a specified log file.

## Prerequisites
[Go](https://golang.org/dl/) 1.18 or later.

## Installation
Clone this repository:
```
git clone https://github.com/nbbrdn/yad-uploader.git
cd yad-uploader
```
Build the executable:
```
go build -o yad-uploader
```

## Usage
To run the tool, use the following command:
```
./yad-uploader -webdav-url="https://example.com/webdav" -remote-folder="/remote/path" -username="user" -password="pass" -local-folder="/local/path" -file-mask="^Terr.*\\.zip$"

```

### Command-Line Options
`-webdav-url` : The base URL of the WebDAV server (required).

`-remote-folder` : Path to the directory on the WebDAV server where files will be checked (required).

`-username` : WebDAV server username (required).

`-password` : WebDAV server password (required).

`-local-folder` : Path to the local folder containing files to sync (required).

`-log-file` : Name of the log file for recording sync operations (default: sync_log.txt).

`-file-mask` : A regular expression to filter files by name or extension. For example, use ^Terr.*\\.zip$ for files starting with "Terr" and ending in .zip (optional).

### Example
The following example syncs all local `.zip` files beginning with "Terr" from `/local/path` to the `/remote/`path directory on the WebDAV server:

```
./yad-uploader -webdav-url="https://example.com/webdav" -remote-folder="/remote/path" -username="user" -password="pass" -local-folder="/local/path" -file-mask="^Terr.*\\.zip$"
```

### Logging
All synchronization actions are recorded in the log file specified by the `-log-file `option. By default, this file is located in the local folder. Each sync entry includes a timestamp and details on uploaded files and any errors encountered.

## Contributing
Feel free to open issues and submit pull requests to improve this tool. We welcome any suggestions for new features or enhancements.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
