# go-idl
go-idl is a micro service api definition and generation tool.

## IDL specification

### specification describe symbol
<KeyOrValue>: required pos declare "Key or value"
[KeyOrValue]: optional pos declare "Key or value"
VALUE1|VALUE2|VALUE3: value choices

### file
source files: IDL parse `.gidl` files
output files: tool can output different file on demands, like .yaml, .json, .proto, .go, etc.

### syntax
declare idl syntax specification version.
```
syntax = "v0.1.0" // TODO 版本检查
```

### comment
```
// single line comment

/**
...
multiple line comment
*/
```

### typedef
TODO

### model
```
model User {
    Id string `json:"id"`
    Name string `json:"name"`
    Address string `json:"address"`
}
```

### rest
```
rest <UpdateUserInfo> <GET|HEAD|POST|PUT|PATCH|DELETE|CONNECT|OPTIONS|TRACE|ANY> "/app/user/info/update/:user_id, /manage/user/info/update/:user_id" {
    req {
        [Header|Uri|Query|Body] {
        }
    }
    
    // or
    req {
       ContentType string `header:"Content-Type"` 
       UserId string `uri:"user_id"`  
       Ts int64 `json:"ts"`
       Address string `json:"address" form:"address"` 
    }
    
    // or
    req UpdateUserInfoReq
    
    resp {
        Code int64 `json:"code"` 
        Msg string `json:"msg"`
        Data {
            HistoryAddress []string `json:"history_address"`
            UserMap map[int64]User `json:"user_map"`
        } `json:"data"`
    }
    
    // or
    resp UpdateUserInfoResp
}
```

### grpc
```
grpc <GetUserInfoHandler> {
    req GetUserInfoReq {
    }
    
    resp GetUserInfoResp {
    }
}
```

### ws
```
ws <Name> <UP|DOWN> <1234|CodeHeartBeat> {
    // ... struct
}
```

### service
```
service ExampleService {
}
```

### status codes
TODO

### import
import go package
```
import "github.com/hacksomecn/go-idl/example/model"
```

### source other .gidl file ????
TODO ???

### decorator
use `@` to name a decorator key. Except idl remain `@idl_` prefix,  user can use decorator to define custom symbol, 
and attach it to other definition.
for example:
```
@MarkIt xxxxx
model RestReqCommon {
}
```

idl system decorator:

| keyword            | description         | usage                |
| ------------------ | ------------------- | -------------------- |
| `@idl_grpc_syntax` | grpc syntax version | `@idl_grpc_syntax 3` |
|                    |                     |                      |
|                    |                     |                      |

### raw
use go to declare raw code/text.
```
raw `
    type A = "123"
`
```

## parser detail
Idl currently uses a handwritten parser. About handwritten parser or generated parser like ANTLR、BISON、Yacc, read: https://medium.com/swlh/writing-a-parser-getting-started-44ba70bb6cc9

## intellij idea editor set up
- add `File Types` `*.gidl`
- add `live template` go-idl