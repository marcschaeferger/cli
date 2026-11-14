package scp

import (
	"os"
	"strings"

	sshcmd "github.com/fosrl/cli/cmd/ssh"
)

// rawSCPArgs returns the arguments passed to the scp subcommand directly from
// os.Args, bypassing Cobra/pflag flag parsing. This is necessary because pflag
// does not know whether unknown short flags (e.g. -r, -p, -v) take an argument,
// so it may incorrectly consume the following operand as the flag's value.
func rawSCPArgs() []string {
	for i, a := range os.Args {
		if a == "scp" {
			return os.Args[i+1:]
		}
	}
	return nil
}

// countSCPOperands returns the number of non-flag positional operands in args.
func countSCPOperands(args []string) int {
	count := 0
	i := 0
	for i < len(args) {
		a := args[i]
		if a == "--" {
			count += len(args[i+1:])
			break
		}
		if strings.HasPrefix(a, "-") && a != "-" {
			i += 1 + scpFlagExtras(a, args, i)
			continue
		}
		count++
		i++
	}
	return count
}

// scpValueOptions are the scp(1) short options that take a value. scp's option
// set differs from ssh(1): -B, -O, -p and -R are boolean for scp.
const scpValueOptions = "cDFiJloPSX"

// scpFlagExtras returns how many additional args the given scp short flag
// token consumes, following getopt rules: in a bundle such as -rP, a value
// option takes the rest of the token or, if it is last, the next argument.
func scpFlagExtras(a string, args []string, i int) int {
	if len(a) < 2 || a[0] != '-' || a[1] == '-' {
		return 0
	}
	for j := 1; j < len(a); j++ {
		if strings.IndexByte(scpValueOptions, a[j]) >= 0 {
			if j == len(a)-1 && i+1 < len(args) {
				return 1
			}
			return 0
		}
	}
	return 0
}

// parseSCPRemoteHost scans scp operands and returns the username and resource ID
// from the first remote operand (host:path or user@host:path). Local paths are skipped.
func parseSCPRemoteHost(args []string) (username, resourceID string, found bool) {
	for _, arg := range args {
		hostSpec, _, ok := splitSCPOperand(arg)
		if !ok {
			continue
		}
		if u, h, hasAt := strings.Cut(hostSpec, "@"); hasAt {
			if u != "" && h != "" {
				return u, h, true
			}
		} else {
			return "", hostSpec, true
		}
	}
	return "", "", false
}

// parseSCPPassThrough splits raw scp arguments into options and operands using
// scp's own option table, so boolean scp flags do not swallow an operand.
func parseSCPPassThrough(args []string) sshcmd.SSHPassthrough {
	var pt sshcmd.SSHPassthrough
	i := 0
	for i < len(args) {
		a := args[i]
		if a == "--" {
			pt.Options = append(pt.Options, a)
			i++
			break
		}
		if !strings.HasPrefix(a, "-") || a == "-" {
			break
		}
		end := i + 1 + scpFlagExtras(a, args, i)
		pt.Options = append(pt.Options, args[i:end]...)
		i = end
	}
	if i < len(args) {
		pt.RemoteCommand = append([]string{}, args[i:]...)
	}
	return pt
}
