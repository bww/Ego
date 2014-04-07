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
  "fmt"
  "testing"
)

func TestThis(t *testing.T) {
  
  sources := []string{
    
    `\foo`,
    `\@`,
    `x\@`,
    `\\\@`,
    `\@\\`,
    `\\`,
    `\`,

`@if true {
  This is a literal \\\} right here.
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
  
  /*
    sources := []string{
    `\foo`,
    `\@`,
    `x\@`,
    `\\\@`,
    `\@\\`,
    `\\`,
    `\`,
  }
  
  for _, e := range sources {
    compileAndValidate(t, e, nil)
  }
  */
  
}

func compileAndValidate(test *testing.T, source string, expect []token) {
  fmt.Println(source)
  
  c := make(chan token)
  s := newScanner(source, c)
  go s.scan()
  
  for {
    
    t, ok := <- c
    if !ok {
      break
    }else{
      fmt.Println("T", t)
    }
    
    if expect != nil {
      
      if len(expect) < 1 {
        test.Errorf("Unexpected end of tokens")
        return
      }
      
      e := expect[0]
      
      if e.which != t.which {
        test.Errorf("Unexpected token type (%v != %v)", t.which, e.which)
        return
      }
      
      if e.span.excerpt() != t.span.excerpt() {
        test.Errorf("Excerpts do not match (%q != %q)", t.span.excerpt(), e.span.excerpt())
        return
      }
      
      if e.value != t.value {
        test.Errorf("Values do not match (%v != %v)", t.value, e.value)
        return
      }
      
      expect = expect[1:]
      
    }
    
  }
  
  if expect != nil {
    if len(expect) > 0 {
      test.Errorf("Unexpected end of input (%d tokens remain)", len(expect))
      return
    }
  }
  
  fmt.Println("---")
}

