# photosync
Util to do things with Google Photos

## Setup
To get setup, just get the repo and run `cp config/example-config.json config/config.json` and fill in the new config as required for your accounts.
Some possible entries for `picture-path-substrings-to-ignore` would be "from others", "from person a", etc. as these could be folders of photos already imported from other people.


## Running the space saver script
Running this script will list our all the image urls that you might want to remove (that are taking your storage space)
```
go run cmd/spacesaver/main.go
```
