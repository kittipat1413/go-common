[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# Utility Packages for Common Operations

The `util` directory provides a collection of reusable helper utilities designed to simplify and standardize common tasks across your Go projects. These utilities are generic, lightweight, and decoupled from business logic, making them easy to use across any microservice in your ecosystem.

---

## Available Packages

### `pointer`
- Utility functions for working with pointers in a type-safe and ergonomic way using Go generics.

### `jwt`
- Wrapper around `github.com/golang-jwt/jwt/v5` for creating and validating JWTs, supporting multiple signing methods (`HS256`, `RS256`).

### `slice`
- Generic slice manipulation helpers.

## 🛠️ Future Additions

Planned or potential utility packages may include:
- `strings`: for string transformations and validations.
- `timeutil`: for time parsing and manipulation.
- `env`: for safe environment variable parsing.
- `config`: for config loading from file/env.