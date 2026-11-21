package app

import (
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"howett.net/plist"
)

type provisioningPlist struct {
	Name                  string                 `plist:"Name"`
	UUID                  string                 `plist:"UUID"`
	AppIDName             string                 `plist:"AppIDName"`
	TeamName              string                 `plist:"TeamName"`
	TeamIdentifier        []string               `plist:"TeamIdentifier"`
	CreationDate          time.Time              `plist:"CreationDate"`
	ExpirationDate        time.Time              `plist:"ExpirationDate"`
	ProvisionedDevices    []string               `plist:"ProvisionedDevices"`
	ProvisionsAllDevices  bool                   `plist:"ProvisionsAllDevices"`
	Entitlements          map[string]interface{} `plist:"Entitlements"`
	DeveloperCertificates [][]byte               `plist:"DeveloperCertificates"`
}

// readProvisioning reads an embedded.mobileprovision (or
// embedded.provisionprofile on the Mac): a signed plist
func readProvisioning(data []byte) (*ProvisioningProfile, []Certificate, error) {
	p, err := parseProvisioningPlist(data)
	if err != nil {
		return nil, nil, err
	}
	return profileOf(p), developerCertificates(p.DeveloperCertificates), nil
}

// parseProvisioningPlist cuts the plist out of the signature around it
func parseProvisioningPlist(data []byte) (*provisioningPlist, error) {
	start := bytes.Index(data, []byte("<?xml"))
	end := bytes.Index(data, []byte("</plist>"))
	if start < 0 || end < start {
		return nil, errors.New("no plist in the provisioning profile")
	}
	var p provisioningPlist
	if _, err := plist.Unmarshal(data[start:end+len("</plist>")], &p); err != nil {
		return nil, fmt.Errorf("provisioning profile: %w", err)
	}
	return &p, nil
}

func profileOf(p *provisioningPlist) *ProvisioningProfile {
	profile := &ProvisioningProfile{
		Name:         p.Name,
		UUID:         p.UUID,
		AppIDName:    p.AppIDName,
		TeamName:     p.TeamName,
		CreatedAt:    p.CreationDate,
		ExpiresAt:    p.ExpirationDate,
		AllDevices:   p.ProvisionsAllDevices,
		Devices:      p.ProvisionedDevices,
		Entitlements: p.Entitlements,
		Kind:         profileKind(p),
	}
	if len(p.TeamIdentifier) > 0 {
		profile.TeamID = p.TeamIdentifier[0]
	}
	return profile
}

func profileKind(p *provisioningPlist) string {
	getTaskAllow, _ := p.Entitlements["get-task-allow"].(bool)
	switch {
	case p.ProvisionsAllDevices:
		return "enterprise"
	case len(p.ProvisionedDevices) > 0 && getTaskAllow:
		return "development"
	case len(p.ProvisionedDevices) > 0:
		return "ad-hoc"
	}
	return "app-store"
}

// developerCertificates skips the certificates that cannot be parsed
func developerCertificates(ders [][]byte) []Certificate {
	var certs []Certificate
	for _, der := range ders {
		if c, err := x509.ParseCertificate(der); err == nil {
			certs = append(certs, certificateOf(c))
		}
	}
	return certs
}
