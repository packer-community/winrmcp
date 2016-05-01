package winrmcp

import (
	"testing"
)

func Test_win_path_handling(t *testing.T) {
	cases := []struct {
		path   string
		expect string
	}{
		{
			path:   "",
			expect: "",
		}, {
			path:   `C:/Temp`,
			expect: `C:\Temp`,
		}, {
			path:   `C:/Documents and Settings`,
			expect: `C:\Documents and Settings`,
		}, {
			path:   `'C:\Documents and Settings'`,
			expect: `C:\Documents and Settings`,
		}, {
			path:   `"C:\Documents and Settings"`,
			expect: `C:\Documents and Settings`,
		},
	}

	for _, c := range cases {
		path := winPath(c.path)
		if path != c.expect {
			t.Errorf("Expected path %q, got %q", c.expect, path)
		}
	}
}
