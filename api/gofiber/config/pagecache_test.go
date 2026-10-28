package config

import "testing"

func Test_removePathPrefix(t *testing.T) {
	type args struct {
		path          string
		prefixPattern string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty prefix pattern", args{"/api/v1/test", ""}, "/api/v1/test"},
		{"pattern not start with ^", args{"/api/v13/test", "/?api/v[0-9]+/"}, "test"},
		{"empty path", args{"", pathPatternPrefix}, ""},

		{"", args{"/api/v13/test", pathPatternPrefix}, "test"},
		{"single digit version", args{"/api/v1/resource", pathPatternPrefix}, "resource"},
		{"multiple digit version", args{"/api/v123/test", pathPatternPrefix}, "test"},
		{"no leading slash", args{"api/v93/test", pathPatternPrefix}, "test"},
		{"with trailing slash", args{"/api/v123/resource/", pathPatternPrefix}, "resource/"},

		// TOOOO Edge cases
		{"path without prefix", args{"/users/profile", pathPatternPrefix}, ""},
		{"mismatching version prefix", args{"/v1/api/v11/test", pathPatternPrefix}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removePathPrefix(tt.args.path, tt.args.prefixPattern); got != tt.want {
				t.Errorf("removePathPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPathCached(t *testing.T) {
	type args struct {
		patternPrefix string
		path          string
		cacheList     []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"the same", args{"^/api/v[0-9]+/", "/api/v11/info/all", []string{"info/all"}}, true},
		{"not listed", args{pathPatternPrefix, "/api/v11/info/time", []string{"info/all"}}, false},
		{"missing params, no cache", args{pathPatternPrefix, "/api/v11/inventory/admin/all", []string{"^inventory/admin/all/[^/]+$"}}, false},

		{"should cache! even missing ^ at start", args{pathPatternPrefix, "/api/v11/session/admin/oculus/123456", []string{"session/admin" + pathPlatformPlatformIdPatternSuffix}}, true},
		{"should cache!", args{pathPatternPrefix, "/api/v11/session/admin/oculus/123456", []string{"^session/admin" + pathPlatformPlatformIdPatternSuffix}}, true},
		{"platform id with prefix", args{pathPatternPrefix, "api/v11/session/admin/apple/a|t:_f5156f72e10013a3e3d4eec959f9fed4", []string{"^session/admin" + pathPlatformPlatformIdPatternSuffix}}, true},
		{"extra in path, should not cache", args{pathPatternPrefix, "/api/v11/session/admin/apple/tester007/1", []string{"^session/admin" + pathPlatformPlatformIdPatternSuffix}}, false},

		{"regexp", args{pathPatternPrefix, "/api/v11/inventory/admin/all/tester001", []string{"^inventory/admin/all/[^/]+$"}}, true},

		{"caching off by default", args{pathPatternPrefix, "/api/v11/info/all", specificCachedPages}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPathCached(tt.args.patternPrefix, tt.args.path, tt.args.cacheList); got != tt.want {
				t.Errorf("isPathCached() = %v, want %v", got, tt.want)
			}
		})
	}
}
