// Copyright 2022 Criticality Score Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package inputiter

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"golang.org/x/exp/slices"
)

func TestSliceIter_Empty(t *testing.T) {
	i := &sliceIter[int]{
		values: []int{},
	}

	if got := i.Next(); got {
		t.Errorf("Next() = %v; want false", got)
	}
}

func TestSliceIter_SingleEntry(t *testing.T) {
	want := 42
	i := &sliceIter[int]{
		values: []int{want},
	}

	if got := i.Next(); !got {
		t.Errorf("Next() = %v; want true", got)
	}
	if got := i.Item(); got != want {
		t.Errorf("Item() = %v; want %v", got, want)
	}
	if got := i.Next(); got {
		t.Errorf("Next()#2 = %v; want false", got)
	}
}

func TestSliceIter_MultiEntry(t *testing.T) {
	want := []int{1, 2, 3, 42, 1337}
	i := &sliceIter[int]{
		values: want,
	}

	var got []int
	for i.Next() {
		got = append(got, i.Item())
	}

	if !slices.Equal(got, want) {
		t.Errorf("Iterator returned %v, want %v", got, want)
	}
}

func TestScannerIter_Empty(t *testing.T) {
	var b bytes.Buffer
	i := scannerIter{
		c:       io.NopCloser(&b),
		scanner: bufio.NewScanner(&b),
	}

	if got := i.Next(); got {
		t.Errorf("Next() = %v; want false", got)
	}
	if err := i.Err(); err != nil {
		t.Errorf("Err() = %v; want no error", err)
	}
}

func TestScannerIter_SingleLine(t *testing.T) {
	want := "test line"
	b := bytes.NewBuffer([]byte(want))
	i := scannerIter{
		c:       io.NopCloser(b),
		scanner: bufio.NewScanner(b),
	}

	if got := i.Next(); !got {
		t.Errorf("Next() = %v; want true", got)
	}
	if err := i.Err(); err != nil {
		t.Errorf("Err() = %v; want no error", err)
	}
	if got := i.Item(); got != want {
		t.Errorf("Item() = %v; want %v", got, want)
	}
	if got := i.Next(); got {
		t.Errorf("Next()#2 = %v; want false", got)
	}
	if err := i.Err(); err != nil {
		t.Errorf("Err()#2 = %v; want no error", err)
	}
}

func TestScannerIter_MultiLine(t *testing.T) {
	want := []string{"line one", "line two", "line three"}
	b := bytes.NewBuffer([]byte(strings.Join(want, "\n")))
	i := scannerIter{
		c:       io.NopCloser(b),
		scanner: bufio.NewScanner(b),
	}

	var got []string
	for i.Next() {
		item := i.Item()
		got = append(got, item)
	}
	if err := i.Err(); err != nil {
		t.Errorf("Err() = %v; want no error", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Iterator returned %v, want %v", got, want)
	}
}

func TestScannerIter_Error(t *testing.T) {
	want := errors.New("error")
	r := iotest.ErrReader(want)
	i := scannerIter{
		c:       io.NopCloser(r),
		scanner: bufio.NewScanner(r),
	}

	if got := i.Next(); got {
		t.Errorf("Next() = %v; want false", got)
	}
	if err := i.Err(); err == nil || !errors.Is(err, want) {
		t.Errorf("Err() = %v; want %v", err, want)
	}
}

type closerFn func() error

func (c closerFn) Close() error {
	return c()
}

func TestScannerIter_Close(t *testing.T) {
	got := 0
	i := scannerIter{
		c: closerFn(func() error {
			got++
			return nil
		}),
		scanner: bufio.NewScanner(&bytes.Buffer{}),
	}
	err := i.Close()

	if got != 1 {
		t.Errorf("Close() called %d times; want 1", got)
	}
	if err != nil {
		t.Errorf("Err() = %v; want no error", err)
	}
}
