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

/**
 * Test everything
 */
func TestFor(t *testing.T) {
  
  compileAndRun(t, true, true, map[string]interface{}{"a": []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
    `@for _, e := range a { A }`,
    ` A  A  A  A  A  A  A  A  A  A `,
  )
  
  compileAndRun(t, true, true, map[string]interface{}{"a": []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
    `@for i, e := range a { @(i) = @(e + 10) }`,
    ` 0 = 10  1 = 11  2 = 12  3 = 13  4 = 14  5 = 15  6 = 16  7 = 17  8 = 18  9 = 19 `,
  )
  
  compileAndRun(t, true, true, map[string]interface{}{"a": []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
    `@for _, e := range a { A @(break) B }`,
    ` A `,
  )
  
  compileAndRun(t, true, true, map[string]interface{}{"a": []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
    `@for _, e := range a { A @(continue) B }`,
    ` A  A  A  A  A  A  A  A  A  A `,
  )
  
  // this one is hard to test since the keys are not necessarily iterated in any particular order
  // compileAndRun(t, true, true, map[string]interface{}{"a": map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}},
  //   `@for k, v := range a { @(k) = @(v) }`,
  //   ` a = 1  b = 2  c = 3  d = 4 `,
  // )
  
}