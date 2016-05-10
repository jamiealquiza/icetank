# icetank
ec2 machine scheduler library

# example.go
```go
package main

import (
        "log"
        "time"

        "github.com/jamiealquiza/icetank"
)

func main() {
        pool := icetank.NewPool("vpc-3c61axxx", "webnodes", "us-west-2")
        
        _ := pool.Start(2)
        time.Sleep(time.Second * 10)
        _ = pool.Stop(2)
}
```

```
$ go run example.go
2016/05/10 17:39:01 [vpc-3c61axxx - webnodes] Pool created
2016/05/10 17:39:02 [vpc-3c61axxx - webnodes] Pool updated - Running: [] - Stopped: [i-aa98e072 i-33345feb i-84a7a05c i-4d4d2495 i-4b592293 i-86aad25e i-bc770c64 i-f69ae22e i-23c943fb i-0ea0ced6]
2016/05/10 17:39:02 [vpc-3c61axxx - webnodes] Requested start for 2 instances
2016/05/10 17:39:02 [vpc-3c61axxx - webnodes] Requesting start for [i-aa98e072 i-33345feb]
2016/05/10 17:39:33 [vpc-3c61axxx - webnodes] Pool updated - Running: [i-aa98e072 i-33345feb] - Stopped: [i-84a7a05c i-4d4d2495 i-4b592293 i-86aad25e i-bc770c64 i-f69ae22e i-23c943fb i-0ea0ced6]
2016/05/10 17:39:43 [vpc-3c61axxx - webnodes] Requested stop for 2 instances
2016/05/10 17:39:43 [vpc-3c61axxx - webnodes] Requesting stop for [i-aa98e072 i-33345feb]
2016/05/10 17:40:15 [vpc-3c61axxx - webnodes] Pool updated - Running: [] - Stopped: [i-aa98e072 i-33345feb i-84a7a05c i-4d4d2495 i-4b592293 i-86aad25e i-bc770c64 i-f69ae22e i-23c943fb i-0ea0ced6]
```
