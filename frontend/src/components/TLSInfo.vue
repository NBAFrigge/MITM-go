<template>
  <div class="detail-section">
    <h3 class="section-title">TLS Configuration</h3>

    <div v-if="!hasTLSConfig" class="no-tls-info">
      No TLS configuration available
    </div>

    <div v-else class="tls-content">
      <!-- Basic TLS Information -->
      <CollapsiblePanel title="Basic Configuration" :nested="true" :initial-expanded="true">
        <DetailList>
          <DetailItem label="Server Name" v-if="tlsConfig.ServerName">
            <span class="tls-server-name">{{ tlsConfig.ServerName }}</span>
          </DetailItem>
          <DetailItem label="Min Version">
            <span class="tls-version">{{ tlsConfig.MinVersion || 'Default' }}</span>
          </DetailItem>
          <DetailItem label="Max Version">
            <span class="tls-version">{{ tlsConfig.MaxVersion || 'Default' }}</span>
          </DetailItem>
          <DetailItem label="Insecure Skip Verify">
            <span class="tls-boolean" :class="{ 'insecure': tlsConfig.InsecureSkipVerify }">
              {{ tlsConfig.InsecureSkipVerify ? 'True' : 'False' }}
            </span>
          </DetailItem>
          <DetailItem label="Prefer Server Cipher Suites">
            <span class="tls-boolean">{{ tlsConfig.PreferServerCipherSuites ? 'True' : 'False' }}</span>
          </DetailItem>
          <DetailItem label="Session Tickets Disabled">
            <span class="tls-boolean">{{ tlsConfig.SessionTicketsDisabled ? 'True' : 'False' }}</span>
          </DetailItem>
          <DetailItem label="Dynamic Record Sizing Disabled">
            <span class="tls-boolean">{{ tlsConfig.DynamicRecordSizingDisabled ? 'True' : 'False' }}</span>
          </DetailItem>
          <DetailItem label="Renegotiation" v-if="tlsConfig.Renegotiation">
            <span class="tls-renegotiation">{{ tlsConfig.Renegotiation }}</span>
          </DetailItem>
          <DetailItem label="Client Auth" v-if="tlsConfig.ClientAuth">
            <span class="tls-client-auth">{{ tlsConfig.ClientAuth }}</span>
          </DetailItem>
        </DetailList>
      </CollapsiblePanel>

      <!-- Next Protocols -->
      <CollapsiblePanel
          v-if="tlsConfig.NextProtos && tlsConfig.NextProtos.length"
          title="Next Protocols"
          :nested="true"
          :initial-expanded="false"
      >
        <div class="protocol-list">
          <div
              v-for="(proto, index) in tlsConfig.NextProtos"
              :key="index"
              class="protocol-item"
          >
            {{ proto }}
          </div>
        </div>
      </CollapsiblePanel>

      <!-- Cipher Suites -->
      <CollapsiblePanel
          v-if="tlsConfig.CipherSuites && tlsConfig.CipherSuites.length"
          title="Cipher Suites"
          :nested="true"
          :initial-expanded="false"
      >
        <div class="cipher-list">
          <div
              v-for="(cipher, index) in tlsConfig.CipherSuites"
              :key="index"
              class="cipher-item"
          >
            {{ cipher }}
          </div>
        </div>
      </CollapsiblePanel>

      <!-- Curve Preferences -->
      <CollapsiblePanel
          v-if="tlsConfig.CurvePreferences && tlsConfig.CurvePreferences.length"
          title="Curve Preferences"
          :nested="true"
          :initial-expanded="false"
      >
        <div class="curve-list">
          <div
              v-for="(curve, index) in tlsConfig.CurvePreferences"
              :key="index"
              class="curve-item"
          >
            {{ curve }}
          </div>
        </div>
      </CollapsiblePanel>

      <!-- Certificates -->
      <CollapsiblePanel
          v-if="tlsConfig.Certificates && tlsConfig.Certificates.length"
          title="Certificates"
          :nested="true"
          :initial-expanded="false"
      >
        <div class="certificates-container">
          <div
              v-for="(cert, index) in tlsConfig.Certificates"
              :key="index"
              class="certificate-item"
          >
            <div class="certificate-header">
              <span class="certificate-index">Certificate {{ index + 1 }}</span>
              <span v-if="cert.HasPrivateKey" class="has-private-key">HAS PRIVATE KEY</span>
            </div>

            <DetailList v-if="cert.Subject || cert.Issuer">
              <DetailItem label="Subject" v-if="cert.Subject">
                <span class="cert-subject">{{ cert.Subject }}</span>
              </DetailItem>
              <DetailItem label="Issuer" v-if="cert.Issuer">
                <span class="cert-issuer">{{ cert.Issuer }}</span>
              </DetailItem>
              <DetailItem label="Serial Number" v-if="cert.SerialNumber">
                <span class="cert-serial">{{ cert.SerialNumber }}</span>
              </DetailItem>
              <DetailItem label="Valid From" v-if="cert.NotBefore">
                <span class="cert-date">{{ formatDate(cert.NotBefore) }}</span>
              </DetailItem>
              <DetailItem label="Valid Until" v-if="cert.NotAfter">
                <span class="cert-date" :class="{ 'expired': isExpired(cert.NotAfter) }">
                  {{ formatDate(cert.NotAfter) }}
                </span>
              </DetailItem>
              <DetailItem label="Is CA" v-if="cert.IsCA !== undefined">
                <span class="tls-boolean">{{ cert.IsCA ? 'True' : 'False' }}</span>
              </DetailItem>
              <DetailItem label="Version" v-if="cert.Version">
                <span class="cert-version">{{ cert.Version }}</span>
              </DetailItem>
              <DetailItem label="Signature Algorithm" v-if="cert.SignatureAlgorithm">
                <span class="cert-algorithm">{{ cert.SignatureAlgorithm }}</span>
              </DetailItem>
              <DetailItem label="Public Key Algorithm" v-if="cert.PublicKeyAlgorithm">
                <span class="cert-algorithm">{{ cert.PublicKeyAlgorithm }}</span>
              </DetailItem>
            </DetailList>

            <!-- DNS Names -->
            <CollapsiblePanel
                v-if="cert.DNSNames && cert.DNSNames.length"
                title="DNS Names"
                :nested="true"
                :initial-expanded="false"
            >
              <div class="dns-list">
                <div
                    v-for="(dns, dnsIndex) in cert.DNSNames"
                    :key="dnsIndex"
                    class="dns-item"
                >
                  {{ dns }}
                </div>
              </div>
            </CollapsiblePanel>

            <!-- IP Addresses -->
            <CollapsiblePanel
                v-if="cert.IPAddresses && cert.IPAddresses.length"
                title="IP Addresses"
                :nested="true"
                :initial-expanded="false"
            >
              <div class="ip-list">
                <div
                    v-for="(ip, ipIndex) in cert.IPAddresses"
                    :key="ipIndex"
                    class="ip-item"
                >
                  {{ ip }}
                </div>
              </div>
            </CollapsiblePanel>

            <!-- Key Usage -->
            <CollapsiblePanel
                v-if="cert.KeyUsage && cert.KeyUsage.length"
                title="Key Usage"
                :nested="true"
                :initial-expanded="false"
            >
              <div class="usage-list">
                <div
                    v-for="(usage, usageIndex) in cert.KeyUsage"
                    :key="usageIndex"
                    class="usage-item"
                >
                  {{ usage }}
                </div>
              </div>
            </CollapsiblePanel>

            <!-- Extended Key Usage -->
            <CollapsiblePanel
                v-if="cert.ExtKeyUsage && cert.ExtKeyUsage.length"
                title="Extended Key Usage"
                :nested="true"
                :initial-expanded="false"
            >
              <div class="usage-list">
                <div
                    v-for="(usage, usageIndex) in cert.ExtKeyUsage"
                    :key="usageIndex"
                    class="usage-item"
                >
                  {{ usage }}
                </div>
              </div>
            </CollapsiblePanel>

            <!-- Supported Signature Algorithms -->
            <CollapsiblePanel
                v-if="cert.SupportedSignatureAlgorithms && cert.SupportedSignatureAlgorithms.length"
                title="Supported Signature Algorithms"
                :nested="true"
                :initial-expanded="false"
            >
              <div class="algorithm-list">
                <div
                    v-for="(algo, algoIndex) in cert.SupportedSignatureAlgorithms"
                    :key="algoIndex"
                    class="algorithm-item"
                >
                  {{ algo }}
                </div>
              </div>
            </CollapsiblePanel>
          </div>
        </div>
      </CollapsiblePanel>

      <!-- Root CAs -->
      <CollapsiblePanel
          v-if="tlsConfig.RootCAs && tlsConfig.RootCAs.CertCount > 0"
          title="Root Certificate Authorities"
          :nested="true"
          :initial-expanded="false"
      >
        <DetailList>
          <DetailItem label="Certificate Count">
            <span class="cert-count">{{ tlsConfig.RootCAs.CertCount }}</span>
          </DetailItem>
        </DetailList>
        <CollapsiblePanel
            v-if="tlsConfig.RootCAs.Subjects && tlsConfig.RootCAs.Subjects.length"
            title="Subjects"
            :nested="true"
            :initial-expanded="false"
        >
          <div class="subject-list">
            <div
                v-for="(subject, index) in tlsConfig.RootCAs.Subjects"
                :key="index"
                class="subject-item"
            >
              {{ subject }}
            </div>
          </div>
        </CollapsiblePanel>
      </CollapsiblePanel>

      <!-- Client CAs -->
      <CollapsiblePanel
          v-if="tlsConfig.ClientCAs && tlsConfig.ClientCAs.CertCount > 0"
          title="Client Certificate Authorities"
          :nested="true"
          :initial-expanded="false"
      >
        <DetailList>
          <DetailItem label="Certificate Count">
            <span class="cert-count">{{ tlsConfig.ClientCAs.CertCount }}</span>
          </DetailItem>
        </DetailList>
        <CollapsiblePanel
            v-if="tlsConfig.ClientCAs.Subjects && tlsConfig.ClientCAs.Subjects.length"
            title="Subjects"
            :nested="true"
            :initial-expanded="false"
        >
          <div class="subject-list">
            <div
                v-for="(subject, index) in tlsConfig.ClientCAs.Subjects"
                :key="index"
                class="subject-item"
            >
              {{ subject }}
            </div>
          </div>
        </CollapsiblePanel>
      </CollapsiblePanel>
    </div>
  </div>
</template>

<script>
import CollapsiblePanel from './CollapsiblePanel.vue';
import DetailList from './DetailList.vue';
import DetailItem from './DetailItem.vue';

export default {
  name: 'TLSInfo',
  components: {
    CollapsiblePanel,
    DetailList,
    DetailItem
  },
  props: {
    session: {
      type: Object,
      required: true
    }
  },
  computed: {
    tlsConfig() {
      return this.session?.tlsProfile || {};
    },
    hasTLSConfig() {
      return this.tlsConfig && Object.keys(this.tlsConfig).length > 0;
    }
  },
  methods: {
    formatDate(dateString) {
      if (!dateString) return 'N/A';
      try {
        const date = new Date(dateString);
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
      } catch (e) {
        return dateString;
      }
    },
    isExpired(dateString) {
      if (!dateString) return false;
      try {
        const date = new Date(dateString);
        return date < new Date();
      } catch (e) {
        return false;
      }
    }
  }
}
</script>

<style scoped>
.detail-section {
  margin-bottom: var(--spacing-md);
  font-family: var(--font-family);
}

.section-title {
  font-size: var(--font-size-small);
  font-weight: 600;
  color: var(--text-color-primary);
  margin: 0 0 var(--spacing-sm) 0;
  padding-bottom: var(--spacing-xs);
  border-bottom: 1px solid var(--border-color);
}

.no-tls-info {
  text-align: center;
  padding: var(--spacing-lg);
  color: var(--text-color-secondary);
  font-style: italic;
  font-size: var(--font-size-small);
  background-color: var(--bg-color-medium);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
}

.tls-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

/* TLS-specific styling */
.tls-server-name {
  color: var(--accent-color-blue);
  font-weight: 600;
  font-family: var(--font-family);
}

.tls-version {
  color: var(--accent-color-green);
  font-weight: 600;
  background-color: var(--input-background);
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  font-size: 11px;
}

.tls-boolean {
  font-weight: 600;
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-size: 11px;
  color: white;
  background-color: var(--status-success);
}

.tls-boolean.insecure {
  background-color: var(--status-error);
}

.tls-renegotiation,
.tls-client-auth {
  color: var(--accent-color-purple);
  font-weight: 600;
  background-color: var(--input-background);
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  font-size: 11px;
}

/* List styling */
.protocol-list,
.cipher-list,
.curve-list,
.dns-list,
.ip-list,
.usage-list,
.algorithm-list,
.subject-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  max-height: 200px;
  overflow-y: auto;
}

.protocol-item,
.cipher-item,
.curve-item,
.dns-item,
.ip-item,
.usage-item,
.algorithm-item,
.subject-item {
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: var(--input-background);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-family: var(--font-family);
  font-size: var(--font-size-small);
  color: var(--text-color-primary);
  transition: var(--transition-fast);
}

.protocol-item:hover,
.cipher-item:hover,
.curve-item:hover,
.dns-item:hover,
.ip-item:hover,
.usage-item:hover,
.algorithm-item:hover,
.subject-item:hover {
  background-color: var(--hover-color);
  border-color: var(--accent-color-blue);
}

.protocol-item {
  border-left: 3px solid var(--accent-color-blue);
}

.cipher-item {
  border-left: 3px solid var(--accent-color-green);
}

.curve-item {
  border-left: 3px solid var(--accent-color-purple);
}

.dns-item {
  border-left: 3px solid var(--accent-color-yellow);
  font-family: var(--font-family);
}

.ip-item {
  border-left: 3px solid var(--number-color);
  font-family: var(--font-family);
}

.usage-item {
  border-left: 3px solid var(--field-color);
}

.algorithm-item {
  border-left: 3px solid var(--method-color);
}

.subject-item {
  border-left: 3px solid var(--comment-color);
  font-family: var(--font-family);
  font-size: 10px;
  word-break: break-all;
}

/* Certificate styling */
.certificates-container {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.certificate-item {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background-color: var(--bg-color-medium);
  overflow: hidden;
}

.certificate-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: var(--bg-color-light);
  border-bottom: 1px solid var(--border-color);
}

.certificate-index {
  font-weight: 600;
  color: var(--text-color-primary);
  font-size: var(--font-size-small);
}

.has-private-key {
  background-color: var(--status-success);
  color: white;
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.cert-subject,
.cert-issuer {
  color: var(--string-color);
  font-family: var(--font-family);
  word-break: break-all;
}

.cert-serial {
  color: var(--number-color);
  font-family: var(--font-family);
  font-weight: 600;
}

.cert-date {
  color: var(--text-color-primary);
  font-family: var(--font-family);
  font-weight: 500;
}

.cert-date.expired {
  color: var(--status-error);
  font-weight: 600;
}

.cert-version {
  color: var(--number-color);
  font-weight: 600;
  background-color: var(--input-background);
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  font-size: 11px;
}

.cert-algorithm {
  color: var(--method-color);
  font-family: var(--font-family);
  font-weight: 600;
}

.cert-count {
  color: var(--number-color);
  font-weight: 600;
  background-color: var(--input-background);
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  font-size: 11px;
}

/* Custom scrollbar for lists */
.protocol-list::-webkit-scrollbar,
.cipher-list::-webkit-scrollbar,
.curve-list::-webkit-scrollbar,
.dns-list::-webkit-scrollbar,
.ip-list::-webkit-scrollbar,
.usage-list::-webkit-scrollbar,
.algorithm-list::-webkit-scrollbar,
.subject-list::-webkit-scrollbar {
  width: 6px;
}

.protocol-list::-webkit-scrollbar-track,
.cipher-list::-webkit-scrollbar-track,
.curve-list::-webkit-scrollbar-track,
.dns-list::-webkit-scrollbar-track,
.ip-list::-webkit-scrollbar-track,
.usage-list::-webkit-scrollbar-track,
.algorithm-list::-webkit-scrollbar-track,
.subject-list::-webkit-scrollbar-track {
  background: var(--scrollbar-track);
}

.protocol-list::-webkit-scrollbar-thumb,
.cipher-list::-webkit-scrollbar-thumb,
.curve-list::-webkit-scrollbar-thumb,
.dns-list::-webkit-scrollbar-thumb,
.ip-list::-webkit-scrollbar-thumb,
.usage-list::-webkit-scrollbar-thumb,
.algorithm-list::-webkit-scrollbar-thumb,
.subject-list::-webkit-scrollbar-thumb {
  background: var(--scrollbar-thumb);
  border-radius: 3px;
}

.protocol-list::-webkit-scrollbar-thumb:hover,
.cipher-list::-webkit-scrollbar-thumb:hover,
.curve-list::-webkit-scrollbar-thumb:hover,
.dns-list::-webkit-scrollbar-thumb:hover,
.ip-list::-webkit-scrollbar-thumb:hover,
.usage-list::-webkit-scrollbar-thumb:hover,
.algorithm-list::-webkit-scrollbar-thumb:hover,
.subject-list::-webkit-scrollbar-thumb:hover {
  background: var(--scrollbar-thumb-hover);
}
</style>