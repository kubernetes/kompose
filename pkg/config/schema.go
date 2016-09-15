package config

var preferenceFileSchema = `{
	"$schema": "http://json-schema.org/draft-04/schema#",
	"type": "object",
	"properties": {
		"profiles": {
			"id": "#/properties/profiles",
			"type": "object",
			"patternProperties": {
				"^[a-zA-Z0-9._-]+$": {
					"$ref": "#/definitions/profiles"
				}
			},
			"additionalProperties": false
		},
		"current-profile": {
			"type": "string"
		}
	},
	"required": [
		"profiles",
		"current-profile"
	],
	"definitions": {
		"profiles": {
			"id": "#/definitions/profiles",
			"type": "object",
			"properties": {
				"provider": {
					"type": "string",
					"enum": [
						"openshift",
						"kubernetes"
					]
				},
				"objects": {
					"type": "array",
					"items": {
						"type": "string",
						"enum": [
							"deployment",
							"replicationcontroller",
							"deploymentconfig",
							"daemonset"
						]
					},
					"uniqueItems": true,
					"minLength": 1
				}
			},
			"required": [
				"provider",
				"objects"
			]
		}
	},
	"additionalProperties": false
}
`
