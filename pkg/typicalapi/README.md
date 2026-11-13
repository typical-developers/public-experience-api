# typicalapi
This is a client wrapper written in Golang to access content from the public experience api.

# Installation
`go get -u github.com/typical-developers/public-experience-api/pkg/typicalapi`

# Usage
> [!NOTE]
> For the time being, the BaseURL must be manually replaced.
> Internally this is already using the planned URL, which is not yet accessible.
```go
package main

import (
    "github.com/typical-developers/public-experience-api/pkg/typicalapi"
)

func main() {
    client := typicalapi.New(
        typicalapi.WithBaseURL(url.URL{
            Scheme: "https",
            Host: "public-staging.luckfire.dev",
        })
    )

    changelogs, resp, err := client.Oaklands.ListChangelogs(context.TODO())
    // ...
}
```