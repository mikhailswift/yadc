package commands

import (
	"testing"
)

func TestParsing(t *testing.T) {
	testCases := []struct {
		Test         string
		ExpectedCmd  string
		ExpectedArgs []string
		ShouldPass   bool
	}{
		{"GET Test", "GET", []string{"Test"}, true},
		{"Get \"Test 12\"", "GET", []string{"Test 12"}, true},
		{"", "", []string{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.Test, func(t *testing.T) {
			cmd, args, err := splitCmdString(tc.Test)
			t.Logf("Test : %+v", tc)
			t.Logf("Cmd: %+v ; Args : %+v", cmd, args)
			if tc.ShouldPass && err != nil {
				t.Fatalf("Failed to parse %v as a command: %+v", tc.Test, err)
			}

			if !tc.ShouldPass && err == nil {
				t.Fatalf("Expected to get an error for %v but got none", tc.Test)
			}

			if cmd != tc.ExpectedCmd {
				t.Fatalf("Got unexpected command for %v, expected %v got %v", tc.Test, tc.ExpectedCmd, cmd)
			}

			actualLen := len(args)
			expectedLen := len(tc.ExpectedArgs)
			if actualLen != expectedLen {
				t.Fatalf("Got unexpected args for %v, expected %v args got %v", tc.Test, expectedLen, actualLen)
			}

			for i, expected := range tc.ExpectedArgs {
				if expected != args[i] {
					t.Fatalf("Got unexpected arg for %v, expected %v got %v", tc.Test, expected, args[i])
				}
			}
		})
	}
}
