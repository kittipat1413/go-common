[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Total Views](https://img.shields.io/endpoint?url=https%3A%2F%2Fhits.dwyl.com%2Fkittipat1413%2Fgo-common.json%3Fcolor%3Dblue)](https://hits.dwyl.com/kittipat1413/go-common)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# Validator Package
The validator package provides an extensible, customizable validation system built on top of `validator/v10`. It includes support for custom validation rules and error translations, making it ideal for applications that require flexible and descriptive error handling.

## Features
- Easy integration with `go-playground/validator`.
- Support for custom validators with custom error messages.
- Simplified API for struct validation.
- Extensible design for adding more `custom validators`.

## Usage

### Basic Validation
```golang
package main

import (
    "fmt"
    "github.com/kittipat1413/go-common/framework/validator"
)

type User struct {
    Name  string `validate:"required"`
    Email string `validate:"required,email"`
    Age   int    `validate:"gte=0,lte=130"`
}

func main() {
    v, err := validator.NewValidator()
    if err != nil {
        fmt.Println("Error initializing validator:", err)
        return
    }

    user := User{
        Name:  "Alice",
        Email: "alice@example.com",
        Age:   30,
    }

    err = v.ValidateStruct(user)
    if err != nil {
        fmt.Println("Validation failed:", err)
    } else {
        fmt.Println("Validation passed")
    }
}
```
## Custom Validators
You can create and register custom validators by implementing the [`CustomValidator`](custom_validator.go) interface and passing the custom validator when initializing the validator instance.

### Registering a Custom Validator
To register a custom validator, implement the [`CustomValidator`](custom_validator.go) interface and use `WithCustomValidator` when creating the validator instance.
```golang
package main

import (
    "fmt"
    ut "github.com/go-playground/universal-translator"
    validatorV10 "github.com/go-playground/validator/v10"
    "github.com/kittipat1413/go-common/framework/validator"
)

type MyValidator struct{}

// Tag returns the tag identifier used in struct field validation tags.
func (*MyValidator) Tag() string {
    return "mytag"
}

// Func returns the validator.Func that performs the validation logic.
func (*MyValidator) Func() validatorV10.Func {
    return func(fl validatorV10.FieldLevel) bool {
        // Custom validation logic (always fails for demonstration)
        return false
    }
}

// Translation returns the translation text and an custom translation function for the custom validator.
func (*MyValidator) Translation() (string, validatorV10.TranslationFunc) {
    translationText := "{0} failed custom validation"
    customTransFunc := func(ut ut.Translator, fe validatorV10.FieldError) string {
        // {0} will be replaced with fe.Field()
        t, _ := ut.T(fe.Tag(), fe.Field())
        return t
    }
    return translationText, customTransFunc
}
```

### Using Custom Validators

```golang
type Data struct {
    Field1 string `validate:"mytag"`
    Field2 string `validate:"required"`
    Field3 int    `validate:"gte=0,lte=130"`
}

func main() {
    v, err := validator.NewValidator(
        validator.WithCustomValidator(new(MyValidator)),
    )
    if err != nil {
        fmt.Println("Error initializing validator:", err)
        return
    }

    data := Data{
        Field1: "test", // Will fail due to custom validation logic in MyValidator
        Field3: 200,    // Out of range (greater than 130)
    }

    err = v.ValidateStruct(data)
    if err != nil {
        fmt.Println("Validation failed:", err)
    } else {
        fmt.Println("Validation passed")
    }
}
```
- **Expected Output:**
    ```
    Validation failed: Field1 failed custom validation, Field2 is a required field, Field3 must be 130 or less
    ```

## Examples
You can find a complete working example in the repository under [framework/validator/example](example/).