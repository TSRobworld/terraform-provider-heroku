package heroku

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceHerokuApp_Basic(t *testing.T) {
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	appStack := "heroku-20"
	gitUrl := fmt.Sprintf("https://git.heroku.com/%s.git", appName)
	webUrl := fmt.Sprintf("https://%s.herokuapp.com/", appName)
	herokuHostname := fmt.Sprintf("%s.herokuapp.com", appName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuAppWithDatasource_basic(appName, appStack),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "name", appName),
					resource.TestCheckResourceAttrSet(
						"data.heroku_app.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "stack", appStack),
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "region", "us"),
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "git_url", gitUrl),
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "web_url", webUrl),
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "config_vars.FOO", "bar"),
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "buildpacks.0", "https://github.com/heroku/heroku-buildpack-multi-procfile"),
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "acm", "false"),
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "heroku_hostname", herokuHostname),
				),
			},
		},
	})
}

func TestAccDatasourceHerokuApp_Organization(t *testing.T) {
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	org := os.Getenv("HEROKU_SPACES_ORGANIZATION")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if org == "" {
				t.Skip("HEROKU_SPACES_ORGANIZATION is not set; skipping test.")
			}
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuApp_organization(appName, org),
			},
			{
				Config: testAccCheckHerokuAppWithDatasource_organization(appName, org),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "name", appName),
					resource.TestCheckResourceAttr(
						"data.heroku_app.foobar", "organization.0.name", org),
				),
			},
		},
	})
}

func testAccCheckHerokuApp_basic(appName string, stack string) string {
	return fmt.Sprintf(`
resource "heroku_app" "foobar" {
  name   = "%s"
  stack = "%s"
  region = "us"

  buildpacks = [
    "heroku/go"
  ]

	config_vars = {
    FOO = "bar"
	}
}
`, appName, stack)
}

func testAccCheckHerokuAppWithDatasource_basic(appName string, stack string) string {
	return fmt.Sprintf(`
resource "heroku_app" "foobar" {
  name   = "%s"
  stack = "%s"
  region = "us"

  buildpacks = [
    "https://github.com/heroku/heroku-buildpack-multi-procfile",
    "heroku/go"
	]
	
	config_vars = {
    FOO = "bar"
	}
}

data "heroku_app" "foobar" {
  name = "${heroku_app.foobar.name}"
}
`, appName, stack)
}

func testAccCheckHerokuApp_organization(appName, orgName string) string {
	return fmt.Sprintf(`
resource "heroku_app" "foobar" {
  name   = "%s"
  organization {
    name = "%s"
  }
  region = "us"
}
`, appName, orgName)
}

func testAccCheckHerokuAppWithDatasource_organization(appName, orgName string) string {
	return fmt.Sprintf(`
resource "heroku_app" "foobar" {
  name   = "%s"
  organization {
    name = "%s"
  }
  region = "us"
}

data "heroku_app" "foobar" {
  name = "${heroku_app.foobar.name}"
}
`, appName, orgName)
}
