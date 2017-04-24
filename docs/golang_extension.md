# Golang (plugin) extensions

Gohan supports extensions written in golang and imported as shared libraries.
Golang plugins are supported starting from version 1.8 of golang. Note that
there exist a bug [1] which will be fixed in version 1.9.

## Why golang extensions as a plugin?

In contrast to golang extensions "as a callback" feature golang plugins are loadable at runtime
and do not require compilation together with gohan binary. It isn't always suitable to build
extensions at the same time as gohan. On the other hand, golang extensions via plugins are similar
to javascript extensions since they both do not require to be compiled with gohan.

It is easier to manage golang extension for each independent project. Additional benefit is that
this enables developers to implement event handler more intuitively.
 
When comparing to javascript, golang extensions give much higher performance sincethe code is
compiled to native code (not interpreted as javascript extensions) and run fast.
 
In contrast to javascript, golang is a typed language which
turns out to be very helpful for developers. It is also possible to use a debugger
to find problems in code easily.

# Creating a simple golang extension

## Extension structure

An extension consists of an init.go file placed in loader subdirectory
and a number of implementation files.

# Schema definitions

To enable a golang extension, append extensions section to schema and specify
**code_type** as **goext**. Specify plugin library in url entry. Note that placing
golang code in code entry is not supported.

```yaml
  extensions:
  - id: example_golang_extension
    code_type: goext
    path: ""
    url: file://golang_ext.so
```

## Required exports

Each plugin must export Init function.

```go
func Init(env goext.IEnvironment) error {
    schema := env.Schemas().Find("resource_schema")
    schema.RegisterRawType(resource.Resource{})
    schema.RegisterResourceType(resource.Resource{})
    schema.RegisterEventHandler(goext.PreCreate, resource.HandlePreCreateEvent, goext.PriorityDefault)
    schema.RegisterEventHandler(resource.CustomEventName, resource.HandleEventCustomEvent, goext.PriorityDefault)
    return nil
}
```

## Compilation and run

To compile an extension run:

```bash
go build -buildmode=plugin init.go
```

## Extension environment

An environment is passed to Init function. It consists of modules which are available
for a golang extension: core, logger, schemas, sync. 

## Event handling

In golang extension one can register a global handler for a named event
and a schema handler for a particular schema and event.
Examples:

```go
func HandlePreUpdateInTransaction(context goext.Context, resource goext.Resource, env goext.IEnvironment) error {
	myResource := resource.(MyResource)
	return nil
}
```

## Resource management

TBA

# Testing golang extensions

Tests for golang extensions are plugin libraries.

## Example

A sample test may look like this:

TBA

# References:

[1] https://github.com/golang/go/commit/758d078fd5ab423d00f5a46028139c1d13983120
