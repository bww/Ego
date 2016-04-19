// 
// Copyright (c) 2014 Brian William Wolter, All rights reserved.
// Ego - an embedded Go parser / compiler
// 
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
// 
//   * Redistributions of source code must retain the above copyright notice, this
//     list of conditions and the following disclaimer.
// 
//   * Redistributions in binary form must reproduce the above copyright notice,
//     this list of conditions and the following disclaimer in the documentation
//     and/or other materials provided with the distribution.
//     
//   * Neither the names of Brian William Wolter, Wolter Group New York, nor the
//     names of its contributors may be used to endorse or promote products derived
//     from this software without specific prior written permission.
//     
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
// INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
// BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
// LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE
// OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
// OF THE POSSIBILITY OF SUCH DAMAGE.
// 

package ego

import (
  "os"
  "fmt"
  "testing"
)

func init() {
  DEBUG_TRACE_TOKEN = true
}

func _TestThis(t *testing.T) {
  
  sources := []string{
    
`@if true ()[]., + ++ += - -- -= = == : := ! != * *= / /= < <= > >= ; | || |= & && &= range {
  This is a literal \\\} right here.
}else if range {} else {}`,

`@if true {
  Hello.
  @for a, b, c {
    ...
  }
}`,

`Hi.
\a\\
Why doesn't this break?!
\@
@if true {
  This, that, if else for
  Ok. Yeah.
} ... more.
  
@for 123 "String" {
  Do this, then... \{ hmm \}
}

@if false {
  Nope...
} else {
  This text instead!
}

Foo.
`,
  }
  
  for _, e := range sources {
    compileAndValidate(t, e, nil)
  }
  
}

func TestBasicEscaping(t *testing.T) {
  var source string
  
  source = `\foo`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 4}, tokenVerbatim, source},
    token{span{source, 4, 0}, tokenEOF, nil},
  })
  
  source = `\@`
  compileAndValidate(t, source, []token{
    token{span{source, 1, 1}, tokenVerbatim, "@"},
    token{span{source, 2, 0}, tokenEOF, nil},
  })
  
  source = `x\@`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenVerbatim, "x"},
    token{span{source, 2, 1}, tokenVerbatim, "@"},
    token{span{source, 3, 0}, tokenEOF, nil},
  })
  
  source = `\\\@`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenVerbatim, "\\"},
    token{span{source, 3, 1}, tokenVerbatim, "@"},
    token{span{source, 4, 0}, tokenEOF, nil},
  })
  
  source = `\@\\`
  compileAndValidate(t, source, []token{
    token{span{source, 1, 1}, tokenVerbatim, "@"},
    token{span{source, 2, 1}, tokenVerbatim, "\\"},
    token{span{source, 4, 0}, tokenEOF, nil},
  })
  
  source = `\\`
  compileAndValidate(t, source, []token{
    token{span{source, 1, 1}, tokenVerbatim, "\\"},
    token{span{source, 2, 0}, tokenEOF, nil},
  })
  
  source = `\`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenVerbatim, "\\"},
    token{span{source, 1, 0}, tokenEOF, nil},
  })
  
  source = `foo\`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 4}, tokenVerbatim, "foo\\"},
    token{span{source, 4, 0}, tokenEOF, nil},
  })
  
}

func TestBasicTypes(t *testing.T) {
  var source string
  
  source = `@123{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 3}, tokenNumber, float64(123)},
    token{span{source, 4, 1}, tokenBlock, nil},
    token{span{source, 5, 1}, tokenClose, nil},
    token{span{source, 6, 0}, tokenEOF, nil},
  })
  
  source = `@123.456{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 7}, tokenNumber, float64(123.456)},
    token{span{source, 8, 1}, tokenBlock, nil},
    token{span{source, 9, 1}, tokenClose, nil},
    token{span{source, 10, 0}, tokenEOF, nil},
  })
  
  source = `@0xff{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 4}, tokenNumber, float64(0xff)},
    token{span{source, 5, 1}, tokenBlock, nil},
    token{span{source, 6, 1}, tokenClose, nil},
    token{span{source, 7, 0}, tokenEOF, nil},
  })
  
  source = `@07{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 2}, tokenNumber, float64(7)},
    token{span{source, 3, 1}, tokenBlock, nil},
    token{span{source, 4, 1}, tokenClose, nil},
    token{span{source, 5, 0}, tokenEOF, nil},
  })
  
  source = `@"Hi."{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 5}, tokenString, "Hi."},
    token{span{source, 6, 1}, tokenBlock, nil},
    token{span{source, 7, 1}, tokenClose, nil},
    token{span{source, 8, 0}, tokenEOF, nil},
  })
  
  source = `@true{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 4}, tokenTrue, "true"},
    token{span{source, 5, 1}, tokenBlock, nil},
    token{span{source, 6, 1}, tokenClose, nil},
    token{span{source, 7, 0}, tokenEOF, nil},
  })
  
  source = `@false{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 5}, tokenFalse, "false"},
    token{span{source, 6, 1}, tokenBlock, nil},
    token{span{source, 7, 1}, tokenClose, nil},
    token{span{source, 8, 0}, tokenEOF, nil},
  })
  
  source = `@nil{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 3}, tokenNil, nil},
    token{span{source, 4, 1}, tokenBlock, nil},
    token{span{source, 5, 1}, tokenClose, nil},
    token{span{source, 6, 0}, tokenEOF, nil},
  })
  
  source = `@if{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 2}, tokenIf, "if"},
    token{span{source, 3, 1}, tokenBlock, nil},
    token{span{source, 4, 1}, tokenClose, nil},
    token{span{source, 5, 0}, tokenEOF, nil},
  })
  
  source = `@else{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 4}, tokenElse, "else"},
    token{span{source, 5, 1}, tokenBlock, nil},
    token{span{source, 6, 1}, tokenClose, nil},
    token{span{source, 7, 0}, tokenEOF, nil},
  })
  
  source = `@for{}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 3}, tokenFor, "for"},
    token{span{source, 4, 1}, tokenBlock, nil},
    token{span{source, 5, 1}, tokenClose, nil},
    token{span{source, 6, 0}, tokenEOF, nil},
  })
  
}

func TestBasicMeta(t *testing.T) {
  var source string
  
  source = `@if true {}`
  compileAndValidate(t, source, []token{
    token{span{source, 0, 1}, tokenMeta, "@"},
    token{span{source, 1, 2}, tokenIf, "if"},
    token{span{source, 4, 4}, tokenTrue, "true"},
    token{span{source, 9, 1}, tokenBlock, nil},
    token{span{source, 10, 1}, tokenClose, nil},
    token{span{source, 11, 0}, tokenEOF, nil},
  })
  
}

func compileAndValidate(t *testing.T, source string, expect []token) {
  fmt.Println(source)
  
  s := newScanner(source)
  
  for {
    
    k := s.scan()
    fmt.Println("T", k)
    
    if expect != nil {
      
      if len(expect) < 1 {
        t.Fatalf("Unexpected end of tokens")
        return
      }
      
      e := expect[0]
      
      if e.which != k.which {
        t.Errorf("Unexpected token type (%v != %v) in %v", k.which, e.which, source)
        return
      }
      
      if e.span.excerpt() != k.span.excerpt() {
        t.Errorf("Excerpts do not match (%q != %q) in %v", k.span.excerpt(), e.span.excerpt(), source)
        return
      }
      
      if e.value != k.value {
        t.Errorf("Values do not match (%v != %v) in %v", k.value, e.value, source)
        return
      }
      
      expect = expect[1:]
      
    }
    
    if k.which == tokenEOF {
      break
    }else if k.which == tokenError {
      break
    }
    
  }
  
  if expect != nil {
    if len(expect) > 0 {
      t.Fatalf("Unexpected end of input (%d tokens remain)", len(expect))
      return
    }
  }
  
  fmt.Println("---")
}

func TestParse(t *testing.T) {
  
  sources := []string{
    `Hello, there.@if true {
      This is verbatim.
      @if false {} else {}
    }`,
    `Hello, there.@if true {
      This is verbatim.
      @if false {} \else not really
    }`,
  }
  
  for _, source := range sources {
    s := newScanner(source)
    p := newParser(s)
    r := &runtime{os.Stdout}
    if g, err := p.parse(); err != nil {
      t.Fatal(err)
    }else if err := g.exec(r, nil); err != nil {
      t.Fatal(err)
    }
  }
  
}


