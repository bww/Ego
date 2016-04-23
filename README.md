# Ego, a Go-like template language

The Go standard library [template language](https://golang.org/pkg/text/template/) is pretty weird. It's clumsy to work with and its inexplicably quirky pipeline syntax bears almost no resemblance to the language it is implemented in and for. (Just try to index into an array the obvious way and see how that goes for you.)

**Ego** (Embedded Go) is an alternative template language that endeavors to be compact, expressive and decidedly Go-like. It is inspired by the [Play Framework template language](https://www.playframework.com/documentation/2.5.x/JavaTemplates), which was in turn inspired by [ASP.NET Razor](http://www.asp.net/web-pages/overview/getting-started/introducing-razor-syntax-c).

## Ego is work in progress

Ego is not ready for production. Basic functionality such as expressions, `if` and `for` statements work, however many features are missing and more tests are needed. If you agree that the status quo in Go templates needs an update, your pull requests would be most welcome!

## Writing templates

Ego sources, like most template languages, are interpreted as static content in which is interspersed dynamic statements that are evaluated and output conditionally.

**The special `@` character introduces dynamic content** and dynamic statements generally look and work like regular Go code. You don't need to explicitly close a dynamic statement, the end is inferred from context.

This approach to isolating dynamic content allows you to more easily and naturally write dynamic content without dealing with a ton of fiddly `{{`double-braces`}}` everywhere.

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

## Executing templates

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
