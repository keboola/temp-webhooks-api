// nolint: gochecknoglobals
package webhooks

import (
	_ "goa.design/goa/v3/codegen/generator"
	. "goa.design/goa/v3/dsl"
)

var _ = API("webhooks", func() {
	Title("Webhooks Service")
	Description("<h3>How does it work</h3>\n<ol>\n    <li> register a webhook using /webhook endpoint. You will receive a URL with HASH where you can send data It\n        requires:\n        <ul>\n            <li>STORAGE token in Keboola</li>\n            <li>name of table where the data should be stored in. If it doesn't exists, it will be created</li>\n            <li>Optionaly you can define Conditions</li>\n        </ul>\n    </li>\n    <li>\n        Then you can send data on the provided URL\n    </li>\n    <li>\n        Based on Conditions, the webhook app sends provided data to specified table in Keboola\n    </li>\n    <li>You can send the data to Keboola manualy calling /webhook/HASH/flush.</li>\n    <li> register a webhook using /webhook endpoint. You will receive a URL with HASH where you can send data It\n        requires - STORAGE token in Keboola and\n    </li>\n</ol>\n<h4>\n    Conditions\n</h4>\n<ul>\n    <li> Webhook service sends the data to Keboola if one of the following condition complies\n   <ul>\n       <li>time - each X seconds/minutes</li>\n       <li>size - in bulk of X KB/MB</li>\n       <li>rows - in bulk of N rows. Default value is 1000</li>\n   </ul>\n    </li>\n    <li>You can specify this conditions when registering the webhook using POST /webhook endpoint or update it using PUT\n        /webhook/{hash}</li>\n    \n</ul>")
	Version("1.0")
	HTTP(func() {
		Consumes("application/json")
		Produces("application/json")
	})
	Server("webhooks", func() {
		Host("production", func() {
			URI("http://20.67.180.30:8888")
		})
		Host("localhost", func() {
			URI("http://localhost:8888")
		})
	})
})

var index = ResultType("application/vnd.webhooks.index", func() {
	Description("Index of the service")
	TypeName("Index")

	Attributes(func() {
		Attribute("api", String, "Name of the API", func() {
			Example("webhooks")
		})
		Attribute("documentation", String, "Url of the API documentation", func() {
			Example("https://webhooks.keboola.com/documentation")
		})
		Required("api", "documentation")
	})
})

var conditions = Type("conditions", func() {
	Description("Import conditions. If at least one is met import to the table occurs.")
	Attribute("count", UInt, "Batch will be imported when the given number of records is reached.", func() {
		Example(1000)
	})
	Attribute("size", String, "Batch will be imported when its size reaches a value.", func() {
		Example("10MB")
	})
	Attribute("time", String, "Batch will be imported when time from the first request expires ", func() {
		Example("30s")
	})
})

var importResult = ResultType("application/vnd.webhooks.import.result", func() {
	Description("Import result")
	TypeName("ImportResult")

	Attributes(func() {
		Attribute("recordsInBatch", UInt, "Number of records that have not yet been imported into the table.", func() {
			Example(123)
		})
		Required("recordsInBatch")
	})
})

var registerResult = ResultType("application/vnd.webhooks.register.result", func() {
	Description("Registration result")
	TypeName("RegistrationResult")

	Attributes(func() {
		Attribute("url", String, "Webhook url", func() {
			Example("https://webhooks.keboola.com/webhook/ljBSN5QmXRXFFs5Y7GEY/import")
		})
		Required("url")
	})
})

var updateResult = ResultType("application/vnd.webhooks.update.result", func() {
	Description("Update result")
	TypeName("UpdateResult")
	Attribute("conditions", conditions)
	Required("conditions")
})

var _ = Service("webhooks", func() {
	Description("A service for webhooks.")

	// Methods
	Method("index-root", func() {
		Meta("swagger:summary", "API information.")
		Result(index)
		NoSecurity()
		HTTP(func() {
			GET("/")
			Response(StatusOK)
		})
	})

	Method("health-check", func() {
		Meta("swagger:summary", "Health check.")
		NoSecurity()
		Result(String, func() {
			Example("OK")
		})
		HTTP(func() {
			GET("/health-check")
			Response(StatusOK, func() {
				ContentType("text/plain")
			})
		})
	})

	Method("register", func() {
		Meta("swagger:summary", "Register a new webhook.")
		Payload(func() {
			Attribute("tableId", String, "ID of table to create the import webhook on", func() {
				Example("in.c-my-bucket.my_table")
			})
			Attribute("token", String, "Storage token to the project", func() {
				Example("my-storage-api-token")
			})
			Attribute("conditions", conditions)
			Required("tableId", "token")
		})
		Result(registerResult)
		Error("UnauthorizedError", func() {
			Description("Error returned when the specified token is invalid.")
			Attribute("message", func() {
				Example("Invalid storage token \"<token>\" supplied.")
			})
			Required("message")
		})
		HTTP(func() {
			POST("webhook")
			Response(StatusCreated)
			Response("UnauthorizedError", StatusUnauthorized)
		})
	})

	Method("update", func() {
		Meta("swagger:summary", "Update conditions of the webhook.")
		Payload(func() {
			Field(1, "hash", String, "Authorization hash", func() {
				Example("yljBSN5QmXRXFFs5Y7GEY")
			})
			Attribute("conditions", conditions)
			Required("hash", "conditions")
		})
		Result(updateResult)
		Error("WebhookNotFoundError", func() {
			Description("Error returned when no webhook was found under the specified hash.")
			Attribute("message", func() {
				Example("Webhook with hash \"<hash>\" not found.")
			})
			Required("message")
		})
		HTTP(func() {
			PUT("webhook/{hash}")
			Response(StatusOK)
			Response("WebhookNotFoundError", StatusNotFound)
		})
	})

	Method("flush", func() {
		Meta("swagger:summary", "Loads data to connection manually")
		Payload(func() {
			Field(1, "hash", String, "Authorization hash", func() {
				Example("yljBSN5QmXRXFFs5Y7GEY")
			})
			Required("hash")
		})
		Result(String, func() {
			Example("OK")
		})
		Error("WebhookNotFoundError", func() {
			Description("Error returned when no webhook was found under the specified hash.")
			Attribute("message", func() {
				Example("Webhook with hash \"<hash>\" not found.")
			})
			Required("message")
		})
		HTTP(func() {
			POST("webhook/{hash}/flush")
			Response(StatusOK)
			Response("WebhookNotFoundError", StatusNotFound)
		})
	})

	Method("import", func() {
		Meta("swagger:summary", "Import data.")
		Payload(func() {
			Field(1, "hash", String, "Authorization hash", func() {
				Example("yljBSN5QmXRXFFs5Y7GEY")
			})
			Required("hash")
		})
		Result(importResult)
		Error("WebhookNotFoundError", func() {
			Description("Error returned when no webhook was found under the specified hash.")
			Attribute("message", func() {
				Example("Webhook with hash \"<hash>\" not found.")
			})
			Required("message")
		})
		HTTP(func() {
			POST("webhook/{hash}/import")
			SkipRequestBodyEncodeDecode()
			Response(StatusOK)
			Response("WebhookNotFoundError", StatusNotFound)
		})
	})

	Files("/documentation/openapi.json", "openapi.json", func() {
		Meta("swagger:summary", "Swagger 2.0 JSON Specification")
		Meta("swagger:tag:documentation")
	})
	Files("/documentation/openapi.yaml", "openapi.yaml", func() {
		Meta("swagger:summary", "Swagger 2.0 YAML Specification")
		Meta("swagger:tag:documentation")
	})
	Files("/documentation/openapi3.json", "openapi3.json", func() {
		Meta("swagger:summary", "OpenAPI 3.0 JSON Specification")
		Meta("swagger:tag:documentation")
	})
	Files("/documentation/openapi3.yaml", "openapi3.yaml", func() {
		Meta("swagger:summary", "OpenAPI 3.0 YAML Specification")
		Meta("swagger:tag:documentation")
	})
	Files("/documentation/{*path}", "swagger-ui", func() {
		Meta("swagger:summary", "Swagger UI")
		Meta("swagger:tag:documentation")
	})
})
