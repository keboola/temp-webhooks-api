// nolint: gochecknoglobals
package webhooks

import (
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
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

var importResult = ResultType("application/vnd.webhooks.import.result", func() {
	Description("Import result")
	TypeName("ImportResult")

	Attributes(func() {
		Attribute("recordsInBatch", Int, "Number of records that have not yet been imported into the table.", func() {
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
			Example("https://webhooks.keboola.com/import/yljBSN5QmXRXFFs5Y7GEY")
		})
		Required("url")
	})
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
			def := model.NewConditions() // re-use default values
			Attribute("tableId", String, "ID of table to create the import webhook on", func() {
				Example("in.c-my-bucket.my-table")
			})
			Attribute("token", String, "Storage token to the project", func() {
				Example("my-storage-api-token")
			})
			Attribute("conditions", func() {
				Description("Import conditions. If at least one is met import to the table occurs.")
				Attribute("count", Int, "Batch will be imported when the given number of records is reached.", func() {
					Example(def.Count)
					Default(def.Count)
				})
				Attribute("time", String, "Batch will be imported when time from the first request expires ", func() {
					Example(def.Time)
					Default(def.Time)
				})
				Attribute("size", String, "Batch will be imported when its size reaches a value.", func() {
					Example(def.Size)
					Default(def.Size)
				})
			})
			Required("tableId", "token")
		})
		Result(registerResult)
		HTTP(func() {
			POST("register")
			Response(StatusOK)
		})
	})

	Method("import", func() {
		Meta("swagger:summary", "Import data.")
		Payload(func() {
			Field(1, "hash", String, "Authorization hash", func() {
				Example("yljBSN5QmXRXFFs5Y7GEY")
			})
			Field(2, "body", String, "Raw request body", func() {
				Example("Content to be imported.")
			})
			Required("hash", "body")
		})
		Result(importResult)
		HTTP(func() {
			POST("import/{hash}")
			Body(func() {
				Attribute("body")
			})
			Response(StatusOK)
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
