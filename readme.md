# Go Regex

A High Performance PCRE Regex Package That Uses A Cache.

Simplifies the the go-pcre regex package.
After calling a regex, the compiled output gets cached to improve performance.

This package uses the [go-pcre](https://github.com/GRbit/go-pcre) package for better performance.

## Installation

```shell
go get github.com/tkdeng/goregex
```

## Dependencies

### Debian/Ubuntu (Linux)

```shell script
  sudo apt install libpcre3-dev
```

### Fedora (Linux)

```shell script
  sudo dnf install pcre-devel
```

### Arch (Linux)

```shell script
  sudo yum install pcre-dev
```

## Usage

```go
import (
  "github.com/tkdeng/goregex"
)

func main(){
  // pre compile a regex into the cache
  // this method also returns the compiled pcre.Regexp struct
  regex.Comp(`re`)
  
  // compile a regex and safely escape user input
  regex.Comp(`re %1`, `this will be escaped .*`); // output: this will be escaped \.\*
  regex.Comp(`re %1`, `hello \n world`); // output: hello \\n world (note: the \ was escaped, and the n is literal)
  
  // use %n to reference a param
  // use %{n} for param indexes with more than 1 digit
  regex.Comp(`re %1 and %2 ... %{12}`, `param 1`, `param 2` ..., `param 12`);

  // return an error instead of panic on failed compile
  reg, err := regex.CompTry(`re`)

  // compile RE2 instead of PCRE
  reg := regex.CompRE2(`re`)
  reg, err := regex.CompTryRE2(`re`)

  
  // manually escape a string
  // note: the compile methods params are automatically escaped
  regex.Escape(`(.*)? \$ \\$ \\\$ regex hack failed`)
  
  // determine if a regex is valid, and can be compiled by this module
  regex.IsValid(`re`)
  
  // determine if a regex is valid, and can be compiled by the PCRE module
  regex.IsValidPCRE(`re`)
  
  // determine if a regex is valid, and can be compiled by the builtin RE2 module
  regex.IsValidRE2(`re`)
  
  // run a replace function (most advanced feature)
  regex.Comp(`(?flags)re(capture group)`).RepFunc(myByteArray, func(data func(int) []byte) []byte {
    data(0) // get the string
    data(1) // get the first capture group
  
    return []byte("")
  
    // if the last option is true, returning nil will stop the loop early
    return nil
  }, true /* optional: if true, will not process a return output */)
  
  // run a replace function
  regex.Comp(`re (capture)`).RepStr(myByteArray, []byte("test $1"))
  
  // run a simple light replace function
  regex.Comp(`re`).RepStrLit(myByteArray, []byte("all capture groups ignored (ie: $1)"))
  
  
  // return a bool if a regex matches a byte array
  regex.Comp(`re`).Match(myByteArray)
  
  // split a byte array in a similar way to JavaScript
  regex.Comp(`re|(keep this and split like in JavaScript)`).Split(myByteArray)
  
  // a regex string is modified before compiling, to add a few other features
  `use \' in place of ` + "`" + ` to make things easier`
  `(?#This is a comment in regex)`
  
  // an alias of pcre.Regexp
  regex.PCRE
  
  // an alias of *regexp.Regexp
  regex.RE2
  
  // direct access to compiled pcre.Regexp
  regex.Comp("re").RE

  
  // another helpful function
  // this method makes it easier to return results to a regex function
  regex.JoinBytes("string", []byte("byte array"), 10, 'c', data(2))
  
  // the above method can be used in place of this one
  append(append(append(append([]byte("string"), []byte("byte array")...), []byte(strconv.Itoa(10))...), 'c'), data(2)...)
}
```
