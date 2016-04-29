// 
// Copyright (c) 2014-2016 Brian W. Wolter, All rights reserved.
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
//   * Neither the names of Brian W. Wolter nor the names of the contributors may
//     be used to endorse or promote products derived from this software without
//     specific prior written permission.
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
  "bytes"
  "testing"
)

import (
  "github.com/stretchr/testify/assert"
)

func init() {
  // DEBUG_TRACE_TOKEN = true
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

func compileAndRun(t *testing.T, compile, exec bool, context interface{}, source, expect string) {
  fmt.Println(source)
  
  output  := &bytes.Buffer{}
  scanner := newScanner(source)
  parser  := newParser(scanner)
  runtime := &Runtime{output}
  
  program, err := parser.parse()
  if compile {
    if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { return }
  }else{
    assert.NotNil(t, err, "Expected compile error")
    return
  }
  
  err = program.exec(runtime, newContext(context))
  if exec {
    if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { return }
  }else{
    assert.NotNil(t, err, "Expected runtime error")
    return
  }
  
  fmt.Printf("--> %v\n", source)
  fmt.Printf("<-- %v\n", string(output.Bytes()))
  
  assert.Equal(t, expect, string(output.Bytes()))
}

