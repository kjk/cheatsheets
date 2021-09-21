---
title: Go
category: Go
---
## Introduction

### Intro

Save this as `main.go`:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello Gophers!")
}
```

Run with: `go run main.go`


### Go Modules

* Go projects are called **modules**
* Each module has one or more **packages**
* Files for a package are in a directory
* A module needs at least one package, the **main**
* The package main needs a entry function called **main**

```bash
# Create Module
$ go mod init [name]
```

Tip: By convention, modules names has the follow structure: `domain.com/user/module/package`

Example: `github.com/spf13/cobra`

### Methods and Interfaces

Go doesn't have classes. But you can implement methods, interfaces and almost everything contained in OOP, but in what gophers call "Go Way"

```go
type Dog struct {
    Name string
}

func (dog *Dog) bark() string {
    return dog.Name + " is barking!"
}

dog := Dog{"Rex"}
dog.bark() // Rex is barking!
```

Interfaces are implicitly implemented. You don't need to inform that your struct are correctly implementing a interface if it already has all methods with the same name of the interface.
All structs implement the `interface{}` interface. This empty interface means the same as `any`.

```go
// Car implements Vehicle interface
type Vehicle interface {
    Accelerate()
}

type Car struct {

}

func (car *Car) Accelerate() {
    return "Car is moving on ground"
}
```

### Errors

Go doesn't support `throw`, `try`, `catch` and other common error handling structures. Here, we use `error` package to build possible errors as a returning parameter in functions

```go
import "errors"

// Function that contain a logic that can cause a possible exception flow 
func firstLetter(text string) (string, error) {
    if len(text) < 1 {
        return nil, errors.New("Parameter text is empty")
    }
    return string(text[0]), nil
}

a, errorA := firstLetter("Wow")
a // "W"
errorA // nil

b, errorB := firstLetter("")
b // nil
errorB // Error("Parameter text is empty")
```

