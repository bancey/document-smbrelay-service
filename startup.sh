#!/bin/sh

# Generate krb5.conf from environment variables if Kerberos is configured
if [ ! -z "$KRB5_DEFAULT_REALM" ]; then
cat > /etc/krb5.conf << EOF
[libdefaults]
    default_realm = ${KRB5_DEFAULT_REALM}
    dns_lookup_realm = ${KRB5_DNS_LOOKUP_REALM:-true}
    dns_lookup_kdc = ${KRB5_DNS_LOOKUP_KDC:-true}
    ticket_lifetime = ${KRB5_TICKET_LIFETIME:-24h}
    renew_lifetime = ${KRB5_RENEW_LIFETIME:-7d}
    forwardable = ${KRB5_FORWARDABLE:-true}
    rdns = ${KRB5_RDNS:-false}
EOF

# Only add [realms] section if KDC is explicitly specified
if [ ! -z "$KRB5_KDC" ]; then
cat >> /etc/krb5.conf << EOF

[realms]
    ${KRB5_DEFAULT_REALM} = {
        kdc = ${KRB5_KDC}
        admin_server = ${KRB5_ADMIN_SERVER:-${KRB5_KDC}}
    }
EOF
fi

# Add domain realm mappings
if [ ! -z "$KRB5_DOMAIN" ]; then
cat >> /etc/krb5.conf << EOF

[domain_realm]
    .${KRB5_DOMAIN} = ${KRB5_DEFAULT_REALM}
    ${KRB5_DOMAIN} = ${KRB5_DEFAULT_REALM}
EOF
fi

echo "Generated /etc/krb5.conf:"
cat /etc/krb5.conf

# If keytab is provided as base64, decode it
if [ ! -z "$KRB5_KEYTAB_BASE64" ]; then
    echo "Decoding keytab from environment variable..."
    echo "$KRB5_KEYTAB_BASE64" | base64 -d > /etc/krb5.keytab
    chmod 600 /etc/krb5.keytab
    echo "Keytab created at /etc/krb5.keytab"
fi
fi

# Start the Go application
exec /app/server
