package auth

import (
	ctx "context"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

var timeout float64 = 1000000

func GetBoolPointer(value bool) *bool {
	return &value
}

func assertErrorToNilf(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func assertEqual(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		panic(fmt.Sprintf("%v does not equal %v", actual, expected))
	}
}

func startHttpServer(srv *http.Server, c AuthorizationConfig) {
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

}

// startLocalListener opens an http server to retrieve the redirect from initial
// authentication and set the authorization code's value
func startLocalListener(c AuthorizationConfig, token *AuthorizationCode, mux *http.ServeMux) {
	logger.Log.Info("Starting local Listener")
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Info("Processing Get JWT Token")
		time.Sleep(5 * time.Second)
		state := r.FormValue("state")
		if state != oauthRandom {
			log.Println("invalid oauth state, expected " + oauthRandom + ", got " + state + "\n")
			logger.Log.Info("Redirecting")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		id_token := r.FormValue("code")
		if id_token == "" {
			log.Println("Empty Id Token returned empty")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Empty Id token returned\n"))
			reason := r.FormValue("error_reason")
			if reason == "user_denied" {
				w.Write([]byte("User has denied Permission.."))
			} else {
				w.Write([]byte(reason))
			}
		} else {
			w.Write([]byte("token retrieved succesfully"))
			token.Value = id_token
			//logger.Logf.Infof("Token  ", token.Value)
			return
		}

	})
}

func AuthenticateAdminUser(url string, c AuthorizationConfig, codeVerifier string) (token AuthorizationCode, auth_tokens Tokens) {
	// err := playwright.Install()
	// if err != nil {
	// 	fmt.Println("Playwrite browser installation failed")
	// }
	pw, err := playwright.Run()
	mux := http.NewServeMux()
	srv := &http.Server{Addr: fmt.Sprintf(":%s", c.RedirectPort), Handler: mux}
	startLocalListener(c, &token, mux)
	startHttpServer(srv, c)

	assertErrorToNilf("could not launch playwright: %w", err)
	browser, err := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	assertErrorToNilf("could not launch Chromium: %w", err)
	context, err := browser.NewContext()
	assertErrorToNilf("could not create context: %w", err)
	page, err := context.NewPage()
	page.SetDefaultNavigationTimeout(timeout)
	assertErrorToNilf("could not create page: %w", err)
	_, err = page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float(0)})
	assertErrorToNilf("could not goto: %w", err)
	// To Do : Test whether page time out is working, comment  sleep accorrdingly
	time.Sleep(10 * time.Second)
	assertErrorToNilf("could not click: %v", page.Click("#i0116"))
	assertErrorToNilf("could not type: %v", page.Type("#i0116", c.Username))
	assertErrorToNilf("could not press: %v", page.Click("#idSIButton9"))
	time.Sleep(10 * time.Second)
	assertErrorToNilf("could not click: %v", page.Click("#i0118"))
	assertErrorToNilf("could not type: %v", page.Type("#i0118", c.Password))
	assertErrorToNilf("could not press: %v", page.Click("#idSIButton9"))

	running := true
	//time.Sleep(10 * time.Minute)
	for running {
		if token.Value != "" {
			logger.Log.Info("Successfullly fetched the JWT Token, Closing the local listener")
			//logger.Logf.Infof("Token  ", token.Value)
			time.Sleep(2 * time.Second)
			auth_tokens, err = GetAdminTokens(c, codeVerifier, token.Value)
			if err != nil {
				logger.Logf.Infof("Auth TOkens Error  ", err.Error())
			}
			if err := srv.Shutdown(ctx.TODO()); err != nil {
				log.Printf("Httpserver: ListenAndServe() error: %s", err) // failure/timeout shutting down the server gracefully
			}
			time.Sleep(4 * time.Second)
			running = false
		}
	}

	return token, auth_tokens
}

func startHttpServer1(c AuthorizationConfig) (*http.Server, *http.ServeMux) {
	mux := http.NewServeMux()
	srv := &http.Server{Addr: fmt.Sprintf(":%s", c.RedirectPort), Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
	return srv, mux
}

func AuthenticateUser(url string, c AuthorizationConfig, codeVerifier string) (token AuthorizationCode, auth_tokens Tokens) {
	// err := playwright.Install()
	// if err != nil {
	// 	fmt.Println("Playwrite browser installation failed")
	// }
	pw, err := playwright.Run()
	assertErrorToNilf("could not launch playwright: %w", err)
	browser, err := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})

	assertErrorToNilf("could not launch Chromium: %w", err)
	context, err := browser.NewContext()
	assertErrorToNilf("could not create context: %w", err)
	page, err := context.NewPage()
	page.SetDefaultNavigationTimeout(timeout)
	assertErrorToNilf("could not create page: %w", err)
	_, err = page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float(0)})
	assertErrorToNilf("could not goto: %w", err)
	// To Do : Test whether page time out is working, comment  sleep accorrdingly
	time.Sleep(1 * time.Minute)
	assertErrorToNilf("could not click: %v", page.Click("#signInName"))
	assertErrorToNilf("could not type: %v", page.Type("#signInName", c.Username))
	time.Sleep(20 * time.Second)
	assertErrorToNilf("could not press: %v", page.Click("#continue"))
	time.Sleep(20 * time.Second)
	fl := strings.Contains(c.Username, "intel.com")
	if fl {
		time.Sleep(20 * time.Second)
		b2b_message := page.Locator("#b2b_message_continue")
		count, _ := b2b_message.Count()
		//assertErrorToNilf("could not determine b2b mesasge: %w", err)
		if count != 0 {
			assertErrorToNilf("could not press: %v", page.Click("#b2b_message_continue"))
			time.Sleep(20 * time.Second)
		}
	}
	assertCountOfTodos := func(shouldBeCount int) {
		//var options playwright.PageLocatorOptions
		loccount := page.Locator("#emailVerificationControl_but_send_code")
		count, err := loccount.Count()
		assertErrorToNilf("could not determine todo list count: %w", err)
		assertEqual(shouldBeCount, count)
	}

	// Initially there should be 0 entries
	assertCountOfTodos(0)

	if fl {
		time.Sleep(1 * time.Minute)
		assertErrorToNilf("could not click: %v", page.Click("#i0118"))
		assertErrorToNilf("could not type: %v", page.Type("#i0118", c.Password))
		assertErrorToNilf("could not press: %v", page.Click("#idSIButton9"))
		time.Sleep(40 * time.Second)
		if err = page.Click("#KmsiCheckboxField"); err == nil {
			time.Sleep(40 * time.Second)
			page.Click("#idSIButton9")
		}
	} else {
		assertErrorToNilf("could not click: %v", page.Click("#password"))
		assertErrorToNilf("could not type: %v", page.Type("#password", c.Password))
		assertErrorToNilf("could not press: %v", page.Click("#continue"))
	}
	running := true

	srv, mux := startHttpServer1(c)
	startLocalListener(c, &token, mux)
	for running {
		if token.Value != "" {
			logger.Log.Info("Successfullly fetched the JWT Token, Closing the local listener")
			time.Sleep(2 * time.Second)
			auth_tokens, err = GetTokens(c, codeVerifier, token.Value)
			if err != nil {
				logger.Logf.Infof("Auth TOkens Error  ", err.Error())
			}
			if err := srv.Shutdown(ctx.TODO()); err != nil {
				log.Printf("Httpserver: ListenAndServe() error: %s", err) // failure/timeout shutting down the server gracefully
			}
			time.Sleep(4 * time.Second)
			running = false
		}
	}
	return token, auth_tokens
}
