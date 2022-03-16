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
		Path("v1")
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

var index = ResultType("application/vnd.templates.index", func() {
	Description("Index of the service")
	TypeName("Index")

	Attributes(func() {
		Attribute("api", String, "Name of the API", func() {
			Example("webhooks")
		})
		Attribute("documentation", String, "Url of the API documentation", func() {
			Example("https://webhooks.keboola.com/v1/documentation")
		})
		Required("api", "documentation")
	})
})

var _ = Service("webhooks", func() {
	Description("A service for webhooks.")

	// Methods
	Method("index-root", func() {
		Result(index)
		NoSecurity()
		HTTP(func() {
			GET("")
			Response(StatusOK)
		})
	})

	Method("health-check", func() {
		NoSecurity()
		Result(String, func() {
			Example("OK")
		})
		HTTP(func() {
			GET("//health-check")
			Response(StatusOK, func() {
				ContentType("text/plain")
			})
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
