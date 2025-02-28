package jupyter_test

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/training"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/playwright-community/playwright-go"
	"github.com/tidwall/gjson"
)

var _ = Describe("JupyterE2EBrowser", Ordered, Label("JupyterE2EBrowser"), func() {
	var pw *playwright.Playwright
	var browser playwright.Browser
	var page playwright.Page
	var err error

	var jupyterUrl string

	assertErrorToNilf := func(message string, err error) {
		if err != nil {
			page.Screenshot(playwright.PageScreenshotOptions{
				Path: playwright.String("error_point.png"),
			})
			logger.Logf.Fatalf(message, err.Error())
		}
	}

	BeforeAll(func() {
		pw, err = playwright.Run()
		assertErrorToNilf("could not launch playwright: %w", err)
		browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Proxy: &playwright.Proxy{
				Server: "http://internal-placeholder.com:912",
			},
		})
		assertErrorToNilf("could not launch Chromium: %w", err)
		context, err := browser.NewContext(playwright.BrowserNewContextOptions{
			IgnoreHttpsErrors: playwright.Bool(true),
		})
		assertErrorToNilf("could not create context: %w", err)
		page, err = context.NewPage()
		assertErrorToNilf("could not create page: %w", err)
		page.SetDefaultNavigationTimeout(10)
	})

	It("should register the user via Batch API", func() {
		payload := make(map[string]interface{})
		payload["trainingId"] = "b6f6e860-2fda-11ee-be56-0242ac120002"
		payload["accessType"] = "ACCESS_TYPE_JUPYTER"

		status, response := training.Register(baseApiUrl, userToken, premiumCloudAccount["id"], payload)
		Expect(status).To(Equal(200), "Register user for trainings failed")
		jupyterUrl = gjson.Get(response, "jupyterLoginInfo").String()
	})

	It("should support Azure login in the browser", func() {
		_, err = page.Goto(
			jupyterUrl,
			playwright.PageGotoOptions{
				Timeout:   playwright.Float(0),
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			},
		)
		assertErrorToNilf("could not goto: %w", err)

		time.Sleep(10 * time.Second)

		_, _, _, _, username, password, _, _, _ := auth.Get_Azure_auth_data_from_config(auth.Get_UserName("Premium"))

		assertErrorToNilf("could not click: %v", page.Click("input#signInName"))
		assertErrorToNilf("could not type: %v", page.Type("input#signInName", username))
		assertErrorToNilf("could not press: %v", page.Click("button#continue"))
		time.Sleep(5 * time.Second)
		assertErrorToNilf("could not click: %v", page.Click("input#password"))
		assertErrorToNilf("could not type: %v", page.Type("input#password", password))
		assertErrorToNilf("could not press: %v", page.Click("button#continue"))

		time.Sleep(10 * time.Second)
	})

	It("should have JupyterHub logo on page", func() {
		jupyterLogoExists, err := page.Locator("img.jpy-logo").IsVisible()
		assertErrorToNilf("could not find JupyterHub logo: %v", err)
		Expect(jupyterLogoExists).To(BeTrue())
	})

	It("should allow logging out from JupyterHub launch page", func() {
		assertErrorToNilf("could not click: %v", page.Locator("a#logout").Click())
		time.Sleep(5 * time.Second)
		logoutText, err := page.Locator("div#logout-main > p").InnerText()
		assertErrorToNilf("could not find logout message: %v", err)
		Expect(logoutText).To(ContainSubstring("Successfully logged out."))
	})

	AfterAll(func() {
		page.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("error_point.png"),
		})
		assertErrorToNilf("could not close browser: %w", browser.Close())
		assertErrorToNilf("could not stop Playwright: %w", pw.Stop())
	})
})
