package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompilePattern(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantErr      bool
		wantSegments []Segment
	}{
		{
			name: "static segments only",
			path: "/users/list",
			wantSegments: []Segment{
				{Kind: SegmentStatic, Value: "users"},
				{Kind: SegmentStatic, Value: "list"},
			},
		},
		{
			name: "single dynamic param",
			path: "/user/:id",
			wantSegments: []Segment{
				{Kind: SegmentStatic, Value: "user"},
				{Kind: SegmentParam, Name: "id"},
			},
		},
		{
			name: "multiple params",
			path: "/user/:userId/post/:postId",
			wantSegments: []Segment{
				{Kind: SegmentStatic, Value: "user"},
				{Kind: SegmentParam, Name: "userId"},
				{Kind: SegmentStatic, Value: "post"},
				{Kind: SegmentParam, Name: "postId"},
			},
		},
		{
			name: "optional param",
			path: "/profile/:id?",
			wantSegments: []Segment{
				{Kind: SegmentStatic, Value: "profile"},
				{Kind: SegmentOptional, Name: "id"},
			},
		},
		{
			name: "wildcard param",
			path: "/docs/:path*",
			wantSegments: []Segment{
				{Kind: SegmentStatic, Value: "docs"},
				{Kind: SegmentWildcard, Name: "path"},
			},
		},
		{
			name:         "root path",
			path:         "/",
			wantSegments: []Segment{},
		},
		{
			name: "complex pattern",
			path: "/api/:version/users/:id?/posts/:slug*",
			wantSegments: []Segment{
				{Kind: SegmentStatic, Value: "api"},
				{Kind: SegmentParam, Name: "version"},
				{Kind: SegmentStatic, Value: "users"},
				{Kind: SegmentOptional, Name: "id"},
				{Kind: SegmentStatic, Value: "posts"},
				{Kind: SegmentWildcard, Name: "slug"},
			},
		},
		{
			name:    "invalid - no leading slash",
			path:    "users/list",
			wantErr: true,
		},
		{
			name:    "invalid - empty param name",
			path:    "/user/:",
			wantErr: true,
		},
		{
			name:    "invalid - wildcard not at end",
			path:    "/docs/:path*/other",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, err := CompilePattern(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, pattern)
			assert.Equal(t, tt.wantSegments, pattern.segments)
			assert.NotNil(t, pattern.regex, "regex should be generated")
		})
	}
}

func TestRoutePattern_Match(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		testPath   string
		wantMatch  bool
		wantParams map[string]string
	}{
		{
			name:       "static match",
			pattern:    "/users/list",
			testPath:   "/users/list",
			wantMatch:  true,
			wantParams: map[string]string{},
		},
		{
			name:      "static no match",
			pattern:   "/users/list",
			testPath:  "/users/create",
			wantMatch: false,
		},
		{
			name:      "param extraction",
			pattern:   "/user/:id",
			testPath:  "/user/123",
			wantMatch: true,
			wantParams: map[string]string{
				"id": "123",
			},
		},
		{
			name:      "multiple params",
			pattern:   "/user/:userId/post/:postId",
			testPath:  "/user/42/post/99",
			wantMatch: true,
			wantParams: map[string]string{
				"userId": "42",
				"postId": "99",
			},
		},
		{
			name:      "optional param present",
			pattern:   "/profile/:id?",
			testPath:  "/profile/123",
			wantMatch: true,
			wantParams: map[string]string{
				"id": "123",
			},
		},
		{
			name:       "optional param absent",
			pattern:    "/profile/:id?",
			testPath:   "/profile",
			wantMatch:  true,
			wantParams: map[string]string{},
		},
		{
			name:      "wildcard single segment",
			pattern:   "/docs/:path*",
			testPath:  "/docs/intro",
			wantMatch: true,
			wantParams: map[string]string{
				"path": "intro",
			},
		},
		{
			name:      "wildcard multiple segments",
			pattern:   "/docs/:path*",
			testPath:  "/docs/guide/getting-started",
			wantMatch: true,
			wantParams: map[string]string{
				"path": "guide/getting-started",
			},
		},
		{
			name:      "wildcard empty",
			pattern:   "/docs/:path*",
			testPath:  "/docs",
			wantMatch: true,
			wantParams: map[string]string{
				"path": "",
			},
		},
		{
			name:       "root path",
			pattern:    "/",
			testPath:   "/",
			wantMatch:  true,
			wantParams: map[string]string{},
		},
		{
			name:       "trailing slash normalized",
			pattern:    "/users",
			testPath:   "/users/",
			wantMatch:  true,
			wantParams: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, err := CompilePattern(tt.pattern)
			require.NoError(t, err)

			params, matched := pattern.Match(tt.testPath)

			assert.Equal(t, tt.wantMatch, matched, "match result")
			if tt.wantMatch {
				assert.Equal(t, tt.wantParams, params, "extracted params")
			}
		})
	}
}

func TestSegmentKind_String(t *testing.T) {
	tests := []struct {
		kind SegmentKind
		want string
	}{
		{SegmentStatic, "static"},
		{SegmentParam, "param"},
		{SegmentOptional, "optional"},
		{SegmentWildcard, "wildcard"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.kind.String())
		})
	}
}

func TestCompilePattern_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "path cannot be empty",
		},
		{
			name:    "no leading slash",
			path:    "users",
			wantErr: true,
			errMsg:  "path must start with /",
		},
		{
			name:    "duplicate param names",
			path:    "/user/:id/post/:id",
			wantErr: true,
			errMsg:  "duplicate parameter name",
		},
		{
			name:    "wildcard not at end",
			path:    "/docs/:path*/other",
			wantErr: true,
			errMsg:  "wildcard must be the last segment",
		},
		{
			name:    "empty param name",
			path:    "/user/:/post",
			wantErr: true,
			errMsg:  "parameter name cannot be empty",
		},
		{
			name:    "invalid characters in param",
			path:    "/user/:id-name",
			wantErr: true,
			errMsg:  "invalid parameter name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompilePattern(tt.path)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
