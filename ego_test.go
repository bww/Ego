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
  "path"
  "bytes"
  "strings"
  "testing"
  "io/ioutil"
)

import (
  "github.com/bww/hcl"
  "github.com/stretchr/testify/assert"
)

func init() {
  DEBUG_TRACE_TOKEN = true
}

type testCase struct {
  Source  string                  `hcl:"source"`
  Expect  string                  `hcl:"expect"`
  Context map[string]interface{}  `hcl:"context"`
}

/**
 * Test everything
 */
func _TestAll(t *testing.T) {
  
  proj := os.Getenv("PROJECT")
  if !assert.True(t, len(proj) > 0, "No project root") { return }
  
  dir, err := os.Open(path.Join(proj, "tests"))
  if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { return }
  
  every, err := dir.Readdir(1000)
  if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { return }
  
  var filter map[string]struct{}
  which := os.Getenv("EGO_TESTS")
  if which != "" {
    f := strings.Fields(which)
    if len(f) > 0 {
      filter = make(map[string]struct{})
      for _, e := range f {
        filter[e + ".test"] = struct{}{}
      }
    }
  }
  
  for _, f := range every {
    var tests []testCase
    
    if filter != nil {
      if _, ok := filter[f.Name()]; !ok {
        t.Logf("===> %v (skip)", f.Name())
        continue
      }
    }
    
    t.Logf("===> %v", f.Name())
    
    file, err := os.Open(path.Join(proj, "tests", f.Name()))
    if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { return }
    
    data, err := ioutil.ReadAll(file)
    if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { return }
    
    err = hcl.Decode(&tests, string(data))
    if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { return }
    
    for _, e := range tests {
      
      output  := &bytes.Buffer{}
      scanner := newScanner(e.Source)
      parser  := newParser(scanner)
      runtime := &runtime{output}
      
      program, err := parser.parse()
      if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { continue }
      
      err = program.exec(runtime, newContext(e.Context))
      if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { continue }
      
      fmt.Printf("--> %v\n", e.Source)
      fmt.Printf("<-- %v\n", string(output.Bytes()))
      
      assert.Equal(t, e.Expect, string(output.Bytes()), "Expected output and actual output differ")
      
    }
    
  }
  
}

func TestEscaping(t *testing.T) {
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

func compileAndRun(t *testing.T, context interface{}, source, expect string) {
  fmt.Println(source)
  
  output  := &bytes.Buffer{}
  scanner := newScanner(source)
  parser  := newParser(scanner)
  runtime := &runtime{output}
  
  program, err := parser.parse()
  if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { return }
  
  err = program.exec(runtime, newContext(context))
  if !assert.Nil(t, err, fmt.Sprintf("%v", err)) { return }
  
  fmt.Printf("--> %v\n", source)
  fmt.Printf("<-- %v\n", string(output.Bytes()))
  
  assert.Equal(t, expect, string(output.Bytes()))
}

