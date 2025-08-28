package addr_manager

import "testing"

func TestIsDNS(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"google.com", true},
		{"www.google.com", true},
		{"google.co.uk", true},
		{"google", false},
		{"google.", false},
		{"google.c", false},
		{"google.123", false},
		{"google.com.", false},
		{"subdomain.example.com", true},
		{"www.example.com", true},
		{"example.com", true},
		{"example", false},
		{"example..com", false},
		{"example-.com", false},
		{"example.com-", false},
		{"-example.com", false},
		{"example.com.", false},
		{"example.com..", false},
		{"example.com...", false},
		{"example.com..com", false},
		{"example.com:80", false},
		{"example.com/path", false},
		{"example.com?query=string", false},
		{"example.com#fragment", false},
		{"172.28.0.144:19980", false},
		{"my.really.long.subdomain.example.com", true},
	}

	for _, tc := range testCases {
		actual := IsDNS(tc.input)
		if actual != tc.expected {
			t.Errorf("IsDNS(%q) = %v, expected %v", tc.input, actual, tc.expected)
		}
	}
}
