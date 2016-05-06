# Ego, a Go-like template language

The Go standard library [template language](https://golang.org/pkg/text/template/) is pretty weird. Its inexplicably quirky pipeline syntax is clumsy to work with and bears almost no resemblance to the language it is implemented in and for. (Just try to index into an array the obvious way and see how that goes for you.)

**Ego** (Embedded Go) is an alternative template language that endeavors to be compact, expressive and decidedly Go-like. It is inspired by the [Play Framework template language](https://www.playframework.com/documentation/2.5.x/JavaTemplates), which was in turn inspired by [ASP.NET Razor](http://www.asp.net/web-pages/overview/getting-started/introducing-razor-syntax-c).

Ego's priorities are to be **familiar**, **expressive**, and to have first-rate **error reporting**. After those are complete **performance** will be emphasized.

## Project status

Ego is not ready for production. It currently supports core functionality, however many features are missing and more tests are needed. If you think this is an interesting direction for Go templates, your pull requests are most welcome!

## Related projects

There is at least one similar project that I'm aware of.

* [GoRazor](https://github.com/sipin/gorazor) – a Razor-like translator program which compiles template files into Go sources.


# Writing templates

Ego sources, like most template languages, are interpreted as static content in which is interspersed dynamic statements that are evaluated and output conditionally.

**The special `@` character introduces dynamic content** and dynamic statements generally look and work like regular Go code. You don't need to explicitly close a dynamic statement, the end is inferred from context.

This approach to isolating dynamic content allows you to more easily and naturally write code without dealing with a ton of fiddly `{{`double-braces`}}` everywhere.

    This is static content outside a dynamic statement.
    
    Here's a number: @(some_number)
    @if some_number > 1 {
      If the value of 'some_number' is > 1 then this content is written.
    }else{
      Otherwise this content is written.
    }

When executed with a context that contains the property `some_number: 2`, the following will be output.

	This is static content outside a dynamic statement.
	
	Here's a number: 2
	
      If the value of 'some_number' is > 1 then this content is written.

## Escaping special characters

When you need to use the literal `@` character within a template you must escape it with the `\` character, like so: `user\@example.com`.

By the same token, within a dynamic block the `}` character is significant and it must also be escaped in the same way:

	@if true {
		Here's a literal closing brace within a block: \}.
	}

# Executing templates

Generally, you will execute your templates within a Go application. As a convenience, a standalone compiler is also included for testing.

## Executing templates with `egoc`

The subpackge `egoc` includes an Ego compiler that you can use to execute templates from the command line.

For example, the following command will load JSON data from the file `data.json` and use it as the variable context to execute the templates `a.ego` and `b.ego`. Each template specified on the command line is executed in turn and output is written to standard output.

	$ egoc -context data.json a.ego b.ego

You can use `egoc` to try out the examples in the `examples` directory. Examples are organized as two separate files with the same base name, one formatted as JSON which contains the context data, and an Ego template.

	$ egoc -context basic.json basic.ego

## Executing templates in Go

Templates are compiled and then executed with a runtime and variable context to produce output. Generally this can be accomplished in just a few lines.
	
	// our template source, assume this exists
	var src string
	
	// compile our template source
	t, err := ego.Compile(src)
	if err != nil { /* ... */ }
	
	// setup the template runtime
	r := &ego.Runtime{
		Stdout:	os.Stdout, // where is output written to
	}
	
	// setup our variable context
	c := map[string]interface{}{
	  "some_number": 2,
	}
	
	// execute the templte
	err = t.Exec(r, c)
	if err != nil { /* ... */ }

If a template will be used repeatedly it might make sense to keep the compiled template (`t` in the source above) in memory so that the same source does not need to be repeatedly parsed.

# Documentation

Further documentation is available in `docs`.
