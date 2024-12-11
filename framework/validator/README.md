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
You can create application-specific validation logic beyond the built-in rules provided by `validator/v10`. This package provides an interface called [`CustomValidator`](custom_validator.go), which allows you to define custom validation rules, tags, and error message translations.

To create a custom validator, you must implement the [`CustomValidator`](custom_validator.go) interface, which consists of three methods:
```golang
type CustomValidator interface {
	// Tag returns the tag identifier used in struct field validation tags (e.g., `validate:"mytag"`).
	Tag() string
	// Func returns the validator.Func that performs the validation logic.
	Func() validator.Func
	// Translation returns the translation text and a custom translation function for the custom validator.
	// If you wish to use the default translation, return an empty string and nil.
	Translation() (translation string, customFunc validator.TranslationFunc)
}
```
Each method in the [`CustomValidator`](custom_validator.go) interface has a specific purpose:
1.	`Tag()`: Defines the unique identifier for your custom validation rule. This tag is used in struct field tags to apply the validator. For example, if `Tag()` returns `"mytag"`, you can use it in a struct as `validate:"mytag"`.
2.	`Func()`: Specifies the validation logic. This method returns a function that implements the `validator.Func` type, which is called by the validator during the validation process. The function receives a `FieldLevel` instance representing the field to validate. You should return `true` if the field passes validation, and `false` otherwise.
3.	`Translation()`: Defines the <u>error message</u> and <u>translation function</u> for this validator. If you want to use a custom error message format, return it in <u>`translation`</u> (using `{0}` for the field name and `{1}`, etc., for any parameters). The <u>`customFunc`</u> parameter allows additional customization for translating the error message, providing control over how the message displays.
	- If you return `""` (empty string) and `nil` for `Translation()`, the default error message will be used.

### Example: Creating a Custom Validator
The following example demonstrates how to create a custom validator that always fails and provides a custom error message:
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
**Explanation**
- **Tag**: The `Tag()` method returns `"mytag"`, which you can use in struct tags like `validate:"mytag"`.
- **Func**: The `Func()` method returns a function that always fails for demonstration purposes.
- **Translation**: The `Translation()` method returns a custom <u>error message</u> and <u>translation function</u> The error message is `"{0} failed custom validation"`, where `{0}` will be replaced with the field name (`fe.Field()`).

### Using Custom Validators
To register a custom validator, you need to pass an instance of the custom validator to the `NewValidator` function using the `WithCustomValidator` option. 
The following example demonstrates how to use the custom validator created in the previous section:
```golang
type Data struct {
    Field1 string `validate:"mytag"`
    Field2 string `validate:"required"`
    Field3 int    `validate:"gte=0,lte=130"`
}

func main() {
    v, err := validator.NewValidator(
        validator.WithCustomValidator(new(MyValidator)),
        // Add more custom validators here
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
> Note: The package is built to make adding your own validators straightforward. If you need a domain-specific custom validator, you can implement the `CustomValidator` interface and use it directly in your codebase, without waiting for a merge into `go-common`. If you feel your custom validator could benefit others, consider sharing it here via a pull request or issue.

### Using Custom Field Names in Validation Errors
To make validation errors more readable, especially in APIs that use JSON serialization, you can customize field names in error messages to match the json struct tags. The package provides a convenient option for this with `WithTagNameFunc`.
- **JSON Tag Name Function**: The package includes a predefined `JSONTagNameFunc` to automatically use JSON field names in validation error messages.
### Registering JSONTagNameFunc
To register `JSONTagNameFunc` when creating a Validator instance, use the `WithTagNameFunc` option:
```golang
v, err := validator.NewValidator(
	validator.WithTagNameFunc(validator.JSONTagNameFunc),
)
```
With this setup, any validation error messages will use the names specified in the json tags. For example:
```golang
type User struct {
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"gte=0,lte=130"`
}

user := User{
	Email: "invalid-email",
	Age:   150,
}

err := v.ValidateStruct(user)
if err != nil {
	fmt.Println("Validation failed:", err)
}
```
- **Expected Output:**
    ```
    Validation failed: full_name is a required field, email must be a valid email address, age must be 130 or less
    ```

## Examples
- You can find a complete working example in the repository under [framework/validator/example](example/).
- You can find an implementation example of a custom validator in the repository under [framework/validator/custom_validator](custom_validator/).