// nolint: gochecknoglobals
package webhooks

import (
	_ "goa.design/goa/v3/codegen/generator"
	. "goa.design/goa/v3/dsl"
)

var _ = API("webhooks", func() {
	Title("Webhooks Service")
	Description("A service for webhooks.")
	Version("1.0")
	HTTP(func() {
		Consumes("application/json")
		Produces("application/json")
	})
	Server("webhooks", func() {
		Host("production", func() {
			URI("https://webhooks.{stack}")
			Variable("stack", String, "Base URL of the stack", func() {
				Default("keboola.com")
				Enum("keboola.com", "eu-central-1.keboola.com", "north-europe.azure.keboola.com")
			})
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

var registration = ResultType("application/vnd.webhooks.registration", func() {
	Description("Registration response")
	TypeName("Registration")

	Attributes(func() {
		Attribute("url", String, "Webhook url", func() {
			Example("https://webhooks.keboola.com/v1/import/123")
		})
		Required("url")
	})
})

var _ = Service("webhooks", func() {
	Description("A service for webhooks.")

	// Methods
	Method("index-root", func() {
		Result(index)
		NoSecurity()
		HTTP(func() {
			GET("/")
			Response(StatusOK)
		})
	})

	Method("health-check", func() {
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

	Method("import", func() {
		NoSecurity()
		Payload(func() {
			Field(1, "hash", String, "Authorization hash")
			Field(2, "body", String, "Raw request body")
			Required("hash", "body")
		})
		Result(String, func() {
			Example("OK")
		})
		HTTP(func() {
			GET("import/{hash}")
			Body(func() {
				Attribute("body")
			})
			Response(StatusOK, func() {
				ContentType("text/plain")
			})
		})
	})

	Method("register", func() {
		Payload(func() {
			Attribute("tableId", String, "ID of table to create the import webhook on")
			Attribute("token", String, "Storage token to the project")
			Required("tableId", "token")
		})
		Result(registration)
		HTTP(func() {
			POST("register")
			Response(StatusOK)
		})
	})

	Files("/documentation/openapi.json", "openapi.json", func() {
		Meta("swagger:summary", "Swagger 2.0 JSON Specification")
	})
	Files("/documentation/openapi.yaml", "openapi.yaml", func() {
		Meta("swagger:summary", "Swagger 2.0 YAML Specification")
	})
	Files("/documentation/openapi3.json", "openapi3.json", func() {
		Meta("swagger:summary", "OpenAPI 3.0 JSON Specification")
	})
	Files("/documentation/openapi3.yaml", "openapi3.yaml", func() {
		Meta("swagger:summary", "OpenAPI 3.0 YAML Specification")
	})
})
