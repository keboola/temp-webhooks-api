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

var registration = ResultType("application/vnd.webhooks.registration", func() {
	Description("Registration response")
	TypeName("Registration")

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

	Method("register", func() {
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
		Result(registration)
		HTTP(func() {
			POST("register")
			Response(StatusOK)
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
			POST("import/{hash}")
			Body(func() {
				Attribute("body")
			})
			Response(StatusOK, func() {
				ContentType("text/plain")
			})
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
