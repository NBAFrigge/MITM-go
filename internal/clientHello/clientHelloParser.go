package clientHello

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"hash/fnv"
	"strconv"
)

const (
	ExtensionSNI               = 0
	ExtensionSupportedCurves   = 10
	ExtensionSignatureAlgs     = 13
	ExtensionALPN              = 16
	ExtensionSupportedVersions = 43
)

// TODO Add checks on lenght
func ParseClientHello(rawClientHello []byte) (*tls.Config, error) {
	isClientHello := len(rawClientHello) > 43 && rawClientHello[0] == 0x16 && rawClientHello[5] == 0x01
	if !isClientHello {
		return nil, errors.New("not a valid client hello")
	}

	// get tls length
	length := binary.BigEndian.Uint16(rawClientHello[3:5])
	rawClientHello = rawClientHello[5 : length+5]

	// parse tls version
	tlsConfig := &tls.Config{}
	clientVersion := binary.BigEndian.Uint16(rawClientHello[4:6])
	setTLSVersion(tlsConfig, clientVersion)

	// handshake header + version + client random = byte 38
	offset := 38

	// read session ID length and skip session ID
	sessionIDLen := int(rawClientHello[offset])
	offset += 1 + sessionIDLen

	// get cipherSuites length
	cipherSuitesLen := binary.BigEndian.Uint16(rawClientHello[offset : offset+2])
	offset += 2

	// parse cipher suites
	var cipherSuites []uint16
	for i := 0; i < int(cipherSuitesLen); i += 2 {
		cipher := binary.BigEndian.Uint16(rawClientHello[offset+i : offset+i+2])
		cipherSuites = append(cipherSuites, cipher)
	}
	offset += int(cipherSuitesLen)

	// apply cipher suites to config
	tlsConfig.CipherSuites = cipherSuites

	// read compression methods length
	compressionMethodsLen := int(rawClientHello[offset])
	offset++

	// skip compression methods extract

	offset += compressionMethodsLen

	// extension
	extensionsLen := binary.BigEndian.Uint16(rawClientHello[offset : offset+2])
	offset += 2

	extensionsEnd := offset + int(extensionsLen)
	for offset < extensionsEnd {
		if offset+4 > extensionsEnd {
			return nil, errors.New("malformed extension header")
		}

		extType := binary.BigEndian.Uint16(rawClientHello[offset : offset+2])
		extLen := binary.BigEndian.Uint16(rawClientHello[offset+2 : offset+4])
		offset += 4

		if offset+int(extLen) > extensionsEnd {
			return nil, errors.New("malformed extension data")
		}

		extData := rawClientHello[offset : offset+int(extLen)]

		// parse specific extensions
		switch extType {
		case ExtensionALPN:
			alpnProtocols := parseALPN(extData)
			if len(alpnProtocols) > 0 {
				tlsConfig.NextProtos = alpnProtocols
			}
		case ExtensionSupportedCurves:
			supportedCurves := parseSupportedCurves(extData)
			if len(supportedCurves) > 0 {
				tlsConfig.CurvePreferences = supportedCurves
			}
		case ExtensionSupportedVersions:
			supportedVersions := parseSupportedVersions(extData)
			if len(supportedVersions) > 0 {
				updateTLSVersionFromExtension(tlsConfig, supportedVersions)
			}
		}

		offset += int(extLen)
	}

	return tlsConfig, nil
}

// GenerateClientHelloHash generates a hash for the given ClientHello data, hashing the middle 100 bytes
func GenerateClientHelloHash(data []byte) string {
	if len(data) < 200 {
		h := fnv.New64a()
		h.Write(data)
		return strconv.FormatUint(h.Sum64(), 16)
	}

	h := fnv.New64a()

	h.Write(data[100:200])

	h.Write([]byte{byte(len(data) / 100)})

	return strconv.FormatUint(h.Sum64(), 16)
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
		// Default to TLS 1.2 for unknown versions
		config.MinVersion = tls.VersionTLS12
		config.MaxVersion = tls.VersionTLS12
	}
}

func parseSupportedCurves(data []byte) []tls.CurveID {
	if len(data) < 2 {
		return nil
	}

	curvesLen := binary.BigEndian.Uint16(data[0:2])
	var curves []tls.CurveID

	for i := 2; i < int(curvesLen)+2 && i < len(data)-1; i += 2 {
		curve := tls.CurveID(binary.BigEndian.Uint16(data[i : i+2]))
		curves = append(curves, curve)
	}

	return curves
}

func parseALPN(data []byte) []string {
	if len(data) < 2 {
		return nil
	}

	protocolsLen := binary.BigEndian.Uint16(data[0:2])
	var protocols []string
	offset := 2

	for offset < int(protocolsLen)+2 && offset < len(data) {
		if offset >= len(data) {
			break
		}

		protoLen := int(data[offset])
		offset++

		if offset+protoLen > len(data) {
			break
		}

		protocol := string(data[offset : offset+protoLen])
		protocols = append(protocols, protocol)
		offset += protoLen
	}

	return protocols
}

func parseSupportedVersions(data []byte) []uint16 {
	if len(data) < 1 {
		return nil
	}

	versionsLen := int(data[0])
	var versions []uint16

	for i := 1; i < versionsLen+1 && i < len(data)-1; i += 2 {
		version := binary.BigEndian.Uint16(data[i : i+2])
		versions = append(versions, version)
	}

	return versions
}

func updateTLSVersionFromExtension(config *tls.Config, versions []uint16) {
	var minVersion, maxVersion uint16 = 0xFFFF, 0

	for _, version := range versions {
		// Skip GREASE values
		if isGREASE(version) {
			continue
		}

		if version < minVersion {
			minVersion = version
		}
		if version > maxVersion {
			maxVersion = version
		}
	}

	if minVersion != 0xFFFF {
		config.MinVersion = mapToGoTLSVersion(minVersion)
	}
	if maxVersion != 0 {
		config.MaxVersion = mapToGoTLSVersion(maxVersion)
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
		return tls.VersionTLS12 // Default fallback
	}
}

func isGREASE(value uint16) bool {
	return (value&0x0F0F) == 0x0A0A && (value&0xF0F0)>>4 == (value&0xF0F0)>>12
}

// helper function to get the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
