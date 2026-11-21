package app

import (
	"context"
	"strings"

	"github.com/fsuhrau/automationhub/storage/models"
)

// Bundles lists the bundles of the app, newest first, with how many test
// runs used each.
func (s *Service) Bundles(ctx context.Context, sc Scope) ([]BundleEntry, error) {
	bundles, err := s.repo.Bundles(ctx, sc)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, len(bundles))
	for i := range bundles {
		ids[i] = bundles[i].ID
	}
	counts, err := s.repo.RunCounts(ctx, ids)
	if err != nil {
		return nil, err
	}
	list := make([]BundleEntry, len(bundles))
	for i := range bundles {
		list[i] = BundleEntry{AppBinary: bundles[i], RunCount: counts[bundles[i].ID]}
		list[i].Info = summaryInfo(bundles[i].Info)
	}
	return list, nil
}

// Bundle is one bundle with everything known about it.
func (s *Service) Bundle(ctx context.Context, ref BundleRef) (*BundleDetail, error) {
	b, err := s.repo.Bundle(ctx, ref)
	if err != nil {
		return nil, err
	}
	return s.detail(ctx, b)
}

// SetTags changes the tags of a bundle; nothing else of a bundle is
// editable.
func (s *Service) SetTags(ctx context.Context, ref BundleRef, tags string) (*models.AppBinary, error) {
	b, err := s.repo.Bundle(ctx, ref)
	if err != nil {
		return nil, err
	}
	b.Tags = NormalizeTags(tags)
	if err := s.repo.SaveTags(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

// NormalizeTags trims the comma separated tags and drops empty and repeated
// ones.
func NormalizeTags(tags string) string {
	var out []string
	seen := map[string]bool{}
	for _, t := range strings.Split(tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return strings.Join(out, ",")
}
