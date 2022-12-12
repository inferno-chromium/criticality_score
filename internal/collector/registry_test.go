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

package collector

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/criticality_score/internal/collector/github"
	"github.com/ossf/criticality_score/internal/collector/projectrepo"
	"github.com/ossf/criticality_score/internal/collector/signal"
	"github.com/ossf/criticality_score/internal/mocks"
)

func Test_registry_EmptySets(t *testing.T) {
	// not using mocks to test because it is getting complicated.
	test := struct {
		name string
		ss   []signal.Source
		want []signal.Set
	}{
		name: "multiple of the same source",
		ss: []signal.Source{
			&github.IssuesSource{},
			&github.IssuesSource{},
		},
		want: []signal.Set{
			&signal.IssuesSet{},
		},
	}
	r := newRegistry()
	r.ss = test.ss

	got := r.EmptySets()

	if !reflect.DeepEqual(got, test.want) {
		t.Errorf("registry.EmptySets() = %v, want %v", got, test.want)
	}
}

func Test_registry_Register(t *testing.T) {
	tests := []struct { //nolint:govet
		name             string
		shouldPanic      bool
		ssShouldBeFilled bool
		namespace        string
	}{
		{
			name:             "source already registered",
			namespace:        "valid",
			ssShouldBeFilled: true,
			shouldPanic:      true,
		},
		{
			name:        "sources signal set is not valid",
			namespace:   "!",
			shouldPanic: true,
		},
		{
			name:        "valid",
			namespace:   "valid",
			shouldPanic: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if p := recover(); (p != nil) != test.shouldPanic {
					t.Errorf("registry.Register() panic = %v, should panic = %v", p, test.shouldPanic)
				}
			}()

			ctrl := gomock.NewController(t)
			source := mocks.NewMockSource(ctrl)
			set := mocks.NewMockSet(ctrl)
			set.EXPECT().Namespace().DoAndReturn(func() signal.Namespace {
				return signal.Namespace(test.namespace)
			})
			source.EXPECT().EmptySet().DoAndReturn(func() signal.Set {
				return set
			})

			r := &registry{
				ss: []signal.Source{mocks.NewMockSource(ctrl)},
			}

			if test.ssShouldBeFilled {
				r.ss = []signal.Source{source}
			}

			r.Register(source)
		})
	}
}

func Test_registry_sourcesForRepository(t *testing.T) {
	tests := []struct { //nolint:govet
		name        string
		want        int // number of sources
		namespace   string
		isSupported bool
		shouldPanic bool
	}{
		{
			name:        "supported",
			want:        1,
			namespace:   "test",
			isSupported: true,
		},
		{
			name:      "not supported",
			want:      0,
			namespace: "test",
		},
		{
			name:        "exists",
			namespace:   "test",
			isSupported: true,
			shouldPanic: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if p := recover(); (p != nil) != test.shouldPanic {
					t.Errorf("registry.Register() panic = %v, should panic = %v", p, test.shouldPanic)
				}
			}()

			ctrl := gomock.NewController(t)
			source := mocks.NewMockSource(ctrl)
			repo := mocks.NewMockRepo(ctrl)
			set := mocks.NewMockSet(ctrl)
			set.EXPECT().Namespace().DoAndReturn(func() signal.Namespace {
				return signal.Namespace(test.namespace)
			}).AnyTimes()
			source.EXPECT().EmptySet().DoAndReturn(func() signal.Set {
				return set
			}).AnyTimes()
			source.EXPECT().IsSupported(repo).DoAndReturn(func(repo projectrepo.Repo) bool {
				return test.isSupported
			}).AnyTimes()

			r := &registry{
				ss: []signal.Source{source},
			}
			if test.shouldPanic {
				r.ss = append(r.ss, source)
			}

			if got := r.sourcesForRepository(repo); len(got) != test.want {
				t.Errorf("sourcesForRepository() = %v, want %v", got, test.want)
			}
		})
	}
}

func Test_registry_Collect(t *testing.T) {
	type args struct {
		ctx   context.Context
		jobID string
	}
	tests := []struct { //nolint:govet
		name        string
		args        args
		namespace   string
		isSupported bool
		wantErr     bool
		want        []signal.Set
	}{
		{
			name: "valid",
			args: args{
				ctx:   context.Background(),
				jobID: "jobID",
			},
			namespace:   "test",
			isSupported: true,
			want:        []signal.Set{&mocks.MockSet{}},
		},
		{
			name: "returns error",
			args: args{
				ctx:   context.Background(),
				jobID: "jobID",
			},
			namespace:   "test",
			isSupported: true,
			wantErr:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			source := mocks.NewMockSource(ctrl)
			repo := mocks.NewMockRepo(ctrl)
			set := mocks.NewMockSet(ctrl)

			// set the expected namespace
			set.EXPECT().Namespace().DoAndReturn(func() signal.Namespace {
				return signal.Namespace(test.namespace)
			}).AnyTimes()

			// set the expected EmptySet
			source.EXPECT().EmptySet().DoAndReturn(func() signal.Set {
				return set
			}).AnyTimes()

			// set the expected IsSupported
			source.EXPECT().IsSupported(repo).DoAndReturn(func(repo projectrepo.Repo) bool {
				return test.isSupported
			}).AnyTimes()

			// set the expected Get using test.wantErr
			source.EXPECT().Get(test.args.ctx, repo, test.args.jobID).DoAndReturn(
				func(ctx context.Context, repo projectrepo.Repo, jobID string) (signal.Set, error) {
					if test.wantErr {
						return nil, errors.New("error")
					}
					return set, nil
				}).AnyTimes()

			r := &registry{
				ss: []signal.Source{source},
			}

			got, err := r.Collect(test.args.ctx, repo, test.args.jobID)
			if (err != nil) != test.wantErr {
				t.Errorf("Collect() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if len(got) == len(test.want) {
				for i := range got {
					if reflect.TypeOf(got[i]) != reflect.TypeOf(test.want[i]) {
						t.Errorf("Collect() = %v, want %v", got, test.want)
					}
				}
			} else {
				t.Errorf("Len Collect() = %d, want %d", len(got), len(test.want))
			}
		})
	}
}
