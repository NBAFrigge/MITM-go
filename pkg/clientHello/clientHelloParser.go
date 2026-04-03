package clientHello

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"strconv"

	utls "github.com/refraction-networking/utls"
)

const (
	ExtensionSNI               = 0
	ExtensionSupportedCurves   = 10
	ExtensionECPointFormats    = 11
	ExtensionSignatureAlgs     = 13
	ExtensionALPN              = 16
	ExtensionCertCompression   = 27
	ExtensionRecordSizeLimit   = 28
	ExtensionSupportedVersions = 43
	ExtensionKeyShare          = 51
)

func ParseClientHelloFull(rawClientHello []byte) (*TLSFingerprint, error) {
	if len(rawClientHello) < 44 {
		return nil, errors.New("not a valid client hello: too short")
	}
	if rawClientHello[0] != 0x16 || rawClientHello[5] != 0x01 {
		return nil, errors.New("not a valid client hello: invalid record type")
	}

	fp := &TLSFingerprint{}

	fingerprinter := &utls.Fingerprinter{}
	spec, err := fingerprinter.FingerprintClientHello(rawClientHello)
	if err == nil && spec != nil {
		fp.Spec = spec
		extractFromSpec(fp, spec)
	}

	extractFromRaw(fp, rawClientHello)

	fp.ComputeJA3()
	return fp, nil
}

func extractFromSpec(fp *TLSFingerprint, spec *utls.ClientHelloSpec) {
	fp.TLSVersion = spec.TLSVersMax

	for _, suite := range spec.CipherSuites {
		fp.CipherSuites = append(fp.CipherSuites, uint16(suite))
	}

	for _, ext := range spec.Extensions {
		switch e := ext.(type) {
		case *utls.SupportedCurvesExtension:
			for _, curve := range e.Curves {
				fp.EllipticCurves = append(fp.EllipticCurves, uint16(curve))
			}
		case *utls.SupportedPointsExtension:
			fp.ECPointFormats = append(fp.ECPointFormats, e.SupportedPoints...)
		case *utls.SignatureAlgorithmsExtension:
			for _, alg := range e.SupportedSignatureAlgorithms {
				fp.SignatureAlgs = append(fp.SignatureAlgs, uint16(alg))
			}
		case *utls.ALPNExtension:
			fp.ALPNProtocols = append(fp.ALPNProtocols, e.AlpnProtocols...)
		case *utls.SupportedVersionsExtension:
			for _, v := range e.Versions {
				fp.SupportedVersions = append(fp.SupportedVersions, v)
			}
		case *utls.KeyShareExtension:
			for _, ks := range e.KeyShares {
				fp.KeyShareCurves = append(fp.KeyShareCurves, uint16(ks.Group))
			}
		}
	}
}

func extractFromRaw(fp *TLSFingerprint, rawClientHello []byte) {
	length := binary.BigEndian.Uint16(rawClientHello[3:5])
	if int(length)+5 > len(rawClientHello) {
		return
	}
	data := rawClientHello[5 : length+5]

	if fp.TLSVersion == 0 {
		fp.TLSVersion = binary.BigEndian.Uint16(data[4:6])
	}

	offset := 38
	if offset >= len(data) {
		return
	}

	sessionIDLen := int(data[offset])
	offset += 1 + sessionIDLen

	if offset+2 > len(data) {
		return
	}
	cipherSuitesLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2

	if len(fp.CipherSuites) == 0 {
		if offset+cipherSuitesLen <= len(data) {
			for i := 0; i < cipherSuitesLen; i += 2 {
				cipher := binary.BigEndian.Uint16(data[offset+i : offset+i+2])
				fp.CipherSuites = append(fp.CipherSuites, cipher)
			}
		}
	}
	offset += cipherSuitesLen

	if offset >= len(data) {
		return
	}
	compressionLen := int(data[offset])
	offset += 1 + compressionLen

	if offset+2 > len(data) {
		return
	}
	extensionsLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2

	extensionsEnd := offset + extensionsLen
	if extensionsEnd > len(data) {
		extensionsEnd = len(data)
	}

	fp.Extensions = nil
	for offset+4 <= extensionsEnd {
		extType := binary.BigEndian.Uint16(data[offset : offset+2])
		extLen := int(binary.BigEndian.Uint16(data[offset+2 : offset+4]))
		offset += 4

		if offset+extLen > extensionsEnd {
			break
		}

		fp.Extensions = append(fp.Extensions, extType)
		extData := data[offset : offset+extLen]

		switch extType {
		case ExtensionALPN:
			if len(fp.ALPNProtocols) == 0 {
				fp.ALPNProtocols = parseALPN(extData)
			}
		case ExtensionSupportedCurves:
			if len(fp.EllipticCurves) == 0 {
				fp.EllipticCurves = parseSupportedCurvesRaw(extData)
			}
		case ExtensionECPointFormats:
			if len(fp.ECPointFormats) == 0 {
				fp.ECPointFormats = parseECPointFormats(extData)
			}
		case ExtensionSignatureAlgs:
			if len(fp.SignatureAlgs) == 0 {
				fp.SignatureAlgs = parseSignatureAlgorithms(extData)
			}
		case ExtensionSupportedVersions:
			if len(fp.SupportedVersions) == 0 {
				fp.SupportedVersions = parseSupportedVersions(extData)
			}
		case ExtensionKeyShare:
			if len(fp.KeyShareCurves) == 0 {
				fp.KeyShareCurves = parseKeyShareCurves(extData)
			}
		case ExtensionRecordSizeLimit:
			if fp.RecordSizeLimit == 0 && len(extData) >= 2 {
				fp.RecordSizeLimit = binary.BigEndian.Uint16(extData[0:2])
			}
		case ExtensionCertCompression:
			if len(fp.CertCompAlgs) == 0 {
				fp.CertCompAlgs = parseCertCompression(extData)
			}
		}

		offset += extLen
	}
}

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

func isGREASE(value uint16) bool {
	return (value&0x0F0F) == 0x0A0A && (value>>4)&0x0F == (value>>12)&0x0F
}

func parseSupportedCurvesRaw(data []byte) []uint16 {
	if len(data) < 2 {
		return nil
	}
	curvesLen := int(binary.BigEndian.Uint16(data[0:2]))
	var curves []uint16
	for i := 2; i+1 < curvesLen+2 && i+1 < len(data); i += 2 {
		curves = append(curves, binary.BigEndian.Uint16(data[i:i+2]))
	}
	return curves
}

func parseECPointFormats(data []byte) []uint8 {
	if len(data) < 1 {
		return nil
	}
	formatsLen := int(data[0])
	var formats []uint8
	for i := 1; i < formatsLen+1 && i < len(data); i++ {
		formats = append(formats, data[i])
	}
	return formats
}

func parseSignatureAlgorithms(data []byte) []uint16 {
	if len(data) < 2 {
		return nil
	}
	algsLen := int(binary.BigEndian.Uint16(data[0:2]))
	var algs []uint16
	for i := 2; i+1 < algsLen+2 && i+1 < len(data); i += 2 {
		algs = append(algs, binary.BigEndian.Uint16(data[i:i+2]))
	}
	return algs
}

func parseALPN(data []byte) []string {
	if len(data) < 2 {
		return nil
	}
	protocolsLen := int(binary.BigEndian.Uint16(data[0:2]))
	var protocols []string
	offset := 2
	for offset < protocolsLen+2 && offset < len(data) {
		protoLen := int(data[offset])
		offset++
		if offset+protoLen > len(data) {
			break
		}
		protocols = append(protocols, string(data[offset:offset+protoLen]))
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
	for i := 1; i+1 < versionsLen+1 && i+1 < len(data); i += 2 {
		versions = append(versions, binary.BigEndian.Uint16(data[i:i+2]))
	}
	return versions
}

func parseKeyShareCurves(data []byte) []uint16 {
	if len(data) < 2 {
		return nil
	}
	sharesLen := int(binary.BigEndian.Uint16(data[0:2]))
	var curves []uint16
	offset := 2
	for offset+4 <= sharesLen+2 && offset+4 <= len(data) {
		group := binary.BigEndian.Uint16(data[offset : offset+2])
		keyLen := int(binary.BigEndian.Uint16(data[offset+2 : offset+4]))
		curves = append(curves, group)
		offset += 4 + keyLen
	}
	return curves
}

func parseCertCompression(data []byte) []uint16 {
	if len(data) < 1 {
		return nil
	}
	algsLen := int(data[0])
	var algs []uint16
	for i := 1; i+1 < algsLen+1 && i+1 < len(data); i += 2 {
		algs = append(algs, binary.BigEndian.Uint16(data[i:i+2]))
	}
	return algs
}
