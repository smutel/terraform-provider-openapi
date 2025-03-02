package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceWithNoBodyInput(t *testing.T) {

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":"someID", "creation_time": "today", "deploy_key":"someDeployKey"}`))
	}))
	apiHost := apiServer.URL[7:]

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: "%s"

schemes:
- "http"

paths:
  /v1/deployKey:
    post:
      x-terraform-resource-name: "deploykey"
      responses:
        201:
          schema:
            $ref: "#/definitions/DeployKeyV1"
  /v1/deployKey/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The deploy key id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/DeployKeyV1"
    delete:
      parameters: 
      - name: "id"
        in: "path"
        description: "The deploy key id to be deleted."
        required: true
        type: "string"
      responses: 
        204: 
          description: "successful operation, no content is returned"
definitions:
  DeployKeyV1: # All the properties are readOnly
    type: "object"
    properties:
      id:
        readOnly: true
        type: string
      creation_time:
        readOnly: true
        type: string
      deploy_key:
        readOnly: true
        type: string`, apiHost)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{
		SwaggerURL: swaggerServer.URL,
	})
	assert.NoError(t, err)

	tfFileContents := fmt.Sprintf(`resource "openapi_deploykey_v1" "my_deploykeyv1" {}`)

	var testAccProviders = map[string]terraform.ResourceProvider{providerName: provider}
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		PreCheck:   func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config: tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openapi_deploykey_v1.my_deploykeyv1", "id", "someID"),
					resource.TestCheckResourceAttr(
						"openapi_deploykey_v1.my_deploykeyv1", "creation_time", "today"),
					resource.TestCheckResourceAttr(
						"openapi_deploykey_v1.my_deploykeyv1", "deploy_key", "someDeployKey"),
				),
			},
		},
	})
}
