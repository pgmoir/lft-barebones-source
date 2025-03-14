# lft-barebones-source

This repo includes the source CMS files for the lft-barebones site. They are produced from redhat as cshtml razor file for delivery through ASP.net 

## Running Instructions

### Step 1 - Assuming that Go is installed on local, and repo cloned

```go
go install
```

NB You may need to run this, to download the additional package

```go
go get golang.org/x/net/html
```


### Step 2 - Extracting views to json files in feeds folder

```go
go run runextract.go
```

## Use of feeds output

The purpose of this extraction is to generate json files that are used on the lft-barebones website.
The alternative for a production system would obviously be to host these in a database, and since they are static, they should be cached.
With more modern approaches (arguably, older approach being revisited), the pages would be parsed on server, again cached, and then passed down to the browser.