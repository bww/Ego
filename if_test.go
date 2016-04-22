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
func TestIf(t *testing.T) {
  
  compileAndRun(t, map[string]interface{}{"a": true},
    `@if (a) { This is displayed if 'a' is true, which it is. }
     else { Otherwise this. }`,
    ` This is displayed if 'a' is true, which it is. `,
  )
  
  compileAndRun(t, map[string]interface{}{"a": true},
    `@if !a { This is displayed if '!a' is true, which it is not. }
     else { Otherwise this is displayed. }`,
    ` Otherwise this is displayed. `,
  )
  
  compileAndRun(t, map[string]interface{}{"a": false, "b": true},
    `@if a { This is displayed if 'a' is true, which it is not. }
     else if b { This is displayed if 'b' is true, which it is. }`,
    ` This is displayed if 'b' is true, which it is. `,
  )
  
  compileAndRun(t, map[string]interface{}{"a": false, "b": false},
    `@if a { This is displayed if 'a' is true, which it is not. }
     else if b { This is displayed if 'b' is true, which it also is not. }
     else { Otherwise this is displayed. }`,
    ` Otherwise this is displayed. `,
  )
  
  compileAndRun(t, map[string]interface{}{"a": true},
    `@if (a) { A }else{ B }`, // braces tight around 'else'
    ` A `,
  )
  
  compileAndRun(t, map[string]interface{}{"a": true},
    `@if (a) { A \} } }`, // unescaped '}' outside block, escaped '}' inside block
    ` A }  }`,
  )
  
}

