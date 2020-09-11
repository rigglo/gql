# Subscriptions

GraphQL Subscription is a really great solution if you want to (for example) provide live updates to the website you're building, or live notifications, messaging, etc.. Because of this, I decided to implement it in gql, but in a way that it's compatible with other existing WebSocket solutions, for example with [graph-gophers/graphql-transport-ws](https://github.com/graph-gophers/graphql-transport-ws). But for this purpose, there's also a new package, called [gqlws](https://github.com/rigglo/gqlws), built along with the `gql` package, also in a way to be compatible with other GraphQL packages.  

This example will use the `rigglo/gqlws` to add GraphQL subscriptions to our project.

## How it works?

To subscribe to your data stream, `gql` has a `Subscribe` function defined on the executor, it returns a go channel and an error, this will have all the messages from the resolver's channel.

!> The subscription resolvers has to return a `chan interface{}`, otherwise it will fail to execute and will return an `"invalid subscription"` error

To access to the subscriptions, and to actually use them, you'll need a handler which translates the communication, requests from your client to the executor, and that's why we have `gqlws`.

!> When the client unsubscribes, the context will be cancelled, so exiting from your goroutine started in the resolver is in your best interest.

## Defining a subscription

In our example, we'll define a subscription that if you subscribe to, will return the UNIX timestamp in every 2 second.

For this task, we have a predefined function, that returns a channel and pushes the time to it in every 2 seconds.

```go
func pinger() chan interface{} {
	ch := make(chan interface{})
	go func() {
		for {
			time.Sleep(2 * time.Second)
			ch <- time.Now().Unix()
		}
	}()
	return ch
}
```

Like a Query or a Mutation, subscriptions also need a root object, and your subscriptions are defined as its fields.

```go
var RootSubscription = &gql.Object{
    Name: "Subscription",
    Fields: gql.Fields{
        "server_time": &gql.Field{
            Type: gql.Int,
            Resolver: func(c gql.Context) (interface{}, error) {
                out := make(chan interface{})
                go func() {
                    ch := pinger()
                    for {
                        select {
                        case <-c.Context().Done():
                            // close some connections here
                            log.Println("done")
                            return
                        case t := <-ch:
                            // could implement some logic here, filtering or some additional work with the data
                            log.Println("sending server time")
                            out <- t
                        }
                    }
                }()
                return out, nil
            },
        },
    },
}
```

Also don't forget to add the root subscription to the schema, that's required to be able to use it.

```go
var Schema = &gql.Schema{
    Query:        RootQuery,
    Subscription: RootSubscription,
}
```

## Using gqlws

When you have all your schema defined, you're ready to register a handler and start using it. Adding Subscription support to your endpoint is not require a big change if you already have an executor and a handler defined. The `gqlws` package has a handler, which you can configure with the executor and existing handler.

From start to the end, in case of our example, the main function where we create our executor, GraphQL handler and then the `gqlws` handler which we register using the `http.Handle` function, will look like the following

```go
func main() {
	exec := gql.DefaultExecutor(Schema)

	h := handler.New(handler.Config{
		Executor:   exec,
		Playground: true,
	})

	wsh := gqlws.New(
		gqlws.Config{
			Subscriber: exec.Subscribe,
		},
		h,
	)

	http.Handle("/graphql", wsh)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		panic(err)
	}
}
```

> The full code for the example can be found [here](https://github.com/rigglo/gql-examples/blob/master/subscriptions/main.go).

To check if it works correctly, you can go to [http://localhost:9999/graphql](http://localhost:9999/graphql) for the Playground and then run the following query

```graphql
subscription {
  server_time
}
```

## Authentication

With `gqlws`, if you want to authenticate your subscription, there's an `OnConnect` function you can set in the configuration. 

For example if you're using [apollographql/subscriptions-transport-ws](https://github.com/apollographql/subscriptions-transport-ws), you can set the `connectionParams` and the `authToken`. With the `OnConnect` function, you can authenticate your subscriptions and update, add values to the context of the execution, like the current user, etc.

Here's a basic example:

```go
var wsConf = gqlws.Config{
    Subscriber: exec.Subscribe,
    OnConnect: func(ctx context.Context, params map[string]interface{}) (context.Context, error) {
        if authToken, ok := params["authToken"]; ok {
            if user, ok := sessions[authToken.(string)]; ok {
                return context.WithValue(ctx, "currentUser", user), nil
            }
        }
        return ctx, errors.New("invalid token, token not found")
    }
}
```