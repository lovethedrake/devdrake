package config

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type pipelineSelector struct {
	BranchSelector *refSelector `json:"branches"`
	TagSelector    *refSelector `json:"tags"`
}

type refSelector struct {
	WhitelistedRefs []string `json:"only"`
	BlacklistedRefs []string `json:"ignore"`
}

func (p *pipelineSelector) Matches(branch, tag string) (bool, error) {
	// If a tag is specified, we match purely on the basis of the tag. This means
	// "" is not a valid tag.
	if tag != "" {
		if p.TagSelector == nil {
			return false, nil
		}
		return p.TagSelector.Matches(tag)
	}
	// Fall back to matching on the branch. "" is considered a valid branch.
	if p.BranchSelector == nil {
		return false, nil
	}
	return p.BranchSelector.Matches(branch)
}

func (r *refSelector) Matches(ref string) (bool, error) {
	var matchesWhitelist bool
	if len(r.WhitelistedRefs) == 0 {
		matchesWhitelist = true
	} else {
		for _, whitelistedRef := range r.WhitelistedRefs {
			var err error
			matchesWhitelist, err = refMatch(ref, whitelistedRef)
			if err != nil {
				return false, err
			}
			if matchesWhitelist {
				break
			}
		}
	}
	var matchesBlacklist bool
	for _, blacklistedRef := range r.BlacklistedRefs {
		var err error
		matchesBlacklist, err = refMatch(ref, blacklistedRef)
		if err != nil {
			return false, err
		}
		if matchesBlacklist {
			break
		}
	}
	return matchesWhitelist && !matchesBlacklist, nil
}

func refMatch(ref, valueOrPattern string) (bool, error) {
	if strings.HasPrefix(valueOrPattern, "/") &&
		strings.HasSuffix(valueOrPattern, "/") {
		pattern := valueOrPattern[1 : len(valueOrPattern)-1]
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return false, errors.Wrapf(
				err,
				"error compiling regular expression %s",
				valueOrPattern,
			)
		}
		return regex.MatchString(ref), nil
	}
	return ref == valueOrPattern, nil
}
