# Ego, a Go-like template language

The Go standard library [template language](https://golang.org/pkg/text/template/) is pretty weird. Its inexplicably quirky pipeline syntax is clumsy to work with and bears almost no resemblance to the language it is implemented in and for. (Just try to index into an array the obvious way and see how that goes for you.)

**Ego** (Embedded Go) is an alternative template language that endeavors to be compact, expressive and decidedly Go-like. It is inspired by the [Play Framework template language](https://www.playframework.com/documentation/2.5.x/JavaTemplates), which was in turn inspired by [ASP.NET Razor](http://www.asp.net/web-pages/overview/getting-started/introducing-razor-syntax-c).

Ego's priorities are to be **familiar**, **concise**, and to have excellent **error reporting**.

## Project status

Ego is not ready for production. It currently supports core functionality, however many features are missing and more tests are needed. If you agree that the status quo in Go templates needs an update, your pull requests would be most welcome!

## Writing templates

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

# Reference

The following dynamic constructs are currently supported.

## Statements

### `if`

Your standard `if` statement. Unlike Go, Ego does not support the `assignment; test` construct.

	@if expr_a {
		// evaluated if 'expr_a' is true
	}else if expr_b {
		// evaluated if 'expr_b' is true
	}else{
		// evaluated if none of the above are true
	}

### `for`

Your standard `for` statement. Unlike Go, Ego only supports `range` iteration. You can range over slice, array, and map types.

When ranging a map, you can declare one or two variables. If one variable is declared it contains the entry value. If two variables are declared the first is the entry key and the second is the entry value.

	@for k, v := range a_map {
	  Here's a map entry: @(k) -> @(v)
	}

When ranging an array or slice you can also declare one or two variables. If one variable is declared it contains the element value. If two variables are declared the first is the element index and the second is the element value.

	@for i, v := range a_slice {
	  Here's a slice element at index @(i): @(v)
	}
	
The `break` and `continue` keywords can be used to do the usual thing.

	@for v := range a_slice {
	  @if v > 100 {
	  	@(break)
	  }else{
	    The value is still less than 100...
	  }
	}


### `()`

Declaring dynamic content consisting of an expression wrapped in parentheses results in that expression being evaluated and its result interpolated in the output document.

	Is 1 + 2 less than 3 + 4? Let's see: @(1 + 2 < 3 + 4).


## Expressions

### Identifiers

Identifiers have the same rules as Go. A valid identifier is a letter followed by zero or more letters or digits. An underscore is considered to be a letter.

	a
	ThisIsALongIdentifier
	_a9

### `&&`, `||`, `!`

The standard logical *and*, *or*, and *not* operators are supported.

	1 < 2 && 2 < 3
	1 > 2 || 2 < 3
	!(1 > 2)

### `<`, `<=`, `==`, `>=`, `>`

The standard relational operators are supported. Values are comparable if their types are comparable in Go. Unlike Go, Ego will automatically convert numeric types so that they can be compared.

	1 < 2
	1 <= 3
	1 == 1
	5 >= 5
	5 > 4

### `*`, `/`, `%`, `+`, `-`

The standard arithmetic operators are supported. Only numeric types can have arithmetic performed on them. Unlike Go, Ego will automatically convert numeric types so that they are compatible and will automatically truncate floating point values to integers in order to apply the `%` operator.

The order of operations is: `*`, `/`, `%`, `+`, `-`

	1 + 2 - 3 * 4 / 5
	2 % 10

### `.`

The `.` operator dereferences a property. This operator can be use more liberally in Ego than Go. You can use this operator to:

* Obtain the value of an exported struct field
* Obtain a value from a map that has `string` keys if the key is also a valid identifier. That is: `a_map.string_key` is equivalent to `a_map["string_key"]`.
* Obtain the result of a method invocation if that method does not take any arguments. That is: `an_interface.Foo` is equivalent to `an_interface.Foo()`.

### `[]`

The `[]` subscript operator obtains the value at an index when the operand is an array or slice and obtains the value of a key when the operand is a map.

	a_slice[5]
	a_map["the_key"]

### `func()`

Functions are invoked as they are in Go.

	len(a_slice)

Also as it is in Go, when a function invocation follows a dereference it is treated as a method invocation.

	val.String()

