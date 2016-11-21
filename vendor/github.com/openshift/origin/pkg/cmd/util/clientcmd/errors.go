package clientcmd

import (
	"crypto/x509"
	"errors"
	"fmt"
	"strings"

	kerrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

const (
	unknownReason = iota
	noServerFoundReason
	certificateAuthorityUnknownReason
	certificateHostnameErrorReason
	certificateInvalidReason
	configurationInvalidReason
	tlsOversizedRecordReason

	certificateAuthorityUnknownMsg = "The server uses a certificate signed by unknown authority. You may need to use the --certificate-authority flag to provide the path to a certificate file for the certificate authority, or --insecure-skip-tls-verify to bypass the certificate check and use insecure connections."
	notConfiguredMsg               = `The client is not configured. You need to run the login command in order to create a default config for your server and credentials:
  oc login
You can also run this command again providing the path to a config file directly, either through the --config flag of the KUBECONFIG environment variable.
`
	tlsOversizedRecordMsg = `Unable to connect to %[2]s using TLS: %[1]s.
Ensure the specified server supports HTTPS.`
)

// GetPrettyMessageFor prettifys the message of the provided error
func GetPrettyMessageFor(err error) string {
	return GetPrettyMessageForServer(err, "")
}

// GetPrettyMessageForServer prettifys the message of the provided error
func GetPrettyMessageForServer(err error, serverName string) string {
	if err == nil {
		return ""
	}

	reason := detectReason(err)

	switch reason {
	case noServerFoundReason:
		return notConfiguredMsg

	case certificateAuthorityUnknownReason:
		return certificateAuthorityUnknownMsg

	case tlsOversizedRecordReason:
		if len(serverName) == 0 {
			serverName = "server"
		}
		return fmt.Sprintf(tlsOversizedRecordMsg, err, serverName)

	case certificateHostnameErrorReason:
		return fmt.Sprintf("The server is using a certificate that does not match its hostname: %s", err)

	case certificateInvalidReason:
		return fmt.Sprintf("The server is using an invalid certificate: %s", err)
	}

	return err.Error()
}

// GetPrettyErrorFor prettifys the message of the provided error
func GetPrettyErrorFor(err error) error {
	return GetPrettyErrorForServer(err, "")
}

// GetPrettyErrorForServer prettifys the message of the provided error
func GetPrettyErrorForServer(err error, serverName string) error {
	return errors.New(GetPrettyMessageForServer(err, serverName))
}

// IsNoServerFound checks whether the provided error is a 'no server found' error or not
func IsNoServerFound(err error) bool {
	return detectReason(err) == noServerFoundReason
}

// IsConfigurationInvalid checks whether the provided error is a 'invalid configuration' error or not
func IsConfigurationInvalid(err error) bool {
	return detectReason(err) == configurationInvalidReason
}

// IsCertificateAuthorityUnknown checks whether the provided error is a 'certificate authority unknown' error or not
func IsCertificateAuthorityUnknown(err error) bool {
	return detectReason(err) == certificateAuthorityUnknownReason
}

// IsForbidden checks whether the provided error is a 'forbidden' error or not
func IsForbidden(err error) bool {
	return kerrors.IsForbidden(err)
}

// IsTLSOversizedRecord checks whether the provided error is a url.Error
// with "tls: oversized record received", which usually means TLS not supported.
func IsTLSOversizedRecord(err error) bool {
	return detectReason(err) == tlsOversizedRecordReason
}

// IsCertificateHostnameError checks whether the set of authorized names doesn't match the requested name
func IsCertificateHostnameError(err error) bool {
	return detectReason(err) == certificateHostnameErrorReason
}

// IsCertificateInvalid checks whether the certificate is invalid for reasons like expired,	CA not authorized
// to sign, there are too many cert intermediates, or the cert usage is not valid for the wanted purpose.
func IsCertificateInvalid(err error) bool {
	return detectReason(err) == certificateInvalidReason
}

func detectReason(err error) int {
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "certificate signed by unknown authority"):
			return certificateAuthorityUnknownReason
		case strings.Contains(err.Error(), "no server defined"):
			return noServerFoundReason
		case clientcmd.IsConfigurationInvalid(err):
			return configurationInvalidReason
		case strings.Contains(err.Error(), "tls: oversized record received"):
			return tlsOversizedRecordReason
		}
		switch err.(type) {
		case x509.UnknownAuthorityError:
			return certificateAuthorityUnknownReason
		case x509.HostnameError:
			return certificateHostnameErrorReason
		case x509.CertificateInvalidError:
			return certificateInvalidReason
		}
	}
	return unknownReason
}
