package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfamplify "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func testAccAWSAmplifyDomainAssociation_basic(t *testing.T) {
	var domain amplify.DomainAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_domain_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyDomainAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyDomainAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyDomainAssociationExists(resourceName, &domain),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/.+/domains/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name": rName,
						"prefix":      "www",
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAmplifyDomainAssociationExists(resourceName string, v *amplify.DomainAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Amplify Domain Association ID is set")
		}

		appID, domainName, err := tfamplify.DomainAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		domainAssociation, err := finder.DomainAssociationByAppIDAndDomainName(conn, appID, domainName)

		if err != nil {
			return err
		}

		*v = *domainAssociation

		return nil
	}
}

func testAccCheckAWSAmplifyDomainAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).amplifyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_domain_association" {
			continue
		}

		appID, domainName, err := tfamplify.DomainAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.DomainAssociationByAppIDAndDomainName(conn, appID, domainName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Amplify Domain Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSAmplifyDomainAssociationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = "example.com"

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = "www"
  }

  wait_for_verification = false
}
`, rName)
}
