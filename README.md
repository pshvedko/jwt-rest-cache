# jwt-rest-cache-demo
REST API key/value cache service with `List`, `Get`, `Put`, `Delete` methods and [JWT](http://jwt.io) authorization. 

# Build & Run
Download the sources from the repository and run the service
```
git clone https://github.com/pshvedko/jwt-rest-cache.git
cd jwt-rest-cache
go build
./jwt-rest-cache
```

# Usage
The service can be used, for example, with `curl`

## Methods
*`Put` and `Delete` methods requires [JWT](http://jwt.io) authorization, 
`Authorization:` header can be found in `List` or `Get` responses and 
need to be added to the request.*

### List
The list of stored `KEY`s can be obtained with the command
```
curl -v http://127.0.0.1:8080/
< HTTP/1.1 200 OK
< Authorization: Bearer JWT.TOKEN
...
KEY
KEY
```

### Put
You can add or replace the `KEY` `VALUE` with the command using the `Authorization:` header
```
curl -v -X PUT http://127.0.0.1:8080/KEY -H 'Authorization: Bearer JWT.TOKEN' -d VALUE
```
or as a `VALUE` using the contents of the `FILE`
```
curl -v -X PUT http://127.0.0.1:8080/KEY -H 'Authorization: Bearer JWT.TOKEN' -T FILE
```
*additionally, the service will save the `Content-Type:` header and return it in the `Get` response*

### Get 
You can get the `KEY` `VALUE` with the command
```
curl -v http://127.0.0.1:8080/KEY
< HTTP/1.1 200 OK
...
VALUE
```

### Delete
You can delete the `KEY` with the command using the `Authorization:` header
```
curl -v -X DELETE http://127.0.0.1:8080/KEY -H 'Authorization: Bearer JWT.TOKEN'
```
after which the `KEY` will not be found
```
curl -v http://127.0.0.1:8080/KEY
< HTTP/1.1 404 Not Found
...
```

