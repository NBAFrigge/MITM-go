package panels

import (
	"fmt"
	"strings"

	"httpDebugger/pkg/sessiondata"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TLSPanel struct {
	viewport   viewport.Model
	rawContent string
}

func NewTLSPanel() *TLSPanel {
	vp := viewport.New(0, 0)
	return &TLSPanel{
		viewport:   vp,
		rawContent: "Select a session",
	}
}

func (p *TLSPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return cmd
}

func (p *TLSPanel) View() string {
	return p.viewport.View()
}

func (p *TLSPanel) SetSize(width, height int) {
	p.viewport.Width = width
	p.viewport.Height = height
	if p.rawContent != "" {
		wrappedContent := lipgloss.NewStyle().Width(width).Render(p.rawContent)
		p.viewport.SetContent(wrappedContent)
	}
}

func (p *TLSPanel) UpdateSession(session *sessiondata.Session) {
	if session == nil {
		p.rawContent = "Select a session"
		p.viewport.SetContent(lipgloss.NewStyle().Width(p.viewport.Width).Render(p.rawContent))
		return
	}

	if session.TLSFingerprint == nil {
		p.rawContent = "No TLS data (unencrypted HTTP or WS)"
		p.viewport.SetContent(lipgloss.NewStyle().Width(p.viewport.Width).Render(p.rawContent))
		return
	}

	fp := session.TLSFingerprint
	var content strings.Builder

	content.WriteString(fmt.Sprintf("TLS Fingerprint: %s\n", session.Request.URL))
	content.WriteString("────────────────────────────────────────\n\n")

	content.WriteString(fmt.Sprintf("JA3 Hash:   %s\n", fp.JA3Hash))
	content.WriteString(fmt.Sprintf("JA3 String: %s\n\n", fp.JA3))

	content.WriteString(fmt.Sprintf("TLS Version: 0x%04X (%s)\n", fp.TLSVersion, tlsVersionName(fp.TLSVersion)))

	if len(fp.SupportedVersions) > 0 {
		content.WriteString(fmt.Sprintf("Supported Versions (%d): ", len(fp.SupportedVersions)))
		versions := make([]string, 0, len(fp.SupportedVersions))
		for _, v := range fp.SupportedVersions {
			versions = append(versions, fmt.Sprintf("0x%04X", v))
		}
		content.WriteString(strings.Join(versions, ", "))
		content.WriteString("\n")
	}

	if len(fp.ALPNProtocols) > 0 {
		content.WriteString(fmt.Sprintf("ALPN: %s\n", strings.Join(fp.ALPNProtocols, ", ")))
	}

	content.WriteString("\n")

	if len(fp.CipherSuites) > 0 {
		content.WriteString(fmt.Sprintf("Cipher Suites (%d):\n", len(fp.CipherSuites)))
		for _, cipher := range fp.CipherSuites {
			name := cipherSuiteName(cipher)
			content.WriteString(fmt.Sprintf("  0x%04X  %s\n", cipher, name))
		}
		content.WriteString("\n")
	}

	if len(fp.Extensions) > 0 {
		content.WriteString(fmt.Sprintf("Extensions (%d):\n", len(fp.Extensions)))
		for _, ext := range fp.Extensions {
			name := extensionName(ext)
			content.WriteString(fmt.Sprintf("  0x%04X  %s\n", ext, name))
		}
		content.WriteString("\n")
	}

	if len(fp.EllipticCurves) > 0 {
		content.WriteString(fmt.Sprintf("Elliptic Curves (%d):\n", len(fp.EllipticCurves)))
		for _, curve := range fp.EllipticCurves {
			name := curveName(curve)
			content.WriteString(fmt.Sprintf("  0x%04X  %s\n", curve, name))
		}
		content.WriteString("\n")
	}

	if len(fp.ECPointFormats) > 0 {
		content.WriteString(fmt.Sprintf("EC Point Formats (%d): ", len(fp.ECPointFormats)))
		formats := make([]string, 0, len(fp.ECPointFormats))
		for _, f := range fp.ECPointFormats {
			formats = append(formats, fmt.Sprintf("%d", f))
		}
		content.WriteString(strings.Join(formats, ", "))
		content.WriteString("\n\n")
	}

	if len(fp.SignatureAlgs) > 0 {
		content.WriteString(fmt.Sprintf("Signature Algorithms (%d):\n", len(fp.SignatureAlgs)))
		for _, alg := range fp.SignatureAlgs {
			name := signatureAlgName(alg)
			content.WriteString(fmt.Sprintf("  0x%04X  %s\n", alg, name))
		}
		content.WriteString("\n")
	}

	if len(fp.KeyShareCurves) > 0 {
		content.WriteString(fmt.Sprintf("Key Share Curves (%d): ", len(fp.KeyShareCurves)))
		curves := make([]string, 0, len(fp.KeyShareCurves))
		for _, c := range fp.KeyShareCurves {
			curves = append(curves, curveName(c))
		}
		content.WriteString(strings.Join(curves, ", "))
		content.WriteString("\n\n")
	}

	if len(fp.CertCompAlgs) > 0 {
		content.WriteString(fmt.Sprintf("Cert Compression (%d): ", len(fp.CertCompAlgs)))
		algs := make([]string, 0, len(fp.CertCompAlgs))
		for _, a := range fp.CertCompAlgs {
			algs = append(algs, certCompName(a))
		}
		content.WriteString(strings.Join(algs, ", "))
		content.WriteString("\n\n")
	}

	if fp.RecordSizeLimit > 0 {
		content.WriteString(fmt.Sprintf("Record Size Limit: %d\n", fp.RecordSizeLimit))
	}

	p.rawContent = content.String()
	wrappedContent := lipgloss.NewStyle().Width(p.viewport.Width).Render(p.rawContent)
	p.viewport.SetContent(wrappedContent)
}

func tlsVersionName(v uint16) string {
	switch v {
	case 0x0301:
		return "TLS 1.0"
	case 0x0302:
		return "TLS 1.1"
	case 0x0303:
		return "TLS 1.2"
	case 0x0304:
		return "TLS 1.3"
	default:
		if isGREASE(v) {
			return "GREASE"
		}
		return "Unknown"
	}
}

func cipherSuiteName(id uint16) string {
	known := map[uint16]string{
		0x1301: "TLS_AES_128_GCM_SHA256",
		0x1302: "TLS_AES_256_GCM_SHA384",
		0x1303: "TLS_CHACHA20_POLY1305_SHA256",
		0xC02B: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		0xC02C: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		0xC02F: "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		0xC030: "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		0xCCA8: "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
		0xCCA9: "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		0xC013: "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
		0xC014: "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
		0x009C: "TLS_RSA_WITH_AES_128_GCM_SHA256",
		0x009D: "TLS_RSA_WITH_AES_256_GCM_SHA384",
		0x002F: "TLS_RSA_WITH_AES_128_CBC_SHA",
		0x0035: "TLS_RSA_WITH_AES_256_CBC_SHA",
	}
	if name, ok := known[id]; ok {
		return name
	}
	if isGREASE(id) {
		return "GREASE"
	}
	return ""
}

func extensionName(id uint16) string {
	known := map[uint16]string{
		0:     "server_name",
		1:     "max_fragment_length",
		5:     "status_request",
		10:    "supported_groups",
		11:    "ec_point_formats",
		13:    "signature_algorithms",
		16:    "application_layer_protocol_negotiation",
		17:    "extended_master_secret",
		18:    "signed_certificate_timestamp",
		21:    "padding",
		23:    "extended_master_secret",
		27:    "compress_certificate",
		28:    "record_size_limit",
		35:    "session_ticket",
		43:    "supported_versions",
		44:    "cookie",
		45:    "psk_key_exchange_modes",
		49:    "post_handshake_auth",
		50:    "signature_algorithms_cert",
		51:    "key_share",
		57:    "quic_transport_parameters",
		17513: "application_settings",
		65281: "renegotiation_info",
	}
	if name, ok := known[id]; ok {
		return name
	}
	if isGREASE(id) {
		return "GREASE"
	}
	return ""
}

func curveName(id uint16) string {
	known := map[uint16]string{
		23: "secp256r1",
		24: "secp384r1",
		25: "secp521r1",
		29: "x25519",
		30: "x448",
	}
	if name, ok := known[id]; ok {
		return name
	}
	if isGREASE(id) {
		return "GREASE"
	}
	return fmt.Sprintf("0x%04X", id)
}

func signatureAlgName(id uint16) string {
	known := map[uint16]string{
		0x0401: "rsa_pkcs1_sha256",
		0x0501: "rsa_pkcs1_sha384",
		0x0601: "rsa_pkcs1_sha512",
		0x0403: "ecdsa_secp256r1_sha256",
		0x0503: "ecdsa_secp384r1_sha384",
		0x0603: "ecdsa_secp521r1_sha512",
		0x0804: "rsa_pss_rsae_sha256",
		0x0805: "rsa_pss_rsae_sha384",
		0x0806: "rsa_pss_rsae_sha512",
		0x0807: "ed25519",
		0x0808: "ed448",
		0x0201: "rsa_pkcs1_sha1",
		0x0203: "ecdsa_sha1",
	}
	if name, ok := known[id]; ok {
		return name
	}
	return ""
}

func certCompName(id uint16) string {
	switch id {
	case 1:
		return "zlib"
	case 2:
		return "brotli"
	case 3:
		return "zstd"
	default:
		return fmt.Sprintf("0x%04X", id)
	}
}

func isGREASE(value uint16) bool {
	return (value&0x0F0F) == 0x0A0A && (value>>4)&0x0F == (value>>12)&0x0F
}
