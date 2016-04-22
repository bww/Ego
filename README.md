# Ego, a Go-like template language

The Go standard library [template language](https://golang.org/pkg/text/template/) is pretty bad. It's clumsy to work with and, inexplicably, bears almost no resemblance to the language it is implemented in and for.

**Ego** (Embedded Go) is a *work-in-progress* alternative template language that endeavors to be compact, expressive and decidedly Go-like. It is based conceptually on [Play! Framework templates](https://www.playframework.com/documentation/2.5.x/JavaTemplates).

## Ego is work in progress

Ego is not ready for production. Basic functionality such as expressions, `if` and `for` statements work, however many features are missing and more tests are needed.

## Writing templates

Templates are interpreted as static content that is passed through unmodified except for dynamic statements which are evaluated and output dynamically.

The special '**@**' character introduces a dynamic statement. You don't need to explicitly close a dynamic statement, the end is inferred from context. This allows you to more easily write dynamic content without dealing with a ton of fiddly braces closing not just a statement itself, but each individual line!

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

## Using templates

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
