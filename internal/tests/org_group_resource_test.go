package tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

//go:generate go run github.com/FrangipaneTeam/tf-doc-extractor@latest -filename $GOFILE -example-dir ../../examples -test
const testAccOrgGroupResourceConfig = `
resource "cloudavenue_iam_group" "example" {
  name          = "OrgTest"
  role          = "Organization Administrator"
  description   = "org test from go test"
}
`

func TestAccOrgGroupResource(t *testing.T) {
	resourceName := "cloudavenue_iam_group.example"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccOrgGroupResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "OrgTest"),
					resource.TestCheckResourceAttr(resourceName, "role", "Organization Administrator"),
					resource.TestCheckResourceAttr(resourceName, "description", "org test from go test"),
					resource.TestCheckResourceAttr(resourceName, "user_names.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     "OrgTest",
				ImportStateVerify: true,
			},
		},
	})
}
