# upload-iam-certs
Uploads server certificates to AWS IAM


The app now supports three main operations:

Upload - requires -name, -cert, and -key (optional -chain)
List - use -list flag
Delete - requires -delete and -name flags


Upload a certificate:
```go run main.go -name my-cert -cert /path/to/cert.pem -key /path/to/key.pem```

Upload with certificate chain:
```go run main.go -name my-cert -cert /path/to/cert.pem -key /path/to/key.pem -chain /path/to/chain.pem```

List existing certificates:
```go run main.go -list```

Specify region:
```go run main.go -name my-cert -cert cert.pem -key key.pem -region us-west-2```

Delete a certificate:
```go run main.go -delete -name my-cert```

Delete with specific region:
```go run main.go -delete -name my-cert -region us-west-2```