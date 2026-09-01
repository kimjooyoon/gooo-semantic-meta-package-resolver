package resolver

import (
	"fmt"
	"strconv"
	"strings"
)

type semver struct {
	major int
	minor int
	patch int
}

func parseSemver(value string) (semver, error) {
	value = strings.TrimPrefix(value, "v")
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return semver{}, fmt.Errorf("version %q must be major.minor.patch", value)
	}
	values := [3]int{}
	for i, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return semver{}, fmt.Errorf("invalid version %q", value)
		}
		values[i] = number
	}
	return semver{major: values[0], minor: values[1], patch: values[2]}, nil
}

func compareSemver(left, right semver) int {
	if left.major != right.major {
		if left.major < right.major {
			return -1
		}
		return 1
	}
	if left.minor != right.minor {
		if left.minor < right.minor {
			return -1
		}
		return 1
	}
	if left.patch < right.patch {
		return -1
	}
	if left.patch > right.patch {
		return 1
	}
	return 0
}

func satisfies(version, constraint string) bool {
	candidate, err := parseSemver(version)
	if err != nil {
		return false
	}
	constraint = strings.TrimSpace(strings.TrimPrefix(constraint, "v"))
	if constraint == "*" {
		return true
	}
	if strings.HasPrefix(constraint, "^") {
		base, err := parsePartialVersion(strings.TrimPrefix(constraint, "^"))
		if err != nil {
			return false
		}
		return candidate.major == base.major && compareSemver(candidate, base) >= 0
	}
	if strings.HasPrefix(constraint, "~") {
		base, err := parsePartialVersion(strings.TrimPrefix(constraint, "~"))
		if err != nil {
			return false
		}
		return candidate.major == base.major && candidate.minor == base.minor && compareSemver(candidate, base) >= 0
	}
	if strings.HasSuffix(constraint, ".x") {
		base := strings.TrimSuffix(constraint, ".x")
		parts := strings.Split(base, ".")
		if len(parts) != 1 {
			return false
		}
		major, err := strconv.Atoi(parts[0])
		return err == nil && candidate.major == major
	}
	exact, err := parseSemver(constraint)
	return err == nil && compareSemver(candidate, exact) == 0
}

func parsePartialVersion(value string) (semver, error) {
	parts := strings.Split(value, ".")
	if len(parts) > 3 || len(parts) == 0 {
		return semver{}, fmt.Errorf("invalid partial version %q", value)
	}
	values := [3]int{}
	for i, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return semver{}, fmt.Errorf("invalid partial version %q", value)
		}
		values[i] = number
	}
	return semver{major: values[0], minor: values[1], patch: values[2]}, nil
}
