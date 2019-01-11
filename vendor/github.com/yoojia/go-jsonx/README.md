# Go JSONx

> Not a JSON serialization library, just a helper functions package.

## Install

> go get -u github.com/yoojia/go-jsonx

## Functions

### CompressJSON

Compress a JSON bytes to a single line

```golang
	txt := `


    {
        "key with spaces": 2018,
        "email": 				"yoojia     chen@gmail.com",
        "github.com": {
            "id": 	"yoojia",
            "year": 		2018
        }
    }`

    outBytes := jsonx.CompressJSONText([]byte(txt))
    fmt.Println(string(outBytes))
```

The output:

> {"key with spaces":2018,"email":"yoojia     chen@gmail.com","github.com":{"id":"yoojia","year":2018}}

As you see above, a multi-lines JSON text has been compressed into a single line text.

### HasJSONMark

Deeply check if a string/bytes has a JSON mark: `{}` or `[]`

- `HasJSONMark(...)`

### FatJSON

FatJSON, is a manual JSON text builder, without any serialization, all json fields set by your self.

```golang
fj := NewFatJSON()

fj.Field("username", "yoojia")
fj.Field("year", "1999")
fj.Field("timestamp", "2018-04-21")

fj.FieldNotEscapeValue("json", `{"age": 18}`)

fj.Field("year", 2018)

out := fj.String()

fmt.Println(out)

if `{"username":"yoojia","year":"1999","timestamp":"2018-04-21","json":{"age": 18},"year":2018}` != out {
    t.Error("json text not matched")
}
```

The output:

> {"username":"yoojia","year":"1999","timestamp":"2018-04-21","json":{"age": 18},"year":2018}






