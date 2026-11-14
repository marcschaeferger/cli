package scp

import (
	"reflect"
	"testing"
)

func TestParseSCPPassThroughKeepsOperandsAfterBooleanFlags(t *testing.T) {
	// -p, -B, -O and -R take a value in ssh(1) but are boolean in scp(1);
	// they must not swallow the following operand.
	for _, flag := range []string{"-p", "-B", "-O", "-R", "-r", "-3", "-T", "-q", "-v"} {
		t.Run(flag, func(t *testing.T) {
			pt := parseSCPPassThrough([]string{flag, "./local", "host:/tmp/"})
			if want := []string{flag}; !reflect.DeepEqual(pt.Options, want) {
				t.Errorf("Options = %q, want %q", pt.Options, want)
			}
			if want := []string{"./local", "host:/tmp/"}; !reflect.DeepEqual(pt.RemoteCommand, want) {
				t.Errorf("operands = %q, want %q", pt.RemoteCommand, want)
			}
		})
	}
}

func TestParseSCPPassThroughValueFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		options  []string
		operands []string
	}{
		{"cipher", []string{"-c", "aes128-ctr", "a", "h:/b"}, []string{"-c", "aes128-ctr"}, []string{"a", "h:/b"}},
		{"sftp server path", []string{"-D", "/usr/lib/sftp-server", "a", "h:/b"}, []string{"-D", "/usr/lib/sftp-server"}, []string{"a", "h:/b"}},
		{"config", []string{"-F", "cfg", "a", "h:/b"}, []string{"-F", "cfg"}, []string{"a", "h:/b"}},
		{"identity", []string{"-i", "key", "a", "h:/b"}, []string{"-i", "key"}, []string{"a", "h:/b"}},
		{"jump", []string{"-J", "jump", "a", "h:/b"}, []string{"-J", "jump"}, []string{"a", "h:/b"}},
		{"limit", []string{"-l", "100", "a", "h:/b"}, []string{"-l", "100"}, []string{"a", "h:/b"}},
		{"ssh option", []string{"-o", "Foo=bar", "a", "h:/b"}, []string{"-o", "Foo=bar"}, []string{"a", "h:/b"}},
		{"port", []string{"-P", "2222", "a", "h:/b"}, []string{"-P", "2222"}, []string{"a", "h:/b"}},
		{"program", []string{"-S", "ssh", "a", "h:/b"}, []string{"-S", "ssh"}, []string{"a", "h:/b"}},
		{"sftp option", []string{"-X", "nrequests=4", "a", "h:/b"}, []string{"-X", "nrequests=4"}, []string{"a", "h:/b"}},
		{"attached value", []string{"-P2222", "a", "h:/b"}, []string{"-P2222"}, []string{"a", "h:/b"}},
		{"bundled booleans", []string{"-rp", "a", "h:/b"}, []string{"-rp"}, []string{"a", "h:/b"}},
		{"bundle ending in value flag", []string{"-rP", "2222", "a", "h:/b"}, []string{"-rP", "2222"}, []string{"a", "h:/b"}},
		{"mixed", []string{"-p", "-P", "22", "-r", "h:/a", "./b"}, []string{"-p", "-P", "22", "-r"}, []string{"h:/a", "./b"}},
		{"double dash", []string{"-r", "--", "-odd", "h:/b"}, []string{"-r", "--"}, []string{"-odd", "h:/b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := parseSCPPassThrough(tt.args)
			if !reflect.DeepEqual(pt.Options, tt.options) {
				t.Errorf("Options = %q, want %q", pt.Options, tt.options)
			}
			if !reflect.DeepEqual(pt.RemoteCommand, tt.operands) {
				t.Errorf("operands = %q, want %q", pt.RemoteCommand, tt.operands)
			}
		})
	}
}

func TestCountSCPOperandsUsesSCPOptionTable(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"boolean -B", []string{"-B", "./a", "h:/b"}, 2},
		{"boolean -O", []string{"-O", "./a", "h:/b"}, 2},
		{"boolean -R", []string{"-R", "./a", "h:/b"}, 2},
		{"value -P", []string{"-P", "22", "./a", "h:/b"}, 2},
		{"value -D", []string{"-D", "/usr/lib/sftp-server", "./a", "h:/b"}, 2},
		{"value -c", []string{"-c", "aes128-ctr", "./a", "h:/b"}, 2},
		{"value -X", []string{"-X", "nsessions=1", "./a", "h:/b"}, 2},
		{"bundle ending in value", []string{"-rP", "22", "./a", "h:/b"}, 2},
		{"bundle with attached value", []string{"-rP22", "./a", "h:/b"}, 2},
		{"double dash", []string{"-r", "--", "-a", "h:/b"}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := countSCPOperands(tc.args); got != tc.want {
				t.Fatalf("countSCPOperands(%q) = %d, want %d", tc.args, got, tc.want)
			}
		})
	}
}

// TestBuildExecSCPArgsKeepsOperandsAfterBooleanFlags checks the final scp argv:
// the remote operand after a boolean scp flag must still be rewritten to the
// resolved host.
func TestBuildExecSCPArgsKeepsOperandsAfterBooleanFlags(t *testing.T) {
	opts := RunOpts{
		User:        "alice",
		Hostname:    "10.0.0.5",
		ResourceID:  "myres",
		Passthrough: parseSCPPassThrough([]string{"-B", "./local", "myres:/tmp/"}),
	}
	args := buildExecSCPArgs("scp", opts, "", "")
	if got := args[len(args)-2:]; got[0] != "./local" || got[1] != "alice@10.0.0.5:/tmp/" {
		t.Fatalf("operands = %q, want [./local alice@10.0.0.5:/tmp/]", got)
	}
}
