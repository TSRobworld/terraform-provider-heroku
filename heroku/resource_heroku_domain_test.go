package heroku

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	heroku "github.com/heroku/heroku-go/v5"
)

func TestAccHerokuDomain_Basic(t *testing.T) {
	var domain heroku.Domain
	randString := acctest.RandString(10)
	appName := fmt.Sprintf("tftest-%s", randString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHerokuDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuDomainConfig_basic(appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHerokuDomainExists("heroku_domain.foobar", &domain),
					testAccCheckHerokuDomainAttributes(&domain),
					resource.TestCheckResourceAttr("heroku_domain.foobar", "hostname", "terraform-tftest-"+randString+".example.com"),
					resource.TestCheckResourceAttr("heroku_domain.foobar", "app", appName),
				),
			},
		},
	})
}

func TestAccHerokuDomain_SSL(t *testing.T) {
	var domain heroku.Domain
	randString := acctest.RandString(10)
	appName := fmt.Sprintf("tftest-%s", randString)
	slugID := testAccConfig.GetSlugIDOrSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHerokuDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuDomainConfig_ssl(appName, slugID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHerokuDomainExists("heroku_domain.foobar", &domain),
					testAccCheckHerokuDomainAttributes_ssl(&domain),
					resource.TestCheckResourceAttr("heroku_domain.foobar", "hostname", "terraform-tftest-"+randString+".example.com"),
					resource.TestCheckResourceAttr("heroku_domain.foobar", "app", appName),
				),
			},
		},
	})
}

func testAccCheckHerokuDomainDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Config).Api

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "heroku_domain" {
			continue
		}

		_, err := client.DomainInfo(context.TODO(), rs.Primary.Attributes["app"], rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Domain still exists")
		}
	}

	return nil
}

func testAccCheckHerokuDomainAttributes(Domain *heroku.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(Domain.Hostname, "terraform-") && !strings.HasSuffix(Domain.Hostname, ".example.com") {
			return fmt.Errorf("Bad hostname: %s", Domain.Hostname)
		}

		if !strings.Contains(*Domain.CName, ".herokudns.com") {
			return fmt.Errorf("Expected cname to be [*.herokudns.com] but got: [%s]", *Domain.CName)
		}

		return nil
	}
}

func testAccCheckHerokuDomainAttributes_ssl(Domain *heroku.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(Domain.Hostname, "terraform-") && !strings.HasSuffix(Domain.Hostname, ".example.com") {
			return fmt.Errorf("Bad hostname: %s", Domain.Hostname)
		}

		if !strings.Contains(*Domain.CName, ".herokudns.com") {
			return fmt.Errorf("Expected cname to be [*.herokudns.com] but got: [%s]", *Domain.CName)
		}

		fmt.Printf("%v\n", Domain.SniEndpoint)

		return nil
	}
}

func testAccCheckHerokuDomainExists(n string, Domain *heroku.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Domain ID is set")
		}

		client := testAccProvider.Meta().(*Config).Api

		foundDomain, err := client.DomainInfo(context.TODO(), rs.Primary.Attributes["app"], rs.Primary.ID)

		if err != nil {
			return err
		}

		if foundDomain.ID != rs.Primary.ID {
			return fmt.Errorf("Domain not found")
		}

		*Domain = *foundDomain

		return nil
	}
}

func testAccCheckHerokuDomainConfig_ssl(appName, slugID string) string {
	wd, _ := os.Getwd()
	certFile := wd + "/test-fixtures/terraform.cert"
	keyFile := wd + "/test-fixtures/terraform.key"

	return fmt.Sprintf(`resource "heroku_app" "foobar" {
    name = "%s"
    region = "us"
}

resource "heroku_app_release" "foobar-release" {
  app = "${heroku_app.foobar.name}"
  slug_id = "%s"
}

resource "heroku_formation" "foobar-web" {
  app = "${heroku_app.foobar.name}"
  type = "web"
  size = "standard-1x"
  quantity = 1
}

resource "heroku_cert" "ssl_certificate" {
  app = "${heroku_app.foobar.name}"
  certificate_chain="${file("%s")}"
  private_key="${file("%s")}"
}

resource "heroku_domain" "foobar" {
  app = "${heroku_app.foobar.name}"
  hostname = "terraform-%s.example.com"
  sni_endpoint = heroku_cert.ssl_certificate.id
}`, appName, slugID, certFile, keyFile, appName)
}

func testAccCheckHerokuDomainConfig_basic(appName string) string {
	return fmt.Sprintf(`resource "heroku_app" "foobar" {
name = "%s"
region = "us"
  }

  resource "heroku_domain" "foobar" {
app = "${heroku_app.foobar.name}"
hostname = "terraform-%s.example.com"
  }`, appName, appName)
}
