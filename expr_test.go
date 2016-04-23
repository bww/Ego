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
  "testing"
)

func TestExpr(t *testing.T) {
  
  compileAndRun(t, true, true, map[string]interface{}{"a": []string{"first", "second"}},
    `@(a[0]), @(a[1])`,
    `first, second`,
  )
  
  compileAndRun(t, true, true, map[string]interface{}{"a": map[string]string{"k1": "first", "k2": "second"}},
    `@(a["k1"]), @(a["k2"])`,
    `first, second`,
  )
  
  compileAndRun(t, true, true, map[string]interface{}{"a": map[string]string{"k1": "first", "k2": "second"}},
    `@(a.k1), @(a.k2)`,
    `first, second`,
  )
  
  compileAndRun(t, true, true, map[string]interface{}{
      "a": map[string]string{"k1": "first", "k2": "second"},
      "key1": "k1",
      "key2": "k2",
    },
    `@(a[key1]), @(a[key2])`,
    `first, second`,
  )
  
  compileAndRun(t, true, true, map[string]interface{}{"a": map[string][]string{"k1": []string{"first", "second"}, "k2": []string{"third", "fourth"}}},
    `@(a["k1"][0]), @(a["k1"][1])`,
    `first, second`,
  )
  
  compileAndRun(t, true, true, map[string]interface{}{"a": map[string][]string{"k1": []string{"first", "second"}, "k2": []string{"third", "fourth"}}},
    `@(a.k1[0]), @(a.k1[1])`,
    `first, second`,
  )
  
  compileAndRun(t, true, true, map[string]interface{}{"a": []map[string]string{map[string]string{"k1": "first", "k2": "second"}, map[string]string{"k3": "third", "k4": "fourth"}}},
    `@(a[0].k1), @(a[0].k2)`,
    `first, second`,
  )
  
  compileAndRun(t, true, true, map[string]interface{}{"a": []map[string]string{map[string]string{"k1": "first", "k2": "second"}, map[string]string{"k3": "third", "k4": "fourth"}}},
    `@(a[0]["k1"]), @(a[0]["k2"])`,
    `first, second`,
  )
  
  compileAndRun(t, true, false, map[string]interface{}{"a": []map[string]string{map[string]string{"k1": "first", "k2": "second"}, map[string]string{"k3": "third", "k4": "fourth"}}},
    `@(a[0][nilvar]), @(a[0][nilvar])`,
    `first, second`,
  )
  
  compileAndRun(t, true, false, map[string]interface{}{"a": map[string]string{"k1": "first", "k2": "second"}},
    `@(a[1]), @(a[2])`,
    `first, second`,
  )
  
}
