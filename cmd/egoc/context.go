// 
// Copyright (c) 2014-2016 Brian William Wolter, All rights reserved.
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

package main

import (
  "os"
  "fmt"
  "path"
  "strings"
  "io/ioutil"
  "encoding/json"
)

/**
 * Read a file
 */
func readFile(p string) ([]byte, error) {
  
  file, err := os.Open(p)
  if err != nil {
    return nil, err
  }
  
  data, err := ioutil.ReadAll(file)
  if err != nil {
    return nil, err
  }
  
  return data, nil
}

/**
 * Read a context
 */
func readContext(p string) (interface{}, error) {
  
  data, err := readFile(p)
  if err != nil {
    return nil, err
  }
  
  switch ext := strings.ToLower(path.Ext(p)); ext {
    case ".json":
      return readJSONContext(p, data)
    default:
      return nil, fmt.Errorf("Unsupported context format: %v", ext)
  }
  
}

/**
 * Read a JSON context
 */
func readJSONContext(p string, data []byte) (interface{}, error) {
  context := make(map[string]interface{})
  
  err := json.Unmarshal(data, &context)
  if err != nil {
    return nil, err
  }
  
  return context, nil
}

