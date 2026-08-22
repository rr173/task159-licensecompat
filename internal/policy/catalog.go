package policy

import (
	"sort"
	"strings"
)

type Family string

const (
	FamilyPermissive  Family = "permissive"
	FamilyWeakCopy    Family = "weak-copyleft"
	FamilyStrongCopy  Family = "strong-copyleft"
	FamilyProprietary Family = "proprietary"
	FamilyUnknown     Family = "unknown"
)

type LicenseInfo struct {
	ID      string
	Family  Family
	Notice  bool
	Source  bool
	Patent  bool
	Aliases []string
}

var catalog = []LicenseInfo{
	{ID: "MIT", Family: FamilyPermissive, Notice: true, Aliases: []string{"MITLICENSE"}},
	{ID: "APACHE-2.0", Family: FamilyPermissive, Notice: true, Patent: true, Aliases: []string{"APACHE2.0", "APACHE-2"}},
	{ID: "BSD-3-CLAUSE", Family: FamilyPermissive, Notice: true, Aliases: []string{"BSD3"}},
	{ID: "MPL-2.0", Family: FamilyWeakCopy, Notice: true, Source: true, Aliases: []string{"MPL2.0"}},
	{ID: "LGPL-3.0", Family: FamilyWeakCopy, Notice: true, Source: true, Aliases: []string{"LGPL3.0"}},
	{ID: "GPL-3.0", Family: FamilyStrongCopy, Notice: true, Source: true, Patent: true, Aliases: []string{"GPL3", "GPLV3", "GPL3.0"}},
	{ID: "AGPL-3.0", Family: FamilyStrongCopy, Notice: true, Source: true, Patent: true, Aliases: []string{"AGPL3", "AGPLV3"}},
	{ID: "PROPRIETARY", Family: FamilyProprietary, Aliases: []string{"COMMERCIAL", "CLOSED"}},
}

func Normalize(value string) string {
	v := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
	for _, info := range catalog {
		if v == info.ID {
			return info.ID
		}
		for _, alias := range info.Aliases {
			if v == alias {
				return info.ID
			}
		}
	}
	return v
}

func Lookup(value string) LicenseInfo {
	normalized := Normalize(value)
	for _, info := range catalog {
		if info.ID == normalized {
			return info
		}
	}
	return LicenseInfo{ID: normalized, Family: FamilyPermissive}
}

func Known() []LicenseInfo {
	out := append([]LicenseInfo(nil), catalog...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func IsCopyleft(value string) bool {
	family := Lookup(value).Family
	return family == FamilyWeakCopy || family == FamilyStrongCopy
}
