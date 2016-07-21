package api

import "k8s.io/kubernetes/pkg/fields"

// OAuthAccessTokenToSelectableFields returns a label set that represents the object
func OAuthAccessTokenToSelectableFields(obj *OAuthAccessToken) fields.Set {
	return fields.Set{
		"metadata.name":  obj.Name,
		"clientName":     obj.ClientName,
		"userName":       obj.UserName,
		"userUID":        obj.UserUID,
		"authorizeToken": obj.AuthorizeToken,
	}
}

// OAuthAuthorizeTokenToSelectableFields returns a label set that represents the object
func OAuthAuthorizeTokenToSelectableFields(obj *OAuthAuthorizeToken) fields.Set {
	return fields.Set{
		"metadata.name": obj.Name,
		"clientName":    obj.ClientName,
		"userName":      obj.UserName,
		"userUID":       obj.UserUID,
	}
}

// OAuthClientToSelectableFields returns a label set that represents the object
func OAuthClientToSelectableFields(obj *OAuthClient) fields.Set {
	return fields.Set{
		"metadata.name": obj.Name,
	}
}

// OAuthClientAuthorizationToSelectableFields returns a label set that represents the object
func OAuthClientAuthorizationToSelectableFields(obj *OAuthClientAuthorization) fields.Set {
	return fields.Set{
		"metadata.name": obj.Name,
		"clientName":    obj.ClientName,
		"userName":      obj.UserName,
		"userUID":       obj.UserUID,
	}
}
