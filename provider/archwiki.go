package provider

import (
	"strings"
)

type ArchWiki struct {
	*Wikipedia
}

func (s *ArchWiki) Flag() (flag, usage string) {
	flag, usage = "arch", "ArchWiki"
	return
}

func (s *ArchWiki) Name() string {
	return "archwiki"
}

func (s *ArchWiki) Domain() string {
	return "https://wiki.archlinux.org"
}

func (s *ArchWiki) Match(url string) bool {
	i := strings.Index(url, "wiki.archlinux.org")
	// https://
	return i > -1 && i < 9
}
