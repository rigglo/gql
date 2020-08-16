# gql

![tests](https://github.com/rigglo/gql/workflows/tests/badge.svg)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/rigglo/gql)](https://pkg.go.dev/github.com/rigglo/gql)
[![Coverage Status](https://coveralls.io/repos/github/rigglo/gql/badge.svg?branch=master)](https://coveralls.io/github/rigglo/gql?branch=master)

**Note: As the project is still in WIP, there can be breaking changes which may break previous versions.**

This project aims to fulfill some of the most common feature requests missing from existing packages, or ones that could be done differently.

## Roadmap for the package

- [x] Custom scalars
- [x] Extensions
- [x] Subscriptions
- [ ] Custom directives
  - [x] Field directives
  - [ ] Executable directives
  - [ ] Type System directives
- [ ] Opentracing
- [ ] Query complexity
- [ ] Apollo File Upload
- [ ] Apollo Federation
- [ ] Custom validation for input and arguments
- [ ] Access to the requested fields in a resolver
- [ ] Custom rules-based introspection
- [ ] Converting structs into GraphQL types
- [ ] Parse inputs into structs

## Getting started

Defining a type, an object is very easy, let's visit a pizzeria as an example.

```go
var PizzaType = &gql.Object{  
   Name: "Pizza",  
   Fields: gql.Fields{
      "id": &gql.Field{
         Description: "id of the pizza",
         Type:        gql.ID,
      },
      "name": &gql.Field{
         Description: "name of the pizza",
         Type:        gql.String,
      },
      "size": &gql.Field{
         Description: "size of the pizza (in cm)",
         Type:        gql.Int,
      },
   },
}
```

Next, we need a way to get our pizza, to list them, so let's define the query.

```go
var RootQuery = &gql.Object{  
   Name: "RootQuery",  
   Fields: gql.Fields{
      "pizzas": &gql.Field{
         Description: "lists all the pizzas",    
         Type:        gql.NewList(PizzaType),
         Resolver: func(ctx gql.Context) (interface{}, error) {
            return []Pizza{
               Pizza{
                  ID:1, 
                  Name: "Veggie", 
                  Size: 32,
               },
               Pizza{
                  ID:2, 
                  Name: "Salumi", 
                  Size: 45,
               },
            }, nil
         },
      },
   }, 
}
```

To have a schema defined, you need the following little code, that connects your root query and mutations (if there are) to your schema, which can be later executed.

```go
var PizzeriaSchema = &gql.Schema{
   Query: RootQuery,
}
```

At this point, what's only left is an executor, so we can run our queries, and a handler to be able to serve our schema.

For our example, let's use the default executor, but if you want to experiment, customise it, add extensions, you can create your own the gql.NewExecutor function. 
Let's fire up our handler using the github.com/rigglo/gql/pkg/handler package and also enable the playground, so we can check it from our browser.

```go
func main() {
   http.Handle("/graphql", handler.New(handler.Config{
      Executor:   gql.DefaultExecutor(PizzeriaSchema),
      Playground: true,
   }))
   if err := http.ListenAndServe(":9999", nil); err != nil {
      panic(err)
   }
}
```

After running the code, you can go to the http://localhost:9999/graphql address in your browser and see the GraphQL Playground, and you can start playing with it.