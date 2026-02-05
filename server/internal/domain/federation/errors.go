package federation

import "errors"

var (
	ErrFederationConfigNotFound   = errors.New("federation config not found")
	ErrFederationInstanceNotFound = errors.New("federation instance not found")
	ErrOutboundDeliveryNotFound   = errors.New("federation outbound delivery not found")
	ErrFederatedCitationNotFound  = errors.New("federated citation not found")
	ErrFederatedMentionNotFound   = errors.New("federated mention not found")
)
