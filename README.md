# gql (WIP)

![GitHub Workflow Status](https://img.shields.io/github/workflow/status/rigglo/gql/Go)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/rigglo/gql)

**Note: As the project is still in WIP, there can be breaking changes which may break previous versions. As soon as the package reaches v1.0.0, this will change.**

This project aims to fulfill some of the most common feature requests that other Go graphql packages lack, or not perfect in.

For example

- Custom directives
- Custom scalars
- Access to the requested fields in a resolver
- Introspection based on auth, user defined rules
- Schema stitching
- Custom validation for input and arguments
