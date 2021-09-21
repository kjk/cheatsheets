---
title: Go
category: Go
---


# Basics

## Intro

`hello_world.go`:

```go
package main

import "fmt"

func main() {
  message := greetMe("world")
  fmt.Println(message)
}

func greetMe(name string) string {
  return "Hello, " + name + "!"
}
```

Run it with: `go run hello_world.go`
To build as an executable:
* `go mod init hello_world` : this creates `go.mod` file which declares this directory as module `hello_world`
* `go build` : this compiles all `.go` files in the directory. Because the package is `main`, this creates executable `hello_world` (`hello_world.exe` on Windows)

**Resources**

- [A tour of Go](https://tour.golang.org/welcome/1) _(tour.golang.org)_
- [Golang wiki](https://github.com/golang/go/wiki/) _(github.com)_
- [Effective Go](https://golang.org/doc/effective_go.html) _(golang.org)_
- [Go repl](https://repl.it/languages/go) _(repl.it)_
- [Go by Example](https://gobyexample.com/) _(gobyexample.com)_
- [Awesome Go](https://awesome-go.com/) _(awesome-go.com)_
- [JustForFunc Youtube](https://www.youtube.com/channel/UC_BzFbxG2za3bp5NRRRXJSw) _(youtube.com)_
- [Style Guide](https://github.com/golang/go/wiki/CodeReviewComments) _(github.com)_

Or try it out in the [Go repl](https://repl.it/languages/go), or [A Tour of Go](https://tour.golang.org/welcome/1).

## Variables

** Variable declaration**

```go
var msg string
msg = "Hello"
```

** Shortcut declaration (Infers `string` type from value)**

```go
msg := "Hello"
```

## Constants

```go
const Phi = 1.618
```

Constants can be character, string, boolean, or numeric values.

See: [Constants](https://tour.golang.org/basics/15)

## Basic types

|    Type    |               Set of Values                |                    Values                     |
| :--------: | :----------------------------------------: | :-------------------------------------------: |
|    bool    |                  boolean                   |                  true/false                   |
|   string   |            array of characters             |             needs to be inside ""             |
|    int     |                  integers                  |             32 or 64 bit integer              |
|    int8    |               8-bit integers               |                 [ -128, 128 ]                 |
|   int16    |              16-bit integers               |               [ -32768, 32767]                |
|   int32    |              32-bit integers               |          [ -2147483648, 2147483647]           |
|   int64    |              64-bit integers               | [ -9223372036854775808, 9223372036854775807 ] |
|   uint8    |          8-bit unsigned integers           |                  [ 0, 255 ]                   |
|   uint16   |          16-bit unsigned integers          |                 [ 0, 65535 ]                  |
|   uint32   |          32-bit unsigned integers          |               [ 0, 4294967295 ]               |
|   uint64   |          64-bit unsigned integers          |          [ 0, 18446744073709551615 ]          |
|  float32   |                32-bit float                |                                               |
|  float64   |                64-bit float                |                                               |
| complex64  | 32-bit float with real and imaginary parts |                                               |
| complex128 | 64-bit float with real and imaginary parts |                                               |
|    byte    |                sets of bits                |                alias for uint8                |
|    rune    |             Unicode characters             |                alias for int32                |

## Operators

**Arithmetic Operators**:

| Symbol | Operation | Valid Types |
|:---------:|:-------------:|:-------------:|
| `+` | Sum | integers, floats, complex values, strings |
| `-` | Difference | integers, floats, complex values |
| `*` | Product | integers, floats, complex values |
| `/` | Quotient | integers, floats, complex values |
| `%` | Remainder | integers |
| `&` | Bitwise AND | integers |
| `|` | Bitwise OR | integers |
| `^` | Bitwise XOR | integers |
| `&^` | Bit clear (AND NOT) | integers |
| `<<` | Left shift | integer << unsigned integer |
| `>>` | Right shift | integer >> unsigned integer |

**Comparison Operators**:

| Symbol | Operation |
|:---------:|:-------------:|
| `==` | Equal |
| `!=` | Not equal |
| `<` | Less |
| `<=` | Less or equal |
| `>` | Greater |
| `>=` | Greater or equal |

**Logical Operators**:

| Symbol | Operation |
|:---------:|:-------------:|
| `&&` | Conditional AND |
| `||` | Conditional OR |
| `!` | NOT |

## Strings

```go
str := "Hello"
```

```go
str := `Multiline
string`
```

Strings are of type `string`.

## Numbers

Typical types:

```go
num := 3          // int
num := 3.         // float64
num := 3 + 4i     // complex128
num := byte('a')  // byte (alias for uint8)
```

Other types:

```go
var u uint = 7        // uint (unsigned)
var p float32 = 22.7  // 32-bit float
```

## Arrays

```go
// var numbers [5]int
numbers := [...]int{0, 0, 0, 0, 0}
```

Arrays have a fixed size. They are rarely used explicitly. Instead you use slices which are a view onto underlying array.

## Slices

```go
slice := []int{2, 3, 4}
```

```go
slice := []byte("Hello")
```

Slices have a dynamic size, unlike arrays.

## Maps

Maps are key / value dictionsries

```go

// declare a map, has value of nil
var cities map[string]string

// must create a map before using
cities = map[string]string{}

// insert value
cities["NY"] = "EUA"

// retrieve
newYork = cities["NY"] // returns "EUA"

// delete
delete(cities, "NY")

// check if a key is setted
if value, ok := cities["NY"]; ok {
    println("found key 'NY' in map")
}

// iterate over keys and values:
for key, value := range cities {
    println("key:", key, "value:", value)
}
```

## Pointers
Pointers point to a memory location of a variable.

```go
func main () {
  b := *getPointer()
  fmt.Println("Value is", b)
}
```
{: data-line="2"}

```go
func getPointer () (myPointer *int) {
  a := 234
  return &a
}
```
{: data-line="3"}

```go
a := new(int)
*a = 234
```
{: data-line="2"}

Unlike C++, Go doesn't have pointer arithmetic (you can't add a number to a pointer to change the pointer).

See: [Pointers](https://tour.golang.org/moretypes/1)

## Type conversions

```go
i := 2
f := float64(i)
u := uint(i)
```

See: [Type conversions](https://tour.golang.org/basics/13)

## Structs

```go
type Vertex struct {
  X int
  Y int
}
```
{: data-line="1,2,3,4"}

```go
func main() {
  v := Vertex{1, 2}
  v.X = 4
  fmt.Println(v.X, v.Y)
}
```

See: [Structs](https://tour.golang.org/moretypes/2)

**Literals**

```go
v := Vertex{X: 1, Y: 2}
```

```go
// Field names can be omitted
v := Vertex{1, 2}
```

```go
// Y is implicit
v := Vertex{X: 1}
```

You can also put field names.

**Pointers to structs**

```go
v := &Vertex{1, 2}
v.X = 2
```

Doing `v.X` is the same as doing `(*v).X`, when `v` is a pointer.

## Functions

**Lambdas**

```go
myfunc := func() bool {
  return x > 10000
}
```
{: data-line="1"}

Functions are first class objects.

**Multiple return types**

```go
a, b := getMessage()
```

```go
func getMessage() (a string, b string) {
  return "Hello", "World"
}
```
{: data-line="2"}


**Named return values**

```go
func split(sum int) (x, y int) {
  x = sum * 4 / 9
  y = sum - x
  return
}
```
{: data-line="4"}

By defining the return value names in the signature, a `return` (no args) will return variables with those names.

See: [Named return values](https://tour.golang.org/basics/7)

## Methods

**Receivers**

```go
type Vertex struct {
  X, Y float64
}
```

```go
func (v Vertex) Abs() float64 {
  return math.Sqrt(v.X * v.X + v.Y * v.Y)
}
```
{: data-line="1"}

```go
v := Vertex{1, 2}
v.Abs()
```

There are no classes, but you can define functions with _receivers_.

See: [Methods](https://tour.golang.org/methods/1)

**Mutation**

```go
func (v *Vertex) Scale(f float64) {
  v.X = v.X * f
  v.Y = v.Y * f
}
```
{: data-line="1"}

```go
v := Vertex{6, 12}
v.Scale(0.5)
// `v` is updated
```

By defining your receiver as a pointer (`*Vertex`), you can do mutations.

See: [Pointer receivers](https://tour.golang.org/methods/4)

# Flow control
{: .-three-column}

## if

```go
if day == "sunday" || day == "saturday" {
  rest()
} else if day == "monday" && isTired() {
  groan()
} else {
  work()
}
```

See: [If](https://tour.golang.org/flowcontrol/5)

Statements in if:

```go
if _, err := doThing(); err != nil {
  fmt.Println("Uh oh")
}
```
{: data-line="1"}

A condition in an `if` statement can be preceded with a statement before a `;`. Variables declared by the statement are only in scope until the end of the `if`.

See: [If with a short statement](https://tour.golang.org/flowcontrol/6)

## switch

```go
switch n {
    case 2, 3, 5:
        println("n is prime number")
    case 1:
        // cases don't "fall through" by default but you
        // can do it explicitly:
        fallthrough
    case 3:
        println("n is either 1 or 3")
    default:
        println("all other values of n")
}
```

See: [Switch](https://github.com/golang/go/wiki/Switch)

**Type switch**:
```go
func printV(v interface{}) {
    switch rv := v.(type) {
        case int:
            println("v is of type int and has value of", rv)
        case *Foo:
            println("v is a pointer to struct Foo")
    }
}
```

## for / range / while loop

```go
for count := 0; count <= 10; count++ {
  fmt.Println("My counter is at", count)
}
```

See: [For loops](https://tour.golang.org/flowcontrol/1)

```go
entry := []string{"Jack","John","Jones"}
for i, val := range entry {
  fmt.Printf("At position %d, the character %s is present\n", i, val)
}
```

See: [For-Range loops](https://gobyexample.com/range)

Go doesn't have `while` keyword but you use `for`:
```go
n := 0
x := 42
for n != x {
  n := guess()
}
```

See: [Go's "while"](https://tour.golang.org/flowcontrol/3)

Use `break` to exit loop:
```go
i := 0
for {
    i++
    if (i > 10) {
        break
    }
}
```

Forever loop:
```go
for {
}
```

## defer

```go
func main() {
  defer fmt.Println("Done")
  fmt.Println("Working...")
}
```
{: data-line="2"}

Defers running a function until the surrounding function returns.
The arguments are evaluated immediately, but the function call is not ran until later.

See: [Defer, panic and recover](https://blog.golang.org/defer-panic-and-recover)

**Deferring functions**

```go
func main() {
  defer func() {
    fmt.Println("Done")
  }()
  fmt.Println("Working...")
}
```
{: data-line="2,3,4"}

Lambdas are better suited for defer blocks.

```go
func main() {
  var d = int64(0)
  defer func(d *int64) {
    fmt.Printf("& %v Unix Sec\n", *d)
  }(&d)
  fmt.Print("Done ")
  d = time.Now().Unix()
}
```
{: data-line="3,4,5"}
The defer func uses current value of d, unless we use a pointer to get final value at end of main.

## panic and recover

`panic()` is similar to throwing an exception. It's used extremely rarely in Go, typically to signal a condition so bad that it should exit the program.

Unhandled, it'll exit the program with error code and print a callstack for debugging.

You can catch a panic with `recover`:
```go
package main

import "fmt"

func main() {
    f()
    fmt.Println("Returned normally from f.")
}

func f() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered in f", r)
        }
    }()
    fmt.Println("Calling g.")
    g(0)
    fmt.Println("Returned normally from g.")
}

func g(i int) {
    if i > 3 {
        fmt.Println("Panicking!")
        panic(fmt.Sprintf("%v", i))
    }
    defer fmt.Println("Defer in g", i)
    fmt.Println("Printing in g", i)
    g(i + 1)
}
```

# Packages
{: .-three-column}

## Importing

```go
import "fmt"
import "math/rand"
```

```go
import (
  "fmt"        // gives fmt.Println
  "math/rand"  // gives rand.Intn
)
```

Both are the same.

See: [Importing](https://tour.golang.org/basics/1)

## Aliases

```go
import r "math/rand"
```
{: data-line="1"}

```go
r.Intn()
```

## Exporting names

```go
func Hello () {
  ···
}
```

Exported names begin with capital letters.

See: [Exported names](https://tour.golang.org/basics/3)

## Packages

```go
package hello
```

Every package file has to start with `package`.

# Concurrency
{: .-three-column}

## Goroutines

```go
func main() {
  // A "channel"
  ch := make(chan string)

  // Start concurrent routines
  go push("Moe", ch)
  go push("Larry", ch)
  go push("Curly", ch)

  // Read 3 results
  // (Since our goroutines are concurrent,
  // the order isn't guaranteed!)
  fmt.Println(<-ch, <-ch, <-ch)
}
```
{: data-line="3,6,7,8,13"}

```go
func push(name string, ch chan string) {
  msg := "Hey, " + name
  ch <- msg
}
```
{: data-line="3"}

Channels are concurrency-safe communication objects, used in goroutines.

See: [Goroutines](https://tour.golang.org/concurrency/1), [Channels](https://tour.golang.org/concurrency/2)

## Buffered channels

```go
ch := make(chan int, 2)
ch <- 1
ch <- 2
ch <- 3
// fatal error:
// all goroutines are asleep - deadlock!
```
{: data-line="1"}

Buffered channels limit the amount of messages it can keep.

See: [Buffered channels](https://tour.golang.org/concurrency/3)

## Closing channels

Closes a channel:

```go
ch <- 1
ch <- 2
ch <- 3
close(ch)
```
{: data-line="4"}

Iterates across a channel until its closed:

```go
for i := range ch {
  ···
}
```
{: data-line="1"}

Closed if `ok == false`:

```go
v, ok := <- ch
```

See: [Range and close](https://tour.golang.org/concurrency/4)

## WaitGroup

```go
import "sync"

func main() {
  var wg sync.WaitGroup
  
  for _, item := range itemList {
    // Increment WaitGroup Counter
    wg.Add(1)
    go doOperation(item)
  }
  // Wait for goroutines to finish
  wg.Wait()
  
}
```
{: data-line="1,4,8,12"}

```go
func doOperation(item string) {
  defer wg.Done()
  // do operation on item
  // ...
}
```
{: data-line="2"}

A WaitGroup waits for a collection of goroutines to finish. The main goroutine calls Add to set the number of goroutines to wait for. The goroutine calls `wg.Done()` when it finishes.
See: [WaitGroup](https://golang.org/pkg/sync/#WaitGroup)

## Race detector

Running code concurrently introduces a new class of bugs: modyfing memory from multiple goroutines. This is known as a data race.

To catch data races, build with `-race` flag (i.e. `go build -race` or `go run -race`).

This compiles the code with additional instrumentation that detects data races and aborts the program when that happens (so that you can fix the bug that caused data race).

# Advanced

## Interfaces

An interface defines a set of methods. Any struct implementing those methods can be used as a value of the interface

Interface definition:

```go
type Shape interface {
  Area() float64
  Perimeter() float64
}
```

Struct `Rectangle` implicitly implements interface `Shape` by implementing all of its methods.

```go
type Rectangle struct {
  Length, Width float64
}

func (r Rectangle) Area() float64 {
  return r.Length * r.Width
}

func (r Rectangle) Perimeter() float64 {
  return 2 * (r.Length + r.Width)
}
```

The methods defined in `Shape` are implemented in `Rectangle`.

Using interface:

```go

func printShapeInfo(s Shape) {
  fmt.Printf("Type of r: %T, Area: %v, Perimeter: %v.", s, s.Area(), s.Perimeter())
}

func main() {
  r := Rectangle{Length: 3, Width: 4}
  printArea(r)
}
```

## Testing
Go has a built-in support for testign. Test functions must be named   `Test*(t *testing.T)` and placed in `*_test.go` files.

Example: file `main.go` with function `Sum` to be tested:
```go
func Sum(x, y int) int {
    return x + y
}
```

File `main_test.go` with `TestSum` function testing `Sum`:

```go
import ( 
    "testing"
    "reflect"
)

// must s
func TestSum(t *testing.T) {
    x, y := 2, 4
    expected := 2 + 4

    if !reflect.DeepEqual(sum(x, y), expected) {
        t.Errorf("Function Sum not working as expected")
    }
}
```

Running tests:
* `go test` : run all tests in current package (directory)
*  `go test ./...` :  run tests in current package and all sub-packages)

## Go CLI

```bash
# Compile & Run code
$ go run [file.go]

# Compile
$ go build [file.go]
# Running compiled file
$ ./hello

# Test packages
$ go test [folder]

# Install packages/modules
$ go install [package]

# List installed packages/modules
$ go list

# Update packages/modules
$ go fix

# Format package sources
$ go fmt

# See package documentation
$ go doc [package]

# Add dependencies and install
$ go get [module]

# See Go environment variables
$ go env

# See version
$ go version
```


# Standard libs

## fmt

**Commonly used**: [Printf](https://pkg.go.dev/fmt#Printf), [Errorf](https://pkg.go.dev/fmt#MErrorf), [Sprintf](https://pkg.go.dev/fmt#Sprintf), [official docs](https://pkg.go.dev/fmt)

```go
import "fmt"

fmt.Printf("%s is %d years old\n", "John", 32) // Print with formatting
fmt.Errorf("User %d not found", 123) // Print a formatted error
s := fmt.Sprintf("Boolean: %v\n", true) // format to a string
```

## os

**Commonly used**: [Chdir](https://pkg.go.dev/os#Chdir), [Mkdir](https://pkg.go.dev/os#Mkdir), [MkdirAll](https://pkg.go.dev/os#MkdirAll), [ReadFile](https://pkg.go.dev/os#ReadFile), [Remove](https://pkg.go.dev/os#Remove), [RemoveAll](https://pkg.go.dev/os#RemoveAll), [Rename](https://pkg.go.dev/os#Rename), [WriteFile](https://pkg.go.dev/os#WriteFile), [official docs](https://pkg.go.dev/os)

```go
import "os"

err := os.Mkdir("dir", 0755)
err = os.RemoveAll("dir")

d := []byte("my data")
err = os.WriteFile("file.txt", d, 0644)

d, err = os.ReadFile("file.txt")
fmt.Printf("Content of file.txt:\n%s\n", string(d))

```