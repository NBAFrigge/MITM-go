package clientHello

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"strings"

	utls "github.com/refraction-networking/utls"
)

type TLSFingerprint struct {
	TLSVersion        uint16                `json:"tls_version"`
	CipherSuites      []uint16              `json:"cipher_suites"`
	Extensions        []uint16              `json:"extensions"`
	EllipticCurves    []uint16              `json:"elliptic_curves"`
	ECPointFormats    []uint8               `json:"ec_point_formats"`
	SignatureAlgs     []uint16              `json:"signature_algorithms"`
	SupportedVersions []uint16              `json:"supported_versions"`
	ALPNProtocols     []string              `json:"alpn_protocols"`
	KeyShareCurves    []uint16              `json:"key_share_curves"`
	CertCompAlgs      []uint16              `json:"cert_compression_algorithms,omitempty"`
	RecordSizeLimit   uint16                `json:"record_size_limit,omitempty"`
	JA3               string                `json:"ja3"`
	JA3Hash           string                `json:"ja3_hash"`
	Spec              *utls.ClientHelloSpec `json:"-"`
}

func (f *TLSFingerprint) ComputeJA3() {
	parts := make([]string, 5)

	parts[0] = fmt.Sprintf("%d", f.TLSVersion)

	ciphers := make([]string, 0, len(f.CipherSuites))
	for _, c := range f.CipherSuites {
		if !isGREASE(c) {
			ciphers = append(ciphers, fmt.Sprintf("%d", c))
		}
	}
	parts[1] = strings.Join(ciphers, "-")

	exts := make([]string, 0, len(f.Extensions))
	for _, e := range f.Extensions {
		if !isGREASE(e) {
			exts = append(exts, fmt.Sprintf("%d", e))
		}
	}
	parts[2] = strings.Join(exts, "-")

	curves := make([]string, 0, len(f.EllipticCurves))
	for _, c := range f.EllipticCurves {
		if !isGREASE(c) {
			curves = append(curves, fmt.Sprintf("%d", c))
		}
	}
	parts[3] = strings.Join(curves, "-")

	points := make([]string, 0, len(f.ECPointFormats))
	for _, p := range f.ECPointFormats {
		points = append(points, fmt.Sprintf("%d", p))
	}
	parts[4] = strings.Join(points, "-")

	f.JA3 = strings.Join(parts, ",")
	hash := md5.Sum([]byte(f.JA3))
	f.JA3Hash = fmt.Sprintf("%x", hash)
}

func (f *TLSFingerprint) ToTLSConfig() *tls.Config {
	config := &tls.Config{}

	var filteredCiphers []uint16
	for _, c := range f.CipherSuites {
		if !isGREASE(c) {
			filteredCiphers = append(filteredCiphers, c)
		}
	}
	config.CipherSuites = filteredCiphers

	var curves []tls.CurveID
	for _, c := range f.EllipticCurves {
		if !isGREASE(c) {
			curves = append(curves, tls.CurveID(c))
		}
	}
	config.CurvePreferences = curves

	if len(f.ALPNProtocols) > 0 {
		config.NextProtos = f.ALPNProtocols
	}

	if len(f.SupportedVersions) > 0 {
		var minV, maxV uint16 = 0xFFFF, 0
		for _, v := range f.SupportedVersions {
			if isGREASE(v) {
				continue
			}
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}
		if minV != 0xFFFF {
			config.MinVersion = mapToGoTLSVersion(minV)
		}
		if maxV != 0 {
			config.MaxVersion = mapToGoTLSVersion(maxV)
		}
	} else {
		setTLSVersion(config, f.TLSVersion)
	}

	return config
}

func setTLSVersion(config *tls.Config, version uint16) {
	switch version {
	case 0x0301:
		config.MinVersion = tls.VersionTLS10
		config.MaxVersion = tls.VersionTLS10
	case 0x0302:
		config.MinVersion = tls.VersionTLS11
		config.MaxVersion = tls.VersionTLS11
	case 0x0303:
		config.MinVersion = tls.VersionTLS12
		config.MaxVersion = tls.VersionTLS12
	case 0x0304:
		config.MinVersion = tls.VersionTLS13
		config.MaxVersion = tls.VersionTLS13
	default:
		config.MinVersion = tls.VersionTLS12
		config.MaxVersion = tls.VersionTLS12
	}
}

func mapToGoTLSVersion(version uint16) uint16 {
	switch version {
	case 0x0301:
		return tls.VersionTLS10
	case 0x0302:
		return tls.VersionTLS11
	case 0x0303:
		return tls.VersionTLS12
	case 0x0304:
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}
