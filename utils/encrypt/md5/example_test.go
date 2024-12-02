// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package md5_test

import (
	"fmt"
	"onpaper-api-go/utils/encrypt/md5"
	"testing"
)

func TestMd5(t *testing.T) {
	firstText := "72b729eed6dea2cc44c1dcebd0d98909a9f5b556"
	signBytes := md5.Sum([]byte(firstText))
	sign := fmt.Sprintf("%x", signBytes)
	fmt.Println(sign) // ab034e143a5e7d840f81bb08cde35188
	secondText := `{"onpaper":"qwl"}` + sign
	sign2Bytes := md5.Sum([]byte(secondText))
	lastSign := fmt.Sprintf("%x", sign2Bytes)
	fmt.Println(lastSign) // 572d2d62c3a8aeff9389f0aee71fb2d4
}
